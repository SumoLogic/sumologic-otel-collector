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
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
)

// waitForWatchToBeEstablished tries to solve an issue with a data race when using
// a fake client with informers.
//
// Basically there is a small amount of time between starting the informers and
// establishing a watch that the notifications might get lost.
// In order to mitigate that wait for a watch to be established for a particular resource.
//
// Related issue: https://github.com/kubernetes/kubernetes/issues/95372
func waitForWatchToBeEstablished(client *fake.Clientset, resource string) <-chan struct{} {
	ch := make(chan struct{})
	client.PrependWatchReactor("*", func(action clienttesting.Action) (handled bool, ret watch.Interface, err error) {
		gvr := action.GetResource()
		ns := action.GetNamespace()

		watch, err := client.Tracker().Watch(gvr, ns)
		if err != nil {
			return false, nil, err
		}

		if action.GetVerb() == "watch" && gvr.String() == resource {
			close(ch)
		}
		return true, watch, nil
	})
	return ch
}

func Test_OwnerProvider_GetOwners(t *testing.T) {
	c, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	require.NoError(t, err)

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	op, err := newOwnerProvider(logger, c, labels.Everything(), fields.Everything(), "kube-system")
	require.NoError(t, err)

	client := c.(*fake.Clientset)
	ch := waitForWatchToBeEstablished(client, "apps/v1, Resource=statefulsets")

	op.Start()
	t.Cleanup(func() {
		op.Stop()
	})

	<-ch

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

	pod := &api_v1.Pod{
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
	}

	_, err = c.CoreV1().Pods("kube-system").
		Create(context.Background(), pod, metav1.CreateOptions{})
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
	}, 5*time.Second, 5*time.Millisecond)
}

func Test_OwnerProvider_GetServices(t *testing.T) {
	const (
		namespace = "kube-system"
	)

	c, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	require.NoError(t, err)

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	op, err := newOwnerProvider(logger, c, labels.Everything(), fields.Everything(), namespace)
	require.NoError(t, err)

	client := c.(*fake.Clientset)
	ch := waitForWatchToBeEstablished(client, "/v1, Resource=endpoints")

	op.Start()
	t.Cleanup(func() {
		op.Stop()
	})

	var (
		pod = &api_v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-pod",
				Namespace: namespace,
				UID:       "f15f0585-a0bc-43a3-96e4-dd2eace75392",
			},
		}
		endpoints1 = &api_v1.Endpoints{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-service",
				Namespace: namespace,
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
								Namespace: namespace,
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
				Namespace: namespace,
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
								Namespace: namespace,
								Kind:      "Pod",
								UID:       pod.UID,
							},
						},
					},
				},
			},
		}
	)

	<-ch

	t.Run("adding endpoints", func(t *testing.T) {
		_, err = c.CoreV1().Endpoints(namespace).
			Create(context.Background(), endpoints1, metav1.CreateOptions{})
		require.NoError(t, err)

		_, err = c.CoreV1().Endpoints(namespace).
			Create(context.Background(), endpoints2, metav1.CreateOptions{})
		require.NoError(t, err)

		_, err = c.CoreV1().Pods(namespace).
			Create(context.Background(), pod, metav1.CreateOptions{})
		require.NoError(t, err)

		assert.Eventually(t, func() bool {
			services := op.GetServices(pod)
			if len(services) != 2 {
				t.Logf("services: %v", services)
				return false
			}

			return assert.Equal(t, []string{"my-service", "my-service-2"}, services)
		}, 5*time.Second, 10*time.Millisecond)
	})

	t.Run("deleting endpoints", func(t *testing.T) {
		err = c.CoreV1().Endpoints(namespace).
			Delete(context.Background(), endpoints1.Name, metav1.DeleteOptions{})
		require.NoError(t, err)
		assert.Eventually(t, func() bool {
			services := op.GetServices(pod)
			if len(services) != 1 {
				t.Logf("services: %v", services)
				return false
			}

			return services[0] == "my-service-2"
		}, 5*time.Second, 10*time.Millisecond)

		err = c.CoreV1().Endpoints(namespace).
			Delete(context.Background(), endpoints2.Name, metav1.DeleteOptions{})
		require.NoError(t, err)

		assert.Eventually(t, func() bool {
			services := op.GetServices(pod)
			if len(services) != 0 {
				t.Logf("services: %v", services)
				return false
			}

			return len(services) == 0
		}, 5*time.Second, 10*time.Millisecond)
	})
}
