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
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

type syslogexporter struct {
	config *Config
	host   component.Host
	logger *zap.Logger
}

func initExporter(cfg *Config, createSettings exporter.CreateSettings) (*syslogexporter, error) {

	s := &syslogexporter{
		config: cfg,
		logger: createSettings.Logger,
	}

	s.logger.Info("Syslog Exporter configured")

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
	)
}

func (s *syslogexporter) logToText(record plog.LogRecord) string {
	return record.Body().AsString()
}

func (s *syslogexporter) pushLogsData(ctx context.Context, ld plog.Logs) error {
	s.logger.Info("Syslog Exporter is pushing data")
	w, err := Dial("udp", "127.0.0.1:514", LOG_ERR, "testtag")
	if err != nil {
		return nil
	}
	defer w.Close()
	rls := ld.ResourceLogs()
	for i := 0; i < rls.Len(); i++ {
		rl := rls.At(i)
		slgs := rl.ScopeLogs()
		for i := 0; i < slgs.Len(); i++ {
			slg := slgs.At(i)
			for j := 0; j < slg.LogRecords().Len(); j++ {
				lr := slg.LogRecords().At(j)
				formattedLine := s.logToText(lr)
				w.Info(formattedLine)
			}
		}
	}
	return nil
}

func (s *syslogexporter) start(ctx context.Context, host component.Host) error {
	s.host = host
	return s.configure(ctx)
}

func (s *syslogexporter) configure(ctx context.Context) error {
	return nil
}
