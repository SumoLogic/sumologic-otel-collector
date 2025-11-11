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
	"sync"
	"time"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type FakeInformer struct {
	*FakeController

	namespace     string
	labelSelector labels.Selector
	fieldSelector fields.Selector
}

func NewFakeInformer(
	logger *zap.Logger,
	_ kubernetes.Interface,
	namespace string,
	labelSelector labels.Selector,
	fieldSelector fields.Selector,
	limit int,
) cache.SharedInformer {
	return &FakeInformer{
		FakeController: &FakeController{},
		namespace:      namespace,
		labelSelector:  labelSelector,
		fieldSelector:  fieldSelector,
	}
}

func (f *FakeInformer) AddEventHandler(handler cache.ResourceEventHandler) (cache.ResourceEventHandlerRegistration, error) {
	return f.AddEventHandlerWithResyncPeriod(handler, time.Second)
}

func (f *FakeInformer) AddEventHandlerWithResyncPeriod(handler cache.ResourceEventHandler, resyncPeriod time.Duration) (cache.ResourceEventHandlerRegistration, error) {
	return nil, nil
}

func (f *FakeInformer) AddEventHandlerWithOptions(handler cache.ResourceEventHandler, options cache.HandlerOptions) (cache.ResourceEventHandlerRegistration, error) {
	if options.ResyncPeriod != nil {
		return f.AddEventHandlerWithResyncPeriod(handler, *options.ResyncPeriod)
	}
	return f.AddEventHandlerWithResyncPeriod(handler, time.Second)
}

func (f *FakeInformer) RemoveEventHandler(handle cache.ResourceEventHandlerRegistration) error {
	return nil
}

func (f *FakeInformer) IsStopped() bool {
	return false
}

func (f *FakeInformer) GetStore() cache.Store {
	return cache.NewStore(func(obj interface{}) (string, error) { return "", nil })
}

func (f *FakeInformer) GetController() cache.Controller {
	return f.FakeController
}

type FakeController struct {
	sync.Mutex
	stopped bool
}

func (c *FakeController) HasSynced() bool {
	return true
}

func (c *FakeController) Run(stopCh <-chan struct{}) {
	<-stopCh
	c.Lock()
	c.stopped = true
	c.Unlock()
}

func (c *FakeController) RunWithContext(ctx context.Context) {
	<-ctx.Done()
	c.Lock()
	c.stopped = true
	c.Unlock()
}

func (c *FakeController) HasStopped() bool {
	c.Lock()
	defer c.Unlock()
	return c.stopped
}

func (c *FakeController) LastSyncResourceVersion() string {
	return ""
}

func (f *FakeInformer) SetWatchErrorHandler(cache.WatchErrorHandler) error {
	return nil
}

func (f *FakeInformer) SetWatchErrorHandlerWithContext(cache.WatchErrorHandlerWithContext) error {
	return nil
}

func (f *FakeInformer) SetTransform(cache.TransformFunc) error {
	return nil
}
