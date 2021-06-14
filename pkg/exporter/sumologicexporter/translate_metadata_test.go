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
	"go.opentelemetry.io/collector/consumer/pdata"
)

func TestTranslateMetadata(t *testing.T) {
	metadata := pdata.NewAttributeMap()
	metadata.InsertString("host.name", "testing-host")
	metadata.InsertString("k8s.cluster.name", "testing-cluster")
	require.Equal(t, 2, metadata.Len())

	translateMetadata(metadata)

	assert.Equal(t, 2, metadata.Len())
	assertAttribute(t, metadata, "host", "testing-host")
	assertAttribute(t, metadata, "host.name", "")
	assertAttribute(t, metadata, "cluster", "testing-cluster")
	assertAttribute(t, metadata, "k8s.cluster.name", "")
}

func TestTranslateMetadataDoesNothingWhenAttributeDoesNotExist(t *testing.T) {
	metadata := pdata.NewAttributeMap()
	require.Equal(t, 0, metadata.Len())

	translateMetadata(metadata)

	assert.Equal(t, 0, metadata.Len())
	assertAttribute(t, metadata, "host", "")
}

func TestTranslateMetadataLeavesOtherAttributesUnchanged(t *testing.T) {
	metadata := pdata.NewAttributeMap()
	metadata.InsertString("one", "one1")
	metadata.InsertString("host.name", "host1")
	metadata.InsertString("three", "three1")
	require.Equal(t, 3, metadata.Len())

	translateMetadata(metadata)

	assert.Equal(t, 3, metadata.Len())
	assertAttribute(t, metadata, "one", "one1")
	assertAttribute(t, metadata, "host", "host1")
	assertAttribute(t, metadata, "three", "three1")
}

func TestTranslateMetadataOverwritesExistingAttribute(t *testing.T) {
	metadata := pdata.NewAttributeMap()
	metadata.InsertString("host", "host1")
	metadata.InsertString("host.name", "hostname1")
	require.Equal(t, 2, metadata.Len())

	translateMetadata(metadata)

	assert.Equal(t, 1, metadata.Len())
	assertAttribute(t, metadata, "host", "hostname1")
}

func TestTranslateMetadataOverwritesMultipleExistingAttributes(t *testing.T) {
	// Note: Current implementation of pdata.AttributeMap does not allow to insert duplicate keys.
	// See https://cloud-native.slack.com/archives/C01N5UCHTEH/p1624020829067500
	metadata := pdata.NewAttributeMap()
	metadata.InsertString("host", "host1")
	metadata.InsertString("host", "host2")
	require.Equal(t, 1, metadata.Len())
	metadata.InsertString("host.name", "hostname1")
	require.Equal(t, 2, metadata.Len())

	translateMetadata(metadata)

	assert.Equal(t, 1, metadata.Len())
	assertAttribute(t, metadata, "host", "hostname1")
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
	translatedValue := translateConfigValue("%{k8s.pod.name}-%{host.name}/%{pod}-%{host}")

	assert.Equal(t, "%{pod}-%{host}/-", translatedValue)
}
