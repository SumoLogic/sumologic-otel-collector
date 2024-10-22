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
	"context"
	"fmt"
	"reflect"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	api_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig"
)

func newFakeAPIClientset(_ k8sconfig.APIConfig) (kubernetes.Interface, error) {
	fakeClient := fake.NewSimpleClientset()
	// Add batch/v1 CronJob resource so that setting up a CronJob informer works in testing.
	// This is required for the `client.Discovery().ServerGroupsAndResources()` function call
	// in `newOwnerProvider()` to work with the fake client.
	fakeClient.Fake.Resources = append(fakeClient.Fake.Resources, &metav1.APIResourceList{
		GroupVersion: "batch/v1",
		APIResources: []metav1.APIResource{
			{
				Name:         "cronjobs",
				SingularName: "cronjob",
				Namespaced:   true,
				Group:        "batch",
				Version:      "batch/v1",
				Kind:         "CronJob",
				ShortNames:   []string{"cj"},
			},
		},
	})
	return fakeClient, nil
}

func podAddAndUpdateTest(t *testing.T, c *WatchClient, handler func(obj interface{})) {
	assert.Equal(t, len(c.Pods), 0)

	// pod without IP
	pod := &api_v1.Pod{}
	handler(pod)
	assert.Equal(t, len(c.Pods), 0)

	pod = &api_v1.Pod{}
	pod.Name = "podA"
	pod.Status.PodIP = "1.1.1.1"
	handler(pod)
	assert.Equal(t, len(c.Pods), 1)
	got := c.Pods["1.1.1.1"]
	assert.Equal(t, got.Address, "1.1.1.1")
	assert.Equal(t, got.Name, "podA")
	assert.Equal(t, got.PodUID, "")

	pod = &api_v1.Pod{}
	pod.Name = "podB"
	pod.Status.PodIP = "1.1.1.1"
	handler(pod)
	assert.Equal(t, len(c.Pods), 1)
	got = c.Pods["1.1.1.1"]
	assert.Equal(t, got.Address, "1.1.1.1")
	assert.Equal(t, got.Name, "podB")
	assert.Equal(t, got.PodUID, "")

	pod = &api_v1.Pod{}
	pod.Name = "podC"
	pod.Status.PodIP = "2.2.2.2"
	pod.UID = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	handler(pod)
	assert.Equal(t, len(c.Pods), 3)
	got = c.Pods["2.2.2.2"]
	assert.Equal(t, got.Address, "2.2.2.2")
	assert.Equal(t, got.Name, "podC")
	assert.Equal(t, got.PodUID, "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	got = c.Pods["aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"]
	assert.Equal(t, got.Address, "2.2.2.2")
	assert.Equal(t, got.Name, "podC")
	assert.Equal(t, got.PodUID, "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
}

func TestDefaultClientset(t *testing.T) {
	c, err := New(
		zap.NewNop(),
		k8sconfig.APIConfig{},
		ExtractionRules{},
		Filters{},
		[]Association{},
		Excludes{},
		nil,
		nil,
		nil,
		"",
		10,
		30*time.Second,
		DefaultPodDeleteGracePeriod,
	)
	assert.Error(t, err)
	assert.Equal(t, "invalid authType for kubernetes: ", err.Error())
	assert.Nil(t, c)

	c, err = New(
		zap.NewNop(),
		k8sconfig.APIConfig{},
		ExtractionRules{},
		Filters{},
		[]Association{},
		Excludes{},
		newFakeAPIClientset,
		nil,
		nil,
		"",
		10,
		30*time.Second,
		DefaultPodDeleteGracePeriod,
	)
	assert.NoError(t, err)
	assert.NotNil(t, c)
}

func TestBadFilters(t *testing.T) {
	c, err := New(
		zap.NewNop(),
		k8sconfig.APIConfig{},
		ExtractionRules{},
		Filters{Fields: []FieldFilter{{Op: selection.Exists}}},
		[]Association{},
		Excludes{},
		newFakeAPIClientset,
		NewFakeInformer,
		newFakeOwnerProvider,
		"",
		10,
		30*time.Second,
		DefaultPodDeleteGracePeriod,
	)
	assert.Error(t, err)
	assert.Nil(t, c)
}

func TestClientStartStop(t *testing.T) {
	c, _ := newTestClient(t)
	ctr := c.informer.GetController()
	require.IsType(t, &FakeController{}, ctr)
	fctr := ctr.(*FakeController)
	require.NotNil(t, fctr)

	done := make(chan struct{})
	assert.False(t, fctr.HasStopped())
	go func() {
		c.Start()
		close(done)
	}()
	c.Stop()
	<-done
	assert.True(t, fctr.HasStopped())
}

func TestConstructorErrors(t *testing.T) {
	er := ExtractionRules{}
	ff := Filters{}
	t.Run("client-provider-call", func(t *testing.T) {
		var gotAPIConfig k8sconfig.APIConfig
		apiCfg := k8sconfig.APIConfig{
			AuthType: "test-auth-type",
		}
		clientProvider := func(c k8sconfig.APIConfig) (kubernetes.Interface, error) {
			gotAPIConfig = c
			return nil, fmt.Errorf("error creating k8s client")
		}
		c, err := New(
			zap.NewNop(),
			apiCfg,
			er,
			ff,
			[]Association{},
			Excludes{},
			clientProvider,
			NewFakeInformer,
			newFakeOwnerProvider,
			"",
			10,
			30*time.Second,
			DefaultPodDeleteGracePeriod,
		)
		assert.Nil(t, c)
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "error creating k8s client")
		assert.Equal(t, apiCfg, gotAPIConfig)
	})
}

func TestPodAdd(t *testing.T) {
	c, _ := newTestClient(t)
	podAddAndUpdateTest(t, c, c.handlePodAdd)
}

func TestPodHostNetwork(t *testing.T) {
	c, _ := newTestClient(t)
	assert.Equal(t, 0, len(c.Pods))

	pod := &api_v1.Pod{}
	pod.Name = "podA"
	pod.Status.PodIP = "1.1.1.1"
	pod.Spec.HostNetwork = true
	c.handlePodAdd(pod)
	assert.Equal(t, len(c.Pods), 1)
	got := c.Pods["1.1.1.1"]
	assert.Equal(t, got.Address, "1.1.1.1")
	assert.Equal(t, got.Name, "podA")
	assert.True(t, got.Ignore)
}

func TestPodAddOutOfSync(t *testing.T) {
	c, _ := newTestClient(t)
	assert.Equal(t, len(c.Pods), 0)

	pod := &api_v1.Pod{}
	pod.Name = "podA"
	pod.Namespace = "namespace"
	pod.Status.PodIP = "1.1.1.1"
	startTime := meta_v1.NewTime(time.Now())
	pod.Status.StartTime = &startTime
	c.handlePodAdd(pod)
	assert.Equal(t, len(c.Pods), 2)
	got := c.Pods["1.1.1.1"]
	assert.Equal(t, got.Address, "1.1.1.1")
	assert.Equal(t, got.Name, "podA")

	pod2 := &api_v1.Pod{}
	pod2.Name = "podB"
	pod2.Namespace = "namespace"
	pod2.Status.PodIP = "1.1.1.1"
	startTime2 := meta_v1.NewTime(time.Now().Add(-time.Second * 10))
	pod2.Status.StartTime = &startTime2
	c.handlePodAdd(pod2)
	assert.Equal(t, len(c.Pods), 3)
	got = c.Pods["1.1.1.1"]
	assert.Equal(t, got.Address, "1.1.1.1")
	assert.Equal(t, got.Name, "podA")
}

func TestPodUpdate(t *testing.T) {
	c, _ := newTestClient(t)
	podAddAndUpdateTest(t, c, func(obj interface{}) {
		// first argument (old pod) is not used right now
		c.handlePodUpdate(&api_v1.Pod{}, obj)
	})
}

func TestPodDelete(t *testing.T) {
	c, _ := newTestClient(t)
	podAddAndUpdateTest(t, c, c.handlePodAdd)
	assert.Equal(t, len(c.Pods), 3)
	assert.Equal(t, c.Pods["1.1.1.1"].Address, "1.1.1.1")

	// delete empty IP pod
	c.handlePodDelete(&api_v1.Pod{})

	// delete non-existent IP
	pod := &api_v1.Pod{}
	pod.Status.PodIP = "9.9.9.9"
	c.handlePodDelete(pod)
	assert.Equal(t, len(c.Pods), 3)
	got := c.Pods["1.1.1.1"]
	assert.Equal(t, got.Address, "1.1.1.1")
	assert.Equal(t, len(c.deleteQueue), 0)

	// delete matching IP with wrong name/different pod
	pod = &api_v1.Pod{}
	pod.Status.PodIP = "1.1.1.1"
	c.handlePodDelete(pod)
	got = c.Pods["1.1.1.1"]
	assert.Equal(t, len(c.Pods), 3)
	assert.Equal(t, got.Address, "1.1.1.1")
	assert.Equal(t, len(c.deleteQueue), 0)

	// delete matching IP and name
	pod = &api_v1.Pod{}
	pod.Name = "podB"
	pod.Status.PodIP = "1.1.1.1"
	tsBeforeDelete := time.Now()
	c.handlePodDelete(pod)
	assert.Equal(t, len(c.Pods), 3)
	assert.Equal(t, len(c.deleteQueue), 1)
	deleteRequest := c.deleteQueue[0]
	assert.Equal(t, deleteRequest.id, PodIdentifier("1.1.1.1"))
	assert.Equal(t, deleteRequest.podName, "podB")
	assert.False(t, deleteRequest.ts.Before(tsBeforeDelete))
	assert.False(t, deleteRequest.ts.After(time.Now()))

	pod = &api_v1.Pod{}
	pod.Name = "podC"
	pod.Status.PodIP = "2.2.2.2"
	pod.UID = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	tsBeforeDelete = time.Now()
	c.handlePodDelete(pod)
	assert.Equal(t, len(c.Pods), 3)
	assert.Equal(t, len(c.deleteQueue), 3)
	deleteRequest = c.deleteQueue[1]
	assert.Equal(t, deleteRequest.id, PodIdentifier("2.2.2.2"))
	assert.Equal(t, deleteRequest.podName, "podC")
	assert.False(t, deleteRequest.ts.Before(tsBeforeDelete))
	assert.False(t, deleteRequest.ts.After(time.Now()))
	deleteRequest = c.deleteQueue[2]
	assert.Equal(t, deleteRequest.id, PodIdentifier("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"))
	assert.Equal(t, deleteRequest.podName, "podC")
	assert.False(t, deleteRequest.ts.Before(tsBeforeDelete))
	assert.False(t, deleteRequest.ts.After(time.Now()))
}

func TestDeleteQueue(t *testing.T) {
	c, _ := newTestClient(t)
	podAddAndUpdateTest(t, c, c.handlePodAdd)
	assert.Equal(t, len(c.Pods), 3)
	assert.Equal(t, c.Pods["1.1.1.1"].Address, "1.1.1.1")

	// delete pod
	pod := &api_v1.Pod{}
	pod.Name = "podB"
	pod.Status.PodIP = "1.1.1.1"
	c.handlePodDelete(pod)
	assert.Equal(t, len(c.Pods), 3)
	assert.Equal(t, len(c.deleteQueue), 1)
}

func TestDeleteLoop(t *testing.T) {
	// go c.deleteLoop(time.Second * 1)
	c, _ := newTestClient(t)

	pod := &api_v1.Pod{}
	pod.Status.PodIP = "1.1.1.1"
	c.handlePodAdd(pod)
	assert.Equal(t, len(c.Pods), 1)
	assert.Equal(t, len(c.deleteQueue), 0)

	c.handlePodDelete(pod)
	assert.Equal(t, len(c.Pods), 1)
	assert.Equal(t, len(c.deleteQueue), 1)

	gracePeriod := time.Millisecond * 500
	go c.deleteLoop(time.Millisecond, gracePeriod)
	go func() {
		time.Sleep(time.Millisecond * 50)
		c.m.Lock()
		assert.Equal(t, len(c.Pods), 1)
		c.m.Unlock()
		c.deleteMut.Lock()
		assert.Equal(t, len(c.deleteQueue), 1)
		c.deleteMut.Unlock()

		time.Sleep(gracePeriod + (time.Millisecond * 50))
		c.m.Lock()
		assert.Equal(t, len(c.Pods), 0)
		c.m.Unlock()
		c.deleteMut.Lock()
		assert.Equal(t, len(c.deleteQueue), 0)
		c.deleteMut.Unlock()
		close(c.stopCh)
	}()
	<-c.stopCh
}

func TestGetIgnoredPod(t *testing.T) {
	c, _ := newTestClient(t)
	pod := &api_v1.Pod{}
	pod.Status.PodIP = "1.1.1.1"
	c.handlePodAdd(pod)
	c.Pods[PodIdentifier(pod.Status.PodIP)].Ignore = true
	got, ok := c.getPod(PodIdentifier(pod.Status.PodIP))
	assert.Nil(t, got)
	assert.False(t, ok)
}

func TestGetPod(t *testing.T) {
	c, _ := newTestClient(t)

	pod := &api_v1.Pod{}
	pod.Status.PodIP = "1.1.1.1"
	pod.UID = "1234"
	pod.Name = "pod_name"
	pod.Namespace = "namespace_name"
	pod.OwnerReferences = []meta_v1.OwnerReference{
		metav1.OwnerReference{
			Name: "A reference",
		},
		metav1.OwnerReference{
			Name: "Another reference",
		},
	}
	c.handlePodAdd(pod)

	expected := &Pod{
		Name:            "pod_name",
		Namespace:       "namespace_name",
		Address:         "1.1.1.1",
		PodUID:          "1234",
		Attributes:      map[string]string{},
		OwnerReferences: &pod.OwnerReferences,
	}

	got, ok := c.getPod(PodIdentifier("1.1.1.1"))
	assert.Equal(t, got, expected)
	assert.True(t, ok)

	got, ok = c.getPod(PodIdentifier("1234"))
	assert.Equal(t, got, expected)
	assert.True(t, ok)

	got, ok = c.getPod(PodIdentifier("pod_name.namespace_name"))
	assert.Equal(t, got, expected)
	assert.True(t, ok)
}

func TestGetPodConcurrent(t *testing.T) {
	c, _ := newTestClient(t)

	pod := &api_v1.Pod{}
	pod.Status.PodIP = "1.1.1.1"
	pod.UID = "1234"
	pod.Name = "pod_name"
	pod.Namespace = "namespace_name"
	pod.OwnerReferences = []meta_v1.OwnerReference{
		{
			Kind: "StatefulSet",
			Name: "snug-sts",
			UID:  "f15f0585-a0bc-43a3-96e4-dd2eace75391",
		},
	}
	c.handlePodAdd(pod)

	expected := &Pod{
		Name:            "pod_name",
		Namespace:       "namespace_name",
		Address:         "1.1.1.1",
		PodUID:          "1234",
		Attributes:      map[string]string{},
		OwnerReferences: &pod.OwnerReferences,
	}

	numThreads := 2
	wg := sync.WaitGroup{}
	for i := 0; i < numThreads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, ok := c.getPod(PodIdentifier("1.1.1.1"))
			assert.Equal(t, got, expected)
			assert.True(t, ok)
		}()
	}
	wg.Wait()
}

func TestGetPodOwnerAttributesConcurrent(t *testing.T) {
	rules := ExtractionRules{
		OwnerLookupEnabled: true,
		StatefulSetName:    true,
		Tags:               NewExtractionFieldTags(),
	}
	c, _ := newTestClientWithRulesAndFilters(t, rules, Filters{})

	pod := &api_v1.Pod{}
	pod.Status.PodIP = "1.1.1.1"
	pod.UID = "1234"
	pod.Name = "pod_name"
	pod.Namespace = "namespace_name"
	pod.OwnerReferences = []meta_v1.OwnerReference{
		{
			Kind: "StatefulSet",
			Name: "snug-sts",
			UID:  "f15f0585-a0bc-43a3-96e4-dd2eace75391",
		},
	}
	c.handlePodAdd(pod)

	expected := map[string]string{
		rules.Tags.StatefulSetName: "snug-sts",
	}

	numThreads := 2
	wg := sync.WaitGroup{}
	for i := 0; i < numThreads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, ok := c.GetPodAttributes(PodIdentifier("1.1.1.1"))
			assert.Equal(t, got, expected)
			assert.True(t, ok)
		}()
	}
	wg.Wait()
}

func TestGetPodWhenNamespaceInExtractedMetadata(t *testing.T) {
	c, _ := newTestClient(t)

	c.Rules.Namespace = true
	c.Rules.Tags.Namespace = "namespace"

	pod := &api_v1.Pod{}
	pod.Status.PodIP = "1.1.1.1"
	pod.UID = "1234"
	pod.Name = "pod_name"
	pod.Namespace = "namespace_name"
	c.handlePodAdd(pod)

	expected := &Pod{
		Name:      "pod_name",
		Namespace: "namespace_name",
		Address:   "1.1.1.1",
		PodUID:    "1234",
		Attributes: map[string]string{
			"namespace": "namespace_name",
		},
		OwnerReferences: &pod.OwnerReferences,
	}

	got, ok := c.getPod(PodIdentifier("1.1.1.1"))
	assert.Equal(t, got, expected)
	assert.True(t, ok)

	got, ok = c.getPod(PodIdentifier("1234"))
	assert.Equal(t, got, expected)
	assert.True(t, ok)

	got, ok = c.getPod(PodIdentifier("pod_name.namespace_name"))
	assert.Equal(t, got, expected)
	assert.True(t, ok)
}

func TestHandlerWrongType(t *testing.T) {
	c, logs := newTestClientWithRulesAndFilters(t, ExtractionRules{}, Filters{})
	assert.Equal(t, logs.Len(), 0)
	c.handlePodAdd(1)
	c.handlePodDelete(1)
	c.handlePodUpdate(1, 2)
	assert.Equal(t, logs.Len(), 3)
	for _, l := range logs.All() {
		assert.Equal(t, l.Message, "object received was not of type api_v1.Pod")
	}
}

func TestNoHostnameExtractionRules(t *testing.T) {
	c, _ := newTestClientWithRulesAndFilters(t, ExtractionRules{}, Filters{})

	podName := "auth-service-abc12-xyz3"

	pod := &api_v1.Pod{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: podName,
		},
		Spec: api_v1.PodSpec{},
		Status: api_v1.PodStatus{
			PodIP: "1.1.1.1",
		},
	}

	c.Rules = ExtractionRules{
		HostName: true,
		Tags:     NewExtractionFieldTags(),
	}

	c.handlePodAdd(pod)
	p, _ := c.getPod(PodIdentifier(pod.Status.PodIP))
	assert.Equal(t, p.Attributes["k8s.pod.hostname"], podName)
}

func TestExtractionRules(t *testing.T) {
	// OwnerLookupEnabled is set to true so the newOwnerProviderFunc can be called in the initializer
	c, _ := newTestClientWithRulesAndFilters(t, ExtractionRules{OwnerLookupEnabled: true}, Filters{})

	pod := &api_v1.Pod{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:              "auth-service-abc12-xyz3",
			Namespace:         "ns1",
			UID:               "33333",
			CreationTimestamp: meta_v1.Now(),
			Labels: map[string]string{
				"label1": "lv1",
				"label2": "k1=v1 k5=v5 extra!",
			},
			Annotations: map[string]string{
				"annotation1": "av1",
			},
		},
		Spec: api_v1.PodSpec{
			NodeName: "node1",
			Hostname: "auth-hostname3",
			Containers: []api_v1.Container{
				{
					Image: "auth-service-image",
					Name:  "auth-service-container-name",
				},
			},
		},
		Status: api_v1.PodStatus{
			PodIP: "1.1.1.1",
			ContainerStatuses: []api_v1.ContainerStatus{
				{
					ContainerID: "111-222-333",
				},
			},
		},
	}

	testCases := []struct {
		name       string
		podOwner   *meta_v1.OwnerReference
		rules      ExtractionRules
		attributes map[string]string
	}{
		{
			name:       "no-rules",
			rules:      ExtractionRules{},
			attributes: nil,
		},
		{
			name: "deployment name without owner lookup",
			rules: ExtractionRules{
				DeploymentName: true,
				Tags:           NewExtractionFieldTags(),
			},
			attributes: map[string]string{},
		},
		{
			name: "deployment name with owner lookup",
			podOwner: &meta_v1.OwnerReference{
				Kind: "ReplicaSet",
				Name: "dearest-deploy-77c99ccb96",
				UID:  "1a1658f9-7818-11e9-90f1-02324f7e0d1e",
			},
			rules: ExtractionRules{
				DeploymentName:     true,
				OwnerLookupEnabled: true,
				Tags:               NewExtractionFieldTags(),
			},
			attributes: map[string]string{
				"k8s.deployment.name": "dearest-deploy",
			},
		},
		{
			name: "statefulset name",
			podOwner: &meta_v1.OwnerReference{
				Kind: "StatefulSet",
				Name: "snug-sts",
				UID:  "f15f0585-a0bc-43a3-96e4-dd2eace75391",
			},
			rules: ExtractionRules{
				StatefulSetName:    true,
				DeploymentName:     true,
				OwnerLookupEnabled: true,
				Tags:               NewExtractionFieldTags(),
			},
			attributes: map[string]string{
				"k8s.statefulset.name": "snug-sts",
			},
		},
		{
			name: "job name and cron job are not added by default",
			podOwner: &meta_v1.OwnerReference{
				Kind: "Job",
				Name: "hello-job",
				UID:  "f15f0585-a0bc-43a3-96e4-dd2ea9975391",
			},
			rules: ExtractionRules{
				OwnerLookupEnabled: true,
				Tags:               NewExtractionFieldTags(),
			},
			attributes: map[string]string{},
		},
		{
			name: "job name is added properly",
			podOwner: &meta_v1.OwnerReference{
				Kind: "Job",
				Name: "hello-job",
				UID:  "f15f0585-a0bc-43a3-96e4-dd2ea9975391",
			},
			rules: ExtractionRules{
				JobName:            true,
				OwnerLookupEnabled: true,
				Tags:               NewExtractionFieldTags(),
			},
			attributes: map[string]string{
				"k8s.job.name": "hello-job",
			},
		},
		{
			name: "job name and cron job name are added properly",
			podOwner: &meta_v1.OwnerReference{
				Kind: "Job",
				Name: "hello-job",
				UID:  "f15f0585-a0bc-43a3-96e4-dd2ea9975391",
			},
			rules: ExtractionRules{
				JobName:            true,
				CronJobName:        true,
				OwnerLookupEnabled: true,
				Tags:               NewExtractionFieldTags(),
			},
			attributes: map[string]string{
				"k8s.job.name":     "hello-job",
				"k8s.cronjob.name": "hello-cronjob",
			},
		},
		{
			name: "metadata",
			podOwner: &meta_v1.OwnerReference{
				Kind: "ReplicaSet",
				Name: "foo-bar-rs",
				UID:  "1a1658f9-7818-11e9-90f1-02324f7e0d1e",
			},
			rules: ExtractionRules{
				ContainerID:        true,
				ContainerImage:     true,
				ContainerName:      true,
				DaemonSetName:      true,
				DeploymentName:     true,
				HostName:           true,
				PodUID:             true,
				PodName:            true,
				ReplicaSetName:     true,
				ServiceName:        true,
				StatefulSetName:    true,
				StartTime:          true,
				Namespace:          true,
				NodeName:           true,
				OwnerLookupEnabled: true,
				Tags:               NewExtractionFieldTags(),
			},
			attributes: map[string]string{
				"k8s.container.id":    "111-222-333",
				"k8s.container.image": "auth-service-image",
				"k8s.container.name":  "auth-service-container-name",
				"k8s.deployment.name": "dearest-deploy",
				"k8s.pod.hostname":    "auth-hostname3",
				"k8s.pod.uid":         "33333",
				"k8s.pod.name":        "auth-service-abc12-xyz3",
				"k8s.pod.startTime":   pod.GetCreationTimestamp().String(),
				"k8s.replicaset.name": "dearest-deploy-77c99ccb96",
				"k8s.service.name":    "foo_bar",
				"k8s.namespace.name":  "ns1",
				"k8s.node.name":       "node1",
			},
		},
		{
			name: "non-default tags",
			rules: ExtractionRules{
				ContainerID:     true,
				ContainerImage:  false,
				ContainerName:   true,
				DaemonSetName:   false,
				DeploymentName:  false,
				HostName:        false,
				PodUID:          false,
				PodName:         false,
				ReplicaSetName:  false,
				ServiceName:     false,
				StatefulSetName: false,
				StartTime:       false,
				Namespace:       false,
				NodeName:        false,
				Tags: ExtractionFieldTags{
					ContainerID:   "cid",
					ContainerName: "cn",
				},
			},
			attributes: map[string]string{
				"cid": "111-222-333",
				"cn":  "auth-service-container-name",
			},
		},
		{
			name: "labels",
			rules: ExtractionRules{
				Annotations: []FieldExtractionRule{
					{
						Name: "a1",
						Key:  "annotation1",
					},
				},
				Labels: []FieldExtractionRule{
					{
						Name: "l1",
						Key:  "label1",
					}, {
						Name:  "l2",
						Key:   "label2",
						Regex: regexp.MustCompile(`k5=(?P<value>[^\s]+)`),
					},
				},
			},
			attributes: map[string]string{
				"l1": "lv1",
				"l2": "v5",
				"a1": "av1",
			},
		},
		{
			name: "generic-labels",
			rules: ExtractionRules{
				OwnerLookupEnabled: true,
				Tags:               NewExtractionFieldTags(),
				Annotations: []FieldExtractionRule{
					{
						Name: "k8s.pod.annotation.%s",
						Key:  "*",
					},
				},
				Labels: []FieldExtractionRule{
					{
						Name: "k8s.pod.label.%s",
						Key:  "*",
					},
				},
				NamespaceAnnotations: []FieldExtractionRule{
					{
						Name: "namespace_annotations_%s",
						Key:  "*",
					},
				},
				NamespaceLabels: []FieldExtractionRule{
					{
						Name: "namespace_labels_%s",
						Key:  "*",
					},
				},
			},
			attributes: map[string]string{
				"k8s.pod.label.label1":             "lv1",
				"k8s.pod.label.label2":             "k1=v1 k5=v5 extra!",
				"k8s.pod.annotation.annotation1":   "av1",
				"namespace_labels_label":           "namespace_label_value",
				"namespace_annotations_annotation": "namespace_annotation_value",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.podOwner != nil {
				pod.OwnerReferences = []meta_v1.OwnerReference{
					*tc.podOwner,
				}
			}
			c.Rules = tc.rules

			// manually call the data removal function here
			// normally the informer does this, but fully emulating the informer in this test is annoying
			transformedPod := removeUnnecessaryPodData(pod, c.Rules)
			c.handlePodAdd(transformedPod)
			attributes, ok := c.GetPodAttributes(PodIdentifier(pod.Status.PodIP))
			require.True(t, ok)

			assert.Equal(t, len(tc.attributes), len(attributes))
			for k, v := range tc.attributes {
				got, ok := attributes[k]
				if assert.True(t, ok, "Attribute '%s' not found.", k) {
					assert.Equal(t, v, got, "Value of '%s' is incorrect", k)
				}
			}
		})
	}
}

func TestFilters(t *testing.T) {
	testCases := []struct {
		name    string
		filters Filters
		labels  string
		fields  string
	}{
		{
			name:    "no-filters",
			filters: Filters{},
		}, {
			name: "namespace",
			filters: Filters{
				Namespace: "default",
			},
		}, {
			name: "node",
			filters: Filters{
				Node: "ec2-test",
			},
			fields: "spec.nodeName=ec2-test",
		}, {
			name: "labels-and-fields",
			filters: Filters{
				Labels: []FieldFilter{
					{
						Key:   "k1",
						Value: "v1",
						Op:    selection.Equals,
					},
					{
						Key:   "k2",
						Value: "v2",
						Op:    selection.NotEquals,
					},
				},
				Fields: []FieldFilter{
					{
						Key:   "k1",
						Value: "v1",
						Op:    selection.Equals,
					},
					{
						Key:   "k2",
						Value: "v2",
						Op:    selection.NotEquals,
					},
				},
			},
			labels: "k1=v1,k2!=v2",
			fields: "k1=v1,k2!=v2",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, _ := newTestClientWithRulesAndFilters(t, ExtractionRules{}, tc.filters)
			inf := c.informer.(*FakeInformer)
			assert.Equal(t, tc.filters.Namespace, inf.namespace)
			assert.Equal(t, tc.labels, inf.labelSelector.String())
			assert.Equal(t, tc.fields, inf.fieldSelector.String())
		})
	}
}

func TestNodeFilterDoesntApplyToOwners(t *testing.T) {
	filters := Filters{
		Node: "ec2-test",
	}
	c, _ := newTestClientWithRulesAndFilters(t, ExtractionRules{OwnerLookupEnabled: true}, filters)

	// verify that the Pod informer has the Node selector set
	inf := c.informer.(*FakeInformer)
	assert.Equal(t, "", inf.labelSelector.String())
	assert.Equal(t, "spec.nodeName=ec2-test", inf.fieldSelector.String())

	// verify that the owner provider does NOT have the Node selector set
	ownerProvider := c.op.(*fakeOwnerCache)
	assert.Equal(t, "", ownerProvider.fieldSelector.String())
}

func TestPodIgnorePatterns(t *testing.T) {
	testCases := []struct {
		ignore bool
		pod    api_v1.Pod
	}{
		{
			ignore: false,
			pod:    api_v1.Pod{},
		}, {
			ignore: true,
			pod: api_v1.Pod{
				Spec: api_v1.PodSpec{
					HostNetwork: true,
				},
			},
		}, {
			ignore: true,
			pod: api_v1.Pod{
				ObjectMeta: meta_v1.ObjectMeta{
					Annotations: map[string]string{
						"opentelemetry.io/k8s-processor/ignore": "True ",
					},
				},
			},
		}, {
			ignore: true,
			pod: api_v1.Pod{
				ObjectMeta: meta_v1.ObjectMeta{
					Annotations: map[string]string{
						"opentelemetry.io/k8s-processor/ignore": "true",
					},
				},
			},
		}, {
			ignore: false,
			pod: api_v1.Pod{
				ObjectMeta: meta_v1.ObjectMeta{
					Annotations: map[string]string{
						"opentelemetry.io/k8s-processor/ignore": "false",
					},
				},
			},
		}, {
			ignore: false,
			pod: api_v1.Pod{
				ObjectMeta: meta_v1.ObjectMeta{
					Annotations: map[string]string{
						"opentelemetry.io/k8s-processor/ignore": "",
					},
				},
			},
		}, {
			ignore: true,
			pod: api_v1.Pod{
				ObjectMeta: meta_v1.ObjectMeta{
					Name: "jaeger-agent",
				},
			},
		}, {
			ignore: true,
			pod: api_v1.Pod{
				ObjectMeta: meta_v1.ObjectMeta{
					Name: "jaeger-collector",
				},
			},
		}, {
			ignore: true,
			pod: api_v1.Pod{
				ObjectMeta: meta_v1.ObjectMeta{
					Name: "jaeger-agent-b2zdv",
				},
			},
		}, {
			ignore: false,
			pod: api_v1.Pod{
				ObjectMeta: meta_v1.ObjectMeta{
					Name: "test-pod-name",
				},
			},
		},
	}

	c, _ := newTestClient(t)
	for _, tc := range testCases {
		assert.Equal(t, tc.ignore, c.shouldIgnorePod(&tc.pod),
			"Should ignore %v, pod.Name: %q, pod annotations %#v", tc.ignore, tc.pod.Name, tc.pod.Annotations,
		)
	}
}

func Test_extractField(t *testing.T) {
	c := WatchClient{}
	type args struct {
		v string
		r FieldExtractionRule
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"no-regex",
			args{
				"str",
				FieldExtractionRule{Regex: nil},
			},
			"str",
		},
		{
			"basic",
			args{
				"str",
				FieldExtractionRule{Regex: regexp.MustCompile("field=(?P<value>.+)")},
			},
			"",
		},
		{
			"basic",
			args{
				"field=val1",
				FieldExtractionRule{Regex: regexp.MustCompile("field=(?P<value>.+)")},
			},
			"val1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := c.extractField(tt.args.v, tt.args.r); got != tt.want {
				t.Errorf("extractField() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_selectorsFromFilters(t *testing.T) {
	tests := []struct {
		name    string
		filters Filters
		wantL   labels.Selector
		wantF   fields.Selector
		wantErr bool
	}{
		{
			"label/invalid-op",
			Filters{
				Labels: []FieldFilter{{Op: "invalid-op"}},
			},
			nil,
			nil,
			true,
		},
		{
			"fields/invalid-op",
			Filters{
				Fields: []FieldFilter{{Op: selection.Exists}},
			},
			nil,
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := selectorsFromFilters(tt.filters)
			if (err != nil) != tt.wantErr {
				t.Errorf("selectorsFromFilters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantL) {
				t.Errorf("selectorsFromFilters() got = %v, want %v", got, tt.wantL)
			}
			if !reflect.DeepEqual(got1, tt.wantF) {
				t.Errorf("selectorsFromFilters() got1 = %v, want %v", got1, tt.wantF)
			}
		})
	}
}

func Test_PodsGetAddedAndDeletedFromCache(t *testing.T) {
	const (
		namespace = "kube-system"
	)

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	client, err := New(
		logger,
		k8sconfig.APIConfig{},
		ExtractionRules{
			ContainerID:        true,
			ContainerImage:     true,
			ContainerName:      true,
			PodUID:             true,
			PodName:            true,
			StartTime:          true,
			Namespace:          true,
			NodeName:           false,
			OwnerLookupEnabled: true,
			Tags:               NewExtractionFieldTags(),
		},
		Filters{},
		[]Association{},
		Excludes{},
		newFakeAPIClientset,
		newSharedInformer,
		newOwnerProvider,
		"_",
		10,
		10*time.Millisecond,
		10*time.Millisecond,
	)
	require.NoError(t, err)

	c := client.(*WatchClient)

	ch := waitForWatchToBeEstablished(c.kc.(*fake.Clientset), "pods")
	go func() {
		c.Start()
	}()
	defer c.Stop()
	<-ch

	eventuallyNPodsInCache := func(t *testing.T, n int) {
		assert.Eventually(t, func() bool {
			c.m.RLock()
			l := len(c.Pods)
			c.m.RUnlock()
			return l == n
		}, 5*time.Second, 10*time.Millisecond)
	}

	t.Run("pod with IP", func(t *testing.T) {
		pod := &api_v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-pod",
				Namespace: namespace,
				UID:       "f15f0585-a0bc-43a3-96e4-dd2eace75392",
			},
			Status: api_v1.PodStatus{
				PodIP: "10.0.0.1",
			},
		}

		_, err = c.kc.CoreV1().Pods(namespace).
			Create(context.Background(), pod, metav1.CreateOptions{})
		require.NoError(t, err)
		eventuallyNPodsInCache(t, 3)

		err = c.kc.CoreV1().Pods(namespace).
			Delete(context.Background(), pod.Name, metav1.DeleteOptions{})
		require.NoError(t, err)
		eventuallyNPodsInCache(t, 0)
	})

	t.Run("pod without an IP", func(t *testing.T) {
		pod := &api_v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-pod",
				Namespace: namespace,
				UID:       "f15f0585-a0bc-43a3-96e4-dd2eace75392",
			},
		}

		_, err = c.kc.CoreV1().Pods(namespace).
			Create(context.Background(), pod, metav1.CreateOptions{})
		require.NoError(t, err)
		eventuallyNPodsInCache(t, 2)

		err = c.kc.CoreV1().Pods(namespace).
			Delete(context.Background(), pod.Name, metav1.DeleteOptions{})
		require.NoError(t, err)
		eventuallyNPodsInCache(t, 0)
	})

	t.Run("with deleted final state unknown", func(t *testing.T) {
		pod := &api_v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-pod",
				Namespace: namespace,
				UID:       "f15f0585-a0bc-43a3-96e4-dd2eace75392",
			},
		}

		_, err = c.kc.CoreV1().Pods(namespace).
			Create(context.Background(), pod, metav1.CreateOptions{})
		require.NoError(t, err)
		eventuallyNPodsInCache(t, 2)

		// Rather than set up a stub Informer just for this case, bypass the
		// informer + fake k8s client entirely. Manually call the delete
		// handler with DeletedFinalStateUnknown.
		c.handlePodDelete(cache.DeletedFinalStateUnknown{
			Key: fmt.Sprintf("%s/my-pod", namespace),
			Obj: pod,
		})
		defer func() {
			err = c.kc.CoreV1().Pods(namespace).
				Delete(context.Background(), pod.Name, metav1.DeleteOptions{})
			require.NoError(t, err)
		}()

		eventuallyNPodsInCache(t, 0)
	})
}

func newTestClientWithRulesAndFilters(t *testing.T, e ExtractionRules, f Filters) (*WatchClient, *observer.ObservedLogs) {
	observedLogger, logs := observer.New(zapcore.WarnLevel)
	logger := zap.New(observedLogger)
	exclude := Excludes{
		Pods: []ExcludePods{
			{Name: regexp.MustCompile(`jaeger-agent`)},
			{Name: regexp.MustCompile(`jaeger-collector`)},
		},
	}
	c, err := New(
		logger,
		k8sconfig.APIConfig{},
		e,
		f,
		[]Association{},
		exclude,
		newFakeAPIClientset,
		NewFakeInformer,
		newFakeOwnerProvider,
		"_",
		10,
		30*time.Second,
		DefaultPodDeleteGracePeriod,
	)
	require.NoError(t, err)
	return c.(*WatchClient), logs
}

func newTestClient(t *testing.T) (*WatchClient, *observer.ObservedLogs) {
	return newTestClientWithRulesAndFilters(t, ExtractionRules{}, Filters{})
}

func TestServiceInfoArrivesLate(t *testing.T) {
	// Concept: we insert a pod with no service associated,
	// then update the service in the cache,
	// then try fetching the pod and see that it doesn't contain the service
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	podUID := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	cache := OwnerCache{
		objectOwners: map[string]*ObjectOwner{},
		podServices:  map[string][]string{},
		namespaces:   map[string]*api_v1.Namespace{},
		logger:       logger,
		stopCh:       make(chan struct{}),
	}
	cache.podServices["pod"] = []string{"firstService", "secondService"}

	var client = &WatchClient{
		logger:    logger,
		op:        &cache,
		delimiter: ", ",
		Rules: ExtractionRules{
			OwnerLookupEnabled: true,
			ServiceName:        true,
			Tags: ExtractionFieldTags{
				ServiceName: "ServiceName",
			},
		},
		Pods: map[PodIdentifier]*Pod{},
	}

	pod := &api_v1.Pod{}
	pod.Name = "pod"
	pod.Status.PodIP = "2.2.2.2"
	pod.UID = types.UID(podUID)

	client.handlePodAdd(pod)

	attributes, ok := client.GetPodAttributes(PodIdentifier(podUID))
	assert.True(t, ok)

	logger.Debug("pod attributes: ", zap.Any("attributes", attributes))
	serviceName, ok := attributes["ServiceName"]
	assert.True(t, ok)

	// After PodAdd, there are two services:
	assert.Equal(t, "firstService, secondService", serviceName)

	cache.podServices["pod"] = []string{"firstService", "secondService", "thirdService"}

	attributes, ok = client.GetPodAttributes(PodIdentifier(podUID))
	assert.True(t, ok)

	logger.Debug("pod attributes: ", zap.Any("attributes", attributes))
	serviceName, ok = attributes["ServiceName"]
	assert.True(t, ok)

	// Desired behavior: we get all three service names in response:
	assert.Equal(t, "firstService, secondService, thirdService", serviceName)
}
