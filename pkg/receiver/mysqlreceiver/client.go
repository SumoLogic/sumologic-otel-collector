package mysqlreceiver 

import (
	"database/sql"
	"encoding/json"
	"github.com/go-sql-driver/mysql"
	"go.uber.org/zap"

)

type client interface {
	Connect() error
	getRecords() (map[int]string, error)
	Close() error
}

type mySQLClient struct {
	connStr string
	client  *sql.DB
	logger    *zap.Logger
}

var _ client = (*mySQLClient)(nil)

func newMySQLClient(conf *Config) client {
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
func (c *mySQLClient) getRecords() (map[int]string,error) {
	query := "select * from Persons;"
	return Query(*c, query)
}

func Query(c mySQLClient, query string) (map[int]string,error) {
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

	columnstype, err := rows.ColumnTypes()
	if err != nil {
		c.logger.Error("Error getting column datatype from table", zap.Error(err))
	} else {
		var columndatatypes []string
		for _, colvalue := range columnstype {
			columndatatypes = append(columndatatypes, colvalue.DatabaseTypeName())
		}
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
		c.logger.Error("Error found in rows",zap.Error(err))
	}	
	myjsonobject := make(map[string]string)
	myEntireRecord := make(map[int]string)
	for j, value := range lines {
		for i, v := range value {
			myjsonobject[columns[i]] = v
		}
		jsonObjRecord, err := json.Marshal(myjsonobject)
		if err != nil {
			c.logger.Error("Error in marshalling json object",zap.Error(err))
		}
		jsonStr := string(jsonObjRecord)
		myEntireRecord[j+1] = jsonStr		
		if err != nil {
			c.logger.Error("Error in converting records into json object",zap.Error(err))
		}
	}
	return myEntireRecord,nil
}

func (c *mySQLClient) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}
