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
	"os"
	"path/filepath"
	"testing"

	"github.com/oklog/ulid/v2"
	"github.com/open-telemetry/opamp-go/protobufs"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/configauth"
	"go.opentelemetry.io/collector/config/configoptional"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/extension/extensiontest"
	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
)

const (
	errMsgRemoteConfigNotAccepted = "OpAMP agent does not accept remote configuration"
	errMsgInvalidConfigName       = "cannot validate config: " +
		"service::pipelines::logs/localfilesource/0aa79379-c764-4d3d-9d66-03f6df029a07: " +
		"references processor \"batch\" which is not configured"
	errMsgInvalidType              = "'spike_limit_percentage' expected type 'uint32'"
	errExpectedUncofiguredEndPoint = "expected unconfigured opamp endpoint to result in default sumo opamp url setting"
	errMsgInvalidCloudwatchConfig  = "source data must be an array or slice, got int"
)

func defaultSetup() (*Config, extension.Settings) {
	cfg := createDefaultConfig().(*Config)
	set := extensiontest.NewNopSettings(extensiontest.NopType)
	set.BuildInfo = component.BuildInfo{Version: "test version", Command: "otelcoltest"}
	return cfg, set
}

func setupWithRemoteConfig(t *testing.T, d string) (*Config, extension.Settings) {
	cfg, set := defaultSetup()
	cfg.RemoteConfigurationDirectory = d
	return cfg, set
}

func TestApplyRemoteConfig(t *testing.T) {
	tests := []struct {
		name         string
		file         string
		expectError  bool
		errorMessage string
	}{
		{"ApplyRemoteConfig", "testdata/opamp.d/opamp-remote-config.yaml", false, ""},
		{"ApplyRemoteApacheConfig", "testdata/opamp.d/opamp-apache-config.yaml", false, ""},
		{"ApplyRemoteHostConfig", "testdata/opamp.d/opamp-host-config.yaml", false, ""},
		{"ApplyRemoteExtensionsConfig", "testdata/opamp.d/opamp-extensions-config.yaml", false, ""},
		{"ApplyRemoteConfigFailed", "testdata/opamp.d/opamp-invalid-remote-config.yaml", true, errMsgInvalidType},
		{"ApplyRemoteConfigMissingProcessor", "testdata/opamp.d/opamp-missing-processor.yaml", true, errMsgInvalidConfigName},
		{"ApplyFilterProcessorConfig", "testdata/opamp.d/opamp-filter-processor.yaml", false, ""},
		{"ApplyKafkaMetricsConfig", "testdata/opamp.d/opamp-kafkametrics-config.yaml", false, ""},
		{"ApplyElasticsearchConfig", "testdata/opamp.d/opamp-elastic-config.yaml", false, ""},
		{"ApplyMysqlConfig", "testdata/opamp.d/opamp-mysql-config.yaml", false, ""},
		{"ApplyattributesprocessorConfig", "testdata/opamp.d/opamp-attributes-processor.yaml", false, ""},
		{"ApplyPostgresqlConfig", "testdata/opamp.d/opamp-postgresql-config.yaml", false, ""},
		{"ApplyRabbitmqConfig", "testdata/opamp.d/opamp-rabbitmq-config.yaml", false, ""},
		{"ApplyRedisConfig", "testdata/opamp.d/opamp-redis-config.yaml", false, ""},
		{"ApplyFirehoseConfig", "testdata/opamp.d/opamp-aws-firehose-config.yaml", false, ""},
		{"ApplyCloudwatchConfig", "testdata/opamp.d/opamp-aws-cloudwatch-receiver-config.yaml", false, ""},
		{"ApplyContainerInsightConfig", "testdata/opamp.d/opamp-aws-container-insight-config.yaml", false, ""},
		{"ApplyEcsContainerMetricsConfig", "testdata/opamp.d/opamp-aws-container-metrics-config.yaml", false, ""},
		{"ApplyS3Config", "testdata/opamp.d/opamp-aws-s3-exporter-config.yaml", false, ""},
		{"ApplyCloudwatchConfigFailure", "testdata/opamp.d/opamp-aws-cloudwatch-receiver-error-config.yaml", true, errMsgInvalidCloudwatchConfig},
		{"ApplyXrayConfig", "testdata/opamp.d/opamp-aws-xray-config.yaml", false, ""},
		{"ApplyKenesisConfig", "testdata/opamp.d/opamp-aws-kenesis-config.yaml", false, ""},
		{"ApplyDebugExporterConfig", "testdata/opamp.d/opamp-debug-exporter-config.yaml", false, ""},
		{"ApplyFileExporterConfig", "testdata/opamp.d/opamp-file-exporter-config.yaml", false, ""},
		{"ApplyKafkaExporterConfig", "testdata/opamp.d/opamp-kafka-exporter-config.yaml", false, ""},
		{"ApplyLoadbalancingExporterConfig", "testdata/opamp.d/opamp-loadbalancing-exporter-config.yaml", false, ""},
		{"ApplyCumulativetodeltaProcessorConfig", "testdata/opamp.d/opamp-cumulativetodelta-processor-config.yaml", false, ""},
		{"ApplyDeltatorateProcessorConfig", "testdata/opamp.d/opamp-deltatorate-processor-config.yaml", false, ""},
		{"ApplyMetricsgenerationProcessorConfig", "testdata/opamp.d/opamp-metricsgeneration-processor-config.yaml", false, ""},
		{"ApplyGroupbyattrsProcessorConfig", "testdata/opamp.d/opamp-groupbyattrs-processor-config.yaml", false, ""},
		{"ApplyGroupbytraceProcessorConfig", "testdata/opamp.d/opamp-groupbytrace-processor-config.yaml", false, ""},
		{"ApplyK8sattributesProcessorConfig", "testdata/opamp.d/opamp-k8sattributes-processor-config.yaml", false, ""},
		{"ApplyLogdedupProcessorConfig", "testdata/opamp.d/opamp-logdedup-processor-config.yaml", false, ""},
		{"ApplyLogstransformProcessorConfig", "testdata/opamp.d/opamp-logstransform-processor-config.yaml", false, ""},
		{"ApplyMetricstransformProcessorConfig", "testdata/opamp.d/opamp-metricstransform-processor-config.yaml", false, ""},
		{"ApplyProbabilisticsamplerProcessorConfig", "testdata/opamp.d/opamp-probabilisticsampler-processor-config.yaml", false, ""},
		{"ApplyAerospikeConfig", "testdata/opamp.d/opamp-aerospike-receiver-config.yaml", false, ""},
		{"ApplyAzureEventHubConfig", "testdata/opamp.d/opamp-azureeventhub-receiver-config.yaml", false, ""},
		{"ApplyBigipConfig", "testdata/opamp.d/opamp-bigip-receiver-config.yaml", false, ""},
		{"ApplyCarbonReceiverConfig", "testdata/opamp.d/opamp-carbon-receiver-config.yaml", false, ""},
		{"ApplyCloudlfareReceiverConfig", "testdata/opamp.d/opamp-cloudflare-receiver-config.yaml", false, ""},
		{"ApplyPrometheusExporterConfig", "testdata/opamp.d/opamp-prometheus-exporter-config.yaml", false, ""},
		{"ApplyOtlphttpConfig", "testdata/opamp.d/opamp-otlphttp-exporter-config.yaml", false, ""},
		{"ApplyECSObserverConfig", "testdata/opamp.d/opamp-ecsobserver-config.yaml", false, ""},
		{"ApplyRedactionProcessorConfig", "testdata/opamp.d/opamp-redaction-processor-config.yaml", false, ""},
		{"ApplyRemotetapProcessorConfig", "testdata/opamp.d/opamp-remotetap-processor-config.yaml", false, ""},
		{"ApplyGeoipProcessorConfig", "testdata/opamp.d/opamp-geoip-processor-config.yaml", false, ""},
		{"ApplySchemaProcessorConfig", "testdata/opamp.d/opamp-schema-processor-config.yaml", false, ""},
		{"ApplySpanProcessorConfig", "testdata/opamp.d/opamp-span-processor-config.yaml", false, ""},
		{"ApplyTailsamplingProcessorConfig", "testdata/opamp.d/opamp-tailsampling-processor-config.yaml", false, ""},
		{"ApplyCloudfoundryReceiverConfig", "testdata/opamp.d/opamp-cloudfoundry-receiver-config.yaml", false, ""},
		{"ApplyHttpcheckReceiverConfig", "testdata/opamp.d/opamp-httpcheck-receiver-config.yaml", false, ""},
		{"ApplyAsapAuthExtensionConfig", "testdata/opamp.d/opamp-asapauth-extension-config.yaml", false, ""},
		{"ApplyBasicAuthExtensionConfig", "testdata/opamp.d/opamp-basicauth-extension-config.yaml", false, ""},
		{"ApplyBearerTokenAuthExtensionConfig", "testdata/opamp.d/opamp-bearertokenauth-extension-config.yaml", false, ""},
		{"ApplyDbStorageExtensionConfig", "testdata/opamp.d/opamp-dbstorage-extension-config.yaml", false, ""},
		{"ApplyDockerObserverExtensionConfig", "testdata/opamp.d/opamp-dockerobserver-extension-config.yaml", false, ""},
		{"ApplyHeadersSetterExtensionConfig", "testdata/opamp.d/opamp-headerssetter-extension-config.yaml", false, ""},
		{"ApplyHostObserverExtensionConfig", "testdata/opamp.d/opamp-hostobserver-extension-config.yaml", false, ""},
		{"ApplyHttpForwarderExtensionConfig", "testdata/opamp.d/opamp-httpforwarder-extension-config.yaml", false, ""},
		{"ApplyJaegerRemoteSamplingExtensionConfig", "testdata/opamp.d/opamp-jaegerremotesampling-extension-config.yaml", false, ""},
		{"ApplyK8sObserverExtensionConfig", "testdata/opamp.d/opamp-k8sobserver-extension-config.yaml", false, ""},
		{"ApplyOauth2ClientauthExtensionConfig", "testdata/opamp.d/opamp-oauth2clientauth-extension-config.yaml", false, ""},
		{"ApplyOidcAuthExtensionConfig", "testdata/opamp.d/opamp-oidcauth-extension-config.yaml", false, ""},
		{"ApplyPprofExtensionConfig", "testdata/opamp.d/opamp-pprof-extension-config.yaml", false, ""},
		{"ApplyZpagesExtensionConfig", "testdata/opamp.d/opamp-zpages-extension-config.yaml", false, ""},
		{"ApplyInfluxdbReceiverConfig", "testdata/opamp.d/opamp-influxdb-receiver-config.yaml", false, ""},
		{"ApplyJaegerReceiverConfig", "testdata/opamp.d/opamp-jaeger-receiver-config.yaml", false, ""},
		{"ApplyJournaldReceiverConfig", "testdata/opamp.d/opamp-journald-receiver-config.yaml", false, ""},
		{"ApplyK8sClusterReceiverConfig", "testdata/opamp.d/opamp-k8scluster-receiver-config.yaml", false, ""},
		{"ApplyK8sEventsReceiverConfig", "testdata/opamp.d/opamp-k8sevents-receiver-config.yaml", false, ""},
		{"ApplyK8sObjectsReceiverConfig", "testdata/opamp.d/opamp-k8sobjects-receiver-config.yaml", false, ""},
		{"ApplyKubeletStatsReceiverConfig", "testdata/opamp.d/opamp-kubeletstats-receiver-config.yaml", false, ""},
		{"ApplyLokiReceiverConfig", "testdata/opamp.d/opamp-loki-receiver-config.yaml", false, ""},
		{"ApplyMemcachedReceiverConfig", "testdata/opamp.d/opamp-memcached-receiver-config.yaml", false, ""},
		{"ApplyMongodbReceiverConfig", "testdata/opamp.d/opamp-mongodb-receiver-config.yaml", false, ""},
		{"ApplyMongodbAtlasReceiverConfig", "testdata/opamp.d/opamp-mongodbatlas-receiver-config.yaml", false, ""},
		{"ApplyNsxtReceiverConfig", "testdata/opamp.d/opamp-nsxt-receiver-config.yaml", false, ""},
		{"ApplyOracledbReceiverConfig", "testdata/opamp.d/opamp-oracledb-receiver-config.yaml", false, ""},
		{"ApplyOtlpJsonFileReceiverConfig", "testdata/opamp.d/opamp-otlpjsonfile-receiver-config.yaml", false, ""},
		{"ApplyPodmanReceiverConfig", "testdata/opamp.d/opamp-podman-receiver-config.yaml", false, ""},
		{"ApplySimplePrometheusReceiverConfig", "testdata/opamp.d/opamp-simpleprometheus-receiver-config.yaml", false, ""},
		{"ApplyPrometheusReceiverConfig", "testdata/opamp.d/opamp-prometheus-receiver-config.yaml", false, ""},
		{"ApplyPulsarReceiverConfig", "testdata/opamp.d/opamp-pulsar-receiver-config.yaml", false, ""},
		{"ApplyPurefaReceiverConfig", "testdata/opamp.d/opamp-purefa-receiver-config.yaml", false, ""},
		{"ApplyPurefbReceiverConfig", "testdata/opamp.d/opamp-purefb-receiver-config.yaml", false, ""},
		{"ApplyReceiverCreatorConfig", "testdata/opamp.d/opamp-receiver-creator-config.yaml", false, ""},
		{"ApplyRiakReceiverConfig", "testdata/opamp.d/opamp-riak-receiver-config.yaml", false, ""},
		{"ApplySaphanaReceiverConfig", "testdata/opamp.d/opamp-saphana-receiver-config.yaml", false, ""},
		{"ApplySignalFxReceiverConfig", "testdata/opamp.d/opamp-signalfx-receiver-config.yaml", false, ""},
		{"ApplySkyWalkingReceiverConfig", "testdata/opamp.d/opamp-skywalking-receiver-config.yaml", false, ""},
		{"ApplySnowflakeReceiverConfig", "testdata/opamp.d/opamp-snowflake-receiver-config.yaml", false, ""},
		{"ApplySolaceReceiverConfig", "testdata/opamp.d/opamp-solace-receiver-config.yaml", false, ""},
		{"ApplySplunkhecReceiverConfig", "testdata/opamp.d/opamp-splunkhec-receiver-config.yaml", false, ""},
		{"ApplySqlQueryReceiverConfig", "testdata/opamp.d/opamp-sqlquery-receiver-config.yaml", false, ""},
		{"ApplySqlServerReceiverConfig", "testdata/opamp.d/opamp-sqlserver-receiver-config.yaml", false, ""},
		{"ApplySshCheckReceiverConfig", "testdata/opamp.d/opamp-sshcheck-receiver-config.yaml", false, ""},
		{"ApplyStatsdReceiverConfig", "testdata/opamp.d/opamp-statsd-receiver-config.yaml", false, ""},
		{"ApplyTcplogReceiverConfig", "testdata/opamp.d/opamp-tcplog-receiver-config.yaml", false, ""},
		{"ApplyUdplogReceiverConfig", "testdata/opamp.d/opamp-udplog-receiver-config.yaml", false, ""},
		{"ApplyVcenterReceiverConfig", "testdata/opamp.d/opamp-vcenter-receiver-config.yaml", false, ""},
		{"ApplyWavefrontReceiverConfig", "testdata/opamp.d/opamp-wavefront-receiver-config.yaml", false, ""},
		{"ApplyZipkinReceiverConfig", "testdata/opamp.d/opamp-zipkin-receiver-config.yaml", false, ""},
		{"ApplyZookeeperReceiverConfig", "testdata/opamp.d/opamp-zookeeper-receiver-config.yaml", false, ""},
		{"ApplyRemoteWindowsEventConfig", "testdata/opamp.d/opamp-windows-event-config.yaml", false, ""},
		{"ApplyActiveDirecotryDSConfig", "testdata/opamp.d/opamp-activedirectoryds-receiver-config.yaml", false, ""},
		{"ApplyIisReceiverConfig", "testdata/opamp.d/opamp-iis-receiver-config.yaml", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := os.MkdirTemp("", "opamp.d")
			assert.NoError(t, err)
			defer os.RemoveAll(d)
			cfg, set := setupWithRemoteConfig(t, d)
			o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
			assert.NoError(t, err)
			path := filepath.Join(tt.file)
			rb, err := os.ReadFile(path)
			assert.NoError(t, err)

			rc := &protobufs.AgentRemoteConfig{
				Config: &protobufs.AgentConfigMap{
					ConfigMap: map[string]*protobufs.AgentConfigFile{
						"default": {
							Body: rb,
						},
					},
				},
				ConfigHash: []byte("b2b1e3e7f45d564db1c0b621bbf67008"),
			}

			// Test with an error in configuration
			if tt.expectError {
				changed, err := o.applyRemoteConfig(rc)
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.errorMessage)
				assert.False(t, changed)
				assert.Equal(t, len(o.effectiveConfig), 0)
			} else {
				// Test with a valid configuration
				changed, err := o.applyRemoteConfig(rc)
				assert.NoError(t, err)
				assert.True(t, changed)
				assert.NotEqual(t, len(o.effectiveConfig), 0)
			}
			// Test with remote configuration disabled
			cfg.AcceptsRemoteConfiguration = false
			changed, err := o.applyRemoteConfig(rc)
			assert.False(t, changed)
			assert.Error(t, err)
			assert.Equal(t, errMsgRemoteConfigNotAccepted, err.Error())
		})
	}
}

func TestGetAgentCapabilities(t *testing.T) {
	cfg, set := defaultSetup()
	o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)

	assert.Equal(t, o.getAgentCapabilities(), protobufs.AgentCapabilities(4102))

	cfg.AcceptsRemoteConfiguration = false
	assert.Equal(t, o.getAgentCapabilities(), protobufs.AgentCapabilities(4))
}

func TestCreateAgentDescription(t *testing.T) {
	cfg, set := defaultSetup()
	o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)

	assert.Nil(t, o.agentDescription)
	assert.NoError(t, o.createAgentDescription())
	assert.NotNil(t, o.agentDescription)
}

func TestLoadEffectiveConfig(t *testing.T) {
	cfg, set := defaultSetup()
	o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)

	assert.Equal(t, len(o.effectiveConfig), 0)

	assert.NoError(t, o.loadEffectiveConfig("testdata"))
	assert.NotEqual(t, len(o.effectiveConfig), 0)
}

func TestSaveEffectiveConfig(t *testing.T) {
	cfg, set := defaultSetup()
	o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)

	d, err := os.MkdirTemp("", "opamp.d")
	assert.NoError(t, err)
	defer os.RemoveAll(d)

	assert.NoError(t, o.saveEffectiveConfig(d))
}

func TestSaveEffectiveConfigWithInvalidConfig(t *testing.T) {
	tests := []struct {
		name         string
		file         string
		errorMessage string
	}{
		{"ApplyInvalidApacheURIConfig", "testdata/opamp.d/opamp-invalid-apache-uri-config.yaml", "query must be 'auto'"},
		{"ApplyInvalidApacheKeysConfig", "testdata/opamp.d/opamp-invalid-apache-keys-config.yaml", "has invalid keys: endpointt"},
		{"ApplyInvalidPipelineConfigUndefinedComponent", "testdata/opamp.d/opamp-invalid-pipeline-undefined-component-config.yaml", "references receiver \"file\" which is not configured"},
		{"ApplyInvalidPipelineConfigNoExporter", "testdata/opamp.d/opamp-invalid-pipeline-no-exporter-config.yaml", "must have at least one exporter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := os.MkdirTemp("", "opamp.d")
			assert.NoError(t, err)
			defer os.RemoveAll(d)
			cfg, set := setupWithRemoteConfig(t, d)
			o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
			assert.NoError(t, err)
			path := filepath.Join(tt.file)
			rb, err := os.ReadFile(path)
			assert.NoError(t, err)

			o.effectiveConfig = map[string]*protobufs.AgentConfigFile{
				tt.name: {
					Body: rb,
				},
			}
			err = o.saveEffectiveConfig(d)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorMessage)
		})
	}
}

func TestInvalidConfigNotPersistedAfterValidationFailure(t *testing.T) {
	tests := []struct {
		name         string
		file         string
		errorMessage string
	}{
		{"InvalidApacheURIConfig", "testdata/opamp.d/opamp-invalid-apache-uri-config.yaml", "query must be 'auto'"},
		{"InvalidApacheKeysConfig", "testdata/opamp.d/opamp-invalid-apache-keys-config.yaml", "has invalid keys: endpointt"},
		{"InvalidPipelineConfigUndefinedComponent", "testdata/opamp.d/opamp-invalid-pipeline-undefined-component-config.yaml", "references receiver \"file\" which is not configured"},
		{"InvalidPipelineConfigNoExporter", "testdata/opamp.d/opamp-invalid-pipeline-no-exporter-config.yaml", "must have at least one exporter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := os.MkdirTemp("", "opamp.d")
			assert.NoError(t, err)
			defer os.RemoveAll(d)

			cfg, set := setupWithRemoteConfig(t, d)
			o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
			assert.NoError(t, err)

			path := filepath.Join(tt.file)
			rb, err := os.ReadFile(path)
			assert.NoError(t, err)

			configFileName := "invalid-config.yaml"
			o.effectiveConfig = map[string]*protobufs.AgentConfigFile{
				configFileName: {
					Body: rb,
				},
			}

			err = o.saveEffectiveConfig(d)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorMessage)

			invalidConfigPath := filepath.Join(d, configFileName)
			_, statErr := os.Stat(invalidConfigPath)
			assert.True(t, os.IsNotExist(statErr), "invalid config file should not exist on disk after validation failure")

			files, err := os.ReadDir(d)
			assert.NoError(t, err)
			assert.Empty(t, files, "directory should be empty after validation failure")
		})
	}
}

func TestMixedValidAndInvalidConfigsValidFirst(t *testing.T) {
	d, err := os.MkdirTemp("", "opamp.d")
	assert.NoError(t, err)
	defer os.RemoveAll(d)

	cfg, set := setupWithRemoteConfig(t, d)
	o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)

	// valid config
	validConfigPath := filepath.Join("testdata/opamp.d/opamp-remote-config.yaml")
	validConfigBody, err := os.ReadFile(validConfigPath)
	assert.NoError(t, err)

	// invalid config
	invalidConfigPath := filepath.Join("testdata/opamp.d/opamp-invalid-apache-uri-config.yaml")
	invalidConfigBody, err := os.ReadFile(invalidConfigPath)
	assert.NoError(t, err)

	o.effectiveConfig = map[string]*protobufs.AgentConfigFile{
		"01-valid-config.yaml": {
			Body: validConfigBody,
		},
		"02-invalid-config.yaml": {
			Body: invalidConfigBody,
		},
	}
	err = o.saveEffectiveConfig(d)

	// Should fail due to invalid config
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query must be 'auto'")

	invalidFilePath := filepath.Join(d, "02-invalid-config.yaml")
	_, statErr := os.Stat(invalidFilePath)
	assert.True(t, os.IsNotExist(statErr), "invalid config file should not exist on disk after validation failure")
}

func TestUpdateAgentIdentity(t *testing.T) {
	cfg, set := defaultSetup()
	o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)

	olduid := o.instanceId
	assert.NotEmpty(t, olduid.String())

	uid := ulid.Make()
	assert.NotEqual(t, uid, olduid)

	o.updateAgentIdentity(uid)
	assert.Equal(t, o.instanceId, uid)
}

func TestComposeEffectiveConfig(t *testing.T) {
	cfg, set := defaultSetup()
	o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)

	ec := o.composeEffectiveConfig()
	assert.NotNil(t, ec)
}

func TestShutdown(t *testing.T) {
	cfg, set := defaultSetup()
	cfg.ClientConfig.Auth = configoptional.None[configauth.Config]()

	o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)

	// Shutdown with no OpAMP client
	assert.NoError(t, o.Shutdown(context.Background()))
}

func TestStart(t *testing.T) {
	d, err := os.MkdirTemp("", "opamp.d")
	assert.NoError(t, err)
	defer os.RemoveAll(d)

	cfg := createDefaultConfig().(*Config)
	cfg.ClientConfig.Auth = configoptional.None[configauth.Config]()
	cfg.RemoteConfigurationDirectory = d
	set := extensiontest.NewNopSettings(extensiontest.NopType)
	o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)

	assert.NoError(t, o.Start(context.Background(), componenttest.NewNopHost()))
}

func TestReload(t *testing.T) {
	d, err := os.MkdirTemp("", "opamp.d")
	assert.NoError(t, err)
	defer os.RemoveAll(d)

	cfg := createDefaultConfig().(*Config)
	cfg.ClientConfig.Auth = configoptional.None[configauth.Config]()
	cfg.RemoteConfigurationDirectory = d
	set := extensiontest.NewNopSettings(extensiontest.NopType)
	o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)

	ctx := context.Background()
	assert.NoError(t, o.Start(ctx, componenttest.NewNopHost()))
	assert.NoError(t, o.Reload(ctx))
}

func TestDefaultEndpointSetOnStart(t *testing.T) {
	cfg := createDefaultConfig().(*Config)
	set := extensiontest.NewNopSettings(extensiontest.NopType)
	o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
	if err != nil {
		t.Fatal(err)
	}
	settings := o.startSettings()
	if settings.OpAMPServerURL != DefaultSumoLogicOpAmpURL {
		t.Error(errExpectedUncofiguredEndPoint)
	}
}

func TestNewOpampAgent(t *testing.T) {
	cfg, set := defaultSetup()
	o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)
	assert.Equal(t, "otelcoltest", o.agentType)
	assert.Equal(t, "test version", o.agentVersion)
	assert.NotEmpty(t, o.instanceId.String())
	assert.Empty(t, o.effectiveConfig)
	assert.Nil(t, o.agentDescription)
}

func TestNewOpampAgentAttributes(t *testing.T) {
	cfg, set := defaultSetup()
	set.Resource.Attributes().PutStr(string(semconv.ServiceNameKey), "otelcol-sumo")
	set.Resource.Attributes().PutStr(string(semconv.ServiceVersionKey), "sumo.0")
	set.Resource.Attributes().PutStr(string(semconv.ServiceInstanceIDKey), "f8999bc1-4c9b-4619-9bae-7f009d2411ec")
	o, err := newOpampAgent(cfg, set.Logger, set.BuildInfo, set.Resource)
	assert.NoError(t, err)
	assert.Equal(t, "otelcol-sumo", o.agentType)
	assert.Equal(t, "sumo.0", o.agentVersion)
	assert.Equal(t, "7RK6DW2K4V8RCSQBKZ02EJ84FC", o.instanceId.String())
}
