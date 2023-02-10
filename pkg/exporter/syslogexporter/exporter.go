// Copyright 2023 OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package syslogexporter

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

type syslogexporter struct {
	config    *Config
	logger    *zap.Logger
	tlsConfig *tls.Config
}

func initExporter(cfg *Config, createSettings exporter.CreateSettings) (*syslogexporter, error) {
	tlsConfig, err := cfg.TLSSetting.LoadTLSConfig()
	if err != nil {
		return nil, err
	}

	s := &syslogexporter{
		config:    cfg,
		logger:    createSettings.Logger,
		tlsConfig: tlsConfig,
	}

	s.logger.Info("Syslog Exporter configured",
		zap.String("endpoint", cfg.Endpoint),
		zap.String("protocol", cfg.Protocol),
		zap.Int("port", cfg.Port),
	)

	return s, nil
}

func newLogsExporter(
	ctx context.Context,
	params exporter.CreateSettings,
	cfg *Config,
) (exporter.Logs, error) {
	s, err := initExporter(cfg, params)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize the logs exporter: %w", err)
	}

	return exporterhelper.NewLogsExporter(
		ctx,
		params,
		cfg,
		s.pushLogsData,
		exporterhelper.WithRetry(cfg.RetrySettings),
		exporterhelper.WithQueue(cfg.QueueSettings),
	)
}

func (se *syslogexporter) logsToMap(record plog.LogRecord) map[string]any {
	attributes := record.Attributes().AsRaw()
	return attributes
}

func (se *syslogexporter) getTimestamp(record plog.LogRecord) time.Time {
	timestamp := record.Timestamp().AsTime()
	return timestamp
}

func (se *syslogexporter) handleWriteErrors(ctx context.Context, errs ...error) {
	var errUnauthorized = errors.New("Unable to write to remote server")
	for _, err := range errs {
		if errors.Is(err, errUnauthorized) {
			se.logger.Warn("Unable to write to remote server")
			return
		}
	}
}

func (se *syslogexporter) pushLogsData(ctx context.Context, ld plog.Logs) error {
	type droppedResourceRecords struct {
		resource pcommon.Resource
		records  []plog.LogRecord
	}
	var (
		errs    []error
		dropped []droppedResourceRecords
	)
	s, err := Connect(se.logger, se.config, se.tlsConfig)
	if err != nil {
		return fmt.Errorf("error connecting to syslog server: %s", err)
	}
	defer s.Close()
	rls := ld.ResourceLogs()
	for i := 0; i < rls.Len(); i++ {
		rl := rls.At(i)
		slgs := rl.ScopeLogs()
		for i := 0; i < slgs.Len(); i++ {
			slg := slgs.At(i)
			for j := 0; j < slg.LogRecords().Len(); j++ {
				lr := slg.LogRecords().At(j)
				formattedLine := se.logsToMap(lr)
				timestamp := se.getTimestamp(lr)
				err = s.Write(formattedLine, timestamp)
				if err != nil {
					dropped = append(dropped, droppedResourceRecords{
						resource: rl.Resource(),
						records:  []plog.LogRecord{lr},
					})
					errs = append(errs, err)
					return err
				}
			}
		}
	}
	if len(dropped) > 0 {
		ld = plog.NewLogs()
		for i := range dropped {
			rls := ld.ResourceLogs().AppendEmpty()
			dropped[i].resource.MoveTo(rls.Resource())

			for j := 0; j < len(dropped[i].records); j++ {
				dropped[i].records[j].MoveTo(
					rls.ScopeLogs().AppendEmpty().LogRecords().AppendEmpty(),
				)
			}
		}
		errs = deduplicateErrors(errs)
		se.handleWriteErrors(ctx, errs...)
		return consumererror.NewLogs(multierr.Combine(errs...), ld)
	}
	return nil
}
