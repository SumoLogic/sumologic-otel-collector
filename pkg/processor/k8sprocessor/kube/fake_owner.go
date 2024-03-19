// Copyright 2019 Omnition Authors
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
	"time"

	"go.uber.org/zap"
	api_v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// fakeOwnerCache is a simple structure which aids querying for owners
type fakeOwnerCache struct {
	logger          *zap.Logger
	objectOwners    map[string]*ObjectOwner
	labelSelector   labels.Selector
	fieldSelector   fields.Selector
	extractionRules ExtractionRules
	namespace       string
}

// NewOwnerProvider creates new instance of the owners api
func newFakeOwnerProvider(logger *zap.Logger,
	client kubernetes.Interface,
	labelSelector labels.Selector,
	fieldSelector fields.Selector,
	extractionRules ExtractionRules,
	namespace string,
	_ time.Duration, _ time.Duration,
) (OwnerAPI, error) {
	ownerCache := fakeOwnerCache{
		labelSelector:   labelSelector,
		fieldSelector:   fieldSelector,
		extractionRules: extractionRules,
		namespace:       namespace,
	}
	ownerCache.objectOwners = map[string]*ObjectOwner{}
	ownerCache.logger = logger

	replicaSet := ObjectOwner{
		UID:       "1a1658f9-7818-11e9-90f1-02324f7e0d1e",
		namespace: "kube-system",
		ownerUIDs: []types.UID{types.UID("94682908-e546-42cc-9972-62bcd09bd9de")},
		kind:      "ReplicaSet",
		name:      "dearest-deploy-77c99ccb96",
	}
	ownerCache.objectOwners[string(replicaSet.UID)] = &replicaSet

	deployment := ObjectOwner{
		UID:       "94682908-e546-42cc-9972-62bcd09bd9de",
		namespace: "kube-system",
		ownerUIDs: []types.UID{},
		kind:      "Deployment",
		name:      "dearest-deploy",
	}
	ownerCache.objectOwners[string(deployment.UID)] = &deployment

	statefulSet := ObjectOwner{
		UID:       "f15f0585-a0bc-43a3-96e4-dd2eace75391",
		namespace: "kube-system",
		ownerUIDs: []types.UID{},
		kind:      "StatefulSet",
		name:      "snug-sts",
	}
	ownerCache.objectOwners[string(statefulSet.UID)] = &statefulSet

	job := ObjectOwner{
		UID:       "f15f0585-a0bc-43a3-96e4-dd2ea9975391",
		namespace: "default",
		ownerUIDs: []types.UID{"f01f0585-a0bc-43a3-9611-dd2ea9975391"},
		kind:      "Job",
		name:      "hello-job",
	}
	ownerCache.objectOwners[string(job.UID)] = &job

	cronjob := ObjectOwner{
		UID:       "f01f0585-a0bc-43a3-9611-dd2ea9975391",
		namespace: "default",
		ownerUIDs: []types.UID{},
		kind:      "CronJob",
		name:      "hello-cronjob",
	}
	ownerCache.objectOwners[string(cronjob.UID)] = &cronjob

	return &ownerCache, nil
}

// Start
func (op *fakeOwnerCache) Start() {}

// Stop
func (op *fakeOwnerCache) Stop() {}

// GetServices fetches list of services for a given pod
func (op *fakeOwnerCache) GetServices(podName string) []string {
	return []string{"foo", "bar"}
}

// GetNamespace returns a namespace
func (op *fakeOwnerCache) GetNamespace(pod *api_v1.Pod) *api_v1.Namespace {
	namespace := api_v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        pod.Namespace,
			Labels:      map[string]string{"label": "namespace_label_value"},
			Annotations: map[string]string{"annotation": "namespace_annotation_value"},
		},
	}
	return &namespace
}

// GetOwners fetches deep tree of owners for a given pod
func (op *fakeOwnerCache) GetOwners(pod *Pod) []*ObjectOwner {
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
	}

	return objectOwners
}
