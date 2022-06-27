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
	"context"
	"sync"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

type mySQLReceiver struct {
	sqlclient client
	logger    *zap.Logger
	config    *Config
	consumer  consumer.Logs
}

func newMySQLReceiver(logger *zap.Logger, conf *Config, next consumer.Logs) (component.LogsReceiver, error) {

	return &mySQLReceiver{
		consumer: next,
		logger:   logger,
		config:   conf,
	}, nil
}

//Produce is used for fetching queries from a channel of queries, using them for extrtacting records for those queries and then pushing those records in channel of records
func (m *mySQLReceiver) produce(records chan<- string, id int, wg *sync.WaitGroup, queryChan <-chan DBQueries) {
	defer wg.Done()
	var recordcount int
	for query := range queryChan {
		channelData, err := m.sqlclient.getRecords(&query)
		if err != nil {
			m.logger.Error("Failed to fetch records", zap.Error(err))
		} else {
			for _, msg := range channelData {
				recordcount++
				records <- msg
			}
		}
	}
	m.logger.Info("Total records extracted and produced:", zap.Int("count", recordcount))
}

//Consume is used for fetching each record from the records channel, converting them into plog.Logs type with the record being passed into the body tag and then the comsumer of the LogsReceiver consuming them
func (m *mySQLReceiver) consume(records <-chan string, id int, wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	var recordcount int
	for msg := range records {
		recordcount++
		logs := m.convertToLog(msg)
		m.consumer.ConsumeLogs(ctx, logs)
	}
	m.logger.Info("Total records converted and consumed:", zap.Int("count", recordcount))
}

// start starts the receiver by initializing the db client connection.
func (m *mySQLReceiver) Start(ctx context.Context, host component.Host) error {
	sqlclient := newMySQLClient(m.config, m.logger)
	err := sqlclient.Connect()
	if err != nil {
		return err
	}
	m.logger.Info("DB Connection successful")
	m.sqlclient = sqlclient
	records := make(chan string)
	queryChan := make(chan DBQueries)
	wp := &sync.WaitGroup{}
	wc := &sync.WaitGroup{}
	maxDBWorkers := 0
	//Considering an ultimate maximum of 10 database workers
	if m.config.SetMaxNoDatabaseWorkers == 0 {
		if len(m.config.DBQueries) < 10 {
			maxDBWorkers = len(m.config.DBQueries)
		} else {
			maxDBWorkers = 10
		}
	} else {
		if (m.config.SetMaxNoDatabaseWorkers) < 10 {
			maxDBWorkers = m.config.SetMaxNoDatabaseWorkers
		} else {
			maxDBWorkers = 10
		}
	}
	wp.Add(maxDBWorkers)
	wc.Add(maxDBWorkers)
	for i := 0; i < maxDBWorkers; i++ {
		go m.produce(records, i, wp, queryChan)
		go m.consume(records, i, wc, ctx)
	}
	for _, dbquery := range m.config.DBQueries {
		queryChan <- dbquery
	}
	close(queryChan)
	wp.Wait()
	close(records)
	wc.Wait()
	m.logger.Info("Records extracted, converted to logs and consumed")
	return nil
}

//This function closes the db connection
func (m *mySQLReceiver) Shutdown(context.Context) error {
	defer m.sqlclient.Close()
	if m.sqlclient == nil {
		return nil
	}
	return nil
}

//This function generates a plog.Logs type log record for each record coming from a database query fetch
func (m *mySQLReceiver) convertToLog(record string) plog.Logs {
	ld := plog.NewLogs()
	rl := ld.ResourceLogs().AppendEmpty()
	sl := rl.ScopeLogs().AppendEmpty()
	lr := sl.LogRecords().AppendEmpty()
	lr.Body().SetStringVal(record)
	return ld
}
