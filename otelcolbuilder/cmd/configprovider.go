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
	"flag"
	"fmt"
	"io"
	"os"

	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/converter/expandconverter"
	"go.opentelemetry.io/collector/confmap/provider/envprovider"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/confmap/provider/httpprovider"
	"go.opentelemetry.io/collector/confmap/provider/httpsprovider"
	"go.opentelemetry.io/collector/confmap/provider/yamlprovider"
	"go.opentelemetry.io/collector/featuregate"
	"go.opentelemetry.io/collector/otelcol"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/configprovider/globprovider"
	"github.com/SumoLogic/sumologic-otel-collector/pkg/configprovider/opampprovider"
)

// This file contains modifications to the collector settings which inject a custom config provider
// Otherwise, it tries to be as close to the upstream defaults as defined here:
// https://github.com/open-telemetry/opentelemetry-collector/blob/65dfc325d974be8ebb7c170b90c6646f9eaef27b/service/command.go#L38

func UseCustomConfigProvider(params *otelcol.CollectorSettings) error {
	// feature flags, which are enabled by default in our distro
	err := featuregate.GlobalRegistry().Set("extension.sumologic.updateCollectorMetadata", true)

	if err != nil {
		return fmt.Errorf("setting feature gate flags failed: %s", err)
	}
	// to create the provider, we need config locations passed in via the command line
	// to get these, we take the command the service uses to start, parse the flags, and read the values
	flagset := flags(featuregate.GlobalRegistry())

	// drop the output from the flagset, we only want to parse
	// by default it prints error messages to stdout :(
	flagset.Init("", flag.ContinueOnError)
	flagset.SetOutput(io.Discard)

	// actually parse the flags and get the config locations
	err = flagset.Parse(os.Args[1:])
	if err != nil {
		// if we fail parsing, we just let the actual command logic deal with it
		// here we only care about config locations
		return nil
	}

	locations := getConfigFlag(flagset)
	opAmpPath := flagset.Lookup(opAmpConfigFlag)

	if len(locations) > 0 && opAmpPath.Value.String() != "" {
		return fmt.Errorf("cannot use --%s and --%s flags together", configFlag, opAmpConfigFlag)
	}

	if len(locations) == 0 && opAmpPath.Value.String() == "" {
		// if no locations, use defaults
		// either this is a command, or the default provider will throw an error
		return nil
	}

	// create the config provider using the locations
	if len(locations) > 0 {
		params.ConfigProviderSettings = NewConfigProviderSettings(locations)
	} else {
		params.ConfigProviderSettings = NewOpAmpConfigProviderSettings(opAmpPath.Value.String())
	}

	return nil
}

func NewConfigProvider(locations []string) (otelcol.ConfigProvider, error) {
	settings := NewConfigProviderSettings(locations)
	return otelcol.NewConfigProvider(settings)
}

// see https://github.com/open-telemetry/opentelemetry-collector/blob/72011ca22dff6614d518768b3bb53a1193c6ad02/service/command.go#L38
// for the logic we're emulating here
// we only add the glob provider, everything else should be the same
func NewConfigProviderSettings(locations []string) otelcol.ConfigProviderSettings {
	return otelcol.ConfigProviderSettings{
		ResolverSettings: confmap.ResolverSettings{
			URIs: locations,
			ProviderFactories: []confmap.ProviderFactory{
				fileprovider.NewFactory(),
				envprovider.NewFactory(),
				yamlprovider.NewFactory(),
				httpprovider.NewFactory(),
				httpsprovider.NewFactory(),
				globprovider.NewFactory(),
			},
			ConverterFactories: []confmap.ConverterFactory{expandconverter.NewFactory()},
		},
	}
}

// NewOpAmpConfigProviderSettings is like NewConfigProviderSettings, but only
// the OpAmp config provider is configured.
func NewOpAmpConfigProviderSettings(location string) otelcol.ConfigProviderSettings {
	return otelcol.ConfigProviderSettings{
		ResolverSettings: confmap.ResolverSettings{
			URIs:               []string{location},
			ProviderFactories:  []confmap.ProviderFactory{opampprovider.NewFactory()},
			ConverterFactories: []confmap.ConverterFactory{expandconverter.NewFactory()},
		},
	}
}
