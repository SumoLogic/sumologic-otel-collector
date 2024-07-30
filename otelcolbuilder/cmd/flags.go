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

// This is a copy of https://github.com/open-telemetry/opentelemetry-collector/blob/65dfc325d974be8ebb7c170b90c6646f9eaef27b/service/flags.go
// that we maintain to be able to independently parse command line flags, until either the necessary names become public,
// or there's a better way of customizing config providers

package main // import "go.opentelemetry.io/collector/service"

import (
	"flag"
	"strings"
)

const (
	configFlag = "config"
)

type configFlagValue struct {
	values []string
	sets   []string
}

func (s *configFlagValue) Set(val string) error {
	s.values = append(s.values, val)
	return nil
}

func (s *configFlagValue) String() string {
	return "[" + strings.Join(s.values, ", ") + "]"
}

// opAmpConfigFlag is a SumoLogic-specific flag for configuring the collector with OpAmp.
// It is mutually exclusive with the --config flag.
const opAmpConfigFlag = "remote-config"

// opAmpConfig houses the contents of the flag
var opAmpConfig string

func flags() *flag.FlagSet {
	flagSet := new(flag.FlagSet)

	cfgs := new(configFlagValue)
	flagSet.Var(cfgs, configFlag, "Locations to the config file(s), note that only a"+
		" single location can be set per flag entry e.g. `--config=file:/path/to/first --config=file:path/to/second`.")

	addOpampConfigFlag(flagSet)

	return flagSet
}

func getConfigFlag(flagSet *flag.FlagSet) []string {
	cfv := flagSet.Lookup(configFlag).Value.(*configFlagValue)
	return append(cfv.values, cfv.sets...)
}

func getOpampConfigFlag(flagSet *flag.FlagSet) string {
	opampPath := flagSet.Lookup(opAmpConfigFlag)
	return opampPath.Value.String()
}

type CommonFlagSet interface {
	StringVar(p *string, name string, value string, usage string)
}

func addOpampConfigFlag(flagSet CommonFlagSet) {
	flagSet.StringVar(&opAmpConfig, opAmpConfigFlag, "", "configure collector with opamp config file")
}
