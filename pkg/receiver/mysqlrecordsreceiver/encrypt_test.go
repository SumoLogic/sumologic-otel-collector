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
	"go.uber.org/zap"
)

const MySecret string = "abc&1*~#^2^#s0^=)^^7%b34"

func TestValidReadMySecret(t *testing.T) {
	var cfg Config
	cfg.EncryptSecretPath = "encrypt_test.go"
	_, err := readMySecret(&cfg)
	require.NoError(t, err)
}

func TestInValidReadMySecret(t *testing.T) {
	var cfg Config
	cfg.EncryptSecretPath = ""
	_, err := readMySecret(&cfg)
	require.Error(t, err)
}

func TestValidEncrytFunc(t *testing.T) {
	StringToEncrypt := "Encrypting this string"
	logger := zap.NewExample()
	encText, _ := Encrypt(StringToEncrypt, MySecret, logger)
	require.EqualValues(t, "Li5E8RFcV/EPZY/neyCXQYjrfa/atA==", encText)
}

func TestInValidEncrytFunc(t *testing.T) {
	StringToEncrypt := "Encrypting this string"
	logger := zap.NewExample()
	encText, _ := Encrypt(StringToEncrypt, MySecret, logger)
	require.NotEqualValues(t, "Li5E8RFcV/EPZY/neyCXQYjrfa/atA", encText)
}

func TestValidDecrytFunc(t *testing.T) {
	StringToDecryot := "Li5E8RFcV/EPZY/neyCXQYjrfa/atA=="
	logger := zap.NewExample()
	decText, _ := Decrypt(StringToDecryot, MySecret, logger)
	require.EqualValues(t, "Encrypting this string", decText)
}

func TestInValidDecrytFunc(t *testing.T) {
	StringToDecryot := "Li5E8RFcV/EPZY/neyCXQYjrfa/atA=="
	logger := zap.NewExample()
	decText, _ := Decrypt(StringToDecryot, MySecret, logger)
	require.NotEqualValues(t, "Encrypting this string!", decText)
}
