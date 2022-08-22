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
	batch_v1 "k8s.io/api/batch/v1"
	api_v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
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
	client.PrependWatchReactor(resource, func(action clienttesting.Action) (handled bool, ret watch.Interface, err error) {
		gvr := action.GetResource()
		ns := action.GetNamespace()

		watch, err := client.Tracker().Watch(gvr, ns)
		if err != nil {
			return false, nil, err
		}

		if action.GetVerb() == "watch" {
			close(ch)
		}
		return true, watch, nil
	})
	return ch
}

func Test_OwnerProvider_GetOwners_ReplicaSet(t *testing.T) {
	c, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	require.NoError(t, err)

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	op, err := newOwnerProvider(
		logger,
		c,
		labels.Everything(),
		fields.Everything(),
		ExtractionRules{
			Namespace:          true,
			OwnerLookupEnabled: true,
			PodUID:             true,
			PodName:            true,
			ReplicaSetName:     true,
			Tags:               NewExtractionFieldTags(),
		},
		"kube-system",
	)
	require.NoError(t, err)

	client := c.(*fake.Clientset)
	replicaSetWatchEstablished := waitForWatchToBeEstablished(client, "replicasets")

	op.Start()
	t.Cleanup(func() {
		op.Stop()
	})

	<-replicaSetWatchEstablished

	replicaSetUID := types.UID("fb9e6935-8936-4959-bd90-4e975a4c2b07")
	replicaSet, err := c.AppsV1().ReplicaSets("kube-system").
		Create(context.Background(),
			&v1.ReplicaSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-rs",
					Namespace: "kube-system",
					UID:       replicaSetUID,
				},
				TypeMeta: metav1.TypeMeta{
					Kind: "ReplicaSet",
				},
			},
			metav1.CreateOptions{},
		)
	require.NoError(t, err)

	pod := &api_v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-pod",
			Namespace: "kube-system",
			UID:       "e98a3d3e-fde9-4b10-8f61-cc37d0357c28",
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind: replicaSet.Kind,
					Name: replicaSet.Name,
					UID:  replicaSet.UID,
				},
			},
		},
	}

	_, err = c.CoreV1().Pods("kube-system").
		Create(context.Background(), pod, metav1.CreateOptions{})
	require.NoError(t, err)

	assert.Eventually(t, func() bool {
		owners := op.GetOwners(&Pod{OwnerReferences: &pod.OwnerReferences})
		if len(owners) != 1 {
			t.Logf("owners: %v", owners)
			return false
		}

		if uid := owners[0].UID; uid != replicaSetUID {
			t.Logf("wrong owner UID: %v", uid)
			return false
		}

		return true
	}, 5*time.Second, 5*time.Millisecond)
}

func Test_OwnerProvider_GetOwners_Deployment(t *testing.T) {
	c, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	require.NoError(t, err)

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	op, err := newOwnerProvider(
		logger,
		c,
		labels.Everything(),
		fields.Everything(),
		ExtractionRules{
			Namespace:          true,
			OwnerLookupEnabled: true,
			PodUID:             true,
			PodName:            true,
			DeploymentName:     true,
			Tags:               NewExtractionFieldTags(),
		},
		"kube-system",
	)
	require.NoError(t, err)

	client := c.(*fake.Clientset)
	replicaSetWatchEstablished := waitForWatchToBeEstablished(client, "replicasets")
	deploymentWatchEstablished := waitForWatchToBeEstablished(client, "deployments")

	op.Start()
	t.Cleanup(func() {
		op.Stop()
	})

	<-replicaSetWatchEstablished
	<-deploymentWatchEstablished

	deploymentUID := types.UID("3849f24d-19c2-4b06-97bd-dcb57201a6a4")
	deployment, err := c.AppsV1().ReplicaSets("kube-system").
		Create(context.Background(),
			&v1.ReplicaSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-deploy",
					Namespace: "kube-system",
					UID:       deploymentUID,
				},
				TypeMeta: metav1.TypeMeta{
					Kind: "Deployment",
				},
			},
			metav1.CreateOptions{},
		)
	require.NoError(t, err)

	replicaSetUID := types.UID("fb9e6935-8936-4959-bd90-4e975a4c2b07")
	replicaSet, err := c.AppsV1().ReplicaSets("kube-system").
		Create(context.Background(),
			&v1.ReplicaSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-rs",
					Namespace: "kube-system",
					UID:       replicaSetUID,
					OwnerReferences: []metav1.OwnerReference{
						{
							Kind: deployment.Kind,
							Name: deployment.Name,
							UID:  deployment.UID,
						},
					},
				},
				TypeMeta: metav1.TypeMeta{
					Kind: "ReplicaSet",
				},
			},
			metav1.CreateOptions{},
		)
	require.NoError(t, err)

	pod := &api_v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-pod",
			Namespace: "kube-system",
			UID:       "e98a3d3e-fde9-4b10-8f61-cc37d0357c28",
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind: replicaSet.Kind,
					Name: replicaSet.Name,
					UID:  replicaSet.UID,
				},
			},
		},
	}

	_, err = c.CoreV1().Pods("kube-system").
		Create(context.Background(), pod, metav1.CreateOptions{})
	require.NoError(t, err)

	assert.Eventually(t, func() bool {
		owners := op.GetOwners(&Pod{OwnerReferences: &pod.OwnerReferences})
		if len(owners) != 2 {
			t.Logf("owners: %v", owners)
			return false
		}

		if uid := owners[1].UID; uid != deploymentUID {
			t.Logf("wrong owner UID: %v", uid)
			return false
		}

		return true
	}, 5*time.Second, 5*time.Millisecond)
}

func Test_OwnerProvider_GetOwners_Statefulset(t *testing.T) {
	c, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	require.NoError(t, err)

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	op, err := newOwnerProvider(
		logger,
		c,
		labels.Everything(),
		fields.Everything(),
		ExtractionRules{
			PodUID:             true,
			PodName:            true,
			StatefulSetName:    true,
			Namespace:          true,
			OwnerLookupEnabled: true,
			Tags:               NewExtractionFieldTags(),
		},
		"kube-system",
	)
	require.NoError(t, err)

	client := c.(*fake.Clientset)
	ch := waitForWatchToBeEstablished(client, "statefulsets")

	op.Start()
	t.Cleanup(func() {
		op.Stop()
	})

	<-ch

	statefulSetUID := types.UID("5513d7a3-3edd-4bd1-b036-7f4c6fb6eb46")
	sts, err := c.AppsV1().StatefulSets("kube-system").
		Create(context.Background(),
			&v1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-sts",
					Namespace: "kube-system",
					UID:       statefulSetUID,
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
		owners := op.GetOwners(&Pod{OwnerReferences: &pod.OwnerReferences})
		if len(owners) != 1 {
			t.Logf("owners: %v", owners)
			return false
		}

		if uid := owners[0].UID; uid != statefulSetUID {
			t.Logf("wrong owner UID: %v", uid)
			return false
		}

		return true
	}, 5*time.Second, 5*time.Millisecond)
}

func Test_OwnerProvider_GetOwners_Daemonset(t *testing.T) {
	c, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	require.NoError(t, err)

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	op, err := newOwnerProvider(
		logger,
		c,
		labels.Everything(),
		fields.Everything(),
		ExtractionRules{
			PodUID:             true,
			PodName:            true,
			DaemonSetName:      true,
			Namespace:          true,
			OwnerLookupEnabled: true,
			Tags:               NewExtractionFieldTags(),
		},
		"kube-system",
	)
	require.NoError(t, err)

	client := c.(*fake.Clientset)
	ch := waitForWatchToBeEstablished(client, "daemonsets")

	op.Start()
	t.Cleanup(func() {
		op.Stop()
	})

	<-ch

	daemonSetUID := types.UID("ac264398-d301-4d32-b75d-0d073b119ccd")
	ds, err := c.AppsV1().DaemonSets("kube-system").
		Create(context.Background(),
			&v1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-daemonset",
					Namespace: "kube-system",
					UID:       daemonSetUID,
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
			UID:       "f15f0585-a0bc-43a3-96e4-dd2eace75396",
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind: ds.Kind,
					Name: ds.Name,
					UID:  ds.UID,
				},
			},
		},
	}

	_, err = c.CoreV1().Pods("kube-system").
		Create(context.Background(), pod, metav1.CreateOptions{})
	require.NoError(t, err)

	assert.Eventually(t, func() bool {
		owners := op.GetOwners(&Pod{OwnerReferences: &pod.OwnerReferences})
		if len(owners) != 1 {
			t.Logf("owners: %v", owners)
			return false
		}

		if uid := owners[0].UID; uid != daemonSetUID {
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

	op, err := newOwnerProvider(
		logger,
		c,
		labels.Everything(),
		fields.Everything(),
		ExtractionRules{
			PodUID:             true,
			PodName:            true,
			Namespace:          true,
			ServiceName:        true,
			OwnerLookupEnabled: true,
			Tags:               NewExtractionFieldTags(),
		},
		namespace,
	)
	require.NoError(t, err)

	client := c.(*fake.Clientset)
	ch := waitForWatchToBeEstablished(client, "endpoints")

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
				UID:       "88125104-a4f6-40ac-906b-fcd385c127f3",
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
				UID:       "07ffe4a1-ca89-4d28-acb5-808b0c0bb20f",
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
			services := op.GetServices(pod.Name)
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
			services := op.GetServices(pod.Name)
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
			services := op.GetServices(pod.Name)
			if len(services) != 0 {
				t.Logf("services: %v", services)
				return false
			}

			return len(services) == 0
		}, 5*time.Second, 10*time.Millisecond)
	})
}

func Test_OwnerProvider_GetOwners_Job(t *testing.T) {
	c, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	require.NoError(t, err)

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	op, err := newOwnerProvider(
		logger,
		c,
		labels.Everything(),
		fields.Everything(),
		ExtractionRules{
			Namespace:          true,
			OwnerLookupEnabled: true,
			PodUID:             true,
			PodName:            true,
			JobName:            true,
			Tags:               NewExtractionFieldTags(),
		},
		"kube-system",
	)
	require.NoError(t, err)

	client := c.(*fake.Clientset)
	ch := waitForWatchToBeEstablished(client, "jobs")

	op.Start()
	t.Cleanup(func() {
		op.Stop()
	})

	<-ch

	jobUID := types.UID("1062885b-a745-4ff7-9617-2566f7e99531")
	job, err := c.BatchV1().Jobs("kube-system").
		Create(context.Background(),
			&batch_v1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-job",
					Namespace: "kube-system",
					UID:       jobUID,
				},
				TypeMeta: metav1.TypeMeta{
					Kind: "Job",
				},
			},
			metav1.CreateOptions{},
		)
	require.NoError(t, err)

	pod := &api_v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-pod",
			Namespace: "kube-system",
			UID:       "e98a3d3e-fde9-4b10-8f61-cc37d0357c28",
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind: job.Kind,
					Name: job.Name,
					UID:  job.UID,
				},
			},
		},
	}

	_, err = c.CoreV1().Pods("kube-system").
		Create(context.Background(), pod, metav1.CreateOptions{})
	require.NoError(t, err)

	assert.Eventually(t, func() bool {
		owners := op.GetOwners(&Pod{OwnerReferences: &pod.OwnerReferences})
		if len(owners) != 1 {
			t.Logf("owners: %v", owners)
			return false
		}

		if uid := owners[0].UID; uid != jobUID {
			t.Logf("wrong owner UID: %v", uid)
			return false
		}

		return true
	}, 5*time.Second, 5*time.Millisecond)
}

func Test_OwnerProvider_GetOwners_CronJob(t *testing.T) {
	c, err := newFakeAPIClientset(k8sconfig.APIConfig{})
	require.NoError(t, err)

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	op, err := newOwnerProvider(
		logger,
		c,
		labels.Everything(),
		fields.Everything(),
		ExtractionRules{
			Namespace:          true,
			OwnerLookupEnabled: true,
			PodUID:             true,
			PodName:            true,
			CronJobName:        true,
			Tags:               NewExtractionFieldTags(),
		},
		"kube-system",
	)
	require.NoError(t, err)

	client := c.(*fake.Clientset)
	jobWatchEstablished := waitForWatchToBeEstablished(client, "jobs")
	cronJobWatchEstablished := waitForWatchToBeEstablished(client, "cronjobs")

	op.Start()
	t.Cleanup(func() {
		op.Stop()
	})

	<-jobWatchEstablished
	<-cronJobWatchEstablished

	cronJobUID := types.UID("fcc51d15-a279-4738-8857-f4e905c84226")
	cronJob, err := c.BatchV1().CronJobs("kube-system").
		Create(context.Background(),
			&batch_v1.CronJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-cronjob",
					Namespace: "kube-system",
					UID:       cronJobUID,
				},
				TypeMeta: metav1.TypeMeta{
					Kind: "CronJob",
				},
			},
			metav1.CreateOptions{},
		)
	require.NoError(t, err)

	jobUID := types.UID("1062885b-a745-4ff7-9617-2566f7e99531")
	job, err := c.BatchV1().Jobs("kube-system").
		Create(context.Background(),
			&batch_v1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-job",
					Namespace: "kube-system",
					UID:       jobUID,
					OwnerReferences: []metav1.OwnerReference{
						{
							Kind: cronJob.Kind,
							Name: cronJob.Name,
							UID:  cronJob.UID,
						},
					},
				},
				TypeMeta: metav1.TypeMeta{
					Kind: "Job",
				},
			},
			metav1.CreateOptions{},
		)
	require.NoError(t, err)

	pod := &api_v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-pod",
			Namespace: "kube-system",
			UID:       "e98a3d3e-fde9-4b10-8f61-cc37d0357c28",
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind: job.Kind,
					Name: job.Name,
					UID:  job.UID,
				},
			},
		},
	}

	_, err = c.CoreV1().Pods("kube-system").
		Create(context.Background(), pod, metav1.CreateOptions{})
	require.NoError(t, err)

	assert.Eventually(t, func() bool {
		owners := op.GetOwners(&Pod{OwnerReferences: &pod.OwnerReferences})
		if len(owners) != 2 {
			t.Logf("owners: %v", owners)
			return false
		}

		if uid := owners[1].UID; uid != cronJobUID {
			t.Logf("wrong owner UID: %v", uid)
			return false
		}

		return true
	}, 5*time.Second, 5*time.Millisecond)
}
