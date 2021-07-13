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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"go.uber.org/zap"
)

// localFsCredentialsStore implements CredentialsStore interface and can be used
// to store and retrieve collector credentials from local file system.
//
// Files are stored locally in collectorCredentialsDirectory.
type localFsCredentialsStore struct {
	collectorCredentialsDirectory string
	logger                        *zap.Logger
}

// Check checks if collector credentials can be found under a name being a hash
// of provided key inside collectorCredentialsDirectory.
func (cr localFsCredentialsStore) Check(key string) bool {
	filenameHash, err := hash(key)
	if err != nil {
		return false
	}
	path := path.Join(cr.collectorCredentialsDirectory, filenameHash)
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}

// Get retrieves collector credentials stored in local file system and then
// decrypts it using a hash of provided key.
func (cr localFsCredentialsStore) Get(key string) (CollectorCredentials, error) {
	filenameHash, err := hash(key)
	if err != nil {
		return CollectorCredentials{}, err
	}

	path := path.Join(cr.collectorCredentialsDirectory, filenameHash)
	creds, err := os.Open(path)
	if err != nil {
		return CollectorCredentials{}, err
	}
	defer creds.Close()

	encryptedCreds, err := ioutil.ReadAll(creds)
	if err != nil {
		return CollectorCredentials{}, err
	}

	collectorCreds, err := decrypt(encryptedCreds, key)
	if err != nil {
		return CollectorCredentials{}, err
	}

	var credentialsInfo CollectorCredentials
	if err = json.Unmarshal(collectorCreds, &credentialsInfo); err != nil {
		return CollectorCredentials{}, err
	}

	cr.logger.Info("Collector registration credentials retrieved from local fs",
		zap.String("path", path),
	)

	return credentialsInfo, nil
}

// Store stores collector credentials in a file in directory as specified
// in CollectorCredentialsDirectory.
// The credentials are encrypted using the provided key.
func (cr localFsCredentialsStore) Store(key string, creds CollectorCredentials) error {
	if err := ensureDirExists(cr.collectorCredentialsDirectory); err != nil {
		return err
	}

	filenameHash, err := hash(key)
	if err != nil {
		return err
	}
	path := path.Join(cr.collectorCredentialsDirectory, filenameHash)
	collectorCreds, err := json.Marshal(creds)
	if err != nil {
		return err
	}

	encryptedCreds, err := encrypt(collectorCreds, key)
	if err != nil {
		return err
	}

	if err = os.WriteFile(path, encryptedCreds, 0600); err != nil {
		return fmt.Errorf("failed to save credentials file '%s': %w",
			path, err,
		)
	}

	cr.logger.Info("Collector registration credentials stored locally",
		zap.String("path", path),
	)

	return nil
}

// ensureDirExists checks if the specified directory exists,
// if it doesn't then it tries to create it.
func ensureDirExists(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		if err := os.Mkdir(path, 0700); err != nil {
			return err
		}
		return nil
	}

	// If the directory doesn't have the execution bit then
	// set it so that we can 'exec' into it.
	if fi.Mode().Perm() != 0700 {
		if err := os.Chmod(path, 0700); err != nil {
			return err
		}
	}

	return nil
}
