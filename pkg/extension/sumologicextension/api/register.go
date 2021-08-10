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

package api

type OpenRegisterRequestPayload struct {
	CollectorName string                 `json:"collectorName"`
	Ephemeral     bool                   `json:"ephemeral"`
	Description   string                 `json:"description"`
	Hostname      string                 `json:"hostname"`
	Category      string                 `json:"category"`
	TimeZone      string                 `json:"timeZone"`
	Clobber       bool                   `json:"clobber"`
	Fields        map[string]interface{} `json:"fields"`
}

type OpenRegisterResponsePayload struct {
	CollectorCredentialId  string `json:"collectorCredentialId"`
	CollectorCredentialKey string `json:"collectorCredentialKey"`
	CollectorId            string `json:"collectorId"`
	CollectorName          string `json:"collectorName"`
}
