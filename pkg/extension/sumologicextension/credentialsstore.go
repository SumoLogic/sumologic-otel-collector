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
	"github.com/SumoLogic/sumologic-otel-collector/pkg/extension/sumologicextension/api"
)

// CollectorCredentials are used for storing the credentials received during
// collector registration.
type CollectorCredentials struct {
	// CollectorName indicates what name was set in the configuration when
	// registration has been made.
	CollectorName string                          `json:"collectorName"`
	Credentials   api.OpenRegisterResponsePayload `json:"collectorCredentials"`
}

// CredentialsStore is an interface to get collector authentication data
type CredentialsStore interface {
	// Check checks if collector credentials exist under the specified key.
	Check(key string) bool

	// Get returns the collector credentials stored under a specified key.
	Get(key string) (CollectorCredentials, error)

	// Store stores the provided collector credentials stored under a specified key.
	Store(key string, creds CollectorCredentials) error
}
