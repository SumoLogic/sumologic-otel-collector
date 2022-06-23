package mysqlrecordsreceiver

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestValidStateFileNameINT(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.DBQueries = make([]DBQueries, 1)
	cfg.DBQueries[0].QueryId = "Q1"
	cfg.DBQueries[0].Query = "Select * from Persons"
	cfg.DBQueries[0].IndexColumnName = "PersonID"
	cfg.DBQueries[0].IndexColumnType = "INT"
	stateFileName := getStateStoreFilename(&cfg.DBQueries[0])
	require.EqualValues(t, "Q1_PersonID_INT.csv", stateFileName)
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
	cfg.DBQueries[0].IndexColumnType = "INT"
	stateFileName := getStateStoreFilename(&cfg.DBQueries[0])
	require.NotEqualValues(t, "garbage", stateFileName)
}

func TestValidINTStateValueI(t *testing.T) {
	factory := NewFactory()
	logger := zap.NewExample()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.DBQueries = make([]DBQueries, 1)
	cfg.DBQueries[0].QueryId = "Q1"
	cfg.DBQueries[0].Query = "Select * from Persons"
	cfg.DBQueries[0].IndexColumnName = "PersonID"
	cfg.DBQueries[0].IndexColumnType = "INT"
	cfg.DBQueries[0].InitialIndexColumnStartValue = "0"
	stateValue := getStateValueINT(&cfg.DBQueries[0], logger)
	require.EqualValues(t, "0", stateValue)
}

func TestValidINTStateValueII(t *testing.T) {
	factory := NewFactory()
	logger := zap.NewExample()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.DBQueries = make([]DBQueries, 1)
	cfg.DBQueries[0].QueryId = "Q1"
	cfg.DBQueries[0].Query = "Select * from Persons"
	cfg.DBQueries[0].IndexColumnName = "PersonID"
	cfg.DBQueries[0].IndexColumnType = "INT"
	cfg.DBQueries[0].InitialIndexColumnStartValue = "1"
	stateValue := getStateValueINT(&cfg.DBQueries[0], logger)
	require.EqualValues(t, "0", stateValue)
}

func TestValidINTStateValueIII(t *testing.T) {
	factory := NewFactory()
	logger := zap.NewExample()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.DBQueries = make([]DBQueries, 1)
	cfg.DBQueries[0].QueryId = "Q1"
	cfg.DBQueries[0].Query = "Select * from Persons"
	cfg.DBQueries[0].IndexColumnName = "PersonID"
	cfg.DBQueries[0].IndexColumnType = "INT"
	cfg.DBQueries[0].InitialIndexColumnStartValue = "58762518"
	stateValue := getStateValueINT(&cfg.DBQueries[0], logger)
	require.EqualValues(t, "58762517", stateValue)
}

func TestValidINTStateValueIV(t *testing.T) {
	logger := zap.NewExample()
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.DBQueries = make([]DBQueries, 1)
	cfg.DBQueries[0].QueryId = "Q1"
	cfg.DBQueries[0].Query = "Select * from Persons"
	cfg.DBQueries[0].IndexColumnName = "PersonID"
	cfg.DBQueries[0].IndexColumnType = "INT"
	stateValue := getStateValueINT(&cfg.DBQueries[0], logger)
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

func TestValidGetStateINTwStateFilePresent(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.DBQueries = make([]DBQueries, 1)
	logger := zap.NewExample()
	cfg.DBQueries[0].QueryId = "Q1"
	cfg.DBQueries[0].Query = "Select * from Persons"
	cfg.DBQueries[0].IndexColumnName = "PersonID"
	cfg.DBQueries[0].IndexColumnType = "INT"
	cfg.DBQueries[0].InitialIndexColumnStartValue = "2"
	getStateValue := getStateValueINT(&cfg.DBQueries[0], logger)
	SaveState(&cfg.DBQueries[0], getStateValue, logger)
	require.FileExists(t, "Q1_PersonID_INT.csv")
	stateValue := GetState(&cfg.DBQueries[0], logger)
	require.EqualValues(t, "1", stateValue)
	require.NoError(t, os.Remove("Q1_PersonID_INT.csv"))
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

func TestValidSaveStateINT(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.DBQueries = make([]DBQueries, 1)
	logger := zap.NewExample()
	cfg.DBQueries[0].QueryId = "Q1"
	cfg.DBQueries[0].Query = "Select * from Persons"
	cfg.DBQueries[0].IndexColumnName = "PersonID"
	cfg.DBQueries[0].IndexColumnType = "INT"
	stateFileName := getStateStoreFilename(&cfg.DBQueries[0])
	require.EqualValues(t, "Q1_PersonID_INT.csv", stateFileName)
	stateValue := getStateValueINT(&cfg.DBQueries[0], logger)
	SaveState(&cfg.DBQueries[0], stateValue, logger)
	require.FileExists(t, "Q1_PersonID_INT.csv")
	require.NoError(t, os.Remove("Q1_PersonID_INT.csv"))
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
