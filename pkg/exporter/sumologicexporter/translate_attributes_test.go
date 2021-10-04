// Copyright 2021 Sumo Logic, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sumologicexporter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/model/pdata"
)

func TestTranslateAttributes(t *testing.T) {
	attributes := pdata.NewAttributeMap()
	attributes.InsertString("host.name", "testing-host")
	attributes.InsertString("host.id", "my-host-id")
	attributes.InsertString("host.type", "my-host-type")
	attributes.InsertString("k8s.cluster.name", "testing-cluster")
	attributes.InsertString("k8s.deployment.name", "my-deployment-name")
	attributes.InsertString("k8s.namespace.name", "my-namespace-name")
	attributes.InsertString("k8s.service.name", "my-service-name, other-service")
	attributes.InsertString("cloud.account.id", "my-account-id")
	attributes.InsertString("cloud.availability_zone", "my-zone")
	attributes.InsertString("cloud.region", "my-region")
	require.Equal(t, 10, attributes.Len())

	attributes = translateAttributes(attributes)

	assert.Equal(t, 10, attributes.Len())
	assertAttribute(t, attributes, "host", "testing-host")
	assertAttribute(t, attributes, "host.name", "")
	assertAttribute(t, attributes, "AccountId", "my-account-id")
	assertAttribute(t, attributes, "cloud.account.id", "")
	assertAttribute(t, attributes, "AvailabilityZone", "my-zone")
	assertAttribute(t, attributes, "clout.availability_zone", "")
	assertAttribute(t, attributes, "Region", "my-region")
	assertAttribute(t, attributes, "cloud.region", "")
	assertAttribute(t, attributes, "InstanceId", "my-host-id")
	assertAttribute(t, attributes, "host.id", "")
	assertAttribute(t, attributes, "InstanceType", "my-host-type")
	assertAttribute(t, attributes, "host.type", "")
	assertAttribute(t, attributes, "Cluster", "testing-cluster")
	assertAttribute(t, attributes, "k8s.cluster.name", "")
	assertAttribute(t, attributes, "deployment", "my-deployment-name")
	assertAttribute(t, attributes, "k8s.deployment.name", "")
	assertAttribute(t, attributes, "namespace", "my-namespace-name")
	assertAttribute(t, attributes, "k8s.namespace.name", "")
	assertAttribute(t, attributes, "service", "my-service-name, other-service")
	assertAttribute(t, attributes, "k8s.service.name", "")
}

func TestTranslateAttributesDoesNothingWhenAttributeDoesNotExist(t *testing.T) {
	attributes := pdata.NewAttributeMap()
	require.Equal(t, 0, attributes.Len())

	attributes = translateAttributes(attributes)

	assert.Equal(t, 0, attributes.Len())
	assertAttribute(t, attributes, "host", "")
}

func TestTranslateAttributesLeavesOtherAttributesUnchanged(t *testing.T) {
	attributes := pdata.NewAttributeMap()
	attributes.InsertString("one", "one1")
	attributes.InsertString("host.name", "host1")
	attributes.InsertString("three", "three1")
	require.Equal(t, 3, attributes.Len())

	attributes = translateAttributes(attributes)

	assert.Equal(t, 3, attributes.Len())
	assertAttribute(t, attributes, "one", "one1")
	assertAttribute(t, attributes, "host", "host1")
	assertAttribute(t, attributes, "three", "three1")
}

func TestTranslateAttributesDoesNotOverwriteExistingAttribute(t *testing.T) {
	attributes := pdata.NewAttributeMap()
	attributes.InsertString("host", "host1")
	attributes.InsertString("host.name", "hostname1")
	require.Equal(t, 2, attributes.Len())

	attributes = translateAttributes(attributes)

	assert.Equal(t, 2, attributes.Len())
	assertAttribute(t, attributes, "host", "host1")
	assertAttribute(t, attributes, "host.name", "hostname1")
}

func TestTranslateAttributesDoesNotOverwriteMultipleExistingAttributes(t *testing.T) {
	// Note: Current implementation of pdata.AttributeMap does not allow to insert duplicate keys.
	// See https://cloud-native.slack.com/archives/C01N5UCHTEH/p1624020829067500
	attributes := pdata.NewAttributeMap()
	attributes.InsertString("host", "host1")
	attributes.InsertString("host", "host2")
	require.Equal(t, 1, attributes.Len())
	attributes.InsertString("host.name", "hostname1")
	require.Equal(t, 2, attributes.Len())

	attributes = translateAttributes(attributes)

	assert.Equal(t, 2, attributes.Len())
	assertAttribute(t, attributes, "host", "host1")
	assertAttribute(t, attributes, "host.name", "hostname1")
}

func assertAttribute(t *testing.T, metadata pdata.AttributeMap, attributeName string, expectedValue string) {
	value, exists := metadata.Get(attributeName)

	if expectedValue == "" {
		assert.False(t, exists)
	} else {
		assert.True(t, exists)
		assert.Equal(t, expectedValue, value.StringVal())

	}
}

func TestTranslateConfigValue(t *testing.T) {
	testcases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "basic",
			input: "%{k8s.pod.name}-%{host.name}",
			want:  "%{pod}-%{host}",
		},
		{
			name:  "basic with sumo convention tags",
			input: "%{k8s.pod.name}-%{host.name}/%{pod}-%{host}",
			want:  "%{pod}-%{host}/%{pod}-%{host}",
		},
		{
			name:  "custom attributes",
			input: "%{_sourceCategory}-%{my_custom_vendor_attr}",
			want:  "%{_sourceCategory}-%{my_custom_vendor_attr}",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, translateConfigValue(tc.input))
		})
	}
}

var (
	bench_pdata_attributes = map[string]pdata.AttributeValue{
		"host.name":               pdata.NewAttributeValueString("testing-host"),
		"host.id":                 pdata.NewAttributeValueString("my-host-id"),
		"host.type":               pdata.NewAttributeValueString("my-host-type"),
		"k8s.cluster.name":        pdata.NewAttributeValueString("testing-cluster"),
		"k8s.deployment.name":     pdata.NewAttributeValueString("my-deployment-name"),
		"k8s.namespace.name":      pdata.NewAttributeValueString("my-namespace-name"),
		"k8s.service.name":        pdata.NewAttributeValueString("my-service-name"),
		"cloud.account.id":        pdata.NewAttributeValueString("my-account-id"),
		"cloud.availability_zone": pdata.NewAttributeValueString("my-zone"),
		"cloud.region":            pdata.NewAttributeValueString("my-region"),
		"abc":                     pdata.NewAttributeValueString("abc"),
		"def":                     pdata.NewAttributeValueString("def"),
		"xyz":                     pdata.NewAttributeValueString("xyz"),
		"jkl":                     pdata.NewAttributeValueString("jkl"),
		"dummy":                   pdata.NewAttributeValueString("dummy"),
	}
	attributes = pdata.NewAttributeMapFromMap(bench_pdata_attributes)
)

func BenchmarkTranslateAttributes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = translateAttributes(attributes)
	}
}

func BenchmarkTranslateAttributesInPlace(b *testing.B) {
	for i := 0; i < b.N; i++ {
		attributes := pdata.NewAttributeMapFromMap(bench_pdata_attributes)
		translateAttributesInPlace(attributes)
	}
}
