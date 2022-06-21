package mysqlrecordsreceiver

import (
	"errors"

	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/confignet"
	"go.uber.org/multierr"
)

type Config struct {
	*config.ReceiverSettings `mapstructure:",squash"`
	AuthenticationMode       string `mapstructure:"authentication_mode,omitempty"`
	Username                 string `mapstructure:"username,omitempty"`
	Password                 string `mapstructure:"password,omitempty"`
	PasswordType             string `mapstructure:"password_type,omitempty"`
	EncryptSecretPath        string `mapstructure:"encrypt_secret_path,omitempty"`
	Database                 string `mapstructure:"database,omitempty"`
	DBHost                   string `mapstructure:"dbhost"`
	DBPort                   string `mapstructure:"dbport,omitempty"`
	Transport                string `mapstructure:"transport,omitempty"`
	AllowNativePasswords     bool   `mapstructure:"allow_native_passwords,omitempty"`
	Region                   string `mapstructure:"region,omitempty"`
	AWSCertificatePath       string `mapstructure:"aws_certificate_path,omitempty"`
	confignet.NetAddr        `mapstructure:",squash"`
	CollectionInterval       string      `mapstructure:"collection_interval,omitempty"`
	DBQueries                []DBQueries `mapstructure:"db_queries"`
	SetConnMaxLifetime       int         `mapstructure:"setconnmaxlifetimemins,omitempty"`
	SetMaxOpenConns          int         `mapstructure:"setmaxopenconns,omitempty"`
	SetMaxIdleConns          int         `mapstructure:"setmaxidleconns,omitempty"`
	SetMaxNoDatabaseWorkers  int         `mapstructure:"setmaxnodatabaseworkers,omitempty"`
}

type DBQueries struct {
	QueryId                      string `mapstructure:"queryid"`
	Query                        string `mapstructure:"query"`
	IndexColumnName              string `mapstructure:"index_column_name,omitempty"`
	InitialIndexColumnStartValue string `mapstructure:"initial_index_column_start_value,omitempty"`
	IndexColumnType              string `mapstructure:"index_column_type,omitempty"`
}

func (cfg *Config) Validate() error {

	var err error

	if cfg.AuthenticationMode != "IAMRDSAuth" && cfg.AuthenticationMode != "BasicAuth" {
		err = multierr.Append(err, errors.New("authentication_mode should be either of 'IAMRDSAuth' or 'BasicAuth'"))
	}

	if len(cfg.PasswordType) != 0 && cfg.PasswordType != "plaintext" && cfg.PasswordType != "encrypted" {
		err = multierr.Append(err, errors.New("password_type should be either of 'plaintext' or 'encrypted'"))
	}

	if cfg.PasswordType == "encrypted" {
		if len(cfg.EncryptSecretPath) == 0 {
			err = multierr.Append(err, errors.New("please specify encrypt_secret_path to read secret for decrpytion"))
		}
	}

	if cfg.AuthenticationMode == "IAMRDSAuth" && len(cfg.Region) == 0 && len(cfg.AWSCertificatePath) == 0 {
		err = multierr.Append(err, errors.New("require aws region and aws certificate path for authentication_mode : 'IAMRDSAuth'. You can download certificate from You can download it from https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/UsingWithRDS.SSL.html"))
	}

	if len(cfg.DBHost) == 0 {
		err = multierr.Append(err, errors.New("dbhost cannot be empty"))
	}

	if len(cfg.Database) == 0 {
		err = multierr.Append(err, errors.New("database cannot be empty"))
	}

	var queryIds []string
	var queryIndexColumnTypes []string
	var size = len(cfg.DBQueries)
	for i := 0; i < size; i++ {
		queryIds = append(queryIds, cfg.DBQueries[i].QueryId)
		queryIndexColumnTypes = append(queryIndexColumnTypes, cfg.DBQueries[i].IndexColumnType)
	}
	queryIdCount := make(map[string]int)
	for _, item := range queryIds {
		_, exist := queryIdCount[item]
		if exist {
			queryIdCount[item] += 1
		} else {
			queryIdCount[item] = 1
		}
	}
	for _, count := range queryIdCount {
		if count > 1 {
			err = multierr.Append(err, errors.New("multiple queries have the same queryId which is not allowed"))
		}
	}
	for _, item := range queryIndexColumnTypes {
		if len(item) != 0 {
			if item != "INT" && item != "TIMESTAMP" {
				err = multierr.Append(err, errors.New("indexcolumtype in queries can only be 'INT' or 'TIMESTAMP'"))
			}
		}
	}

	return err
}
