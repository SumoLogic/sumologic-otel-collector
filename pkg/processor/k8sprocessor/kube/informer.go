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

	"go.uber.org/zap"
	api_v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// InformerProvider defines a function type that returns a new SharedInformer. It is used to
// allow passing custom shared informers to the watch client.
type InformerProvider func(
	logger *zap.Logger,
	client kubernetes.Interface,
	namespace string,
	labelSelector labels.Selector,
	fieldSelector fields.Selector,
) cache.SharedInformer

func newSharedInformer(
	logger *zap.Logger,
	client kubernetes.Interface,
	namespace string,
	ls labels.Selector,
	fs fields.Selector,
) cache.SharedInformer {
	informer := cache.NewSharedInformer(
		&cache.ListWatch{
			ListFunc:  informerListFuncWithSelectors(logger, client, namespace, ls, fs),
			WatchFunc: informerWatchFuncWithSelectors(client, namespace, ls, fs),
		},
		&api_v1.Pod{},
		watchSyncPeriod,
	)
	return informer
}

func informerListFuncWithSelectors(logger *zap.Logger, client kubernetes.Interface, namespace string, ls labels.Selector, fs fields.Selector) cache.ListFunc {
	return func(opts metav1.ListOptions) (runtime.Object, error) {
		podList := &api_v1.PodList{}
		for {
			opts.LabelSelector = ls.String()
			opts.FieldSelector = fs.String()
			opts.Limit = 200
			// Unset resource version to get the latest data
			opts.ResourceVersion = ""
			pods, err := client.CoreV1().Pods(namespace).List(context.Background(), opts)
			if err != nil {
				return nil, err
			}
			podList.Items = append(podList.Items, pods.Items...)
			if pods.Continue == "" {
				break
			}
			opts.Continue = pods.Continue
			logger.Debug("Fetching more pods", zap.String("continue", pods.Continue))
		}
		return podList, nil
	}

}

func informerWatchFuncWithSelectors(client kubernetes.Interface, namespace string, ls labels.Selector, fs fields.Selector) cache.WatchFunc {
	return func(opts metav1.ListOptions) (watch.Interface, error) {
		opts.LabelSelector = ls.String()
		opts.FieldSelector = fs.String()
		opts.ResourceVersion = ""
		return client.CoreV1().Pods(namespace).Watch(context.Background(), opts)
	}
}
