// Copyright 2021, OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package activedirectoryinvreceiver

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	adsi "github.com/go-adsi/adsi"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

type ADReceiver struct {
	config   *ADConfig
	logger   *zap.Logger
	consumer consumer.Logs
	wg       *sync.WaitGroup
	doneChan chan bool
}

func newLogsReceiver(cfg *ADConfig, logger *zap.Logger, consumer consumer.Logs) *ADReceiver {

	return &ADReceiver{
		config:   cfg,
		logger:   logger,
		consumer: consumer,
		wg:       &sync.WaitGroup{},
		doneChan: make(chan bool),
	}
}

func (l *ADReceiver) Start(ctx context.Context, _ component.Host) error {
	l.logger.Debug("starting to poll for active directory inventory records")
	l.wg.Add(1)
	go l.startPolling(ctx)
	return nil
}

func (l *ADReceiver) Shutdown(_ context.Context) error {
	l.logger.Debug("shutting down logs receiver")
	close(l.doneChan)
	l.wg.Wait()
	return nil
}

func (l *ADReceiver) startPolling(ctx context.Context) {
	defer l.wg.Done()
	t := time.NewTicker(l.config.PollInterval * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-l.doneChan:
			return
		case <-t.C:
			err := l.poll(ctx)
			if err != nil {
				l.logger.Error("there was an error during the poll", zap.Error(err))
			}
		}
	}
}

func (r *ADReceiver) poll(ctx context.Context) error {
	go func() {
		client, err := adsi.NewClient()
		if err != nil {
			r.logger.Error("Failed to create client:", zap.Error(err))
			return
		}
		ldapPath := fmt.Sprintf("LDAP://%s", r.config.DN)
		root, err := client.Open(ldapPath)
		if err != nil {
			r.logger.Error("Failed to open root object:", zap.Error(err))
			return
		}
		rootContainer, err := root.ToContainer()
		if err != nil {
			r.logger.Error("Failed to open root object:", zap.Error(err))
			return
		}
		defer rootContainer.Close()
		logs := plog.NewLogs()
		rl := logs.ResourceLogs().AppendEmpty()
		resourceLogs := &rl
		_ = resourceLogs.ScopeLogs().AppendEmpty()
		r.traverse(rootContainer, resourceLogs)
		err = r.consumer.ConsumeLogs(ctx, logs)
		if err != nil {
			r.logger.Error("Error consuming log", zap.Error(err))
		}
	}()

	return nil
}

func (l *ADReceiver) printAttrs(user *adsi.Object, resourceLogs *plog.ResourceLogs) {
	attrs := l.config.Attributes
	attributes := ""
	for _, attr := range attrs {
		values, err := user.Attr(attr)
		if err == nil && len(values) > 0 {
			attributes += fmt.Sprintf("%s: %v\n", attr, values)
		}
	}
	logRecord := resourceLogs.ScopeLogs().At(0).LogRecords().AppendEmpty()
	logRecord.Body().SetStr(attributes)
}

func (l *ADReceiver) traverse(node *adsi.Container, resourceLogs *plog.ResourceLogs) {
	nodeObject, err := node.ToObject()
	if err != nil {
		log.Printf("Error creating objects: %v\n", err)
		return
	}
	l.printAttrs(nodeObject, resourceLogs)
	children, err := node.Children()
	if err != nil {
		log.Printf("Error retrieving children: %v\n", err)
		return
	}
	for child, err := children.Next(); err == nil; child, err = children.Next() {
		childContainer, err := child.ToContainer()
		if err != nil {
			log.Println("Failed to traverse child object:", err)
			return
		}
		l.traverse(childContainer, resourceLogs)
	}
	children.Close()
}
