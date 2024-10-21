// Copyright 2020 OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kube

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	api_v1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sprocessor/observability"
)

// WatchClient is the main interface provided by this package to a kubernetes cluster.
type WatchClient struct {
	m           sync.RWMutex
	deleteMut   sync.Mutex
	logger      *zap.Logger
	kc          kubernetes.Interface
	informer    cache.SharedInformer
	deleteQueue []deleteRequest
	stopCh      chan struct{}
	op          OwnerAPI
	delimiter   string

	// A map containing Pod related data, used to associate them with resources.
	// Key can be either an IP address or Pod UID
	Pods         map[PodIdentifier]*Pod
	Rules        ExtractionRules
	Filters      Filters
	Associations []Association
	Exclude      Excludes
}

// New initializes a new k8s Client.
func New(
	logger *zap.Logger,
	apiCfg k8sconfig.APIConfig,
	rules ExtractionRules,
	filters Filters,
	associations []Association,
	exclude Excludes,
	newClientSet APIClientsetProvider,
	newInformer InformerProvider,
	newOwnerProviderFunc OwnerProvider,
	delimiter string,
	deleteInterval time.Duration,
	gracePeriod time.Duration,
) (Client, error) {
	c := &WatchClient{
		logger:       logger,
		Rules:        rules,
		Filters:      filters,
		Associations: associations,
		Exclude:      exclude,
		stopCh:       make(chan struct{}),
		delimiter:    delimiter,
		Pods:         map[PodIdentifier]*Pod{},
	}
	go c.deleteLoop(deleteInterval, gracePeriod)

	if newClientSet == nil {
		newClientSet = k8sconfig.MakeClient
	}

	kc, err := newClientSet(apiCfg)
	if err != nil {
		return nil, err
	}
	c.kc = kc

	labelSelector, fieldSelector, err := selectorsFromFilters(c.Filters)
	if err != nil {
		return nil, err
	}

	if c.Rules.OwnerLookupEnabled {
		if newOwnerProviderFunc == nil {
			newOwnerProviderFunc = newOwnerProvider
		}

		c.op, err = newOwnerProviderFunc(logger, c.kc, labelSelector, fieldSelector, rules, c.Filters.Namespace, deleteInterval, gracePeriod)
		if err != nil {
			return nil, err
		}
	}

	logger.Info(
		"k8s filtering",
		zap.String("labelSelector", labelSelector.String()),
		zap.String("fieldSelector", fieldSelector.String()),
	)
	if newInformer == nil {
		newInformer = newSharedInformer
	}

	// the Node filter only applies to Pods, so we add it here
	if filters.Node != "" {
		fieldSelector = addNodeSelector(fieldSelector, filters.Node)
	}

	c.informer = newInformer(logger, c.kc, c.Filters.Namespace, labelSelector, fieldSelector)
	return c, err
}

// Start registers pod event handlers and starts watching the kubernetes cluster for pod changes.
func (c *WatchClient) Start() {
	if c.op != nil {
		c.op.Start()
	}

	_, err := c.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handlePodAdd,
		UpdateFunc: c.handlePodUpdate,
		DeleteFunc: c.handlePodDelete,
	})
	if err != nil {
		c.logger.Error("error adding event handler to pod informer", zap.Error(err))
	}
	err = c.informer.SetTransform(
		func(object interface{}) (interface{}, error) {
			originalPod, success := object.(*api_v1.Pod)
			if !success {
				return object.(cache.DeletedFinalStateUnknown), nil
			} else {
				return removeUnnecessaryPodData(originalPod, c.Rules), nil
			}
		},
	)
	if err != nil {
		c.logger.Warn("error setting Pod data transformer, continuing without it", zap.Error(err))
	}
	c.informer.Run(c.stopCh)
}

// Stop signals the the k8s watcher/informer to stop watching for new events.
func (c *WatchClient) Stop() {
	close(c.stopCh)

	if c.op != nil {
		c.op.Stop()
	}
}

func (c *WatchClient) handlePodAdd(obj interface{}) {
	observability.RecordPodAdded()
	if pod, ok := obj.(*api_v1.Pod); ok {
		c.addOrUpdatePod(pod)
	} else {
		c.logger.Error("object received was not of type api_v1.Pod", zap.Any("received", obj))
	}
	c.m.RLock()
	podTableSize := len(c.Pods)
	c.m.RUnlock()
	observability.RecordPodTableSize(int64(podTableSize))
}

func (c *WatchClient) handlePodUpdate(old, new interface{}) {
	observability.RecordPodUpdated()
	if pod, ok := new.(*api_v1.Pod); ok {
		// TODO: update or remove based on whether container is ready/unready?.
		c.addOrUpdatePod(pod)
	} else {
		c.logger.Error("object received was not of type api_v1.Pod", zap.Any("received", new))
	}
	c.m.RLock()
	podTableSize := len(c.Pods)
	c.m.RUnlock()
	observability.RecordPodTableSize(int64(podTableSize))
}

func (c *WatchClient) handlePodDelete(obj interface{}) {
	observability.RecordPodDeleted()

	var pod *api_v1.Pod

	switch obj := obj.(type) {
	case *api_v1.Pod:
		pod = obj
	case cache.DeletedFinalStateUnknown:
		prev, ok := obj.Obj.(*api_v1.Pod)
		if !ok {
			c.logger.Error(
				"object received was DeletedFinalStateUnknown but did not contain api_v1.Pod",
				zap.Any("received", obj),
			)
			return
		}
		pod = prev
	default:
		c.logger.Error("object received was not of type api_v1.Pod", zap.Any("received", obj))
		return
	}

	c.forgetPod(pod)
}

func (c *WatchClient) deleteLoop(interval time.Duration, gracePeriod time.Duration) {
	// This loop runs after N seconds and deletes pods from cache.
	// It iterates over the delete queue and deletes all that aren't
	// in the grace period anymore.
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			var cutoff int
			now := time.Now()
			c.deleteMut.Lock()
			for i, d := range c.deleteQueue {
				if d.ts.Add(gracePeriod).After(now) {
					break
				}
				cutoff = i + 1
			}
			toDelete := c.deleteQueue[:cutoff]
			c.deleteQueue = c.deleteQueue[cutoff:]
			c.deleteMut.Unlock()

			c.m.Lock()
			for _, d := range toDelete {
				if p, ok := c.Pods[d.id]; ok {
					// Sanity check: make sure we are deleting the same pod
					// and the underlying state (ip<>pod mapping) has not changed.
					if p.Name == d.podName {
						delete(c.Pods, d.id)
					}
				}
			}
			podTableSize := len(c.Pods)
			c.m.Unlock()
			observability.RecordPodTableSize(int64(podTableSize))

		case <-c.stopCh:
			return
		}
	}
}

// getPod takes an IP address or Pod UID and returns the pod the identifier is associated with.
func (c *WatchClient) getPod(identifier PodIdentifier) (*Pod, bool) {
	c.m.RLock()
	defer c.m.RUnlock()
	pod, ok := c.Pods[identifier]
	if !ok {
		observability.RecordIPLookupMiss()
		return nil, ok
	}
	if pod.Ignore {
		return nil, false
	}
	return pod, true
}

// GetPodAttributes takes an IP address or Pod UID and returns the metadata attributes of the Pod the
// identifier is associated with
func (c *WatchClient) GetPodAttributes(identifier PodIdentifier) (map[string]string, bool) {
	pod, ok := c.getPod(identifier)
	if !ok {
		return nil, false
	}
	ownerAttributes := c.getPodOwnerMetadataAttributes(pod)

	// we need to take a lock here because pod.Attributes may be modified concurrently
	// TODO: clean up the locking in this function and the ones it calls
	c.m.RLock()
	defer c.m.RUnlock()
	attributes := make(map[string]string, len(pod.Attributes))
	for key, value := range pod.Attributes {
		attributes[key] = value
	}
	for key, value := range ownerAttributes {
		attributes[key] = value
	}
	return attributes, ok
}

func (c *WatchClient) extractPodAttributes(pod *api_v1.Pod) map[string]string {
	tags := map[string]string{}
	if c.Rules.PodName {
		tags[c.Rules.Tags.PodName] = pod.Name
	}

	if c.Rules.Namespace {
		tags[c.Rules.Tags.Namespace] = pod.GetNamespace()
	}

	if c.Rules.StartTime {
		ts := pod.GetCreationTimestamp()
		if !ts.IsZero() {
			tags[c.Rules.Tags.StartTime] = ts.String()
		}
	}

	if c.Rules.PodUID {
		uid := pod.GetUID()
		tags[c.Rules.Tags.PodUID] = string(uid)
	}

	if c.Rules.NodeName {
		if len(pod.Spec.NodeName) == 0 {
			c.logger.Debug("missing Node name for Pod, cache may be out of sync", zap.String("pod", pod.Name))
		}
		tags[c.Rules.Tags.NodeName] = pod.Spec.NodeName
	}

	if c.Rules.HostName {
		// Basing on v1.17 Kubernetes docs, when a hostname is specified, it takes precedence over
		// the associated metadata name, see:
		// https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#pod-s-hostname-and-subdomain-fields
		if pod.Spec.Hostname == "" {
			tags[c.Rules.Tags.HostName] = pod.Name
		} else {
			tags[c.Rules.Tags.HostName] = pod.Spec.Hostname
		}
	}

	// Owner metadata is updated on every query.

	if len(pod.Status.ContainerStatuses) > 0 {
		cs := pod.Status.ContainerStatuses[0]
		if c.Rules.ContainerID {
			if len(cs.ContainerID) == 0 {
				c.logger.Debug("missing container ID for Pod, cache may be out of sync", zap.String("pod", pod.Name), zap.String("container_name", cs.Name))
			}
			tags[c.Rules.Tags.ContainerID] = cs.ContainerID
		}
	}

	if len(pod.Spec.Containers) > 0 {
		container := pod.Spec.Containers[0]

		if c.Rules.ContainerName {
			tags[c.Rules.Tags.ContainerName] = container.Name
		}
		if c.Rules.ContainerImage {
			tags[c.Rules.Tags.ContainerImage] = container.Image
		}
	}

	for _, r := range c.Rules.Labels {
		c.extractLabelsIntoTags(r, pod.Labels, tags)
	}

	if (len(c.Rules.NamespaceLabels) > 0 || len(c.Rules.NamespaceAnnotations) > 0) && c.Rules.OwnerLookupEnabled {
		namespace := c.op.GetNamespace(pod)
		if namespace != nil {
			for _, r := range c.Rules.NamespaceLabels {
				c.extractLabelsIntoTags(r, namespace.Labels, tags)
			}

			for _, r := range c.Rules.NamespaceAnnotations {
				c.extractLabelsIntoTags(r, namespace.Annotations, tags)
			}
		}
	}

	for _, r := range c.Rules.Annotations {
		c.extractLabelsIntoTags(r, pod.Annotations, tags)
	}
	return tags
}

func (c *WatchClient) getPodOwnerMetadataAttributes(pod *Pod) map[string]string {
	c.m.RLock()
	defer c.m.RUnlock()
	attributes := map[string]string{}
	if c.Rules.OwnerLookupEnabled {
		c.logger.Debug("pod owner lookup",
			zap.String("pod.Name", pod.Name),
			zap.Any("pod.OwnerReferences", pod.OwnerReferences),
		)
		owners := c.op.GetOwners(pod)

		for _, owner := range owners {
			switch owner.kind {
			case "DaemonSet":
				if c.Rules.DaemonSetName {
					attributes[c.Rules.Tags.DaemonSetName] = owner.name
				}
			case "Deployment":
				if c.Rules.DeploymentName {
					attributes[c.Rules.Tags.DeploymentName] = owner.name
				}
			case "ReplicaSet":
				if c.Rules.ReplicaSetName {
					attributes[c.Rules.Tags.ReplicaSetName] = owner.name
				}
			case "StatefulSet":
				if c.Rules.StatefulSetName {
					attributes[c.Rules.Tags.StatefulSetName] = owner.name
				}
			case "Job":
				if c.Rules.JobName {
					attributes[c.Rules.Tags.JobName] = owner.name
				}
			case "CronJob":
				if c.Rules.CronJobName {
					attributes[c.Rules.Tags.CronJobName] = owner.name
				}

			default:
				// Do nothing
			}
		}

		if c.Rules.ServiceName {
			services := c.op.GetServices(pod.Name)
			attributes[c.Rules.Tags.ServiceName] = strings.Join(services, c.delimiter)
		}
	}
	return attributes
}

// This function removes all data from the Pod except what is required by extraction rules
func removeUnnecessaryPodData(pod *api_v1.Pod, rules ExtractionRules) *api_v1.Pod {

	// name, namespace, uid, start time and ip are needed for identifying Pods
	transformedPod := api_v1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name:      pod.GetName(),
			Namespace: pod.GetNamespace(),
			UID:       pod.GetUID(),
		},
		Status: api_v1.PodStatus{
			PodIP:     pod.Status.PodIP,
			StartTime: pod.Status.StartTime,
		},
	}

	if rules.StartTime {
		transformedPod.SetCreationTimestamp(pod.GetCreationTimestamp())
	}

	if rules.PodUID {
		transformedPod.SetUID(pod.GetUID())
	}

	if rules.NodeName {
		transformedPod.Spec.NodeName = pod.Spec.NodeName
	}

	if rules.HostName {
		transformedPod.Spec.Hostname = pod.Spec.Hostname
	}

	if rules.ContainerID {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			transformedPod.Status.ContainerStatuses = append(
				transformedPod.Status.ContainerStatuses,
				api_v1.ContainerStatus{ContainerID: containerStatus.ContainerID},
			)
		}
	}

	if rules.ContainerName || rules.ContainerImage {
		for _, container := range pod.Spec.Containers {
			transformedPod.Spec.Containers = append(
				transformedPod.Spec.Containers,
				api_v1.Container{Name: container.Name, Image: container.Image},
			)
		}
	}

	if len(rules.Labels) > 0 {
		transformedPod.Labels = pod.Labels
	}

	if len(rules.Annotations) > 0 {
		transformedPod.Annotations = pod.Annotations
	}

	if rules.OwnerLookupEnabled {
		transformedPod.SetOwnerReferences(pod.GetOwnerReferences())
	}

	return &transformedPod
}

func (c *WatchClient) extractLabelsIntoTags(r FieldExtractionRule, labels map[string]string, tags map[string]string) {
	if r.Key == "*" {
		// Special case, extract everything
		for label, value := range labels {
			tags[fmt.Sprintf(r.Name, label)] = c.extractField(value, r)
		}
	} else {
		if v, ok := labels[r.Key]; ok {
			tags[r.Name] = c.extractField(v, r)
		}
	}
}

func (c *WatchClient) extractField(v string, r FieldExtractionRule) string {
	// Check if a subset of the field should be extracted with a regular expression
	// instead of the whole field.
	if r.Regex == nil {
		return v
	}

	matches := r.Regex.FindStringSubmatch(v)
	if len(matches) == 2 {
		return matches[1]
	}
	return ""
}

func (c *WatchClient) addOrUpdatePod(pod *api_v1.Pod) {
	newPod := &Pod{
		Name:            pod.Name,
		Namespace:       pod.Namespace,
		Address:         pod.Status.PodIP,
		PodUID:          string(pod.UID),
		StartTime:       pod.Status.StartTime,
		OwnerReferences: &pod.OwnerReferences,
	}

	if c.shouldIgnorePod(pod) {
		newPod.Ignore = true
	} else {
		newPod.Attributes = c.extractPodAttributes(pod)
	}

	c.m.Lock()
	defer c.m.Unlock()

	identifiers := []PodIdentifier{
		PodIdentifier(pod.UID),
		PodIdentifier(pod.Status.PodIP),
	}

	if newPod.Name != "" && newPod.Namespace != "" {
		identifiers = append(identifiers, generatePodIDFromName(newPod))
	}

	for _, identifier := range identifiers {
		if identifier != "" {
			// compare initial scheduled timestamp for existing pod and new pod with same identifier
			// and only replace old pod if scheduled time of new pod is newer or equal.
			// This should fix the case where scheduler has assigned the same attributes (like IP address)
			// to a new pod but update event for the old pod came in later.
			if p, ok := c.Pods[identifier]; ok {
				if p.StartTime != nil && pod.Status.StartTime.Before(p.StartTime) {
					continue
				}
			}
			c.Pods[identifier] = newPod
		}
	}
}

type Namer interface {
	GetName() string
	GetNamespace() string
}

func generatePodIDFromName(p Namer) PodIdentifier {
	return PodIdentifier(fmt.Sprintf("%s.%s", p.GetName(), p.GetNamespace()))
}

func (c *WatchClient) forgetPod(pod *api_v1.Pod) {
	p, ok := c.getPod(PodIdentifier(pod.Status.PodIP))
	if ok && p.Name == pod.Name {
		c.appendDeleteQueue(PodIdentifier(pod.Status.PodIP), pod.Name)
	}

	p, ok = c.getPod(PodIdentifier(pod.UID))
	if ok && p.Name == pod.Name {
		c.appendDeleteQueue(PodIdentifier(pod.UID), pod.Name)
	}

	id := generatePodIDFromName(pod)
	p, ok = c.getPod(id)
	if ok && p.Name == pod.Name {
		c.appendDeleteQueue(id, pod.Name)
	}
}

func (c *WatchClient) appendDeleteQueue(podID PodIdentifier, podName string) {
	c.deleteMut.Lock()
	c.deleteQueue = append(c.deleteQueue, deleteRequest{
		id:      podID,
		podName: podName,
		ts:      time.Now(),
	})
	c.deleteMut.Unlock()
}

func (c *WatchClient) shouldIgnorePod(pod *api_v1.Pod) bool {
	// Host network mode is not supported right now with IP based
	// tagging as all pods in host network get same IP addresses.
	// Such pods are very rare and usually are used to monitor or control
	// host traffic (e.g, linkerd, flannel) instead of service business needs.
	// We plan to support host network pods in future.
	if pod.Spec.HostNetwork {
		return true
	}

	// Check if user requested the pod to be ignored through annotations
	if v, ok := pod.Annotations[ignoreAnnotation]; ok {
		if strings.ToLower(strings.TrimSpace(v)) == "true" {
			return true
		}
	}

	// Check if user requested the pod to be ignored through configuration
	for _, excludedPod := range c.Exclude.Pods {
		if excludedPod.Name.MatchString(pod.Name) {
			return true
		}
	}

	return false
}

// selectors from Filters creates K8s selectors from Filters
// this notably does not include the Node filter, which only applies to Pods and is handled separately in addNodeSelector
func selectorsFromFilters(filters Filters) (labels.Selector, fields.Selector, error) {
	labelSelector := labels.Everything()
	for _, f := range filters.Labels {
		r, err := labels.NewRequirement(f.Key, f.Op, []string{f.Value})
		if err != nil {
			return nil, nil, err
		}
		labelSelector = labelSelector.Add(*r)
	}

	var selectors []fields.Selector
	for _, f := range filters.Fields {
		switch f.Op {
		case selection.Equals:
			selectors = append(selectors, fields.OneTermEqualSelector(f.Key, f.Value))
		case selection.NotEquals:
			selectors = append(selectors, fields.OneTermNotEqualSelector(f.Key, f.Value))
		default:
			return nil, nil, fmt.Errorf("field filters don't support operator: '%s'", f.Op)
		}
	}

	return labelSelector, fields.AndSelectors(selectors...), nil
}

func addNodeSelector(selector fields.Selector, node string) fields.Selector {
	nodeSelector := fields.OneTermEqualSelector(podNodeField, node)
	if selector.Empty() {
		return nodeSelector
	}
	return fields.AndSelectors(selector, nodeSelector)
}
