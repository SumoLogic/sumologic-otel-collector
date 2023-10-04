package activedirectoryinvreceiver

import (
	"context"
	"fmt"
	"log"
	"strings"

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
}

func newLogsReceiver(cfg *ADConfig, logger *zap.Logger, consumer consumer.Logs) *ADReceiver {

	return &ADReceiver{
		config:   cfg,
		logger:   logger,
		consumer: consumer,
	}
}

func (r *ADReceiver) Start(ctx context.Context, host component.Host) error {
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

			logs := plog.NewLogs()
			rl := logs.ResourceLogs().AppendEmpty()
			resourceLogs := &rl
			_ = resourceLogs.ScopeLogs().AppendEmpty()
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

func (r *ADReceiver) Shutdown(ctx context.Context) error {
	// Implement your shutdown logic here
	return nil
}
