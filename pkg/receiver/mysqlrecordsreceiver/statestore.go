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
	"encoding/csv"
	"errors"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
)

func getStateStoreFilename(dbquery *DBQueries) string {
	var fileextension = ".csv"
	storeFilename := dbquery.QueryId + "_" + dbquery.IndexColumnName + "_" + dbquery.IndexColumnType + fileextension
	return storeFilename
}

func getStateValueNUMBER(dbquery *DBQueries, logger *zap.Logger) string {
	var startval int = 0
	var stateValue string
	if dbquery.InitialIndexColumnStartValue == "" {
		logger.Info("initial_index_column_start_value int not specified, considering default as 0 for:", zap.String("queryId", dbquery.QueryId))
		stateValue = strconv.Itoa(startval)
	} else if dbquery.InitialIndexColumnStartValue == "0" {
		stateValue = dbquery.InitialIndexColumnStartValue
	} else {
		startval, err := strconv.Atoi(dbquery.InitialIndexColumnStartValue)
		if err != nil {
			stateValue = strconv.Itoa(startval)
			logger.Info("Problem parsing initial_index_column_start_value int", zap.String("queryId", dbquery.QueryId))
			logger.Info("Check collector config file. Considering default 0 for:", zap.String("queryId", dbquery.QueryId))
		} else {
			stateValue = strconv.Itoa(startval - 1)
		}
	}
	return stateValue
}

func getStateValueTIMESTAMP(dbquery *DBQueries, logger *zap.Logger) string {
	var startDate time.Time = time.Now()
	var stateValue string
	if dbquery.InitialIndexColumnStartValue == "" {
		logger.Info("initial_index_column_start_value date not specified, considering default as now - 48hrs for:", zap.String("queryId", dbquery.QueryId))
		startDate = startDate.Add(-48 * time.Hour)
		stateValue = startDate.String()
	} else if dbquery.InitialIndexColumnStartValue != "" {
		startDate, err := time.Parse("2006-01-02 15:04:05", dbquery.InitialIndexColumnStartValue)
		if err != nil {
			startDate = startDate.Add(-48 * time.Hour)
			stateValue = startDate.String()
			logger.Info("Problem parsing initial_index_column_start_value date", zap.String("queryId", dbquery.QueryId))
			logger.Info("Check collector config file. Considering default now - 48hrs for:", zap.String("queryId", dbquery.QueryId))
		} else {
			startDate = startDate.Add(-1 * time.Second)
			stateValue = startDate.String()
		}
	}
	return stateValue
}

func GetState(dbquery *DBQueries, logger *zap.Logger) string {
	var storeFilename = getStateStoreFilename(dbquery)
	var stateValue = ""

	_, err := os.Stat(storeFilename)
	if errors.Is(err, os.ErrNotExist) {
		// State File does not exists, so use start value as mentioned in YAML configuration.
		// If start value is not configured, we set some default value to it
		if dbquery.IndexColumnType == "NUMBER" {
			return getStateValueNUMBER(dbquery, logger)
		} else if dbquery.IndexColumnType == "TIMESTAMP" {
			return getStateValueTIMESTAMP(dbquery, logger)
		}
	} else {
		// State file exists.
		csvFile, err := os.Open(storeFilename)
		if err != nil {
			logger.Info("Error opening state file, using start value as mentioned in collector config file.")
			if dbquery.IndexColumnType == "NUMBER" {
				return getStateValueNUMBER(dbquery, logger)
			} else if dbquery.IndexColumnType == "TIMESTAMP" {
				return getStateValueTIMESTAMP(dbquery, logger)
			}
		} else {
			// Able to read state file, so extract state value.
			// State is maintained in 4th column in csv file of now
			if dbquery.IndexColumnType == "NUMBER" {
				configFileStateValue := getStateValueNUMBER(dbquery, logger)
				reader := csv.NewReader(csvFile)
				records, err := reader.ReadAll()
				if err != nil {
					logger.Error("Failed to read stateFile", zap.Error(err))
				}
				stateFileStateValue := records[1][3]
				if configFileStateValue != "" && configFileStateValue != stateFileStateValue {
					stateValue = configFileStateValue
				} else {
					stateValue = stateFileStateValue
				}
			} else if dbquery.IndexColumnType == "TIMESTAMP" {
				configFileStateValue := getStateValueTIMESTAMP(dbquery, logger)
				reader := csv.NewReader(csvFile)
				records, err := reader.ReadAll()
				if err != nil {
					logger.Error("Failed to read stateFile", zap.Error(err))
				}
				stateFileStateValue := records[1][3]
				if configFileStateValue != "" && configFileStateValue != stateFileStateValue {
					stateValue = configFileStateValue
				} else {
					stateValue = stateFileStateValue
				}
			}
		}
	}
	return stateValue
}

func SaveState(dbquery *DBQueries, stateValue string, logger *zap.Logger) {
	var storeFilename = getStateStoreFilename(dbquery)
	stateData := [][]string{
		{"queryid", "indexcolumnname", "indexcolumntype", "statevalue"},
		{dbquery.QueryId, dbquery.IndexColumnName, dbquery.IndexColumnType, stateValue},
	}

	csvFile, err := os.Create(storeFilename)
	if err != nil {
		logger.Error("Failed in creating state file.", zap.Error(err))
	}

	csvwriter := csv.NewWriter(csvFile)

	for _, empRow := range stateData {
		err = csvwriter.Write(empRow)
		if err != nil {
			logger.Error("Failed in writing in state file.", zap.Error(err))
		}
	}
	csvwriter.Flush()
	csvFile.Close()
}
