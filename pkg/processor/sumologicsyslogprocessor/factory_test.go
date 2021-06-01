// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sumologicsyslogprocessor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configcheck"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.uber.org/zap"
)

func TestCreateDefaultConfig(t *testing.T) {
	cfg := createDefaultConfig()
	assert.NotNil(t, cfg, "failed to create default config")
	assert.NoError(t, configcheck.ValidateConfig(cfg))
}

func TestLogProcessor(t *testing.T) {
	factory := NewFactory()

	cfg := factory.CreateDefaultConfig().(*Config)
	// Manually set required fields
	cfg.FacilityAttr = "testAttrName"

	params := component.ProcessorCreateParams{Logger: zap.NewNop()}
	lp, err := factory.CreateLogsProcessor(context.Background(), params, cfg, consumertest.NewNop())
	assert.NotNil(t, lp)
	assert.NoError(t, err, "cannot create log processor")
}
