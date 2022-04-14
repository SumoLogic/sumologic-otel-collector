package mysqlreceiver 
 
import (
	"context"
	"go.opentelemetry.io/collector/consumer"
	"fmt"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
)

var records map[string]string

type mySQLReceiver struct {
	sqlclient client
	logger    *zap.Logger
	config    *Config
	data	  map[string]string
	consumer consumer.Logs
}

// func (m *mySQLReceiver) SetRecords() (map[string]string, error){
// 	innodbStats, innoErr := m.sqlclient.getInnodbStats()
// 	if innoErr != nil {
// 		m.logger.Error("Failed to fetch InnoDB stats", zap.Error(innoErr))
// 	}
// 	return innodbStats,innoErr
// }

func newMySQLReceiver (logger *zap.Logger, conf *Config, next consumer.Logs) (component.LogsReceiver, error) {

	return &mySQLReceiver{
		consumer: next,
		logger: logger,
		config: conf,
		data: records,
	},nil	
}

func (m *mySQLReceiver) Start(_ context.Context, host component.Host) error {
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
	records = data
	fmt.Println(records)
	//printed : map[Address:Beltola City:Guwahati FirstName:Jayanta LastName:Kashyap PersonID:1]
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

