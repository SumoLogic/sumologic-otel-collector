// Copyright 2021 OpenTelemetry Authors
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

// Program otelcontribcol is an extension to the OpenTelemetry Collector
// that includes additional components, some vendor-specific, contributed
// from the wider community.
package main

import (
	"context"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/service"
)

func TestBuiltCollectorWithConfigurationFiles(t *testing.T) {
	testcases := []struct {
		name       string
		configFile string
		wantErr    error
	}{
		{
			name:       "filelog with sumologicexporter without sumologicextension",
			configFile: "testdata/filelog_sumologicexporter_endpoint.yaml",
		},
		{
			name:       "telegrafreceiver with sumologicexporter without sumologicextension",
			configFile: "testdata/telegrafreceiver_sumologicexporter_endpoint.yaml",
		},
		{
			name:       "filelog reallife example with sumologicexporter without sumologicextension",
			configFile: "testdata/filelog_reallife_complicated_sumologicexporter.yaml",
		},
		{
			name:       "filterprocessor log filtering",
			configFile: "testdata/filterprocessor_logs.yaml",
		},
		{
			name:       "routing processor for traces",
			configFile: "testdata/routing_processor.yaml",
		},
		{
			name:       "metricfrequencyprocessor with telegrafreceiver and sumologicexporter",
			configFile: "testdata/metricfrequencyprocessor.yaml",
		},
		{
			name:       "filelog with sumologicexporter with persistent queue enabled",
			configFile: "testdata/filelog_sumologicexporter_with_persistent_queue_enabled.yaml",
		},
		{
			name:       "telegrafreceiver with routingprocessor",
			configFile: "testdata/telegrafreceiver_routingprocessor.yaml",
		},
		{
			name:       "resource and attributes processors with support for regexp for delete and hash actions",
			configFile: "testdata/attribute_attraction_pattern.yaml",
		},
		{
			name:       "sumologic_schema processor can be used in logs, metrics and traces pipelines",
			configFile: "testdata/sumologicschemaprocessor.yaml",
		},
		{
			name:       "multiple config files can be handled by the glob config provider",
			configFile: "glob:testdata/multiple/*.yaml",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc
			t.Parallel()

			factories, err := components()
			require.NoError(t, err)

			locations := []string{tc.configFile}
			setFlags := []string{}
			cp, err := NewConfigProvider(locations, setFlags)
			require.NoError(t, err)

			t.Log("Creating new app...")
			app, err := service.New(service.CollectorSettings{
				BuildInfo:      component.NewDefaultBuildInfo(),
				Factories:      factories,
				ConfigProvider: cp,
			})
			require.NoError(t, err)

			go func() {
				bo := backoff.NewExponentialBackOff()
				bo.InitialInterval = 25 * time.Millisecond
				bo.MaxInterval = 3 * time.Second
				bo.Multiplier = 1.2

				for {
					switch state := app.GetState(); state {
					case service.Running:
						t.Log("App is in the running state, calling .Shutdown()...")
						time.Sleep(time.Second)
						app.Shutdown()
						return

					default:
						time.Sleep(bo.NextBackOff())
						continue
					}
				}
			}()

			err = app.Run(context.Background())
			if tc.wantErr != nil {
				assert.Equal(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)

				// When adding a new testcase that would happen to include the
				// logging exporter in the pipeline use the below error assert
				// until this issue is resolved upstream.
				// https://github.com/open-telemetry/opentelemetry-collector/issues/4153
				//
				// require.Truef(t, errors.Is(err, syscall.EBADF) || err == nil,
				// 	"error expected to be nil or syscall.BADF but was: %v", err,
				// )
			}
		})
	}
}
