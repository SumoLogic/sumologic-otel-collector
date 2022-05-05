package mysqlreceiver

import (
	"encoding/csv"
	"errors"
	"os"
	"strconv"
	"time"
	"go.uber.org/zap"

)

func getStateStoreFilename(c mySQLClient) string {
	var fileextension = ".csv"
	var storeFilename = c.conf.IndexColumnName + "_" + c.conf.IndexColumnType + fileextension
	return storeFilename
}

func GetState(c mySQLClient) string {
	var storeFilename = getStateStoreFilename(c)
	var stateValue = ""

	_, err := os.Stat(storeFilename)
	if errors.Is(err, os.ErrNotExist) {
		// State File does not exists, so use start value as mentioned in YAML configuration.
		// If start value is not configured, we set some default value to it
		if c.conf.IndexColumnType == "INT" {
			var startval int = 0
			startval, err = strconv.Atoi(c.conf.InitialIndexColumnStartValue)
			if err != nil {
				//startval = 0
				stateValue = strconv.Itoa(startval)
				c.logger.Info("Problem parsing start value int, check InitialIndexColumnStartValue in collector config file. Considering default 0")
			} else {
				stateValue = strconv.Itoa(startval - 1)
			}
		} else if c.conf.IndexColumnType == "TIMESTAMP" {
			var startDate time.Time = time.Now()
			startDate, err := time.Parse("2006-01-02 15:04:05", c.conf.InitialIndexColumnStartValue)
			if err != nil {
				var startDate time.Time = time.Now()
				startDate = startDate.Add(-35 * time.Hour) // If start value is not configured, default to current-35 hours
				stateValue = startDate.String()
				c.logger.Info("Problem parsing start value date, check InitialIndexColumnStartValue in collector config file. Considering default now - 35hrs")
			} else {
				startDate = startDate.Add(-1 * time.Second)
				stateValue = startDate.String()
			}
		}
	} else {
		// State file exists.
		csvFile, err := os.Open(storeFilename)
		if err != nil {
			c.logger.Info("Error opening state file, using start value as mentioned in collector config file.")
			if c.conf.IndexColumnType == "INT" {
				var startval int = 0
				startval, err = strconv.Atoi(c.conf.InitialIndexColumnStartValue)
				if err != nil {
					c.logger.Info("Problem parsing start value int, check InitialIndexColumnStartValue in collector config file. Considering default 0")
					stateValue = strconv.Itoa(startval)
				} else {
					stateValue = strconv.Itoa(startval - 1)
				}
			} else if c.conf.IndexColumnType == "TIMESTAMP" {
				startDate, err := time.Parse("2006-01-02 15:04:05", c.conf.InitialIndexColumnStartValue)
				if err != nil {
					var startDate time.Time = time.Now()
					startDate = startDate.Add(-35 * time.Hour) // If start value is not configured, default to current-35 hours
					stateValue = startDate.String()
					c.logger.Info("Problem parsing start value date, check InitialIndexColumnStartValue in collector config file. Considering default now - 35hrs")
				} else {
					startDate = startDate.Add(-1 * time.Microsecond)
					stateValue = startDate.String()
				}
			}
		} else {
			// Able to read state file, so extract state value.
			// State is maintained in 3rd column in csv file of now
			reader := csv.NewReader(csvFile)
			records, _ := reader.ReadAll()
			stateValue = records[1][2]
		}
	}
	return stateValue
}

func SaveState(c mySQLClient, stateValue string) {
	var storeFilename = getStateStoreFilename(c)
	stateData := [][]string{
		{"indexcolumnname", "indexcolumntype", "statevalue"},
		{c.conf.IndexColumnName, c.conf.IndexColumnType, stateValue},
	}

	csvFile, err := os.Create(storeFilename)
	if err != nil {
		c.logger.Error("Failed in creating state file.", zap.Error(err))	
	}

	csvwriter := csv.NewWriter(csvFile)

	for _, empRow := range stateData {
		_ = csvwriter.Write(empRow)
	}
	csvwriter.Flush()
	csvFile.Close()
}
