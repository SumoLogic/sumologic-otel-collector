package mysqlreceiver 

import (
	"database/sql"
	"fmt"

	// registers the mysql driver
	"github.com/go-sql-driver/mysql"
)

type client interface {
	Connect() error
	//getGlobalStats() (map[string]string, error)
	getInnodbStats() (map[string]string, error)
	Close() error
}

type mySQLClient struct {
	connStr string
	client  *sql.DB
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
		return fmt.Errorf("unable to connect to database: %w", err)
	}
	c.client = clientDB
	return nil
}

// getInnodbStats queries the db for innodb metrics.
func (c *mySQLClient) getInnodbStats() (map[string]string, error) {
	query := "SELECT name, count FROM information_schema.innodb_metrics WHERE name LIKE '%buffer_pool_size%';"
	return Query(*c, query)
}

func Query(c mySQLClient, query string) (map[string]string, error) {
	rows, err := c.client.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	stats := map[string]string{}
	for rows.Next() {
		var key, val string
		if err := rows.Scan(&key, &val); err != nil {
			return nil, err
		}
		stats[key] = val
	}

	return stats, nil
}

func (c *mySQLClient) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}
