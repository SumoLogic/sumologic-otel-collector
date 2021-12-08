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
	"bytes"
	"context"
	"fmt"
	"github.com/google/uuid"
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
	yamlContentType = "text/yaml"
)

type OpAMPExtension struct {
	logger      *zap.Logger
	conf        *Config
	endpoint    string
	name        string
	version     string
	instanceid  string
	currentHash []byte
	client      client.OpAMPClient
}

func newOpAMPExtension(conf *Config, logger *zap.Logger, name string, version string) (*OpAMPExtension, error) {
	// Unfortunately, we don't seem to have access to original instance uuid so need to use another one
	instanceuuid, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	return &OpAMPExtension{
		conf:        conf,
		logger:      logger,
		name:        name,
		version:     version,
		endpoint:    conf.ServerEndpoint,
		currentHash: []byte{},
		instanceid:  instanceuuid.String(),
	}, nil
}

func (se *OpAMPExtension) Debugf(format string, v ...interface{}) {
	se.logger.Debug(fmt.Sprintf(format, v...))
}

func (se *OpAMPExtension) Errorf(format string, v ...interface{}) {
	se.logger.Error(fmt.Sprintf(format, v...))
}

func (se *OpAMPExtension) Start(ctx context.Context, host component.Host) error {
	se.logger.Info("Starting OpAMP Extension")
	settings := client.StartSettings{
		OpAMPServerURL: se.endpoint,
		InstanceUid:    se.instanceid,
		AgentType:      se.name,
		AgentVersion:   se.version,
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
	var err error

	err = se.client.SetEffectiveConfig(se.readCurrentConfig())
	if err != nil {
		return err
	}

	err = se.client.Start(settings)
	return err
}

func (se *OpAMPExtension) readCurrentConfig() *protobufs.EffectiveConfig {
	se.currentHash = se.readLastConfigHash()
	if se.currentHash == nil {
		return nil
	}

	currentConfig, err := ioutil.ReadFile(se.conf.ConfigPath)
	if err != nil {
		return nil
	}

	configMap := map[string]*protobufs.AgentConfigFile{}

	configMap[""] = &protobufs.AgentConfigFile{
		Body:        currentConfig,
		ContentType: yamlContentType,
	}

	return &protobufs.EffectiveConfig{
		Hash:      se.readLastConfigHash(),
		ConfigMap: &protobufs.AgentConfigMap{ConfigMap: configMap},
	}
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
	if config == nil {
		se.logger.Error("Empty remote config received, ignoring it")
		return nil, nil
	}

	newConfig, err := se.buildConfig(config)
	if err != nil {
		se.logger.Error("Failed building config", zap.Error(err))
		return nil, nil
	}

	if !se.validateConfig(newConfig) {
		se.logger.Error("The new config is not valid, skipping it")
		return nil, nil
	}

	if !bytes.Equal(config.ConfigHash, se.currentHash) || !se.usingConfiguredFile() {
		changed, err := se.updateConfigIfDiffers(newConfig)
		if err != nil {
			se.logger.Error("failed updating config", zap.Error(err))
			return nil, err
		}
		se.writeConfigHash(config.ConfigHash)
		if changed {
			se.logger.Info("New config detected. Going to shutdown the collector and restart it")
		}
		if !se.usingConfiguredFile() {
			se.logger.Info("Detected different config path used by collector, going to restart the process")
		}

		go func() {
			err := se.handleRestart(ctx)
			if err != nil {
				se.logger.Error("Restarting the agent failed", zap.Error(err))
			}
		}()
	}

	return nil, nil
}

func (se *OpAMPExtension) usingConfiguredFile() bool {
	// This is supposed to solve bootstraping issue, to some extent
	for i, arg := range os.Args {
		if i > 0 && os.Args[i-1] == "--config" {
			// todo: this requires better comparison
			if arg == se.conf.ConfigPath {
				return true
			}
		}
	}
	return false
}

func (se *OpAMPExtension) buildConfig(config *protobufs.AgentRemoteConfig) (string, error) {
	var efficientConfig string
	for _, v := range config.Config.ConfigMap {
		if len(efficientConfig) > 0 {
			efficientConfig += "\n"
		}

		switch v.ContentType {
		case yamlContentType:
			efficientConfig += string(v.Body)
		default:
			return "", fmt.Errorf("unsupported config content type encountered: %s", v.ContentType)
		}
	}

	// TODO: look at the se.ensureopamp flag and use it to make sure we do not remove ourselves from opamp
	// by accident

	return efficientConfig, nil
}

func (se *OpAMPExtension) validateConfig(newConfig string) bool {
	return true
}

func (se *OpAMPExtension) configHashPath() string {
	return fmt.Sprintf("%s.hash", se.conf.ConfigPath)
}

func (se *OpAMPExtension) readLastConfigHash() []byte {
	content, err := ioutil.ReadFile(se.configHashPath())
	if err != nil {
		se.logger.Warn("Could not read config hash", zap.Error(err))
		return nil
	}
	return content
}

func (se *OpAMPExtension) writeConfigHash(configHash []byte) {
	err := ioutil.WriteFile(se.configHashPath(), configHash, 0640)
	if err != nil {
		se.logger.Warn("Could not write config hash", zap.Error(err))
	}
}

func (se *OpAMPExtension) updateConfigIfDiffers(newConfig string) (bool, error) {
	content, err := ioutil.ReadFile(se.conf.ConfigPath)

	if err == nil {
		current := string(content)
		if current == newConfig {
			se.logger.Info("New config received but it does not differ from the existing one")
			return false, nil
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
		return false, fmt.Errorf("could not write new config file to %s", se.conf.ConfigPath)
	}
	return true, nil
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
