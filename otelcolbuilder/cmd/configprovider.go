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
	"errors"
	"os"
	"strings"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/configprovider/globprovider"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/converter/expandconverter"
	"go.opentelemetry.io/collector/confmap/converter/overwritepropertiesconverter"
	"go.opentelemetry.io/collector/confmap/provider/envprovider"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/confmap/provider/yamlprovider"
	"go.opentelemetry.io/collector/service"
)

// This file contains modifications to the collector settings which inject a custom config provider
// Otherwise, it tries to be as close to the upstream defaults as defined here:
// https://github.com/open-telemetry/opentelemetry-collector/blob/72011ca22dff6614d518768b3bb53a1193c6ad02/service/command.go#L38

func UseCustomConfigProvider(params *service.CollectorSettings) error {

	// to create the provider, we need config locations passed in via the command line
	// to get these, we take the command the service uses to start, parse the flags, and read the values
	var err error
	tempCmd := service.NewCommand(*params)
	_, flags, err := tempCmd.Find(os.Args[1:])
	if err != nil {
		return err
	}

	err = tempCmd.ParseFlags(flags)
	if err != nil {
		// either the flags are completely invalid or don't use config at all
		// in both cases we don't need to do anything
		return nil
	}

	locations, err := tempCmd.Flags().GetStringArray("config")
	if err != nil {
		return err
	}
	if len(locations) == 0 {
		return errors.New("at least one config flag must be provided")
	}

	setFlags, err := tempCmd.Flags().GetStringArray("set")
	if err != nil {
		return err
	}

	// Not sure why this is necessary, config locations other than the first have an extra space at the start
	var cleanedLocations []string
	for _, location := range locations {
		cleanedLocations = append(cleanedLocations, strings.TrimSpace(location))
	}

	// create the config provider using the locations
	params.ConfigProvider, err = NewConfigProvider(cleanedLocations, setFlags)
	if err != nil {
		return err
	}

	return nil
}

func NewConfigProvider(locations []string, setFlags []string) (service.ConfigProvider, error) {
	settings := NewConfigProviderSettings(locations, setFlags)
	return service.NewConfigProvider(settings)
}

// see https://github.com/open-telemetry/opentelemetry-collector/blob/72011ca22dff6614d518768b3bb53a1193c6ad02/service/command.go#L38
// for the logic we're emulating here
// we only add the glob provider, everything else should be the same
func NewConfigProviderSettings(locations []string, setFlags []string) service.ConfigProviderSettings {
	return service.ConfigProviderSettings{
		Locations:    locations,
		MapProviders: makeMapProvidersMap(fileprovider.New(), envprovider.New(), yamlprovider.New(), globprovider.New()),
		MapConverters: []confmap.Converter{
			overwritepropertiesconverter.New(setFlags),
			expandconverter.New(),
		},
	}
}

func makeMapProvidersMap(providers ...confmap.Provider) map[string]confmap.Provider {
	ret := make(map[string]confmap.Provider, len(providers))
	for _, provider := range providers {
		ret[provider.Scheme()] = provider
	}
	return ret
}
