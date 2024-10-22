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
	"time"

	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	api_v1 "k8s.io/api/core/v1"
	discovery_v1 "k8s.io/api/discovery/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sprocessor/observability"
)

const endpointSliceServiceLabel = "kubernetes.io/service-name"

// OwnerProvider allows to dynamically assign constructor
type OwnerProvider func(
	logger *zap.Logger,
	client kubernetes.Interface,
	labelSelector labels.Selector,
	fieldSelector fields.Selector,
	extractionRules ExtractionRules,
	namespace string,
	deleteInterval time.Duration,
	gracePeriod time.Duration,
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

	deleteQueue []ownerCacheEviction
	deleteMu    sync.Mutex

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
	namespace string,
	deleteInterval time.Duration,
	gracePeriod time.Duration,
) (OwnerAPI, error) {

	ownerCache := newOwnerCache(logger)
	go ownerCache.deleteLoop(deleteInterval, gracePeriod)

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
			ownerCache.deleteObject,
			nil,
			nil,
		)
	}

	// Only enable ReplicaSet informer when ReplicaSet or DeploymentName extraction rule is enabled
	if extractionRules.ReplicaSetName || extractionRules.DeploymentName {
		logger.Debug("adding informer for ReplicaSet", zap.String("api_version", "apps/v1"))
		ownerCache.addOwnerInformer("ReplicaSet",
			factory.Apps().V1().ReplicaSets().Informer(),
			ownerCache.cacheObject,
			ownerCache.deleteObject,
			nil,
			nil,
		)
	}

	// Only enable Deployment informer when Deployment extraction rule is enabled
	if extractionRules.DeploymentName {
		logger.Debug("adding informer for Deployment", zap.String("api_version", "apps/v1"))
		ownerCache.addOwnerInformer("Deployment",
			factory.Apps().V1().Deployments().Informer(),
			ownerCache.cacheObject,
			ownerCache.deleteObject,
			nil,
			nil,
		)
	}

	// Only enable StatefulSet informer when StatefulSet extraction rule is enabled
	if extractionRules.StatefulSetName {
		logger.Debug("adding informer for StatefulSet", zap.String("api_version", "apps/v1"))
		ownerCache.addOwnerInformer("StatefulSet",
			factory.Apps().V1().StatefulSets().Informer(),
			ownerCache.cacheObject,
			ownerCache.deleteObject,
			nil,
			nil,
		)
	}

	// Only enable Endpoint informer when Endpoint extraction rule is enabled
	if extractionRules.ServiceName {
		logger.Debug("adding informer for EndpointSlice", zap.String("api_version", "discovery.k8s.io/v1"))
		ownerCache.addOwnerInformer("EndpointSlice",
			factory.Discovery().V1().EndpointSlices().Informer(),
			ownerCache.cacheEndpointSlice,
			ownerCache.deleteEndpointSlice,
			ownerCache.updateEndpointSlice,
			func(object interface{}) (interface{}, error) {
				originalES, success := object.(*discovery_v1.EndpointSlice)
				if !success {
					return object.(cache.DeletedFinalStateUnknown), nil
				} else {
					return removeUnnecessaryEndpointSliceData(originalES), nil
				}
			},
		)
	}

	// Only enable Job informer when Job or CronJob extraction rule is enabled
	if extractionRules.JobName || extractionRules.CronJobName {
		logger.Debug("adding informer for Job", zap.String("api_version", "batch/v1"))
		ownerCache.addOwnerInformer("Job",
			factory.Batch().V1().Jobs().Informer(),
			ownerCache.cacheObject,
			ownerCache.deleteObject,
			nil,
			nil,
		)
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
					ownerCache.deleteObject,
					nil,
					nil,
				)
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
	var ns *api_v1.Namespace

	switch obj := obj.(type) {
	case *api_v1.Namespace:
		ns = obj
	case cache.DeletedFinalStateUnknown:
		prev, ok := obj.Obj.(*api_v1.Namespace)
		if !ok {
			op.logger.Error(
				"object received was DeletedFinalStateUnknown but did not contain api_v1.Namespace",
				zap.Any("received", obj),
			)
			return
		}
		ns = prev
	default:
		op.logger.Error("object received was not of type api_v1.Namespace", zap.Any("received", obj))
		return
	}

	op.nsMutex.Lock()
	delete(op.namespaces, ns.Name)
	op.nsMutex.Unlock()
}

func (op *OwnerCache) addNamespaceInformer(factory informers.SharedInformerFactory) {
	op.logger.Debug("adding informer for Namespace", zap.String("api_version", "v1"))
	informer := factory.Core().V1().Namespaces().Informer()
	_, err := informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			observability.RecordOtherAdded("Namespace")
			op.upsertNamespace(obj)
		},
		UpdateFunc: func(_, obj interface{}) {
			observability.RecordOtherUpdated("Namespace")
			op.upsertNamespace(obj)
		},
		DeleteFunc: op.deferredDelete(func(obj interface{}) {
			observability.RecordOtherDeleted("Namespace")
			op.deleteNamespace(obj)
		}),
	})
	if err != nil {
		op.logger.Error("error adding event handler to namespace informer", zap.Error(err))
	}

	op.informers = append(op.informers, informer)
}

// deferredDelete returns a function that will handle deleting an object from
// the owner cache eventually through the owner cache deleteQueue. Takes an
// evict function that should contain the logic for processing the deletion.
func (op *OwnerCache) deferredDelete(evict func(obj any)) func(any) {
	return func(obj any) {
		op.deleteMu.Lock()
		op.deleteQueue = append(op.deleteQueue, ownerCacheEviction{
			ts:    time.Now(),
			evict: func() { evict(obj) },
		})
		op.deleteMu.Unlock()
	}
}

func (op *OwnerCache) addOwnerInformer(
	kind string,
	informer cache.SharedIndexInformer,
	addFunc func(kind string, obj interface{}),
	deleteFunc func(obj interface{}),
	updateFunc func(oldObj, newObj interface{}),
	transformFunc cache.TransformFunc,
) {
	// if updatefunc is not specified, use addFunc
	if updateFunc == nil {
		updateFunc = func(_, obj interface{}) {
			addFunc(kind, obj)
		}
	}
	_, err := informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			addFunc(kind, obj)
			observability.RecordOtherAdded(kind)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			updateFunc(oldObj, newObj)
			observability.RecordOtherUpdated(kind)
		},
		DeleteFunc: op.deferredDelete(func(obj any) {
			deleteFunc(obj)
			observability.RecordOtherDeleted(kind)
		}),
	})
	if err != nil {
		op.logger.Error("error adding event handler to owner informer", zap.Error(err), zap.String("kind", kind))
	}

	if transformFunc != nil {
		if err = informer.SetTransform(transformFunc); err != nil {
			op.logger.Error("error adding transform to owner informer", zap.Error(err), zap.String("kind", kind))
		}
	}

	op.informers = append(op.informers, informer)
}

func (op *OwnerCache) deleteObject(obj interface{}) {
	var metaObj meta_v1.Object

	switch obj := obj.(type) {
	case meta_v1.Object:
		metaObj = obj
	case cache.DeletedFinalStateUnknown:
		prev, ok := obj.Obj.(meta_v1.Object)
		if !ok {
			op.logger.Error(
				"object received was DeletedFinalStateUnknown but did not contain meta_v1.Object",
				zap.Any("received", obj),
			)
			return
		}
		metaObj = prev
	default:
		op.logger.Error(
			"object received was not of type meta_v1.Object",
			zap.Any("received", obj),
		)
		return
	}

	op.ownersMutex.Lock()
	delete(op.objectOwners, string(metaObj.GetUID()))
	ownerTableSize := len(op.objectOwners)
	op.ownersMutex.Unlock()
	observability.RecordOwnerTableSize(int64(ownerTableSize))
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
	ownerTableSize := len(op.objectOwners)
	op.ownersMutex.Unlock()
	observability.RecordOwnerTableSize(int64(ownerTableSize))
}

func (op *OwnerCache) addServiceToPod(pod string, serviceName string) {
	op.podServicesMutex.Lock()
	defer op.podServicesMutex.Unlock()

	services, ok := op.podServices[pod]
	if !ok {
		// If there's no services/endpoints for a given pod then just update the cache
		// with the provided enpoint.
		op.podServices[pod] = []string{serviceName}
		observability.RecordServiceTableSize(int64(len(op.podServices)))
		return
	}

	if idx := slices.Index(services, serviceName); idx >= 0 {
		return
	}

	services = append(services, serviceName)
	sort.Strings(services)
	op.podServices[pod] = services
}

func (op *OwnerCache) deleteServiceFromPod(pod string, serviceName string) {
	op.podServicesMutex.Lock()
	defer op.podServicesMutex.Unlock()

	services, ok := op.podServices[pod]
	if !ok {
		return
	}

	services = slices.DeleteFunc(services, func(s string) bool {
		return s == serviceName
	})

	if len(services) == 0 {
		delete(op.podServices, pod)
		observability.RecordServiceTableSize(int64(len(op.podServices)))
	} else {
		sort.Strings(services)
		op.podServices[pod] = services
	}
}

func (op *OwnerCache) genericEndpointSliceOp(obj interface{}, serviceFunc func(pod string, serviceName string)) {
	var endpointSlice *discovery_v1.EndpointSlice
	switch obj := obj.(type) {
	case *discovery_v1.EndpointSlice:
		endpointSlice = obj
	case cache.DeletedFinalStateUnknown:
		prev, ok := obj.Obj.(*discovery_v1.EndpointSlice)
		if !ok {
			op.logger.Error(
				"object received was DeletedFinalStateUnknown but did not contain EndpointSlice",
				zap.Any("received", obj),
			)
			return
		}
		endpointSlice = prev
	default:
		op.logger.Error(
			"object received was not of type EndpointSlice",
			zap.Any("received", obj),
		)
		return
	}

	serviceName := getServiceName(endpointSlice)

	for _, endpoint := range endpointSlice.Endpoints {
		if endpoint.TargetRef != nil && endpoint.TargetRef.Kind == "Pod" {
			podName := endpoint.TargetRef.Name
			serviceFunc(podName, serviceName)
		}
	}
}

func (op *OwnerCache) updateEndpointSlice(oldObj interface{}, newObj interface{}) {
	// for updates, we're guaranteed the objects will be the right type
	oldEndpointSlice := oldObj.(*discovery_v1.EndpointSlice)
	newEndpointSlice := newObj.(*discovery_v1.EndpointSlice)

	// add the new endpointslice first, the logic is the same
	op.cacheEndpointSlice("EndpointSlice", newObj)

	// we also need to remove the Service from Pods which were deleted from the endpointslice
	serviceName := getServiceName(newEndpointSlice)

	newPodNames := []string{}
	for _, endpoint := range newEndpointSlice.Endpoints {
		if endpoint.TargetRef != nil && endpoint.TargetRef.Kind == "Pod" {
			podName := endpoint.TargetRef.Name
			newPodNames = append(newPodNames, podName)
		}
	}

	// for each Pod name which was in the old slice, but not in the new slice, schedule a delete
	for _, endpoint := range oldEndpointSlice.Endpoints {
		if endpoint.TargetRef != nil && endpoint.TargetRef.Kind == "Pod" {
			podName := endpoint.TargetRef.Name
			if slices.Index(newPodNames, podName) == -1 {
				// not a deferred delete, as this is a dynamic property which can change often
				op.deleteServiceFromPod(podName, serviceName)
			}
		}
	}
}

func (op *OwnerCache) deleteEndpointSlice(obj interface{}) {
	op.genericEndpointSliceOp(obj, op.deleteServiceFromPod)
}

func (op *OwnerCache) cacheEndpointSlice(kind string, obj interface{}) {
	op.genericEndpointSliceOp(obj, op.addServiceToPod)
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
	defer op.podServicesMutex.RUnlock()
	oo, found := op.podServices[podName]
	if !found {
		return []string{}
	}

	return append([]string(nil), oo...)
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
		} else {
			// try to find the owner
			var ownerReference meta_v1.OwnerReference
			for _, or := range *pod.OwnerReferences {
				if or.UID == uid {
					ownerReference = or
				}
			}
			// ownerReference may be empty
			op.logger.Debug(
				"missing owner data for Pod, cache may be out of sync",
				zap.String("pod", pod.Name),
				zap.String("owner_id", string(uid)),
				zap.String("owner_kind", ownerReference.Kind),
				zap.String("owner_name", ownerReference.Name),
			)
		}
		op.ownersMutex.RUnlock()
	}

	return objectOwners
}

// deleteLoop runs along side the owner cache, checking for and deleting cache
// entries that have been marked for deletion for over the duration of the
// grace period.
func (op *OwnerCache) deleteLoop(interval time.Duration, gracePeriod time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for _, d := range op.nextDeleteQueue(gracePeriod) {
				d.evict()
			}
		case <-op.stopCh:
			return
		}
	}
}

// nextDeleteQueue pops the evictions older than the gracePeriod from the
// cache's deleteQueue
func (op *OwnerCache) nextDeleteQueue(gracePeriod time.Duration) []ownerCacheEviction {
	var cutoff int
	now := time.Now()
	op.deleteMu.Lock()
	defer op.deleteMu.Unlock()
	for i, d := range op.deleteQueue {
		if d.ts.Add(gracePeriod).After(now) {
			break
		}
		cutoff = i + 1
	}
	toDelete := op.deleteQueue[:cutoff]
	op.deleteQueue = op.deleteQueue[cutoff:]
	return toDelete
}

type ownerCacheEviction struct {
	ts    time.Time
	evict func()
}

// This function removes all data from the EndpointSlice except what is required by the client
func removeUnnecessaryEndpointSliceData(endpointSlice *discovery_v1.EndpointSlice) *discovery_v1.EndpointSlice {
	// name and namespace are needed by the informer store
	transformedEndpointSlice := discovery_v1.EndpointSlice{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      endpointSlice.GetName(),
			Namespace: endpointSlice.GetNamespace(),
		},
	}

	// we need a particular label to get the Service name
	serviceName := endpointSlice.GetLabels()[endpointSliceServiceLabel]
	transformedEndpointSlice.SetLabels(map[string]string{
		endpointSliceServiceLabel: serviceName,
	})

	// and for each endpoint, we need the targetRef
	transformedEndpointSlice.Endpoints = make([]discovery_v1.Endpoint, len(endpointSlice.Endpoints))
	for i, endpoint := range endpointSlice.Endpoints {
		if endpoint.TargetRef != nil && endpoint.TargetRef.Kind == "Pod" {
			transformedEndpointSlice.Endpoints[i].TargetRef = endpoint.TargetRef
		}
	}

	return &transformedEndpointSlice
}

// Get the Service name from an EndpointSlice based on a standard label.
// see: https://kubernetes.io/docs/concepts/services-networking/endpoint-slices/#ownership
func getServiceName(endpointSlice *discovery_v1.EndpointSlice) string {
	epLabels := endpointSlice.GetLabels()
	serviceName := epLabels[endpointSliceServiceLabel]
	return serviceName
}
