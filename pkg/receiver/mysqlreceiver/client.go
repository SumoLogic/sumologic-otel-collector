package mysqlreceiver

import (
	"database/sql"
	"encoding/json"

	"strconv"
	"strings"

	"github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

type client interface {
	Connect() error
	getRecords() (map[string]string, error)
	Close() error
}

type mySQLClient struct {
	connStr string
	client  *sql.DB
	logger  *zap.Logger
	conf    *Config
}

var _ client = (*mySQLClient)(nil)

func newMySQLClient(conf *Config, logger *zap.Logger) client {
	driverConf := mysql.Config{
		User:                 conf.Username,
		Passwd:               conf.Password,
		Net:                  conf.Transport,
		Addr:                 conf.Endpoint,
		DBName:               conf.Database,
		AllowNativePasswords: conf.AllowNativePasswords,
	}
	connStr := driverConf.FormatDSN()

	return &mySQLClient{
		connStr: connStr,
		conf:    conf,
		logger:  logger,
	}
}

func (c *mySQLClient) Connect() error {
	clientDB, err := sql.Open("mysql", c.connStr)
	if err != nil {
		c.logger.Error("Unable to connect to database", zap.Error(err))
		return err
	}
	c.client = clientDB
	return nil
}

// getRecords queries the db for records
func (c *mySQLClient) getRecords() (map[string]string, error) {
	return Query(*c)
}

func Query(c mySQLClient) (map[string]string, error) {

	myEntireRecords := make(map[string]string)
	for _, dbquery := range c.conf.DBQueries {
		if len(strings.TrimSpace(dbquery.Query)) == 0 {
			c.logger.Error("Query is empty, check collector config file.")
			continue
		} else if len(strings.TrimSpace(dbquery.IndexColumnName)) == 0 {
			c.logger.Info("IndexColumnName missing from collector config file, so fetching all records.")
		} else if dbquery.IndexColumnType != "TIMESTAMP" && dbquery.IndexColumnType != "INT" {
			c.logger.Error("Configured non supported Indexcolummtype, supported values are TIMESTAMP or INT. Check collector configuration file.")
			continue
		} else if len(strings.TrimSpace(dbquery.IndexColumnName)) != 0 {
			if dbquery.IndexColumnType == "TIMESTAMP" {
				if strings.Contains(dbquery.Query, "where") {
					dbquery.Query += " and INDEXCOLUMNNAME > \"STATEVALUE\" order by INDEXCOLUMNNAME asc;"
				} else {
					dbquery.Query += " where INDEXCOLUMNNAME > \"STATEVALUE\" order by INDEXCOLUMNNAME asc;"
				}
			} else if dbquery.IndexColumnType == "INT" {
				if strings.Contains(dbquery.Query, "where") {
					dbquery.Query += " and INDEXCOLUMNNAME > STATEVALUE order by INDEXCOLUMNNAME asc;"
				} else {
					dbquery.Query += " where INDEXCOLUMNNAME > STATEVALUE order by INDEXCOLUMNNAME asc;"
				}
			}
			c.logger.Info("IndexColumnName specified, fetching records incrementally.")
		}
		if len(strings.TrimSpace(dbquery.IndexColumnName)) == 0 {
			queryFetchResult, _, err := ExecuteQueryandFetchRecords(c, dbquery.Query, dbquery.QueryId)
			for key, element := range queryFetchResult {
				myEntireRecords[key] = element
			}
			if err != nil {
				c.logger.Error("Error in executing query and fetching records", zap.Error(err))
				continue
			}
			if len(queryFetchResult) == 0 {
				c.logger.Info("No database records found for query with : ", zap.String("queryId", dbquery.QueryId))
			} else {
				c.logger.Info("Database records found for query with : ", zap.String("queryId", dbquery.QueryId))
			}
		} else {
			var currentState = GetState(dbquery, c.logger)
			dbquery.Query = strings.Replace(dbquery.Query, "STATEVALUE", currentState, -1)
			dbquery.Query = strings.Replace(dbquery.Query, "INDEXCOLUMNNAME", dbquery.IndexColumnName, -1)
			queryFetchResult, lastIndex, err := ExecuteQueryandFetchRecords(c, dbquery.Query, dbquery.QueryId)
			for key, element := range queryFetchResult {
				myEntireRecords[key] = element
			}
			if err != nil {
				c.logger.Error("Error in executing query and fetching records", zap.Error(err))
				continue
			}
			if len(queryFetchResult) == 0 {
				c.logger.Info("No new records found for query with : ", zap.String("queryId", dbquery.QueryId))
			} else {
				c.logger.Info("New database records found for query with : ", zap.String("queryId", dbquery.QueryId))
				lastRecordFetched := myEntireRecords[lastIndex]
				var lastRecordFetchedVal map[string]interface{}
				err := json.Unmarshal([]byte(lastRecordFetched), &lastRecordFetchedVal)
				if err != nil {
					c.logger.Error("Problem converting sql query resultset into json format.")
				}
				var lastRecordStateNumber = lastRecordFetchedVal[dbquery.IndexColumnName].(string)
				SaveState(dbquery, lastRecordStateNumber, c.logger)
			}
		}
	}
	return myEntireRecords, nil
}

func ExecuteQueryandFetchRecords(c mySQLClient, query string, queryid string) (map[string]string, string, error) {
	rows, err := c.client.Query(query)
	if err != nil {
		c.logger.Error("Error in executing sql query", zap.Error(err))
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		c.logger.Error("Error getting column names from table", zap.Error(err))
	}

	values := make([]sql.RawBytes, len(columns))

	// rows.Scan wants '[]interface{}' as an argument, so we must copy the references into such a slice
	// See http://code.google.com/p/go-wiki/wiki/InterfaceSlice for details
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	lines := make([][]string, 0)

	// now let's loop through the table lines and append them to the slice declared above
	for rows.Next() {
		// read the row on the table
		// each column value will be stored in the slice
		err = rows.Scan(scanArgs...)
		if err != nil {
			c.logger.Error("Error scanning rows from table", zap.Error(err))
		}

		var value string
		var line []string

		for _, col := range values {
			// Here we can check if the value is nil (NULL value)
			if col == nil {
				value = "NULL"
			} else {
				value = string(col)
				line = append(line, value)
			}
		}
		lines = append(lines, line)
	}
	err = rows.Err()
	if err != nil {
		c.logger.Error("Error found in rows", zap.Error(err))
	}
	myjsonobject := make(map[string]string)
	myEntireRecord := make(map[string]string)
	var lastIndex string = ""
	for j, value := range lines {
		for i, v := range value {
			myjsonobject[columns[i]] = v
		}
		jsonObjRecord, err := json.Marshal(myjsonobject)
		if err != nil {
			c.logger.Error("Error in marshalling json object", zap.Error(err))
		}
		jsonStr := string(jsonObjRecord)
		index := queryid + "_record" + strconv.Itoa(j+1)
		myEntireRecord[index] = jsonStr
		lastIndex = index
		if err != nil {
			c.logger.Error("Error in converting records into json object", zap.Error(err))
		}
	}
	return myEntireRecord, lastIndex, nil
}

func (c *mySQLClient) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}
