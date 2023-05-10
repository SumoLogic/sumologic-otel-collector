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
	"os"
	"reflect"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	conventions "go.opentelemetry.io/collector/semconv/v1.18.0"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"k8s.io/apimachinery/pkg/selection"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sprocessor/kube"
)

func TestWithAPIConfig(t *testing.T) {
	p := &kubernetesprocessor{}
	apiConfig := k8sconfig.APIConfig{AuthType: "test-auth-type"}
	err := WithAPIConfig(apiConfig)(p)
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "invalid authType for kubernetes: test-auth-type")

	apiConfig = k8sconfig.APIConfig{AuthType: "kubeConfig"}
	err = WithAPIConfig(apiConfig)(p)
	assert.NoError(t, err)
	assert.Equal(t, apiConfig, p.apiConfig)
}

func TestWithFilterNamespace(t *testing.T) {
	p := &kubernetesprocessor{}
	assert.NoError(t, WithFilterNamespace("testns")(p))
	assert.Equal(t, p.filters.Namespace, "testns")
}

func TestWithFilterNode(t *testing.T) {
	p := &kubernetesprocessor{}
	assert.NoError(t, WithFilterNode("testnode", "")(p))
	assert.Equal(t, p.filters.Node, "testnode")

	p = &kubernetesprocessor{}
	assert.NoError(t, WithFilterNode("testnode", "NODE_NAME")(p))
	assert.Equal(t, p.filters.Node, "")

	os.Setenv("NODE_NAME", "nodefromenv")
	p = &kubernetesprocessor{}
	assert.NoError(t, WithFilterNode("testnode", "NODE_NAME")(p))
	assert.Equal(t, p.filters.Node, "nodefromenv")

	os.Unsetenv("NODE_NAME")
}

func TestWithPassthrough(t *testing.T) {
	p := &kubernetesprocessor{}
	assert.NoError(t, WithPassthrough()(p))
	assert.True(t, p.passthroughMode)
}

func TestWithExtractAnnotations(t *testing.T) {
	tests := []struct {
		name      string
		args      []FieldExtractConfig
		want      []kube.FieldExtractionRule
		wantError string
	}{
		{
			"empty",
			[]FieldExtractConfig{},
			[]kube.FieldExtractionRule{},
			"",
		},
		{
			"bad",
			[]FieldExtractConfig{
				{
					TagName: "tag1",
					Key:     "key1",
					Regex:   "[",
				},
			},
			[]kube.FieldExtractionRule{},
			"error parsing regexp: missing closing ]: `[`",
		},
		{
			"basic",
			[]FieldExtractConfig{
				{
					TagName: "tag1",
					Key:     "key1",
					Regex:   "field=(?P<value>.+)",
				},
			},
			[]kube.FieldExtractionRule{
				{
					Name:  "tag1",
					Key:   "key1",
					Regex: regexp.MustCompile(`field=(?P<value>.+)`),
				},
			},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &kubernetesprocessor{}
			option := WithExtractAnnotations(tt.args...)
			err := option(p)
			if tt.wantError != "" {
				assert.Error(t, err)
				assert.Equal(t, err.Error(), tt.wantError)
				return
			}

			assert.NoError(t, err)
			got := p.rules.Annotations
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithExtractAnnotations() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithExtractLabels(t *testing.T) {
	tests := []struct {
		name      string
		args      []FieldExtractConfig
		want      []kube.FieldExtractionRule
		wantError string
	}{
		{
			"empty",
			[]FieldExtractConfig{},
			[]kube.FieldExtractionRule{},
			"",
		},
		{
			"bad",
			[]FieldExtractConfig{{
				TagName: "t1",
				Key:     "k1",
				Regex:   "[",
			}},
			[]kube.FieldExtractionRule{},
			"error parsing regexp: missing closing ]: `[`",
		},
		{
			"basic",
			[]FieldExtractConfig{
				{
					TagName: "tag1",
					Key:     "key1",
					Regex:   "field=(?P<value>.+)",
				},
			},
			[]kube.FieldExtractionRule{
				{
					Name:  "tag1",
					Key:   "key1",
					Regex: regexp.MustCompile(`field=(?P<value>.+)`),
				},
			},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &kubernetesprocessor{}
			option := WithExtractLabels(tt.args...)
			err := option(p)
			if tt.wantError != "" {
				assert.Error(t, err)
				assert.Equal(t, err.Error(), tt.wantError)
				return
			}

			assert.NoError(t, err)
			got := p.rules.Labels
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithExtractLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithExtractNamespaceLabels(t *testing.T) {
	tests := []struct {
		name      string
		args      []FieldExtractConfig
		want      []kube.FieldExtractionRule
		wantError string
	}{
		{
			"empty",
			[]FieldExtractConfig{},
			[]kube.FieldExtractionRule{},
			"",
		},
		{
			"bad",
			[]FieldExtractConfig{{
				TagName: "t1",
				Key:     "k1",
				Regex:   "[",
			}},
			[]kube.FieldExtractionRule{},
			"error parsing regexp: missing closing ]: `[`",
		},
		{
			"basic",
			[]FieldExtractConfig{
				{
					TagName: "tag1",
					Key:     "key1",
					Regex:   "field=(?P<value>.+)",
				},
			},
			[]kube.FieldExtractionRule{
				{
					Name:  "tag1",
					Key:   "key1",
					Regex: regexp.MustCompile(`field=(?P<value>.+)`),
				},
			},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &kubernetesprocessor{}
			option := WithExtractNamespaceLabels(tt.args...)
			err := option(p)
			if tt.wantError != "" {
				assert.Error(t, err)
				assert.Equal(t, err.Error(), tt.wantError)
				return
			}

			assert.NoError(t, err)
			got := p.rules.NamespaceLabels
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithExtractNamespaceLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithExtractMetadata(t *testing.T) {
	p := &kubernetesprocessor{}
	assert.NoError(t, WithExtractMetadata()(p))
	assert.True(t, p.rules.Namespace)
	assert.True(t, p.rules.PodName)
	assert.True(t, p.rules.PodUID)
	assert.True(t, p.rules.StartTime)
	assert.True(t, p.rules.DeploymentName)
	assert.True(t, p.rules.NodeName)

	p = &kubernetesprocessor{}
	err := WithExtractMetadata("randomfield")(p)
	assert.Error(t, err)
	assert.Equal(t, err.Error(), `"randomfield" is not a supported metadata field`)

	assert.NoError(t, WithExtractMetadata("namespace")(p))
	assert.True(t, p.rules.Namespace)
	assert.False(t, p.rules.PodName)
	assert.False(t, p.rules.PodUID)
	assert.False(t, p.rules.StartTime)
	assert.False(t, p.rules.DeploymentName)
	assert.False(t, p.rules.NodeName)
}

func TestWithExtractMetadataSemanticConventions(t *testing.T) {
	p := &kubernetesprocessor{}
	fields := []string{
		conventions.AttributeContainerID,
		conventions.AttributeContainerImageName,
		conventions.AttributeContainerName,
		conventions.AttributeK8SCronJobName,
		conventions.AttributeK8SDaemonSetName,
		conventions.AttributeK8SDeploymentName,
		conventions.AttributeHostName,
		conventions.AttributeK8SJobName,
		conventions.AttributeK8SNamespaceName,
		conventions.AttributeK8SNodeName,
		conventions.AttributeK8SPodUID,
		conventions.AttributeK8SPodName,
		conventions.AttributeK8SReplicaSetName,
		metadataOtelSemconvServiceName,
		conventions.AttributeK8SStatefulSetName,
		metadataOtelPodStartTime,
	}
	assert.NoError(t, WithExtractMetadata(fields...)(p))
	assert.True(t, p.rules.ContainerID)
	assert.True(t, p.rules.ContainerImage)
	assert.True(t, p.rules.ContainerName)
	assert.True(t, p.rules.CronJobName)
	assert.True(t, p.rules.DaemonSetName)
	assert.True(t, p.rules.DeploymentName)
	assert.True(t, p.rules.HostName)
	assert.True(t, p.rules.JobName)
	assert.True(t, p.rules.Namespace)
	assert.True(t, p.rules.NodeName)
	assert.True(t, p.rules.PodName)
	assert.True(t, p.rules.PodUID)
	assert.True(t, p.rules.ReplicaSetName)
	assert.True(t, p.rules.ServiceName)
	assert.True(t, p.rules.StatefulSetName)
	assert.True(t, p.rules.StartTime)
}

func TestWithExtractMetadataDeprecatedOption(t *testing.T) {
	core, observedLogs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)
	p := &kubernetesprocessor{logger: logger}
	assert.NoError(t, WithExtractMetadata("clusterName")(p))
	assert.Equal(t, 1, observedLogs.Len())
	assert.Equal(t, 1, observedLogs.FilterMessage("clusterName metadata field has been deprecated and will be removed soon").Len())
	observedLogs.TakeAll()

	assert.NoError(t, WithExtractTags(map[string]string{"clustername": "cluster"})(p))
	assert.Equal(t, 1, observedLogs.Len())
	assert.Equal(t, 1, observedLogs.FilterMessage("clusterName metadata field has been deprecated and will be removed soon").Len())
}

func TestWithFilterLabels(t *testing.T) {
	tests := []struct {
		name      string
		args      []FieldFilterConfig
		want      []kube.FieldFilter
		wantError string
	}{
		{
			"empty",
			[]FieldFilterConfig{},
			[]kube.FieldFilter{},
			"",
		},
		{
			"default",
			[]FieldFilterConfig{
				{
					Key:   "k1",
					Value: "v1",
				},
			},
			[]kube.FieldFilter{
				{
					Key:   "k1",
					Value: "v1",
					Op:    selection.Equals,
				},
			},
			"",
		},
		{
			"equals",
			[]FieldFilterConfig{
				{
					Key:   "k1",
					Value: "v1",
					Op:    "equals",
				},
			},
			[]kube.FieldFilter{
				{
					Key:   "k1",
					Value: "v1",
					Op:    selection.Equals,
				},
			},
			"",
		},
		{
			"not-equals",
			[]FieldFilterConfig{
				{
					Key:   "k1",
					Value: "v1",
					Op:    "not-equals",
				},
			},
			[]kube.FieldFilter{
				{
					Key:   "k1",
					Value: "v1",
					Op:    selection.NotEquals,
				},
			},
			"",
		},
		{
			"exists",
			[]FieldFilterConfig{
				{
					Key: "k1",
					Op:  "exists",
				},
			},
			[]kube.FieldFilter{
				{
					Key: "k1",
					Op:  selection.Exists,
				},
			},
			"",
		},
		{
			"does-not-exist",
			[]FieldFilterConfig{
				{
					Key: "k1",
					Op:  "does-not-exist",
				},
			},
			[]kube.FieldFilter{
				{
					Key: "k1",
					Op:  selection.DoesNotExist,
				},
			},
			"",
		},
		{
			"unknown",
			[]FieldFilterConfig{
				{
					Key:   "k1",
					Value: "v1",
					Op:    "unknown-op",
				},
			},
			[]kube.FieldFilter{},
			"'unknown-op' is not a valid label filter operation for key=k1, value=v1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &kubernetesprocessor{}
			option := WithFilterLabels(tt.args...)
			err := option(p)
			if tt.wantError != "" {
				assert.Error(t, err)
				assert.Equal(t, err.Error(), tt.wantError)
				return
			}

			assert.NoError(t, err)
			got := p.filters.Labels
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithExtractLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithFilterFields(t *testing.T) {

	tests := []struct {
		name      string
		args      []FieldFilterConfig
		want      []kube.FieldFilter
		wantError string
	}{
		{
			"empty",
			[]FieldFilterConfig{},
			[]kube.FieldFilter{},
			"",
		},
		{
			"default",
			[]FieldFilterConfig{
				{
					Key:   "k1",
					Value: "v1",
				},
			},
			[]kube.FieldFilter{
				{
					Key:   "k1",
					Value: "v1",
					Op:    selection.Equals,
				},
			},
			"",
		},
		{
			"equals",
			[]FieldFilterConfig{
				{
					Key:   "k1",
					Value: "v1",
					Op:    "equals",
				},
			},
			[]kube.FieldFilter{
				{
					Key:   "k1",
					Value: "v1",
					Op:    selection.Equals,
				},
			},
			"",
		},
		{
			"not-equals",
			[]FieldFilterConfig{
				{
					Key:   "k1",
					Value: "v1",
					Op:    "not-equals",
				},
			},
			[]kube.FieldFilter{
				{
					Key:   "k1",
					Value: "v1",
					Op:    selection.NotEquals,
				},
			},
			"",
		},
		{
			"exists",
			[]FieldFilterConfig{
				{
					Key: "k1",
					Op:  "exists",
				},
			},
			[]kube.FieldFilter{
				{
					Key: "k1",
					Op:  selection.Exists,
				},
			},
			"'exists' is not a valid field filter operation for key=k1, value=",
		},
		{
			"does-not-exist",
			[]FieldFilterConfig{
				{
					Key: "k1",
					Op:  "does-not-exist",
				},
			},
			[]kube.FieldFilter{
				{
					Key: "k1",
					Op:  selection.DoesNotExist,
				},
			},
			"'does-not-exist' is not a valid field filter operation for key=k1, value=",
		},
		{
			"unknown",
			[]FieldFilterConfig{
				{
					Key:   "k1",
					Value: "v1",
					Op:    "unknown-op",
				},
			},
			[]kube.FieldFilter{},
			"'unknown-op' is not a valid field filter operation for key=k1, value=v1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &kubernetesprocessor{}
			option := WithFilterFields(tt.args...)
			err := option(p)
			if tt.wantError != "" {
				assert.Error(t, err)
				assert.Equal(t, err.Error(), tt.wantError)
				return
			}

			assert.NoError(t, err)
			got := p.filters.Fields
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithExtractLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_extractFieldRules(t *testing.T) {
	type args struct {
		fieldType string
		fields    []FieldExtractConfig
	}
	tests := []struct {
		name    string
		args    args
		want    []kube.FieldExtractionRule
		wantErr bool
	}{
		{
			"default",
			args{"labels", []FieldExtractConfig{
				{
					Key: "key",
				},
			}},
			[]kube.FieldExtractionRule{
				{
					Name: "k8s.labels.key",
					Key:  "key",
				},
			},
			false,
		},
		{
			"basic",
			args{"field", []FieldExtractConfig{
				{
					TagName: "name",
					Key:     "key",
				},
			}},
			[]kube.FieldExtractionRule{
				{
					Name: "name",
					Key:  "key",
				},
			},
			false,
		},
		{
			"regex-without-match",
			args{"field", []FieldExtractConfig{
				{
					TagName: "name",
					Key:     "key",
					Regex:   "^h$",
				},
			}},
			[]kube.FieldExtractionRule{},
			true,
		},
		{
			"badregex",
			args{"field", []FieldExtractConfig{
				{
					TagName: "name",
					Key:     "key",
					Regex:   "[",
				},
			}},
			[]kube.FieldExtractionRule{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractFieldRules(tt.args.fieldType, tt.args.fields...)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractFieldRules() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractFieldRules() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithExtractPodAssociation(t *testing.T) {
	tests := []struct {
		name string
		args []PodAssociationConfig
		want []kube.Association
	}{
		{
			"empty",
			[]PodAssociationConfig{},
			[]kube.Association{},
		},
		{
			"basic",
			[]PodAssociationConfig{
				{
					From: "label",
					Name: "ip",
				},
			},
			[]kube.Association{
				{
					From: "label",
					Name: "ip",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &kubernetesprocessor{}
			option := WithExtractPodAssociations(tt.args...)
			assert.NoError(t, option(p))
			assert.Equal(t, tt.want, p.podAssociations)
		})
	}
}

func TestWithExcludes(t *testing.T) {
	tests := []struct {
		name string
		args ExcludeConfig
		want kube.Excludes
	}{
		{
			"default",
			ExcludeConfig{},
			kube.Excludes{},
		},
		{
			"configured",
			ExcludeConfig{
				Pods: []ExcludePodConfig{
					{Name: "ignore_pod1"},
					{Name: "ignore_pod2"},
				},
			},
			kube.Excludes{
				Pods: []kube.ExcludePods{
					{Name: regexp.MustCompile(`ignore_pod1`)},
					{Name: regexp.MustCompile(`ignore_pod2`)},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &kubernetesprocessor{}
			option := WithExcludes(tt.args)
			require.NoError(t, option(p))
			assert.Equal(t, tt.want, p.podIgnore)
		})
	}
}
