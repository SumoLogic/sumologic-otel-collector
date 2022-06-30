# MySQL Records Receiver

This receiver queries MySQL for database records and creates a log record (plog.Logs type) for each database record.

Supported pipeline types: logs

BasicAuth Password Encryption Use Case:

- The receiver supports password encryption for 'BasicAuth' authentication_mode
- To generate an encrypted password:
  - Specify encrypt_secret_path to a secret file containing a 24 character string
  - Include the following in the otel config:
    service:
      telemetry:
        logs:
          level: debug
    The encrypted password will only be printed in the console with a debug log level. Once generated, the user can remove the telemetry field so as to enable logging at the default info level. To use the encrypted password, the user needs to specify password_type as 'encrypted' and also the encrypt_secret_path to the same secret file.

State Management Use Case:

- The receiver supports saving the state of a query fetch into a csv file where a unique/auto-increment field is present in a table of a database.
- The unique/auto-increment field can either be of type 'NUMBER' or 'TIMESTAMP', where a 'NUMBER' should be a non-negative integer and a 'TIMESTAMP' should be of the     default timestamp storage format in mysql, i.e. '2006-01-02 15:04:05'.
- This is basically the delta mode state management feature of the receiver where the current value/state of the unique/auto-increment field is saved in a csv file which can be retreived later so as to fetch records after the saved state value.

## Prerequisites

This receiver supports MySQL version 8.0

## Configuration

```yaml
receivers:
  mysqlrecords:
    # authentication_mode is used for identifying the way of connecting to a mysql database instance
    # it has two possible values namely, 'BasicAuth' and 'IAMRDSAuth'
    # this is a mandatory field
    authentication_mode: BasicAuth

    # this is the username of the database user
    # this is a mandatory field
    username: testuser

    # this is the database name
    # this is a mandatory field
    database: testdatabase

    # this is the host name of the database instance
    # this is a mandatory field
    dbhost: testhost

    # for a RDS MySQL instance, this is the value of the region where the instance is present
    # this is a mandatory field when authentication_mode: 'IAMRDSAuth' and is not required in 'BasicAuth'.
    region: us-east-1

    # this is the path for the pem file containing certificates for different AWS regions
    # details : https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/UsingWithRDS.SSL.html
    # this is a mandatory field when authentication_mode: 'IAMRDSAuth' and is not required in 'BasicAuth'.
    aws_certificate_path: global-bundle.pem

    # this is the password of the database user
    # this will be skipped while using authentication_mode : 'IAMRDSAuth' as an authentication token is used as a password in this case
    password: testpass

    # password_type refers to how the password of the user is entered in the receiver configuration
    # it has two possible values, namely, 'plaintext' and 'encrypted'
    # the default value of password_type is 'plaintext'
    # it has to be mandatorily passed on with value 'encrypted' so as to decrypt an encrypted password with a secret string stored in file in encrypt_secret_path
    password_type: encrypted

    # this is the path to a secret file containing a 24 character string that is used for encrypting a plaintext password
    # when specified with a plaintext password, it will result in a console output with an encrypted password for the plaintext which can be used instead used as a password in the config
    # this is a mandatory path required while passing an encrypted password in the config
    encrypt_secret_path: /path/to/secret/file.txt

    # this is the database port, will be considered 3306 by default if not specified
    dbport: 3306

    # this is the structure for database queries which are required to query from a database instance
    db_queries:

      # this is a user-defined value which a user needs to put in as an identifier for each query that the user wants to run for the receiver
      # it has to be unique for each query
      # this is a mandatory field for the db_queries struct
      - queryid: Q1

        # this is the query string the user wants to run for the receiver
        # this is a mandatory field for the db_queries struct
        query: select * from persons

        # STATE MANAGEMENT Feature

        # index_column_name is the name of the unique/auto-increment field present in the table
        index_column_name: PersonID

        # this is the value for the type of the unique/auto-increment field mentioned above
        # it has two possible values namely, 'NUMBER' and 'TIMESTAMP'
        # this is mandatory feild that needs to be specified by an user trying to save the state of the index_column_name of a database query
        index_column_type: NUMBER

        # while doing state managemenent of a query, a user can explicitly define the identifier value in a table, after which the records should be fetched in
        # this is the explicitly defined identifier value for a particular database query
        # for 'NUMBER' type the default value is 0 and for 'TIMESTAMP' the default value is currentTime - 48hrs
        initial_index_column_start_value: 5

    # this is required to ensure connections are closed by the driver safely before connection is closed by MySQL server, OS, or other middlewares
    # default is 3
    setconnmaxlifetimemins: 3

    # this is highly recommended to limit the number of connection used by the application. There is no recommended limit number because it depends on application and MySQL server
    # default is 5
    setmaxopenconns: 5

    # this is recommended to be set same to setmaxopenconns, when it is smaller than setmaxopenconns, connections can be opened and closed much more frequently than you expect.
    # default is 5
    setmaxidleconns: 5

    # this indicates the number of producer and comsumer workers/threads which will used to fetch, convert and consume database records
    # by default it considers the value to be the number of queries that are to be run in the receiver
    # user can configure a maximum of 10 workers
    setmaxnodatabaseworkers: 4

    # this is the protocol value required for establishing a database connection
    # default is 'tcp'
    transport: tcp

    # default is true
    allow_native_passwords: true

    # this is the collection interval for collecting database records
    # default is 10s
    collection_interval: 10s
```

The full list of settings exposed for this receiver are documented [here](./config.go) with detailed sample configurations [here](./configExamples).
