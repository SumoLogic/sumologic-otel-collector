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

package main

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
)

// TestNewOpAmpConfigProviderSettingsHasNoExpandConverter verifies that the
// OpAMP config provider does not include the expandconverter. The
// expandconverter uses os.Expand which interprets $1, $2, etc. as shell
// variables. When a config contains regex replacement patterns like "$1"
// (e.g. replace_pattern with capture groups), the expandconverter would
// fail with "environment variable "1" has invalid name" because confmap's
// escapeDollarSigns does not escape $N where N is a digit.
//
// Regression test for: SUMO-282987
func TestNewOpAmpConfigProviderSettingsHasNoExpandConverter(t *testing.T) {
	settings := NewOpAmpConfigProviderSettings("opamp:/some/path")
	assert.Empty(t, settings.ResolverSettings.ConverterFactories,
		"expandconverter must not be present: it double-processes $1 patterns causing "+
			"\"environment variable '1' has invalid name\" when using remote Source Templates "+
			"with hashing regex processing rules")
}

// TestOpAmpConfigWithDollarPatternsResolvesWithoutError verifies that an
// OpAMP-delivered config containing "$1" regex capture group references (as
// used in replace_pattern for hashing processing rules) resolves without
// error.
//
// Previously, the expandconverter in NewOpAmpConfigProviderSettings caused:
//
//	"environment variable "1" has invalid name: must match regex ..."
//
// because os.Expand (used by expandconverter) treats "$1" as a shell
// variable reference, and confmap's escapeDollarSigns does not escape
// digit-only variable names.
//
// Regression test for: SUMO-282987
func TestOpAmpConfigWithDollarPatternsResolvesWithoutError(t *testing.T) {
	opampConfigPath, err := filepath.Abs(
		filepath.Join("testdata", "opamp_replace_pattern", "opamp.yaml"),
	)
	require.NoError(t, err)

	settings := NewOpAmpConfigProviderSettings("opamp:" + opampConfigPath)

	resolver, err := confmap.NewResolver(settings.ResolverSettings)
	require.NoError(t, err)

	_, err = resolver.Resolve(context.Background())
	require.NoError(t, err,
		"config containing $1 replacement patterns must resolve without error "+
			"when using OpAMP config provider (SUMO-282987)")

	require.NoError(t, resolver.Shutdown(context.Background()))
}
