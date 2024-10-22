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

package k8sprocessor

import (
	"time"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/tools/cache"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sprocessor/kube"
)

// fakeClient is used as a replacement for WatchClient in test cases.
type fakeClient struct {
	Pods         map[kube.PodIdentifier]*kube.Pod
	Rules        kube.ExtractionRules
	Filters      kube.Filters
	Associations []kube.Association
	Informer     cache.SharedInformer
	StopCh       chan struct{}
}

func selectors() (labels.Selector, fields.Selector) {
	var selectors []fields.Selector
	return labels.Everything(), fields.AndSelectors(selectors...)
}

// newFakeClient instantiates a new FakeClient object and satisfies the ClientProvider type
func newFakeClient(
	logger *zap.Logger,
	apiCfg k8sconfig.APIConfig,
	rules kube.ExtractionRules,
	filters kube.Filters,
	associations []kube.Association,
	exclude kube.Excludes,
	_ kube.APIClientsetProvider,
	_ kube.InformerProvider,
	_ kube.OwnerProvider,
	_ string,
	_ int,
	_ time.Duration,
	_ time.Duration,
) (kube.Client, error) {
	cs := fake.NewSimpleClientset()

	ls, fs := selectors()
	return &fakeClient{
		Pods:         map[kube.PodIdentifier]*kube.Pod{},
		Rules:        rules,
		Filters:      filters,
		Associations: associations,
		Informer:     kube.NewFakeInformer(logger, cs, "", ls, fs, 10),
		StopCh:       make(chan struct{}),
	}, nil
}

func (f *fakeClient) GetPodAttributes(identifier kube.PodIdentifier) (map[string]string, bool) {
	p, ok := f.Pods[identifier]
	if !ok {
		return map[string]string{}, ok
	}
	return p.Attributes, ok
}

// Start is a noop for FakeClient.
func (f *fakeClient) Start() {
	if f.Informer != nil {
		f.Informer.Run(f.StopCh)
	}
}

// Stop is a noop for FakeClient.
func (f *fakeClient) Stop() {
	close(f.StopCh)
}
