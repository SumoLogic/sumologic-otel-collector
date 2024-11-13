// Copyright The OpenTelemetry Authors
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

package globprovider_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/SumoLogic/sumologic-otel-collector/pkg/configprovider/globprovider"

	"go.opentelemetry.io/collector/confmap"
)

func TestValidateProviderScheme(t *testing.T) {
	assert.NoError(t, ValidateProviderScheme(globprovider.NewWithSettings(confmap.ProviderSettings{})))
}

func TestEmptyName(t *testing.T) {
	fp := globprovider.NewWithSettings(confmap.ProviderSettings{})
	_, err := fp.Retrieve(context.Background(), "", nil)
	require.Error(t, err)
	require.NoError(t, fp.Shutdown(context.Background()))
}

func TestUnsupportedScheme(t *testing.T) {
	fp := globprovider.NewWithSettings(confmap.ProviderSettings{})
	_, err := fp.Retrieve(context.Background(), "https://", nil)
	assert.Error(t, err)
	assert.NoError(t, fp.Shutdown(context.Background()))
}


func absolutePath(t *testing.T, relativePath string) string {
	dir, err := os.Getwd()
	require.NoError(t, err)
	return filepath.Join(dir, relativePath)
}

// TODO: Replace this with the upstream exporter version after we upgrade to v0.58.0
var schemeValidator = regexp.MustCompile("^[A-Za-z][A-Za-z0-9+.-]+$")

// ValidateProviderScheme enforces that given confmap.Provider.Scheme() object is following the restriction defined by the collector:
//   - Checks that the scheme name follows the restrictions defined https://datatracker.ietf.org/doc/html/rfc3986#section-3.1
//   - Checks that the scheme name has at leas two characters per the confmap.Provider.Scheme() comment.
func ValidateProviderScheme(p confmap.Provider) error {
	scheme := p.Scheme()
	if len(scheme) < 2 {
		return fmt.Errorf("scheme must be at least 2 characters long: %q", scheme)
	}

	if !schemeValidator.MatchString(scheme) {
		return fmt.Errorf(
			`scheme names consist of a sequence of characters beginning with a letter and followed by any combination of
			letters, digits, \"+\", \".\", or \"-\": %q`,
			scheme,
		)
	}

	return nil
}

func TestSetRemotelyManagedMergeFlow(t *testing.T) {
	fp := globprovider.NewWithSettings(confmap.ProviderSettings{}).(*globprovider.Provider)
	assert.False(t, fp.GetRemotelyManagedMergeFlow(), "Expected remotelyManagedMergeFlow to be false by default")
	fp.SetRemotelyManagedMergeFlow(true)
	assert.True(t, fp.GetRemotelyManagedMergeFlow(), "Expected remotelyManagedMergeFlow to be true after calling SetRemotelyManagedMergeFlow(true)")
}
