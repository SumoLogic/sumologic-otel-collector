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

package sumologicextension

import (
	"time"

	"go.opentelemetry.io/collector/config"
)

// Config has the configuration for the sumologic extension.
type Config struct {
	config.ExtensionSettings `mapstructure:"-"`

	// Credentials contains Access Key and Access ID for Sumo Logic service.
	// Please refer to https://help.sumologic.com/Manage/Security/Access-Keys
	// for detailed instructions how to obtain them.
	Credentials credentials `mapstructure:",squash"`

	// CollectorName is the name under which collector will be registered.
	// Please note that registering a collector under a name which is already
	// used is not allowed.
	CollectorName string `mapstructure:"collector_name"`
	// CollectorDescription is the description which will be used when the
	// collector is being registered.
	CollectorDescription string `mapstructure:"collector_description"`
	// CollectorCategory is the collecot rcategory which will be used when the
	// collector is being registered.
	CollectorCategory string `mapstructure:"collector_category"`

	ApiBaseUrl string `mapstructure:"api_base_url"`

	HeartBeatInterval time.Duration `mapstructure:"heartbeat_interval"`

	CollectorCredentialsPath string `mapstructure:"collector_credentials_path"`
}

type credentials struct {
	AccessID  string `mapstructure:"access_id"`
	AccessKey string `mapstructure:"access_key"`
}
