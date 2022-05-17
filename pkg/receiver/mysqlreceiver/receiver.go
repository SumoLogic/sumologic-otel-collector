package mysqlreceiver

import (
	"context"
	"fmt"
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

func (m *mySQLReceiver) produce(link chan<- string, id int, wg *sync.WaitGroup, queryChan <-chan DBQueries) {
	defer wg.Done()
	fmt.Printf("Starting producer %d\n", id)
	//check whether multiple producers can use the same sql client object
	for query := range queryChan {
		channelData, err := m.sqlclient.getRecords(&query)
		if err != nil {
			m.logger.Error("Failed to fetch records", zap.Error(err))
		} else {
			for _, msg := range channelData {
				link <- msg
			}
		}
	}
	fmt.Printf("Producer Exited %d\n", id)
}

func (m *mySQLReceiver) consume(link <-chan string, id int, wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	fmt.Printf("Starting consumer %d\n", id)
	for msg := range link {
		logs := m.convertToLog(msg)
		m.consumer.ConsumeLogs(ctx, logs)
	}
	fmt.Printf("Exiting consumer %d\n", id)
}

// var queryExecuters []*producers

// type producers struct {
// 	executerChannel chan plog.Logs
// 	quitChannel     chan bool
// 	identifier      int
// }

// func (m *mySQLReceiver) execute(jobQ chan<- string, workerPool chan *producers, allDone chan<- bool) {

// 	fmt.Println("Starting executer")
// 	channelData, err := m.sqlclient.getRecords()
// 	fmt.Println(channelData)
// 	if err != nil {
// 		m.logger.Error("Failed to fetch records", zap.Error(err))
// 	}
// 	for _, j := range channelData {
// 		jobQ <- j
// 		fmt.Println("Printing JOBQ")
// 		fmt.Println(j)
// 	}
// 	close(jobQ)
// 	for _, w := range queryExecuters {
// 		w.quitChannel <- true
// 	}
// 	close(workerPool)
// 	allDone <- true
// }

// func (m *mySQLReceiver) produce(jobQ <-chan string, p *producers, workerPool chan *producers) {

// 	fmt.Println("Starting producer")
// 	for {
// 		select {
// 		case queryResult := <-jobQ:
// 			{
// 				workerPool <- p
// 				log := m.convertToLog(queryResult)
// 				fmt.Println("Printing PLOG")
// 				fmt.Println(log)
// 				p.executerChannel <- log
// 				fmt.Println("Printing Executer channel data")
// 				fmt.Println(p.executerChannel)
// 			}
// 		case <-p.quitChannel:
// 			return
// 		}
// 	}
// }

// func (m *mySQLReceiver) consume(workerPool <-chan *producers, ctx context.Context) {

// 	fmt.Println("Starting consumer")
// 	for {
// 		worker := <-workerPool
// 		if ok := <-worker.quitChannel; ok {
// 			log := <-worker.executerChannel
// 			fmt.Println("Printing received Executer channel data")
// 			fmt.Println(log)
// 			m.consumer.ConsumeLogs(ctx, log)
// 		}
// 	}
// }

// start starts the receiver by initializing the db client connection.
func (m *mySQLReceiver) Start(ctx context.Context, host component.Host) error {
	sqlclient := newMySQLClient(m.config, m.logger)
	err := sqlclient.Connect()
	if err != nil {
		return err
	}
	fmt.Println("SQL Connected")
	m.sqlclient = sqlclient

	link := make(chan string)
	queryChan := make(chan DBQueries)
	wp := &sync.WaitGroup{}
	wc := &sync.WaitGroup{}

	fmt.Println("WG Created")

	fmt.Println("All Queries Inserted")
	wp.Add(1)
	wc.Add(1)
	maxProducers := 1
	//maxConsumers := 1

	for i := 0; i < maxProducers; i++ {
		go m.produce(link, i, wp, queryChan)
		go m.consume(link, i, wc, ctx)
	}

	// for i := 0; i < maxConsumers; i++ {
	// 	go m.consume(link, i, wc, ctx)
	// }

	for i, dbquery := range m.config.DBQueries {
		fmt.Println(i)
		queryChan <- dbquery
	}

	close(queryChan)
	wp.Wait()
	close(link)
	wc.Wait()

	// jobQ := make(chan string)
	// allDone := make(chan bool)
	// workerPool := make(chan *producers)

	// producerCount, consumerCount := 1, 1

	// for i := 0; i < producerCount; i++ {
	// 	queryExecuters = append(queryExecuters, &producers{
	// 		executerChannel: make(chan plog.Logs),
	// 		quitChannel:     make(chan bool),
	// 		identifier:      i,
	// 	})
	// 	go m.produce(jobQ, queryExecuters[i], workerPool)
	// }

	// go m.execute(jobQ, workerPool, allDone)

	// for i := 0; i < consumerCount; i++ {
	// 	go m.consume(workerPool, ctx)
	// }
	// <-allDone

	// data, err := m.sqlclient.getRecords()
	// if err != nil {
	// 	m.logger.Error("Failed to fetch records", zap.Error(err))
	// 	return err
	// }
	// for _, element := range data {
	// 	logs := m.convertToLog(element)
	// 	m.consumer.ConsumeLogs(ctx, logs)
	// }
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
