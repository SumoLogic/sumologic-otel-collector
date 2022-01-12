package kube

import (
	"context"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	v1 "k8s.io/api/apps/v1"
	api_v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
)

func Test_OwnerProvider_GetOwners(t *testing.T) {
	c, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	require.NoError(t, err)

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	op, err := newOwnerProvider(logger, c, labels.Everything(), fields.Everything(), "kube-system")
	require.NoError(t, err)

	op.Start()
	t.Cleanup(func() {
		op.Stop()
	})

	sts, err := c.AppsV1().StatefulSets("kube-system").
		Create(context.Background(),
			&v1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-sts",
					Namespace: "kube-system",
					UID:       "f15f0585-a0bc-43a3-96e4-dd2eace75391",
				},
				TypeMeta: metav1.TypeMeta{
					Kind: "StatefulSet",
				},
			},
			metav1.CreateOptions{},
		)
	require.NoError(t, err)

	pod, err := c.CoreV1().Pods("kube-system").
		Create(context.Background(),
			&api_v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-pod",
					Namespace: "kube-system",
					UID:       "f15f0585-a0bc-43a3-96e4-dd2eace75392",
					OwnerReferences: []metav1.OwnerReference{
						{
							Kind: sts.Kind,
							Name: sts.Name,
							UID:  sts.UID,
						},
					},
				},
			},
			metav1.CreateOptions{},
		)
	require.NoError(t, err)

	assert.Eventually(t, func() bool {
		owners := op.GetOwners(pod)
		if len(owners) != 1 {
			t.Logf("owners: %v", owners)
			return false
		}

		if uid := owners[0].UID; uid != "f15f0585-a0bc-43a3-96e4-dd2eace75391" {
			t.Logf("wrong owner UID: %v", uid)
			return false
		}

		return true
	}, 5*time.Second, 50*time.Millisecond)
}

func Test_OwnerProvider_GetServices(t *testing.T) {
	c, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	require.NoError(t, err)

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	op, err := newOwnerProvider(logger, c, labels.Everything(), fields.Everything(), "kube-system")
	require.NoError(t, err)

	op.Start()
	t.Cleanup(func() {
		op.Stop()
	})

	var (
		pod = &api_v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-pod",
				Namespace: "kube-system",
				UID:       "f15f0585-a0bc-43a3-96e4-dd2eace75392",
			},
		}
		endpoints1 = &api_v1.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-service",
				Namespace: pod.Namespace,
				UID:       "f15f0585-a0bc-43a3-96e4-dd2eace75390",
			},
			TypeMeta: metav1.TypeMeta{
				Kind: "Endpoint",
			},
			Subsets: []api_v1.EndpointSubset{
				{
					Addresses: []api_v1.EndpointAddress{
						{
							TargetRef: &api_v1.ObjectReference{
								Name:      pod.Name,
								Namespace: pod.Namespace,
								Kind:      "Pod",
								UID:       pod.UID,
							},
						},
					},
				},
			},
		}
		endpoints2 = &api_v1.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-service-2",
				Namespace: pod.Namespace,
				UID:       "f15f0585-a0bc-43a3-96e4-dd2eace75391",
			},
			TypeMeta: metav1.TypeMeta{
				Kind: "Endpoint",
			},
			Subsets: []api_v1.EndpointSubset{
				{
					Addresses: []api_v1.EndpointAddress{
						{
							TargetRef: &api_v1.ObjectReference{
								Name:      pod.Name,
								Namespace: pod.Namespace,
								Kind:      "Pod",
								UID:       pod.UID,
							},
						},
					},
				},
			},
		}
	)

	// Below time.Sleep seems to "solve" the problem of this test being flaky, namely
	// when we don't receive the endpoint creation "notification(s)" - sometimes one
	// sometimes both for endpoint1 and endpoint2 - which triggers ownerCache.cacheEndpoint.
	//
	// After adding some debug logging one can observe the following i.e. ownerCache receiving
	// only one notification about creation of one endpoint.
	//
	// === RUN   Test_OwnerProvider_GetOwners
	// 2022-01-11T20:23:49.052+0100    INFO    kube/owner.go:87        Staring K8S resource informers  {"#infomers": 7}
	// --- PASS: Test_OwnerProvider_GetOwners (0.05s)
	// === RUN   Test_OwnerProvider_GetServices
	// 2022-01-11T20:23:49.104+0100    INFO    kube/owner.go:87        Staring K8S resource informers  {"#infomers": 7}
	// === RUN   Test_OwnerProvider_GetServices/adding_endpoints
	// 2022-01-11T20:23:49.106+0100    DEBUG   kube/owner.go:335       cacheEndpoint   {"endpoint": "my-service"}
	// 2022-01-11T20:23:49.107+0100    DEBUG   kube/owner.go:305       genericEndpointOp
	// 2022-01-11T20:23:49.107+0100    DEBUG   kube/owner.go:310       genericEndpointOp address       {"addr.TargetRef.Name": "my-pod", "ep.Name": "my-service"}
	// 2022-01-11T20:23:49.157+0100    DEBUG   kube/owner.go:354       get services    {"pod": "my-pod", "namespace": "kube-system", "pod_services": ["my-service"], "found": true}
	// 2022-01-11T20:23:49.207+0100    DEBUG   kube/owner.go:354       get services    {"pod": "my-pod", "namespace": "kube-system", "pod_services": ["my-service"], "found": true}
	// 2022-01-11T20:23:49.257+0100    DEBUG   kube/owner.go:354       get services    {"pod": "my-pod", "namespace": "kube-system", "pod_services": ["my-service"], "found": true}
	// ...
	//
	// This is most probably an issue with either the client-go's fake.NewSimpleClientset() which
	// is used in the tests here or the way the framework is being used.
	//
	//
	// Running these tests with an even increasing -count proved that this decreases
	// the flakiness of tests but doesn't eliminate it completely.
	// time.Sleep(200 * time.Millisecond)

	t.Run("adding endpoints", func(t *testing.T) {
		_, err = c.CoreV1().Endpoints("kube-system").
			Create(context.Background(), endpoints1, metav1.CreateOptions{})
		require.NoError(t, err)

		_, err = c.CoreV1().Endpoints("kube-system").
			Create(context.Background(), endpoints2, metav1.CreateOptions{})
		require.NoError(t, err)

		_, err = c.CoreV1().Pods("kube-system").
			Create(context.Background(), pod, metav1.CreateOptions{})
		require.NoError(t, err)

		assert.Eventually(t, func() bool {
			services := op.GetServices(pod)
			if len(services) != 2 {
				t.Logf("services: %v", services)
				return false
			}

			return assert.Equal(t, []string{"my-service", "my-service-2"}, services)
		}, 5*time.Second, 50*time.Millisecond)
	})

	t.Run("deleting endpoints", func(t *testing.T) {
		err = c.CoreV1().Endpoints("kube-system").
			Delete(context.Background(), endpoints1.Name, metav1.DeleteOptions{})
		require.NoError(t, err)
		assert.Eventually(t, func() bool {
			services := op.GetServices(pod)
			if len(services) != 1 {
				t.Logf("services: %v", services)
				return false
			}

			return services[0] == "my-service-2"
		}, 5*time.Second, 50*time.Millisecond)

		err = c.CoreV1().Endpoints("kube-system").
			Delete(context.Background(), endpoints2.Name, metav1.DeleteOptions{})
		require.NoError(t, err)
		assert.Eventually(t, func() bool {
			services := op.GetServices(pod)
			if len(services) != 0 {
				t.Logf("services: %v", services)
				return false
			}

			return len(services) == 0
		}, 5*time.Second, 50*time.Millisecond)
	})
}
