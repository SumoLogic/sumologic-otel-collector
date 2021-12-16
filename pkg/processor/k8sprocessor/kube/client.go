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
) (Client, error) {
	c := &WatchClient{
		logger:       logger,
		Rules:        rules,
		Filters:      filters,
		Associations: associations,
		Exclude:      exclude,
		stopCh:       make(chan struct{}),
		delimiter:    delimiter,
	}
	go c.deleteLoop(time.Second*30, defaultPodDeleteGracePeriod)

	c.Pods = map[PodIdentifier]*Pod{}
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

		c.op, err = newOwnerProviderFunc(logger, c.kc, labelSelector, fieldSelector, c.Filters.Namespace)
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

	c.informer = newInformer(c.kc, c.Filters.Namespace, labelSelector, fieldSelector)
	return c, err
}

// Start registers pod event handlers and starts watching the kubernetes cluster for pod changes.
func (c *WatchClient) Start() {
	if c.op != nil {
		c.op.Start()
	}

	c.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.handlePodAdd,
		UpdateFunc: c.handlePodUpdate,
		DeleteFunc: c.handlePodDelete,
	})
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
	podTableSize := len(c.Pods)
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
	podTableSize := len(c.Pods)
	observability.RecordPodTableSize(int64(podTableSize))
}

func (c *WatchClient) handlePodDelete(obj interface{}) {
	observability.RecordPodDeleted()
	if pod, ok := obj.(*api_v1.Pod); ok {
		c.forgetPod(pod)
	} else {
		c.logger.Error("object received was not of type api_v1.Pod", zap.Any("received", obj))
	}
	podTableSize := len(c.Pods)
	observability.RecordPodTableSize(int64(podTableSize))
}

func (c *WatchClient) deleteLoop(interval time.Duration, gracePeriod time.Duration) {
	// This loop runs after N seconds and deletes pods from cache.
	// It iterates over the delete queue and deletes all that aren't
	// in the grace period anymore.
	for {
		select {
		case <-time.After(interval):
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
			observability.RecordPodTableSize(int64(podTableSize))
			c.m.Unlock()

		case <-c.stopCh:
			return
		}
	}
}

// GetPod takes an IP address or Pod UID and returns the pod the identifier is associated with.
func (c *WatchClient) GetPod(identifier PodIdentifier) (*Pod, bool) {
	c.m.RLock()
	pod, ok := c.Pods[identifier]
	c.m.RUnlock()
	if ok {
		if pod.Ignore {
			return nil, false
		}
		return pod, ok
	}
	observability.RecordIPLookupMiss()
	return nil, false
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

	if c.Rules.ClusterName {
		clusterName := pod.GetClusterName()
		if clusterName != "" {
			tags[c.Rules.Tags.ClusterName] = clusterName
		}
	}

	if c.Rules.OwnerLookupEnabled {
		c.logger.Info("pod owner lookup", zap.Any("pod", pod.Name), zap.Any("owner refs", pod.OwnerReferences))
		owners := c.op.GetOwners(pod)
		c.logger.Info("pod owner lookup #2", zap.Any("owners", owners))

		for _, owner := range owners {
			switch owner.kind {
			case "DaemonSet":
				if c.Rules.DaemonSetName {
					tags[c.Rules.Tags.DaemonSetName] = owner.name
				}
			case "Deployment":
				if c.Rules.DeploymentName {
					tags[c.Rules.Tags.DeploymentName] = owner.name
				}
			case "ReplicaSet":
				if c.Rules.ReplicaSetName {
					tags[c.Rules.Tags.ReplicaSetName] = owner.name
				}
			case "StatefulSet":
				if c.Rules.StatefulSetName {
					tags[c.Rules.Tags.StatefulSetName] = owner.name
				}
			case "Job":
				if c.Rules.JobName {
					tags[c.Rules.Tags.JobName] = owner.name
				}
			case "CronJob":
				if c.Rules.CronJobName {
					tags[c.Rules.Tags.CronJobName] = owner.name
				}

			default:
				// Do nothing
			}
		}

		if c.Rules.ServiceName {
			tags[c.Rules.Tags.ServiceName] = strings.Join(c.op.GetServices(pod), c.delimiter)
		}

	}

	if len(pod.Status.ContainerStatuses) > 0 {
		cs := pod.Status.ContainerStatuses[0]
		if c.Rules.ContainerID {
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

	if c.Rules.PodUID {
		tags[c.Rules.Tags.PodUID] = string(pod.UID)
	}

	for _, r := range c.Rules.Labels {
		c.extractLabelsIntoTags(r, pod.Labels, tags)
	}

	if len(c.Rules.NamespaceLabels) > 0 && c.Rules.OwnerLookupEnabled {
		namespace := c.op.GetNamespace(pod)
		if namespace != nil {
			for _, r := range c.Rules.NamespaceLabels {
				c.extractLabelsIntoTags(r, namespace.Labels, tags)
			}
		}
	}

	for _, r := range c.Rules.Annotations {
		c.extractLabelsIntoTags(r, pod.Annotations, tags)
	}
	return tags
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
		Name:      pod.Name,
		Namespace: pod.Namespace,
		Address:   pod.Status.PodIP,
		PodUID:    string(pod.UID),
		StartTime: pod.Status.StartTime,
	}

	if c.shouldIgnorePod(pod) {
		newPod.Ignore = true
	} else {
		newPod.Attributes = c.extractPodAttributes(pod)
	}

	c.m.Lock()
	defer c.m.Unlock()

	if pod.UID != "" {
		c.Pods[PodIdentifier(pod.UID)] = newPod
	}
	if pod.Status.PodIP != "" {
		// compare initial scheduled timestamp for existing pod and new pod with same IP
		// and only replace old pod if scheduled time of new pod is newer? This should fix
		// the case where scheduler has assigned the same IP to a new pod but update event for
		// the old pod came in later.
		if p, ok := c.Pods[PodIdentifier(pod.Status.PodIP)]; ok {
			if p.StartTime != nil && pod.Status.StartTime.Before(p.StartTime) {
				return
			}
		}
		c.Pods[PodIdentifier(pod.Status.PodIP)] = newPod
	}
	// Use pod_name.namespace_name identifier
	if newPod.Name != "" && newPod.Namespace != "" {
		c.Pods[PodIdentifier(fmt.Sprintf("%s.%s", newPod.Name, newPod.Namespace))] = newPod
	}
}

func (c *WatchClient) forgetPod(pod *api_v1.Pod) {
	c.m.RLock()
	p, ok := c.GetPod(PodIdentifier(pod.Status.PodIP))
	c.m.RUnlock()

	if ok && p.Name == pod.Name {
		c.appendDeleteQueue(PodIdentifier(pod.Status.PodIP), pod.Name)
	}

	c.m.RLock()
	p, ok = c.GetPod(PodIdentifier(pod.UID))
	c.m.RUnlock()

	if ok && p.Name == pod.Name {
		c.appendDeleteQueue(PodIdentifier(pod.UID), pod.Name)
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

	if filters.Node != "" {
		selectors = append(selectors, fields.OneTermEqualSelector(podNodeField, filters.Node))
	}
	return labelSelector, fields.AndSelectors(selectors...), nil
}
