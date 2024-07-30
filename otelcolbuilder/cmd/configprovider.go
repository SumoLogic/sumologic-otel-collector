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
	"slices"

	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/converter/expandconverter"
	"go.opentelemetry.io/collector/confmap/provider/envprovider"
	"go.opentelemetry.io/collector/featuregate"
	"go.opentelemetry.io/collector/otelcol"

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
	// we need to check if any config locations have been set alongside the opamp config
	flagset := flags()

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
	opAmpPath := getOpampConfigFlag(flagset)

	if len(locations) > 0 && opAmpPath != "" {
		return fmt.Errorf("cannot use --%s and --%s flags together", configFlag, opAmpConfigFlag)
	}

	if opAmpPath == "" {
		// let the default config provider handle things, we don't need to do anything
		return nil
	}

	// opamp path is set, use a custom provider with only opamp
	// remove the opamp flags from os.Args, so flag parsing doesn't throw an error later
	opampFlagIndex := slices.Index(os.Args, fmt.Sprintf("--%s", opAmpConfigFlag))
	os.Args = slices.Delete(os.Args, opampFlagIndex, opampFlagIndex+2)
	params.ConfigProviderSettings = NewOpAmpConfigProviderSettings(opAmpPath)

	return nil
}

// NewOpAmpConfigProviderSettings is like NewConfigProviderSettings, but only
// the OpAmp config provider is configured.
func NewOpAmpConfigProviderSettings(location string) otelcol.ConfigProviderSettings {
	return otelcol.ConfigProviderSettings{
		ResolverSettings: confmap.ResolverSettings{
			URIs:               []string{location},
			ProviderFactories:  []confmap.ProviderFactory{opampprovider.NewFactory(), envprovider.NewFactory()},
			ConverterFactories: []confmap.ConverterFactory{expandconverter.NewFactory()},
		},
	}
}
