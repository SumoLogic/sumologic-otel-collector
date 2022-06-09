package mysqlreceiver

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	"github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

type client interface {
	Connect() error
	getRecords(dbquery *DBQueries) (map[string]string, error)
	Close() error
}

type mySQLClient struct {
	connStr string
	client  *sql.DB
	logger  *zap.Logger
	conf    *Config
}

var _ client = (*mySQLClient)(nil)

func createIAMRDSTLSConf(pempath string, logger *zap.Logger) tls.Config {
	rootCertPool := x509.NewCertPool()
	globalpem, err := ioutil.ReadFile(pempath)
	if err != nil {
		logger.Error("error in reading pem file", zap.Error(err))
	}
	if ok := rootCertPool.AppendCertsFromPEM(globalpem); !ok {
		logger.Error("error in loading certificates from pem file", zap.Error(err))
	}
	return tls.Config{
		RootCAs: rootCertPool,
	}
}

func generateIAMAuthToken(endpoint string, conf *Config, logger *zap.Logger) (token string) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		logger.Error("configuration error:", zap.Error(err))
	}
	fmt.Println(cfg.Credentials)
	authenticationToken, err := auth.BuildAuthToken(
		context.TODO(), endpoint, conf.Region, conf.Username, cfg.Credentials)
	if err != nil {
		logger.Error("failed to create authentication token:", zap.Error(err))
	}
	return authenticationToken
}

func newMySQLClient(conf *Config, logger *zap.Logger) client {
	var port string
	var basicauthpassword string
	var connStr string
	basicauthpassword = conf.Password
	if (len(conf.PasswordType) == 0 || conf.PasswordType == "plaintext") && len(conf.EncryptSecretPath) != 0 {
		MySecret1, err := readMySecret(conf)
		if err != nil {
			logger.Error("error in reading encryption secret from file", zap.Error(err))
		}
		encText, err := Encrypt(conf.Password, MySecret1, logger)
		if err != nil {
			logger.Error("error encrypting your classified text", zap.Error(err))
		}
		logger.Info("The plaintext password can be replaced with this encrpyted password for security purposes. Also, password_type should be entered as encrypted in config file.", zap.String("encryptedPassword", encText))
	}
	if conf.PasswordType == "encrypted" {
		MySecret2, err := readMySecret(conf)
		if err != nil {
			logger.Error("error in reading encryption secret from file", zap.Error(err))
		}
		decText, err := Decrypt(conf.Password, MySecret2, logger)
		if err != nil {
			logger.Error("error decrypting your encrypted text: ", zap.Error(err))
		}
		basicauthpassword = decText
	}
	if len(conf.DBPort) != 0 {
		port = conf.DBPort
	} else {
		logger.Info("dbport empty, considering default 3306")
		port = "3306"
	}
	endpoint := conf.DBHost + ":" + port
	if conf.AuthenticationMode == "IAMRDSAuth" {
		authenticationToken := generateIAMAuthToken(endpoint, conf, logger)
		tlsConf := createIAMRDSTLSConf(conf.AWSCertificatePath, logger)
		tlserr := mysql.RegisterTLSConfig("custom", &tlsConf)
		if tlserr != nil {
			logger.Error("Error %s when RegisterTLSConfig\n", zap.Error(tlserr))
		}
		driverConf := mysql.Config{
			User:                    conf.Username,
			Passwd:                  authenticationToken,
			Net:                     conf.Transport,
			Addr:                    endpoint,
			DBName:                  conf.Database,
			AllowNativePasswords:    conf.AllowNativePasswords,
			TLSConfig:               "custom",
			AllowCleartextPasswords: true,
		}
		connStr = driverConf.FormatDSN()
	} else {
		driverConf := mysql.Config{
			User:                 conf.Username,
			Passwd:               basicauthpassword,
			Net:                  conf.Transport,
			Addr:                 endpoint,
			DBName:               conf.Database,
			AllowNativePasswords: conf.AllowNativePasswords,
		}
		connStr = driverConf.FormatDSN()
	}
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
	if c.conf.SetConnMaxLifetime != 0 {
		clientDB.SetConnMaxLifetime(time.Minute * time.Duration(c.conf.SetConnMaxLifetime))
	} else {
		clientDB.SetConnMaxLifetime(time.Minute * 3)
	}
	if c.conf.SetConnMaxLifetime != 0 {
		clientDB.SetMaxOpenConns(c.conf.SetMaxOpenConns)
	} else {
		clientDB.SetMaxOpenConns(5)
	}
	if c.conf.SetConnMaxLifetime != 0 {
		clientDB.SetMaxIdleConns(c.conf.SetMaxIdleConns)
	} else {
		clientDB.SetMaxIdleConns(5)
	}
	c.client = clientDB
	return nil
}

// getRecords queries the db for records
func (c *mySQLClient) getRecords(dbquery *DBQueries) (map[string]string, error) {
	myEntireRecords := make(map[string]string)
	if len(strings.TrimSpace(dbquery.Query)) == 0 {
		c.logger.Error("Query is empty, check collector config file for:", zap.String("queryId", dbquery.QueryId))
		return nil, nil
	} else if len(strings.TrimSpace(dbquery.IndexColumnName)) == 0 {
		c.logger.Info("IndexColumnName missing from collector config file, so fetching all records for:", zap.String("queryId", dbquery.QueryId))
	} else if len(strings.TrimSpace(dbquery.IndexColumnName)) != 0 && len(strings.TrimSpace(dbquery.IndexColumnType)) == 0 {
		c.logger.Error("IndexColummType should be specified with a IndexColumnName for a query. Supported values are TIMESTAMP or INT.", zap.String("queryId", dbquery.QueryId))
		return nil, nil
	} else if dbquery.IndexColumnType != "TIMESTAMP" && dbquery.IndexColumnType != "INT" {
		c.logger.Error("Configured non supported Indexcolummtype, supported values are TIMESTAMP or INT. Check collector configuration file for:", zap.String("queryId", dbquery.QueryId))
		return nil, nil
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
		c.logger.Info("IndexColumnName specified, fetching records incrementally for:", zap.String("queryId", dbquery.QueryId))
	}
	if len(strings.TrimSpace(dbquery.IndexColumnName)) == 0 {
		queryFetchResult, _, err := ExecuteQueryandFetchRecords(*c, dbquery.Query, dbquery.QueryId)
		for key, element := range queryFetchResult {
			myEntireRecords[key] = element
		}
		if err != nil {
			c.logger.Error("Error in executing query and fetching records for:", zap.Error(err), zap.String("queryId", dbquery.QueryId))
			return nil, nil
		}
		if len(queryFetchResult) == 0 {
			c.logger.Info("No database records found for query with:", zap.String("queryId", dbquery.QueryId))
		} else {
			c.logger.Info("Database records found for query with:", zap.String("queryId", dbquery.QueryId))
		}
	} else {
		var currentState = GetState(dbquery, c.logger)
		dbquery.Query = strings.Replace(dbquery.Query, "STATEVALUE", currentState, -1)
		dbquery.Query = strings.Replace(dbquery.Query, "INDEXCOLUMNNAME", dbquery.IndexColumnName, -1)
		queryFetchResult, lastIndex, err := ExecuteQueryandFetchRecords(*c, dbquery.Query, dbquery.QueryId)
		for key, element := range queryFetchResult {
			myEntireRecords[key] = element
		}
		if err != nil {
			c.logger.Error("Error in executing query and fetching records", zap.Error(err), zap.String("queryId", dbquery.QueryId))
			return nil, nil
		}
		if len(queryFetchResult) == 0 {
			c.logger.Info("No new records found for query with : ", zap.String("queryId", dbquery.QueryId))
		} else {
			c.logger.Info("New database records found for query with : ", zap.String("queryId", dbquery.QueryId))
			lastRecordFetched := myEntireRecords[lastIndex]
			var lastRecordFetchedVal map[string]interface{}
			err := json.Unmarshal([]byte(lastRecordFetched), &lastRecordFetchedVal)
			if err != nil {
				c.logger.Error("Problem converting sql query resultset into json format for:", zap.String("queryId", dbquery.QueryId))
				return nil, nil
			}
			var lastRecordStateNumber = lastRecordFetchedVal[dbquery.IndexColumnName].(string)
			SaveState(dbquery, lastRecordStateNumber, c.logger)
		}
	}
	return myEntireRecords, nil
}

func ExecuteQueryandFetchRecords(c mySQLClient, query string, queryid string) (map[string]string, string, error) {
	rows, err := c.client.Query(query)
	if err != nil {
		c.logger.Error("Error in executing sql query", zap.String("error", "You have an error in your SQL syntax; check the manual that corresponds to your MySQL server version for the right syntax to use for the query."), zap.String("queryId", queryid))
		return nil, "", nil
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		c.logger.Error("Error getting column names from table", zap.String("queryId", queryid))
		return nil, "", nil
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
			c.logger.Error("Error scanning rows from table", zap.String("queryId", queryid))
			return nil, "", nil
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
		c.logger.Error("Error found in rows", zap.String("queryId", queryid))
		return nil, "", nil
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
			c.logger.Error("Error in marshalling json object", zap.String("queryId", queryid))
			return nil, "", nil
		}
		jsonStr := string(jsonObjRecord)
		index := queryid + "_record" + strconv.Itoa(j+1)
		myEntireRecord[index] = jsonStr
		lastIndex = index
		if err != nil {
			c.logger.Error("Error in converting records into json object", zap.String("queryId", queryid))
			return nil, "", nil
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
