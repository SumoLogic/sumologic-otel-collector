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
	"os"
	"testing"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/extension/sumologicextension/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCredentialsStoreLocalFs(t *testing.T) {
	dir, err := os.MkdirTemp("", "otelcol-sumo-credentials-store-local-fs-test-*")
	require.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})

	const key = "my_storage_key"

	creds := CollectorCredentials{
		CollectorName: "name",
		Credentials: api.OpenRegisterResponsePayload{
			CollectorCredentialId:  "credentialId",
			CollectorCredentialKey: "credentialKey",
			CollectorId:            "id",
		},
	}

	sut := localFsCredentialsStore{
		collectorCredentialsDirectory: dir,
		logger:                        zap.NewNop(),
	}

	require.NoError(t, sut.Store(key, creds))

	require.True(t, sut.Check(key))

	actual, err := sut.Get(key)
	require.NoError(t, err)
	assert.Equal(t, creds, actual)
}
