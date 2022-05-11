package mysqlreceiver

import (
	"context"

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

// start starts the receiver by initializing the db client connection.
func (m *mySQLReceiver) Start(ctx context.Context, host component.Host) error {
	sqlclient := newMySQLClient(m.config, m.logger)
	err := sqlclient.Connect()
	if err != nil {
		return err
	}
	m.sqlclient = sqlclient

	data, err := m.sqlclient.getRecords()
	if err != nil {
		m.logger.Error("Failed to fetch records", zap.Error(err))
		return err
	}
	for _, element := range data {
		logs := m.convertToLog(element)
		m.consumer.ConsumeLogs(ctx, logs)
	}
	return nil
}

// shutdown closes the db connection
func (m *mySQLReceiver) Shutdown(context.Context) error {
	if m.sqlclient == nil {
		return nil
	}
	defer m.sqlclient.Close()
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
