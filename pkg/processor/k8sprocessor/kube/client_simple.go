// Copyright 2020 Sumo Logic Inc
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

	"go.uber.org/zap"
	api_v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sprocessor/observability"
)

type InformerClient struct {
	logger    *zap.Logger
	kc        kubernetes.Interface
	informer  cache.SharedIndexInformer
	stopCh    chan struct{}
	op        OwnerAPI
	delimiter string

	Rules        ExtractionRules
	Filters      Filters
	Associations []Association
	Exclude      Excludes
}

const PodIPIndexName = "ip"
const PodUIDIndexName = "uid"

func podIPIndexFunc(obj interface{}) ([]string, error) {
	if pod, ok := obj.(*api_v1.Pod); ok {
		if pod.Status.PodIP != "" {
			return []string{pod.Status.PodIP}, nil
		} else {
			return []string{}, nil
		}
	} else {
		return []string{}, fmt.Errorf("invalid argument %v, must be core.v1.Pod", obj)
	}

}

func podUIDIndexFunc(obj interface{}) ([]string, error) {
	if pod, ok := obj.(*api_v1.Pod); ok {
		if pod.ObjectMeta.UID != "" {
			return []string{string(pod.ObjectMeta.UID)}, nil
		} else {
			return []string{}, nil
		}
	} else {
		return []string{}, fmt.Errorf("invalid argument %v, must be core.v1.Pod", obj)
	}

}

// New initializes a new k8s Client.
func NewInformerClient(
	logger *zap.Logger,
	apiCfg k8sconfig.APIConfig,
	rules ExtractionRules,
	filters Filters,
	associations []Association,
	exclude Excludes,
	newClientSet APIClientsetProvider,
	newInformer IndexInformerProvider,
	newOwnerProviderFunc OwnerProvider,
	delimiter string,
) (*InformerClient, error) {
	c := &InformerClient{
		logger:       logger,
		Rules:        rules,
		Filters:      filters,
		Associations: associations,
		Exclude:      exclude,
		stopCh:       make(chan struct{}),
		delimiter:    delimiter,
	}

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
		newInformer = NewIndexedPodInformer
	}

	c.informer = newInformer(c.kc, c.Filters.Namespace, labelSelector, fieldSelector)
	return c, err
}

// Start registers pod event handlers and starts watching the kubernetes cluster for pod changes.
func (c *InformerClient) Start() {
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
func (c *InformerClient) Stop() {
	close(c.stopCh)

	if c.op != nil {
		c.op.Stop()
	}
}

func (c *InformerClient) handlePodAdd(obj interface{}) {
	observability.RecordPodAdded()
}

func (c *InformerClient) handlePodUpdate(old, new interface{}) {
	observability.RecordPodUpdated()
}

func (c *InformerClient) handlePodDelete(obj interface{}) {
	observability.RecordPodDeleted()
}

func (c *InformerClient) GetPod(identifier PodIdentifier) (*Pod, bool) {
	pod, ok := c.getPodFromInformerStore(identifier)
	if !ok {
		observability.RecordIPLookupMiss()
		return nil, false
	}

	podRecord := c.getPodRecord(pod)
	if podRecord.Ignore {
		return nil, false
	}
	return podRecord, ok

}

func (c *InformerClient) getPodFromInformerStore(identifier PodIdentifier) (*api_v1.Pod, bool) {
	indexer := c.informer.GetIndexer()
	stringId := string(identifier)
	if obj, exists, err := indexer.GetByKey(stringId); exists && err == nil {
		pod := obj.(api_v1.Pod)
		return &pod, exists
	} else if err != nil {
		c.logger.Error("error fetching pod from cache", zap.Any("message", err))
	}

	if objs, err := indexer.ByIndex(PodUIDIndexName, stringId); err == nil {
		if len(objs) > 1 {
			c.logger.Error("multiple pod entries for uid", zap.String("pod uid", stringId))
		}
		if len(objs) == 1 {
			pod := objs[0].(api_v1.Pod)
			return &pod, true
		}
	} else if err != nil {
		c.logger.Error("error fetching pod from cache", zap.Any("message", err))
	}

	if objs, err := indexer.ByIndex(PodIPIndexName, stringId); err == nil {
		// take the most recent value
		if len(objs) > 0 {
			pods := make([]api_v1.Pod, len(objs))
			for i := range objs {
				pods[i] = objs[i].(api_v1.Pod)
			}

			// simple linear search, not worth sorting, as it is unlikely to ever be more than 10 elements
			mostRecent := pods[0]
			if len(pods) == 1 {
				return &mostRecent, true
			}
			for _, pod := range pods {
				if pod.Status.StartTime.After(mostRecent.Status.StartTime.Time) {
					mostRecent = pod
				}
			}

			return &mostRecent, true
		}
	} else if err != nil {
		c.logger.Error("error fetching pod from cache", zap.Any("message", err))
	}

	return nil, false
}

func (c *InformerClient) extractPodAttributes(pod *api_v1.Pod) map[string]string {
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
		c.logger.Debug("pod owner lookup",
			zap.String("pod.Name", pod.Name),
			zap.Any("pod.OwnerReferences", pod.OwnerReferences),
		)
		owners := c.op.GetOwners(pod)

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
			if services := c.op.GetServices(pod); len(services) > 0 {
				tags[c.Rules.Tags.ServiceName] = strings.Join(services, c.delimiter)
			}
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

func (c *InformerClient) extractLabelsIntoTags(r FieldExtractionRule, labels map[string]string, tags map[string]string) {
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

func (c *InformerClient) extractField(v string, r FieldExtractionRule) string {
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

func (c *InformerClient) getPodRecord(pod *api_v1.Pod) *Pod {
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
	return newPod
}

func (c *InformerClient) shouldIgnorePod(pod *api_v1.Pod) bool {
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
