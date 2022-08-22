// Copyright 2019 OpenTelemetry Authors
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
	"sort"
	"sync"

	"go.uber.org/zap"
	api_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sprocessor/observability"
)

// OwnerProvider allows to dynamically assign constructor
type OwnerProvider func(
	logger *zap.Logger,
	client kubernetes.Interface,
	labelSelector labels.Selector,
	fieldSelector fields.Selector,
	extractionRules ExtractionRules,
	namespace string,
) (OwnerAPI, error)

// ObjectOwner keeps single entry
type ObjectOwner struct {
	UID       types.UID
	namespace string
	kind      string
	name      string
	ownerUIDs []types.UID
}

// OwnerAPI describes functions that could allow retrieving owner info
type OwnerAPI interface {
	GetOwners(pod *Pod) []*ObjectOwner
	GetNamespace(pod *api_v1.Pod) *api_v1.Namespace
	GetServices(podName string) []string
	Start()
	Stop()
}

// OwnerCache is a simple structure which aids querying for owners
type OwnerCache struct {
	objectOwners map[string]*ObjectOwner
	ownersMutex  sync.RWMutex

	podServices      map[string][]string
	podServicesMutex sync.RWMutex

	namespaces map[string]*api_v1.Namespace
	nsMutex    sync.RWMutex

	logger *zap.Logger

	stopCh    chan struct{}
	informers []cache.SharedIndexInformer
}

func newOwnerCache(logger *zap.Logger) OwnerCache {
	return OwnerCache{
		objectOwners: map[string]*ObjectOwner{},
		podServices:  map[string][]string{},
		namespaces:   map[string]*api_v1.Namespace{},
		logger:       logger,
		stopCh:       make(chan struct{}),
	}
}

// Start runs the informers
func (op *OwnerCache) Start() {
	op.logger.Info("Staring K8S resource informers", zap.Int("#infomers", len(op.informers)))
	for _, informer := range op.informers {
		go informer.Run(op.stopCh)
	}
}

// Stop shutdowns the informers
func (op *OwnerCache) Stop() {
	close(op.stopCh)
}

func newOwnerProvider(
	logger *zap.Logger,
	client kubernetes.Interface,
	labelSelector labels.Selector,
	fieldSelector fields.Selector,
	extractionRules ExtractionRules,
	namespace string) (OwnerAPI, error) {

	ownerCache := newOwnerCache(logger)

	factory := informers.NewSharedInformerFactoryWithOptions(client, watchSyncPeriod,
		informers.WithNamespace(namespace),
		informers.WithTweakListOptions(func(opts *meta_v1.ListOptions) {
			opts.LabelSelector = labelSelector.String()
			opts.FieldSelector = fieldSelector.String()
		}))

	ownerCache.addNamespaceInformer(factory)

	// Only enable DaemonSet informer when DaemonSet extraction rule is enabled
	if extractionRules.DaemonSetName {
		logger.Debug("adding informer for DaemonSet", zap.String("api_version", "apps/v1"))
		ownerCache.addOwnerInformer("DaemonSet",
			factory.Apps().V1().DaemonSets().Informer(),
			ownerCache.cacheObject,
			ownerCache.deleteObject)
	}

	// Only enable ReplicaSet informer when ReplicaSet or DeploymentName extraction rule is enabled
	if extractionRules.ReplicaSetName || extractionRules.DeploymentName {
		logger.Debug("adding informer for ReplicaSet", zap.String("api_version", "apps/v1"))
		ownerCache.addOwnerInformer("ReplicaSet",
			factory.Apps().V1().ReplicaSets().Informer(),
			ownerCache.cacheObject,
			ownerCache.deleteObject)
	}

	// Only enable Deployment informer when Deployment extraction rule is enabled
	if extractionRules.DeploymentName {
		logger.Debug("adding informer for Deployment", zap.String("api_version", "apps/v1"))
		ownerCache.addOwnerInformer("Deployment",
			factory.Apps().V1().Deployments().Informer(),
			ownerCache.cacheObject,
			ownerCache.deleteObject)
	}

	// Only enable StatefulSet informer when StatefulSet extraction rule is enabled
	if extractionRules.StatefulSetName {
		logger.Debug("adding informer for StatefulSet", zap.String("api_version", "apps/v1"))
		ownerCache.addOwnerInformer("StatefulSet",
			factory.Apps().V1().StatefulSets().Informer(),
			ownerCache.cacheObject,
			ownerCache.deleteObject)
	}

	// Only enable Endpoint informer when Endpoint extraction rule is enabled
	if extractionRules.ServiceName {
		logger.Debug("adding informer for Endpoint", zap.String("api_version", "v1"))
		ownerCache.addOwnerInformer("Endpoint",
			factory.Core().V1().Endpoints().Informer(),
			ownerCache.cacheEndpoint,
			ownerCache.deleteEndpoint)
	}

	// Only enable Job informer when Job or CronJob extraction rule is enabled
	if extractionRules.JobName || extractionRules.CronJobName {
		logger.Debug("adding informer for Job", zap.String("api_version", "batch/v1"))
		ownerCache.addOwnerInformer("Job",
			factory.Batch().V1().Jobs().Informer(),
			ownerCache.cacheObject,
			ownerCache.deleteObject)
	}

	// Only enable CronJob informer when CronJob extraction rule is enabled
	// and when a particular API is available on the cluster
	if extractionRules.CronJobName {
		// Other resources used in remaining informers are all available in all supported
		// cluster versions. Only CronJob from batch/v1 is available starting with k8s 1.21
		// hence make this conditional on the supported batch API group version.
		apiGroups, apiResList, err := client.Discovery().ServerGroupsAndResources()
		if err != nil {
			ownerCache.logger.Debug(
				"failed to get server resources with client-go",
				zap.Error(err),
			)
		} else {
			enableCronJobInformer := func(informer cache.SharedIndexInformer) {
				ownerCache.addOwnerInformer("CronJob",
					informer,
					ownerCache.cacheObject,
					ownerCache.deleteObject)
			}

			handleAPIResources := func(informer cache.SharedIndexInformer, apiResources []meta_v1.APIResource) bool {
				for _, apiR := range apiResources {
					if apiR.Name == "cronjobs" && apiR.Kind == "CronJob" {
						logger.Debug("adding informer for CronJob", zap.String("api_version", apiR.Version))
						enableCronJobInformer(informer)
						return true
					}
				}
				return false
			}

			var preferredBatchVersion string
			for _, g := range apiGroups {
				if g.Name == "batch" {
					preferredBatchVersion = g.PreferredVersion.GroupVersion
					break
				}
			}

			if preferredBatchVersion != "" {
			outerLoop:
				for _, v := range apiResList {
					if v.GroupVersion == preferredBatchVersion {
						switch v.GroupVersion {
						case "batch/v1":
							informer := factory.Batch().V1().CronJobs().Informer()
							if enabled := handleAPIResources(informer, v.APIResources); enabled {
								break outerLoop
							}
						case "batch/v1beta1":
							informer := factory.Batch().V1beta1().CronJobs().Informer()
							if enabled := handleAPIResources(informer, v.APIResources); enabled {
								break outerLoop
							}
						}
					}
				}
			}
		}
	}

	return &ownerCache, nil
}

func (op *OwnerCache) upsertNamespace(obj interface{}) {
	namespace := obj.(*api_v1.Namespace)
	op.nsMutex.Lock()
	op.namespaces[namespace.Name] = namespace
	op.nsMutex.Unlock()
}

func (op *OwnerCache) deleteNamespace(obj interface{}) {
	namespace := obj.(*api_v1.Namespace)
	op.nsMutex.Lock()
	delete(op.namespaces, namespace.Name)
	op.nsMutex.Unlock()
}

func (op *OwnerCache) addNamespaceInformer(factory informers.SharedInformerFactory) {
	op.logger.Debug("adding informer for Namespace", zap.String("api_version", "v1"))
	informer := factory.Core().V1().Namespaces().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			observability.RecordOtherAdded()
			op.upsertNamespace(obj)
		},
		UpdateFunc: func(_, obj interface{}) {
			observability.RecordOtherUpdated()
			op.upsertNamespace(obj)
		},
		DeleteFunc: func(obj interface{}) {
			observability.RecordOtherDeleted()
			op.deleteNamespace(obj)
		},
	})

	op.informers = append(op.informers, informer)
}

func (op *OwnerCache) addOwnerInformer(
	kind string,
	informer cache.SharedIndexInformer,
	cacheFunc func(kind string, obj interface{}),
	deleteFunc func(obj interface{})) {
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			cacheFunc(kind, obj)
			observability.RecordOtherAdded()
		},
		UpdateFunc: func(_, obj interface{}) {
			cacheFunc(kind, obj)
			observability.RecordOtherUpdated()
		},
		DeleteFunc: func(obj interface{}) {
			deleteFunc(obj)
			observability.RecordOtherDeleted()
		},
	})

	op.informers = append(op.informers, informer)
}

func (op *OwnerCache) deleteObject(obj interface{}) {
	op.ownersMutex.Lock()
	delete(op.objectOwners, string(obj.(meta_v1.Object).GetUID()))
	op.ownersMutex.Unlock()
}

func (op *OwnerCache) cacheObject(kind string, obj interface{}) {
	meta := obj.(meta_v1.Object)

	oo := ObjectOwner{
		UID:       meta.GetUID(),
		namespace: meta.GetNamespace(),
		ownerUIDs: []types.UID{},
		kind:      kind,
		name:      meta.GetName(),
	}
	for _, or := range meta.GetOwnerReferences() {
		oo.ownerUIDs = append(oo.ownerUIDs, or.UID)
	}

	op.ownersMutex.Lock()
	op.objectOwners[string(oo.UID)] = &oo
	op.ownersMutex.Unlock()
}

func (op *OwnerCache) addEndpointToPod(pod string, endpoint string) {
	op.podServicesMutex.Lock()
	defer op.podServicesMutex.Unlock()

	services, ok := op.podServices[pod]
	if !ok {
		// If there's no services/endpoints for a given pod then just update the cache
		// with the provided enpoint.
		op.podServices[pod] = []string{endpoint}
		return
	}

	for _, it := range services {
		if it == endpoint {
			return
		}
	}

	services = append(services, endpoint)
	sort.Strings(services)
	op.podServices[pod] = services
}

func (op *OwnerCache) deleteEndpointFromPod(pod string, endpoint string) {
	op.podServicesMutex.Lock()
	defer op.podServicesMutex.Unlock()

	services, ok := op.podServices[pod]
	if !ok {
		return
	}

	for i := 0; len(services) > 0; {
		service := services[i]
		if service == endpoint {
			// Remove the ith entry by...
			l := len(services)
			last := services[l-1]
			// ...moving it at the very end (swapping it with the last entry)...
			services[l-1], services[i] = service, last
			// ... and by truncating the slice by one elem
			services = services[:l-1]
		} else {
			i++
		}

		if i == len(services)-1 {
			break
		}
	}

	if len(services) == 0 {
		delete(op.podServices, pod)
	} else {
		sort.Strings(services)
		op.podServices[pod] = services
	}
}

func (op *OwnerCache) genericEndpointOp(obj interface{}, endpointFunc func(pod string, endpoint string)) {
	ep := obj.(*api_v1.Endpoints)

	for _, it := range ep.Subsets {
		for _, addr := range it.Addresses {
			if addr.TargetRef != nil && addr.TargetRef.Kind == "Pod" {
				endpointFunc(addr.TargetRef.Name, ep.Name)
			}
		}
		for _, addr := range it.NotReadyAddresses {
			if addr.TargetRef != nil && addr.TargetRef.Kind == "Pod" {
				endpointFunc(addr.TargetRef.Name, ep.Name)
			}
		}
	}
}

func (op *OwnerCache) deleteEndpoint(obj interface{}) {
	op.genericEndpointOp(obj, op.deleteEndpointFromPod)
}

func (op *OwnerCache) cacheEndpoint(kind string, obj interface{}) {
	op.genericEndpointOp(obj, op.addEndpointToPod)
}

// GetNamespaces returns a cached namespace object (if one is found) or nil otherwise
func (op *OwnerCache) GetNamespace(pod *api_v1.Pod) *api_v1.Namespace {
	op.nsMutex.RLock()
	namespace, found := op.namespaces[pod.Namespace]
	op.nsMutex.RUnlock()

	if found {
		return namespace
	}
	return nil
}

// GetServices returns a slice with matched services - in case no services are found, it returns an empty slice

func (op *OwnerCache) GetServices(podName string) []string {
	op.podServicesMutex.RLock()
	oo, found := op.podServices[podName]
	op.podServicesMutex.RUnlock()

	if found {
		return oo
	}
	return []string{}
}

// GetOwners goes through the cached data and assigns relevant metadata for pod
func (op *OwnerCache) GetOwners(pod *Pod) []*ObjectOwner {
	objectOwners := []*ObjectOwner{}

	visited := map[types.UID]bool{}
	queue := []types.UID{}

	for _, or := range *pod.OwnerReferences {
		if _, uidVisited := visited[or.UID]; !uidVisited {
			queue = append(queue, or.UID)
			visited[or.UID] = true
		}
	}

	for len(queue) > 0 {
		uid := queue[0]
		queue = queue[1:]

		op.ownersMutex.RLock()
		oo, found := op.objectOwners[string(uid)]
		if found {
			objectOwners = append(objectOwners, oo)

			for _, ownerUID := range oo.ownerUIDs {
				if _, uidVisited := visited[ownerUID]; !uidVisited {
					queue = append(queue, ownerUID)
					visited[ownerUID] = true
				}
			}
		}
		op.ownersMutex.RUnlock()
	}

	return objectOwners
}
