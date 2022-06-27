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
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestValidStateFileNameNUMBER(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.DBQueries = make([]DBQueries, 1)
	cfg.DBQueries[0].QueryId = "Q1"
	cfg.DBQueries[0].Query = "Select * from Persons"
	cfg.DBQueries[0].IndexColumnName = "PersonID"
	cfg.DBQueries[0].IndexColumnType = "NUMBER"
	stateFileName := getStateStoreFilename(&cfg.DBQueries[0])
	require.EqualValues(t, "Q1_PersonID_NUMBER.csv", stateFileName)
}

func TestValidStateFileNameTIMESTAMP(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.DBQueries = make([]DBQueries, 1)
	cfg.DBQueries[0].QueryId = "Q1"
	cfg.DBQueries[0].Query = "Select * from Persons"
	cfg.DBQueries[0].IndexColumnName = "DateTime"
	cfg.DBQueries[0].IndexColumnType = "TIMESTAMP"
	stateFileName := getStateStoreFilename(&cfg.DBQueries[0])
	require.EqualValues(t, "Q1_DateTime_TIMESTAMP.csv", stateFileName)
}

func TestInValidStateFileName(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.DBQueries = make([]DBQueries, 1)
	cfg.DBQueries[0].QueryId = "Q1"
	cfg.DBQueries[0].Query = "Select * from Persons"
	cfg.DBQueries[0].IndexColumnName = "PersonID"
	cfg.DBQueries[0].IndexColumnType = "NUMBER"
	stateFileName := getStateStoreFilename(&cfg.DBQueries[0])
	require.NotEqualValues(t, "garbage", stateFileName)
}

func TestValidNUMBERStateValueI(t *testing.T) {
	factory := NewFactory()
	logger := zap.NewExample()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.DBQueries = make([]DBQueries, 1)
	cfg.DBQueries[0].QueryId = "Q1"
	cfg.DBQueries[0].Query = "Select * from Persons"
	cfg.DBQueries[0].IndexColumnName = "PersonID"
	cfg.DBQueries[0].IndexColumnType = "NUMBER"
	cfg.DBQueries[0].InitialIndexColumnStartValue = "0"
	stateValue := getStateValueNUMBER(&cfg.DBQueries[0], logger)
	require.EqualValues(t, "0", stateValue)
}

func TestValidNUMBERStateValueII(t *testing.T) {
	factory := NewFactory()
	logger := zap.NewExample()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.DBQueries = make([]DBQueries, 1)
	cfg.DBQueries[0].QueryId = "Q1"
	cfg.DBQueries[0].Query = "Select * from Persons"
	cfg.DBQueries[0].IndexColumnName = "PersonID"
	cfg.DBQueries[0].IndexColumnType = "NUMBER"
	cfg.DBQueries[0].InitialIndexColumnStartValue = "1"
	stateValue := getStateValueNUMBER(&cfg.DBQueries[0], logger)
	require.EqualValues(t, "0", stateValue)
}

func TestValidNUMBERStateValueIII(t *testing.T) {
	factory := NewFactory()
	logger := zap.NewExample()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.DBQueries = make([]DBQueries, 1)
	cfg.DBQueries[0].QueryId = "Q1"
	cfg.DBQueries[0].Query = "Select * from Persons"
	cfg.DBQueries[0].IndexColumnName = "PersonID"
	cfg.DBQueries[0].IndexColumnType = "NUMBER"
	cfg.DBQueries[0].InitialIndexColumnStartValue = "58762518"
	stateValue := getStateValueNUMBER(&cfg.DBQueries[0], logger)
	require.EqualValues(t, "58762517", stateValue)
}

func TestValidNUMBERStateValueIV(t *testing.T) {
	logger := zap.NewExample()
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.DBQueries = make([]DBQueries, 1)
	cfg.DBQueries[0].QueryId = "Q1"
	cfg.DBQueries[0].Query = "Select * from Persons"
	cfg.DBQueries[0].IndexColumnName = "PersonID"
	cfg.DBQueries[0].IndexColumnType = "NUMBER"
	stateValue := getStateValueNUMBER(&cfg.DBQueries[0], logger)
	require.EqualValues(t, "0", stateValue)
}

func TestValidTIMESTAMPStateValueI(t *testing.T) {
	factory := NewFactory()
	logger := zap.NewExample()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.DBQueries = make([]DBQueries, 1)
	cfg.DBQueries[0].QueryId = "Q1"
	cfg.DBQueries[0].Query = "Select * from Persons"
	cfg.DBQueries[0].IndexColumnName = "PersonID"
	cfg.DBQueries[0].IndexColumnType = "TIMESTAMP"
	cfg.DBQueries[0].InitialIndexColumnStartValue = "2006-01-02 15:04:05"
	stateValue := getStateValueTIMESTAMP(&cfg.DBQueries[0], logger)
	require.EqualValues(t, "2006-01-02 15:04:04 +0000 UTC", stateValue)
}

func TestValidTIMESTAMPStateValueII(t *testing.T) {
	factory := NewFactory()
	logger := zap.NewExample()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.DBQueries = make([]DBQueries, 1)
	var expectedDate time.Time = time.Now()
	expectedDate = expectedDate.Add(-35 * time.Hour)
	expectedDateString := expectedDate.String()
	expectedDateString = expectedDateString[0:19]
	cfg.DBQueries[0].QueryId = "Q1"
	cfg.DBQueries[0].Query = "Select * from Persons"
	cfg.DBQueries[0].IndexColumnName = "PersonID"
	cfg.DBQueries[0].IndexColumnType = "TIMESTAMP"
	stateValue := getStateValueTIMESTAMP(&cfg.DBQueries[0], logger)
	stateValue = stateValue[0:19]
	require.EqualValues(t, expectedDateString, stateValue)
}

func TestValidGetStateNUMBERwStateFilePresent(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.DBQueries = make([]DBQueries, 1)
	logger := zap.NewExample()
	cfg.DBQueries[0].QueryId = "Q1"
	cfg.DBQueries[0].Query = "Select * from Persons"
	cfg.DBQueries[0].IndexColumnName = "PersonID"
	cfg.DBQueries[0].IndexColumnType = "NUMBER"
	cfg.DBQueries[0].InitialIndexColumnStartValue = "2"
	getStateValue := getStateValueNUMBER(&cfg.DBQueries[0], logger)
	SaveState(&cfg.DBQueries[0], getStateValue, logger)
	require.FileExists(t, "Q1_PersonID_NUMBER.csv")
	stateValue := GetState(&cfg.DBQueries[0], logger)
	require.EqualValues(t, "1", stateValue)
	require.NoError(t, os.Remove("Q1_PersonID_NUMBER.csv"))
}

func TestValidGetStateTIMESTAMPwStateFilePresent(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.DBQueries = make([]DBQueries, 1)
	logger := zap.NewExample()
	cfg.DBQueries[0].QueryId = "Q1"
	cfg.DBQueries[0].Query = "Select * from Persons"
	cfg.DBQueries[0].IndexColumnName = "DateTime"
	cfg.DBQueries[0].IndexColumnType = "TIMESTAMP"
	cfg.DBQueries[0].InitialIndexColumnStartValue = "2006-01-02 15:04:05"
	getStateValue := getStateValueTIMESTAMP(&cfg.DBQueries[0], logger)
	SaveState(&cfg.DBQueries[0], getStateValue, logger)
	require.FileExists(t, "Q1_DateTime_TIMESTAMP.csv")
	stateValue := GetState(&cfg.DBQueries[0], logger)
	require.EqualValues(t, "2006-01-02 15:04:04 +0000 UTC", stateValue)
	require.NoError(t, os.Remove("Q1_DateTime_TIMESTAMP.csv"))
}

func TestValidSaveStateNUMBER(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.DBQueries = make([]DBQueries, 1)
	logger := zap.NewExample()
	cfg.DBQueries[0].QueryId = "Q1"
	cfg.DBQueries[0].Query = "Select * from Persons"
	cfg.DBQueries[0].IndexColumnName = "PersonID"
	cfg.DBQueries[0].IndexColumnType = "NUMBER"
	stateFileName := getStateStoreFilename(&cfg.DBQueries[0])
	require.EqualValues(t, "Q1_PersonID_NUMBER.csv", stateFileName)
	stateValue := getStateValueNUMBER(&cfg.DBQueries[0], logger)
	SaveState(&cfg.DBQueries[0], stateValue, logger)
	require.FileExists(t, "Q1_PersonID_NUMBER.csv")
	require.NoError(t, os.Remove("Q1_PersonID_NUMBER.csv"))
}

func TestValidSaveStateTIMESTAMP(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.DBQueries = make([]DBQueries, 1)
	logger := zap.NewExample()
	cfg.DBQueries[0].QueryId = "Q1"
	cfg.DBQueries[0].Query = "Select * from Persons"
	cfg.DBQueries[0].IndexColumnName = "DateTime"
	cfg.DBQueries[0].IndexColumnType = "TIMESTAMP"
	stateFileName := getStateStoreFilename(&cfg.DBQueries[0])
	require.EqualValues(t, "Q1_DateTime_TIMESTAMP.csv", stateFileName)
	stateValue := getStateValueTIMESTAMP(&cfg.DBQueries[0], logger)
	SaveState(&cfg.DBQueries[0], stateValue, logger)
	require.FileExists(t, "Q1_DateTime_TIMESTAMP.csv")
	require.NoError(t, os.Remove("Q1_DateTime_TIMESTAMP.csv"))
}
