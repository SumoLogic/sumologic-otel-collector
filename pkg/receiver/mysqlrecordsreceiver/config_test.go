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
package mysqlrecordsreceiver

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidConfigforBasicAuth(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.AuthenticationMode = "BasicAuth"
	cfg.Username = "mysqluser"
	cfg.Password = "userpass"
	cfg.DBPort = "3306"
	cfg.DBHost = "localhost"
	cfg.Database = "information_schema"
	require.NoError(t, cfg.Validate())
}

func TestInValidConfigforBasicAuthWOAuthenticationMode(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.Username = "mysqluser"
	cfg.Password = "userpass"
	cfg.DBPort = "3306"
	cfg.DBHost = "localhost"
	cfg.Database = "information_schema"
	require.Error(t, cfg.Validate())
}

func TestInValidConfigforBasicAuthWODBHost(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.AuthenticationMode = "BasicAuth"
	cfg.Username = "mysqluser"
	cfg.Password = "userpass"
	cfg.DBPort = "3306"
	cfg.Database = "information_schema"
	require.Error(t, cfg.Validate())
}

func TestInValidConfigforBasicAuthWODatabase(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.AuthenticationMode = "BasicAuth"
	cfg.Username = "mysqluser"
	cfg.Password = "userpass"
	cfg.DBPort = "3306"
	require.Error(t, cfg.Validate())
}

func TestValidConfigforBasicAuthWPlaintextPass(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.AuthenticationMode = "BasicAuth"
	cfg.Username = "mysqluser"
	cfg.Password = "userpass"
	cfg.PasswordType = "plaintext"
	cfg.DBPort = "3306"
	cfg.DBHost = "localhost"
	cfg.Database = "information_schema"
	require.NoError(t, cfg.Validate())
}

func TestValidConfigforBasicAuthWPlaintextPassWEncryptionPassPath(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.AuthenticationMode = "BasicAuth"
	cfg.Username = "mysqluser"
	cfg.Password = "userpass"
	cfg.PasswordType = "plaintext"
	cfg.EncryptSecretPath = "/path/to/Secret"
	cfg.DBPort = "3306"
	cfg.DBHost = "localhost"
	cfg.Database = "information_schema"
	require.NoError(t, cfg.Validate())
}

func TestValidConfigforBasicAuthWEncryptedPass(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.AuthenticationMode = "BasicAuth"
	cfg.Username = "mysqluser"
	cfg.Password = "userpass"
	cfg.PasswordType = "encrypted"
	cfg.EncryptSecretPath = "/path/to/Secret"
	cfg.DBPort = "3306"
	cfg.DBHost = "localhost"
	cfg.Database = "information_schema"
	require.NoError(t, cfg.Validate())
}

func TestValidConfigforBasicAuthWOEncryptedPassPath(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.AuthenticationMode = "BasicAuth"
	cfg.Username = "mysqluser"
	cfg.Password = "userpass"
	cfg.PasswordType = "encrypted"
	cfg.DBPort = "3306"
	cfg.DBHost = "localhost"
	cfg.Database = "information_schema"
	require.Error(t, cfg.Validate())
}

func TestValidConfigforIAMRDSAuth(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.AuthenticationMode = "IAMRDSAuth"
	cfg.Username = "mysqlrdsuser"
	cfg.DBPort = "3306"
	cfg.DBHost = "localhost"
	cfg.AWSCertificatePath = "/path/to/AWSCertificate"
	cfg.Database = "information_schema"
	require.NoError(t, cfg.Validate())
}

func TestInValidConfigforIAMRDSAuthWOAWSCertPath(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.AuthenticationMode = "IAMRDSAuth"
	cfg.Username = "mysqlrdsuser"
	cfg.DBPort = "3306"
	cfg.DBHost = "localhost"
	cfg.Database = "information_schema"
	require.Error(t, cfg.Validate())
}

func TestValidConfigforBasicAuthWDBQueries(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.DBQueries = make([]DBQueries, 1)
	cfg.DBQueries[0].QueryId = "Q1"
	cfg.DBQueries[0].Query = "Show tables"
	cfg.AuthenticationMode = "BasicAuth"
	cfg.Username = "mysqluser"
	cfg.Password = "userpass"
	cfg.DBPort = "3306"
	cfg.DBHost = "localhost"
	cfg.Database = "information_schema"
	require.NoError(t, cfg.Validate())
}

func TestInValidConfigforBasicAuthWDBQueriesWSameQueryIDs(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.DBQueries = make([]DBQueries, 2)
	cfg.DBQueries[0].QueryId = "Q1"
	cfg.DBQueries[0].Query = "Show tables"
	cfg.DBQueries[1].QueryId = "Q1"
	cfg.DBQueries[1].Query = "Show tables"
	cfg.AuthenticationMode = "BasicAuth"
	cfg.Username = "mysqluser"
	cfg.Password = "userpass"
	cfg.DBPort = "3306"
	cfg.DBHost = "localhost"
	cfg.Database = "information_schema"
	require.Error(t, cfg.Validate())
}

func TestValidConfigforBasicAuthWDBQueriesWNUMBERIndexColumnType(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.DBQueries = make([]DBQueries, 1)
	cfg.DBQueries[0].QueryId = "Q1"
	cfg.DBQueries[0].Query = "Show tables"
	cfg.DBQueries[0].IndexColumnType = "NUMBER"
	cfg.AuthenticationMode = "BasicAuth"
	cfg.Username = "mysqluser"
	cfg.Password = "userpass"
	cfg.DBPort = "3306"
	cfg.DBHost = "localhost"
	cfg.Database = "information_schema"
	require.NoError(t, cfg.Validate())
}

func TestValidConfigforBasicAuthWDBQueriesWTIMESTAMPIndexColumnType(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.DBQueries = make([]DBQueries, 1)
	cfg.DBQueries[0].QueryId = "Q1"
	cfg.DBQueries[0].Query = "Show tables"
	cfg.DBQueries[0].IndexColumnType = "TIMESTAMP"
	cfg.AuthenticationMode = "BasicAuth"
	cfg.Username = "mysqluser"
	cfg.Password = "userpass"
	cfg.DBPort = "3306"
	cfg.DBHost = "localhost"
	cfg.Database = "information_schema"
	require.NoError(t, cfg.Validate())
}

func TestInValidConfigforBasicAuthWDBQueriesWInValidIndexColumnType(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.DBQueries = make([]DBQueries, 1)
	cfg.DBQueries[0].QueryId = "Q1"
	cfg.DBQueries[0].Query = "Show tables"
	cfg.DBQueries[0].IndexColumnType = "garbage"
	cfg.AuthenticationMode = "BasicAuth"
	cfg.Username = "mysqluser"
	cfg.Password = "userpass"
	cfg.DBPort = "3306"
	cfg.DBHost = "localhost"
	cfg.Database = "information_schema"
	require.Error(t, cfg.Validate())
}
