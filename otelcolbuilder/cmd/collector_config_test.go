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
	"net"
	"testing"
	"time"

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
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc
			t.Parallel()

			factories, err := components()
			require.NoError(t, err)

			t.Log("Creating new app...")
			app, err := service.New(service.CollectorSettings{
				BuildInfo: component.DefaultBuildInfo(),
				Factories: factories,
			})
			require.NoError(t, err)

			args := []string{
				"--config=" + tc.configFile,
				"--metrics-addr=" + GetAvailableLocalAddress(t),
				"--log-level=debug",
			}

			cmd := service.NewCommand(app)
			cmd.SetArgs(args)

			go func() {
				for ch := app.GetStateChannel(); ; {
					state := <-ch
					if state == service.Running {
						break
					}
				}

				t.Log("App is in the running state, calling .Shutdown()...")
				time.Sleep(time.Second)
				app.Shutdown()
			}()

			t.Logf("Calling .Run() on the app with the following args: %v", args)

			err = cmd.ExecuteContext(context.Background())
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

// GetAvailableLocalAddress finds an available local port and returns an endpoint
// describing it. The port is available for opening when this function returns
// provided that there is no race by some other code to grab the same port
// immediately.
func GetAvailableLocalAddress(t *testing.T) string {
	ln, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err, "Failed to get a free local port")
	// There is a possible race if something else takes this same port before
	// the test uses it, however, that is unlikely in practice.
	defer ln.Close()
	return ln.Addr().String()
}
