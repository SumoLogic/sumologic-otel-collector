package mysqlreceiver

import (
	"encoding/csv"
	"errors"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
)

func getStateStoreFilename(dbquery DBQueries) string {
	var fileextension = ".csv"
	storeFilename := dbquery.QueryId + "_" + dbquery.IndexColumnName + "_" + dbquery.IndexColumnType + fileextension
	return storeFilename
}

func GetState(dbquery DBQueries, logger *zap.Logger) string {
	var storeFilename = getStateStoreFilename(dbquery)
	var stateValue = ""

	_, err := os.Stat(storeFilename)
	if errors.Is(err, os.ErrNotExist) {
		// State File does not exists, so use start value as mentioned in YAML configuration.
		// If start value is not configured, we set some default value to it
		if dbquery.IndexColumnType == "INT" {
			var startval int = 0
			startval, err = strconv.Atoi(dbquery.InitialIndexColumnStartValue)
			if err != nil {
				stateValue = strconv.Itoa(startval)
				logger.Info("Problem parsing start value int, check InitialIndexColumnStartValue in collector config file. Considering default 0")
			} else {
				stateValue = strconv.Itoa(startval - 1)
			}
		} else if dbquery.IndexColumnType == "TIMESTAMP" {
			var startDate time.Time = time.Now()
			startDate, err := time.Parse("2006-01-02 15:04:05", dbquery.InitialIndexColumnStartValue)
			if err != nil {
				var startDate time.Time = time.Now()
				startDate = startDate.Add(-35 * time.Hour)
				stateValue = startDate.String()
				logger.Info("Problem parsing start value date, check InitialIndexColumnStartValue in collector config file. Considering default now - 35hrs")
			} else {
				startDate = startDate.Add(-1 * time.Second)
				stateValue = startDate.String()
			}
		}
	} else {
		// State file exists.
		csvFile, err := os.Open(storeFilename)
		if err != nil {
			logger.Info("Error opening state file, using start value as mentioned in collector config file.")
			if dbquery.IndexColumnType == "INT" {
				var startval int = 0
				startval, err = strconv.Atoi(dbquery.InitialIndexColumnStartValue)
				if err != nil {
					logger.Info("Problem parsing start value int, check InitialIndexColumnStartValue in collector config file. Considering default 0")
					stateValue = strconv.Itoa(startval)
				} else {
					stateValue = strconv.Itoa(startval - 1)
				}
			} else if dbquery.IndexColumnType == "TIMESTAMP" {
				startDate, err := time.Parse("2006-01-02 15:04:05", dbquery.InitialIndexColumnStartValue)
				if err != nil {
					var startDate time.Time = time.Now()
					startDate = startDate.Add(-35 * time.Hour)
					stateValue = startDate.String()
					logger.Info("Problem parsing start value date, check InitialIndexColumnStartValue in collector config file. Considering default now - 35hrs")
				} else {
					startDate = startDate.Add(-1 * time.Microsecond)
					stateValue = startDate.String()
				}
			}
		} else {
			// Able to read state file, so extract state value.
			// State is maintained in 4th column in csv file of now
			reader := csv.NewReader(csvFile)
			records, _ := reader.ReadAll()
			stateValue = records[1][3]
		}
	}
	return stateValue
}

func SaveState(dbquery DBQueries, stateValue string, logger *zap.Logger) {
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
		_ = csvwriter.Write(empRow)
	}
	csvwriter.Flush()
	csvFile.Close()
}
