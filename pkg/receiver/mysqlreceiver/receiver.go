package mysqlreceiver 
 
import (
	"context"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
	"go.opentelemetry.io/collector/model/pdata"
)

//var records map[int]string

type mySQLReceiver struct {
	sqlclient client
	logger    *zap.Logger
	config    *Config
	//data	  map[int]string
	consumer consumer.Logs
}

func newMySQLReceiver (logger *zap.Logger, conf *Config, next consumer.Logs) (component.LogsReceiver, error) {

	return &mySQLReceiver{
		consumer: next,
		logger: logger,
		config: conf,
		//data: records,
	},nil	
}

func (m *mySQLReceiver) Start(ctx context.Context, host component.Host) error {
	sqlclient := newMySQLClient(m.config)

	err := sqlclient.Connect()
	if err != nil {
		return err
	}
	m.sqlclient = sqlclient

	data,err := m.sqlclient.getRecords()
	if err != nil {
		m.logger.Error("Failed to fetch records", zap.Error(err))
		return err
	}
	//records = data
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
	m.sqlclient.Close()
	return nil
}

func (m *mySQLReceiver) convertToLog(record string) pdata.Logs {
	ld := pdata.NewLogs()
	rl := ld.ResourceLogs().AppendEmpty()
	sl := rl.ScopeLogs().AppendEmpty()
	lr := sl.LogRecords().AppendEmpty()
	pdataObjectMap := pdata.NewMapFromRaw(map[string]interface{}{"record": record})
	pdataObjectMap.CopyTo(lr.Attributes())
	return ld
}
