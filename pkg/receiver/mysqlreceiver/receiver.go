package mysqlreceiver

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

func (m *mySQLReceiver) produce(records chan<- string, id int, wg *sync.WaitGroup, queryChan <-chan DBQueries) {
	defer wg.Done()
	for query := range queryChan {
		channelData, err := m.sqlclient.getRecords(&query)
		if err != nil {
			m.logger.Error("Failed to fetch records", zap.Error(err))
		} else {
			for _, msg := range channelData {
				records <- msg
			}
		}
	}
}

func (m *mySQLReceiver) consume(records <-chan string, id int, wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	for msg := range records {
		logs := m.convertToLog(msg)
		m.consumer.ConsumeLogs(ctx, logs)
	}
}

// start starts the receiver by initializing the db client connection.
func (m *mySQLReceiver) Start(ctx context.Context, host component.Host) error {

	sqlclient := newMySQLClient(m.config, m.logger)
	err := sqlclient.Connect()
	if err != nil {
		return err
	}
	m.sqlclient = sqlclient

	records := make(chan string)
	queryChan := make(chan DBQueries)
	wp := &sync.WaitGroup{}
	wc := &sync.WaitGroup{}
	maxDBWorkers := m.config.SetMaxNoDatabaseWorkers

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
	return nil
}

// shutdown closes the db connection
func (m *mySQLReceiver) Shutdown(context.Context) error {
	defer m.sqlclient.Close()
	if m.sqlclient == nil {
		return nil
	}
	return nil
}

func (m *mySQLReceiver) convertToLog(record string) plog.Logs {

	ld := plog.NewLogs()
	rl := ld.ResourceLogs().AppendEmpty()
	sl := rl.ScopeLogs().AppendEmpty()
	lr := sl.LogRecords().AppendEmpty()
	lr.Body().SetStringVal(record)
	return ld
}
