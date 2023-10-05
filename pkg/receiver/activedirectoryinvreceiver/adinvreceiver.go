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
	"strings"
	"sync"
	"time"

	"github.com/go-ldap/ldap/v3"
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
		l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", r.config.Host, 389))
		if err != nil {
			log.Fatal(err)
		}
		defer l.Close()

		fqdn := strings.Split(r.config.DC, ".")
		domain, hld := fqdn[0], fqdn[1]

		err = l.Bind(fmt.Sprintf("cn=%s,ou=%s,dc=%s, dc=%s", r.config.CN, r.config.OU, domain, hld), r.config.Password)
		if err != nil {
			log.Fatal(err)
		}

		searchRequest := ldap.NewSearchRequest(
			"dc=exampledomain,dc=com",
			ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
			"(&(objectClass=organizationalPerson))",
			[]string{"dn", "cn", "sAMAccountName", "mail", "department", "manager", "memberOf"},
			nil,
		)

		sr, err := l.Search(searchRequest)
		if err != nil {
			log.Fatal(err)
		}

		logs := plog.NewLogs()
		rl := logs.ResourceLogs().AppendEmpty()
		resourceLogs := &rl
		resourceAttributes := resourceLogs.Resource().Attributes()
		resourceAttributes.PutStr("host.name", r.config.Host)
		_ = resourceLogs.ScopeLogs().AppendEmpty()

		for _, entry := range sr.Entries {
			if ctx.Err() != nil {
				// If the collector has been shutdown
				break
			}
			var attributes string
			attributes += fmt.Sprintf("CN: %s, ", entry.GetAttributeValue("cn"))
			attributes += fmt.Sprintf("sAMAccountName: %s, ", entry.GetAttributeValue("sAMAccountName"))
			attributes += fmt.Sprintf("mail: %s, ", entry.GetAttributeValue("mail"))
			attributes += fmt.Sprintf("department: %s, ", entry.GetAttributeValue("department"))
			attributes += fmt.Sprintf("manager: %s, ", entry.GetAttributeValue("manager"))
			attributes += fmt.Sprintf("memberOf: %v, ", entry.GetAttributeValues("memberOf"))
			fmt.Println(attributes)
			logRecord := resourceLogs.ScopeLogs().At(0).LogRecords().AppendEmpty()
			logRecord.Body().SetStr(attributes)
			err := r.consumer.ConsumeLogs(ctx, logs)
			if err != nil {
				r.logger.Error("Error consuming log", zap.Error(err))
			}
		}
	}()

	return nil
}
