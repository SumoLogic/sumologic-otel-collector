// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package opampextension

import (
	"context"
	"fmt"
	"github.com/open-telemetry/opamp-go/client"
	"github.com/open-telemetry/opamp-go/protobufs"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"
	"time"
)

const (
	agentType    = "otelcol"
	agentVersion = "0.0.0"
)

type OpAMPExtension struct {
	logger   *zap.Logger
	conf     *Config
	endpoint string
	client   client.OpAMPClient
}

func newOpAMPExtension(conf *Config, logger *zap.Logger) (*OpAMPExtension, error) {
	return &OpAMPExtension{
		conf:     conf,
		logger:   logger,
		endpoint: conf.ServerEndpoint,
	}, nil
}

func (se *OpAMPExtension) Debugf(format string, v ...interface{}) {
	se.logger.Debug(fmt.Sprintf(format, v...))
}

func (se *OpAMPExtension) Errorf(format string, v ...interface{}) {
	se.logger.Error(fmt.Sprintf(format, v...))
}

func (se *OpAMPExtension) instanceUid() string {
	return "12345"
}

func (se *OpAMPExtension) Start(ctx context.Context, host component.Host) error {
	se.logger.Info("Starting OpAMP Extension")
	settings := client.StartSettings{
		OpAMPServerURL: se.endpoint,
		InstanceUid:    se.instanceUid(),
		AgentType:      agentType,
		AgentVersion:   agentVersion,
		Callbacks: client.CallbacksStruct{
			OnRemoteConfigFunc: se.onAgentRemoteConfig,
			OnConnectFunc: func() {
				se.logger.Debug("Connected")
			},
			OnConnectFailedFunc: func(err error) {
				se.logger.Error("Failed connecting to OpAMP", zap.Error(err))
			},
		},
	}

	se.client = client.New(se)
	err := se.client.Start(settings)
	return err
}

// Shutdown is invoked during service shutdown.
func (se *OpAMPExtension) Shutdown(ctx context.Context) error {
	return se.client.Stop(ctx)
}

func (se *OpAMPExtension) onAgentRemoteConfig(ctx context.Context, config *protobufs.AgentRemoteConfig) (*protobufs.EffectiveConfig, error) {
	//for k, v := range config.Config.ConfigMap {
	//	println("Received config key: ", k)
	//	println("Received content-type: ", v.ContentType)
	//	println("Received body: ", string(v.Body))
	//}
	newConfig := se.buildConfig(config)
	if !se.validateConfig(newConfig) {
		se.logger.Error("The new config is not valid, skipping it")
		return nil, nil
	}

	if se.updateConfigIfDiffers(newConfig) {
		se.logger.Info("New config detected. Going to shutdown the collector and restart it")
		go func() {
			err := se.handleRestart(ctx)
			if err != nil {
				se.logger.Error("Restarting the agent failed", zap.Error(err))
			}
		}()
	}

	return nil, nil
}

func (se *OpAMPExtension) buildConfig(config *protobufs.AgentRemoteConfig) string {
	var efficientConfig string
	for _, v := range config.Config.ConfigMap {
		if len(efficientConfig) > 0 {
			efficientConfig += "\n"
		}
		// NOTE: this always assumes the content-type is string. How should we
		// determine the appropriate type? content-type=string?
		efficientConfig += string(v.Body)
	}

	// TODO: look at the se.ensureopamp flag and use it to make sure we do not remove ourselves from opamp
	// by accident

	return efficientConfig
}

func (se *OpAMPExtension) validateConfig(newConfig string) bool {
	return true
}

func (se *OpAMPExtension) updateConfigIfDiffers(newConfig string) bool {
	content, err := ioutil.ReadFile(se.conf.ConfigPath)

	if err == nil {
		current := string(content)
		if current == newConfig {
			se.logger.Info("New config received but it does not differ from the existing one")
			return false
		}
	} else {
		// TODO: check for various types of errors, but so far just assume the file could not be read
		se.logger.Warn("Could not read existing config", zap.Error(err))
	}

	backupPath := se.conf.ConfigPath + ".old"
	err = ioutil.WriteFile(backupPath, content, 0640)
	if err != nil {
		se.logger.Warn("Could not write backup config", zap.String("backupPath", backupPath))
	}
	err = ioutil.WriteFile(se.conf.ConfigPath, []byte(newConfig), 0640)
	if err != nil {
		se.logger.Error("Could not write new config file", zap.String("path", se.conf.ConfigPath))
		return false
	}
	return true
}

func (se *OpAMPExtension) handleRestart(ctx context.Context) error {
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		return err
	}
	// NOTE: we could use wd and absolute paths instead
	//wd, _ := os.Getwd()

	newArgs := make([]string, len(os.Args))
	for i, arg := range os.Args {
		if i > 0 && os.Args[i-1] == "--config" {
			// Update the config path with the new location
			newArgs[i] = se.conf.ConfigPath
		} else {
			newArgs[i] = arg
		}
	}

	argv0, err := lookPath()
	if err != nil {
		return err
	}

	se.logger.Info("Restarting the process since new configuration needs to be applied")
	// TODO: some more graceful Stop() or so?
	se.client.Stop(ctx)
	time.Sleep(3 * time.Second)

	se.logger.Info("Running new instance")
	err = syscall.Exec(argv0, newArgs, os.Environ())
	if err != nil {
		return err
	}
	err = p.Signal(syscall.SIGTERM)
	if err != nil {
		return err
	}

	return nil
}

func lookPath() (argv0 string, err error) {
	argv0, err = exec.LookPath(os.Args[0])
	if nil != err {
		return
	}
	if _, err = os.Stat(argv0); nil != err {
		return
	}
	return
}

func (se *OpAMPExtension) ComponentID() config.ComponentID {
	return se.conf.ExtensionSettings.ID()
}
