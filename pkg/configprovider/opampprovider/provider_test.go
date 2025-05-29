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

package opampprovider

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.opentelemetry.io/collector/confmap"
)

func TestValidateProviderScheme(t *testing.T) {
	if err := ValidateProviderScheme(NewWithSettings(confmap.ProviderSettings{})); err != nil {
		t.Error(err)
	}
}

func TestEmptyURI(t *testing.T) {
	p := NewWithSettings(confmap.ProviderSettings{})
	_, err := p.Retrieve(context.Background(), "", nil)
	if err == nil {
		t.Error("expected non-nil error")
	}
	if err := p.Shutdown(context.Background()); err != nil {
		t.Error(err)
	}
}

func TestUnsupportedScheme(t *testing.T) {
	p := NewWithSettings(confmap.ProviderSettings{})
	_, err := p.Retrieve(context.Background(), "https://foo", nil)
	if err == nil {
		t.Error("expected non-nil error")
	}
	if err := p.Shutdown(context.Background()); err != nil {
		t.Error(err)
	}
}

func TestNonExistent(t *testing.T) {
	p := NewWithSettings(confmap.ProviderSettings{})

	_, err := p.Retrieve(context.Background(), "opamp:/tmp/does/not/exist", nil)
	if err == nil {
		t.Error("expected non-nil error")
	}
	if err := p.Shutdown(context.Background()); err != nil {
		t.Error(err)
	}
}

func TestInvalidYAML(t *testing.T) {
	p := NewWithSettings(confmap.ProviderSettings{})
	_, err := p.Retrieve(context.Background(), "opamp:"+filepath.Join("testdata", "invalid-yaml.txt"), nil)
	if err == nil {
		t.Error("expected non-nil error")
	}
	if err := p.Shutdown(context.Background()); err != nil {
		t.Error(err)
	}
}

func TestMissingRemoteConfigurationDir(t *testing.T) {
	p := NewWithSettings(confmap.ProviderSettings{})
	_, err := p.Retrieve(context.Background(), "opamp:"+filepath.Join("testdata", "missing_config_dir.yaml"), nil)
	if err == nil {
		t.Error("expected non-nil error")
	}
	if err := p.Shutdown(context.Background()); err != nil {
		t.Error(err)
	}
}

func TestValid(t *testing.T) {
	p := NewWithSettings(confmap.ProviderSettings{})
	defer func() {
		if err := p.Shutdown(context.Background()); err != nil {
			t.Error(err)
		}
	}()

	configPath := "opamp:" + absolutePath(t, filepath.Join("testdata", "valid.yaml"))
	t.Logf("loading opamp config file: %s", configPath)

	ret, err := p.Retrieve(context.Background(), configPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	conf, err := ret.AsConf()
	if err != nil {
		t.Fatal(err)
	}
	got := conf.ToStringMap()
	exp := confmap.NewFromStringMap(map[string]any{
		"processors::batch/first":                           nil,
		"processors::batch/second":                          nil,
		"exporters::otlp/first::endpoint":                   "localhost:4317",
		"exporters::otlp/second::endpoint":                  "localhost:4318",
		"extensions::opamp::remote_configuration_directory": "./testdata/multiple",
		"extensions::opamp::endpoint":                       "wss://example.com/v1/opamp",
	})
	want := exp.ToStringMap()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Retrieve() mismatch (-want +got):\n%s", diff)
	}
}

func TestRemotelyManagedFlowDisabled(t *testing.T) {
	p := NewWithSettings(confmap.ProviderSettings{})
	defer func() {
		if err := p.Shutdown(context.Background()); err != nil {
			t.Error(err)
		}
	}()

	configPath := "opamp:" + absolutePath(t, filepath.Join("testdata", "config_new_merge_disabled.yaml"))
	t.Logf("loading opamp config file: %s", configPath)

	ret, err := p.Retrieve(context.Background(), configPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	conf, err := ret.AsConf()
	if err != nil {
		t.Fatal(err)
	}
	got := conf.ToStringMap()
	exp := confmap.NewFromStringMap(map[string]any{
		"extensions::sumologic::childKey":                   "value",
		"extensions::sumologic::collector_fields::cluster":  "cluster-1",
		"extensions::sumologic::collector_fields::zone":     "eu",
		"extensions::sumologic::collector_fields1::cluster": "cluster-1",
		"extensions::sumologic::collector_fields1::zone":    "eu",
		"processor": "someprocessor",
		"extensions::opamp::remote_configuration_directory": "../globprovider/testdata/mergefunc",
		"extensions::opamp::endpoint":                       "wss://example.com/v1/opamp",
		"extensions::opamp::disable_tag_replacement":        true,
	})
	want := exp.ToStringMap()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Retrieve() mismatch (-want +got):\n%s", diff)
	}
}

func TestRemotelyManagedFlowEnabled(t *testing.T) {
	p := NewWithSettings(confmap.ProviderSettings{})
	defer func() {
		if err := p.Shutdown(context.Background()); err != nil {
			t.Error(err)
		}
	}()

	configPath := "opamp:" + absolutePath(t, filepath.Join("testdata", "config_new_merge_enabled.yaml"))
	t.Logf("loading opamp config file: %s", configPath)

	ret, err := p.Retrieve(context.Background(), configPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	conf, err := ret.AsConf()
	if err != nil {
		t.Fatal(err)
	}
	got := conf.ToStringMap()
	exp := confmap.NewFromStringMap(map[string]any{
		"extensions::sumologic::childKey":                   "value",
		"extensions::sumologic::collector_fields::zone":     "eu",
		"extensions::sumologic::collector_fields1::cluster": "cluster-1",
		"extensions::sumologic::collector_fields1::zone":    "eu",
		"processor": "someprocessor",
		"extensions::opamp::remote_configuration_directory": "../globprovider/testdata/mergefunc",
		"extensions::opamp::endpoint":                       "wss://example.com/v1/opamp",
	})
	want := exp.ToStringMap()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Retrieve() mismatch (-want +got):\n%s", diff)
	}
}

func absolutePath(t *testing.T, relativePath string) string {
	t.Helper()
	pth, err := filepath.Abs(relativePath)
	if err != nil {
		t.Fatal(err)
	}
	return pth
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
