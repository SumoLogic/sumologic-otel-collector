package jobreceiver

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/operator/helper"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza/split"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/receiver/jobreceiver/asset"
	"github.com/SumoLogic/sumologic-otel-collector/pkg/receiver/jobreceiver/output/event"
	"github.com/SumoLogic/sumologic-otel-collector/pkg/receiver/jobreceiver/output/logentries"
)

func TestConfigValidate(t *testing.T) {
	cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
	require.NoError(t, err)

	factory := NewFactory()

	testCases := []struct {
		Name     string
		Expected func() component.Config
	}{
		{
			Name: "minimal",
			Expected: func() component.Config {
				c := factory.CreateDefaultConfig().(*Config)
				c.Schedule.Interval = time.Hour
				c.Exec.Command = "echo"
				c.Exec.Arguments = []string{"hello world"}
				return c
			},
		},
		{
			Name: "log_ntp",
			Expected: func() component.Config {
				c := factory.CreateDefaultConfig().(*Config)
				c.Schedule.Interval = time.Hour
				c.Exec.Command = "check_ntp_time"
				c.Exec.RuntimeAssets = []asset.Spec{
					{
						Name: "monitoring-plugins",
						URL:  "https://assets.bonsai.sensu.io/asset.zip",
					},
				}
				c.Exec.Arguments = []string{"-H", "time.nist.gov"}
				c.Exec.Timeout = time.Second * 8
				c.Output.InputConfig.OperatorType = "log_entries"
				c.Output.InputConfig.Attributes = map[string]helper.ExprStringConfig{"label": "foo"}
				c.Output.InputConfig.Resource = map[string]helper.ExprStringConfig{"label": "bar"}
				c.Output.Builder = &logentries.LogEntriesConfig{
					IncludeCommandName: true,
					IncludeStreamName:  true,
					MaxLogSize:         16 * 1000,
					Encoding:           "utf-8",
					Multiline: split.Config{
						LineStartPattern: "$start",
					},
				}
				return c
			},
		},
		{
			Name: "event_ntp",
			Expected: func() component.Config {
				c := factory.CreateDefaultConfig().(*Config)
				c.Schedule.Interval = time.Hour
				c.Exec.Command = "check_ntp_time"
				c.Exec.Arguments = []string{"-H", "time.nist.gov"}
				c.Exec.Timeout = time.Second * 8
				eventCfg := c.Output.Builder.(*event.EventConfig)
				require.NotNil(t, eventCfg)
				eventCfg.IncludeCommandName = false
				eventCfg.IncludeCommandStatus = false
				eventCfg.IncludeDuration = false
				eventCfg.MaxBodySize = 32 << 10
				return c
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			actual := factory.CreateDefaultConfig()
			expected := tc.Expected()
			sub, err := cm.Sub(component.NewIDWithName(Type, tc.Name).String())
			require.NoError(t, err)
			require.NoError(t, sub.Unmarshal(&actual))
			assert.Equal(t, expected, actual)
		})
	}
}
