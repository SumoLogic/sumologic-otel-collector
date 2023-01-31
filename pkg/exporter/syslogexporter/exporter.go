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
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log/syslog"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

type syslogexporter struct {
	config    *Config
	host      component.Host
	logger    *zap.Logger
	tlsConfig *tls.Config
}

func initExporter(cfg *Config, createSettings exporter.CreateSettings) (*syslogexporter, error) {
	var tlsConfig *tls.Config
	if cfg.CACertificate != "" {
		serverCert, err := ioutil.ReadFile(cfg.CACertificate)
		if err != nil {
			return nil, err
		}

		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(serverCert)
		tlsConfig = &tls.Config{
			RootCAs: pool,
		}
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

func (se *syslogexporter) logToText(record plog.LogRecord) string {
	return record.Body().AsString()
}

func (se *syslogexporter) pushLogsData(ctx context.Context, ld plog.Logs) error {
	se.logger.Info("Syslog Exporter is pushing data")
	addr := fmt.Sprintf("%s:%d", se.config.Endpoint, se.config.Port)

	s, err := Connect(se.config.Protocol, addr, syslog.LOG_ERR, "testtag", se.tlsConfig)
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
				formattedLine := se.logToText(lr)
				err = s.WriteAndRetry(syslog.LOG_INFO, formattedLine)
				if err != nil {
					//TODO: add handling of failures as it is in sumologic exporter
					return err
				}
			}
		}
	}
	return nil
}
