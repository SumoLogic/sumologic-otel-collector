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
	"runtime"
	"testing"

	"go.opentelemetry.io/collector/config/configauth"
	"go.opentelemetry.io/collector/config/configoptional"

	"github.com/oklog/ulid/v2"
	"github.com/open-telemetry/opamp-go/protobufs"
	"github.com/stretchr/testify/assert"

	"slices"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
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

func shouldRunOnCurrentOS(allowedOS []string) bool {
	if len(allowedOS) == 0 {
		return true
	}

	currentOS := runtime.GOOS
	return slices.Contains(allowedOS, currentOS)
}

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
		os           []string // OS restrictions - if nil, runs on all OS
	}{
		{"ApplyRemoteConfig", "testdata/opamp.d/opamp-remote-config.yaml", false, "", nil},
		{"ApplyRemoteApacheConfig", "testdata/opamp.d/opamp-apache-config.yaml", false, "", nil},
		{"ApplyRemoteHostConfig", "testdata/opamp.d/opamp-host-config.yaml", false, "", nil},
		{"ApplyRemoteWindowsEventConfig", "testdata/opamp.d/opamp-windows-event-config.yaml", false, "", []string{"windows"}},
		{"ApplyRemoteExtensionsConfig", "testdata/opamp.d/opamp-extensions-config.yaml", false, "", nil},
		{"ApplyRemoteConfigFailed", "testdata/opamp.d/opamp-invalid-remote-config.yaml", true, errMsgInvalidType, nil},
		{"ApplyRemoteConfigMissingProcessor", "testdata/opamp.d/opamp-missing-processor.yaml", true, errMsgInvalidConfigName, nil},
		{"ApplyFilterProcessorConfig", "testdata/opamp.d/opamp-filter-processor.yaml", false, "", nil},
		{"ApplyKafkaMetricsConfig", "testdata/opamp.d/opamp-kafkametrics-config.yaml", false, "", nil},
		{"ApplyElasticsearchConfig", "testdata/opamp.d/opamp-elastic-config.yaml", false, "", nil},
		{"ApplyMysqlConfig", "testdata/opamp.d/opamp-mysql-config.yaml", false, "", nil},
		{"ApplyattributesprocessorConfig", "testdata/opamp.d/opamp-attributes-processor.yaml", false, "", nil},
		{"ApplyPostgresqlConfig", "testdata/opamp.d/opamp-postgresql-config.yaml", false, "", nil},
		{"ApplyRabbitmqConfig", "testdata/opamp.d/opamp-rabbitmq-config.yaml", false, "", nil},
		{"ApplyRedisConfig", "testdata/opamp.d/opamp-redis-config.yaml", false, "", nil},
		{"ApplyFirehoseConfig", "testdata/opamp.d/opamp-aws-firehose-config.yaml", false, "", nil},
		{"ApplyCloudwatchConfig", "testdata/opamp.d/opamp-aws-cloudwatch-receiver-config.yaml", false, "", nil},
		{"ApplyContainerInsightConfig", "testdata/opamp.d/opamp-aws-container-insight-config.yaml", false, "", nil},
		{"ApplyEcsContainerMetricsConfig", "testdata/opamp.d/opamp-aws-container-metrics-config.yaml", false, "", nil},
		{"ApplyS3Config", "testdata/opamp.d/opamp-aws-s3-exporter-config.yaml", false, "", nil},
		{"ApplyCloudwatchConfigFailure", "testdata/opamp.d/opamp-aws-cloudwatch-receiver-error-config.yaml", true, errMsgInvalidCloudwatchConfig, nil},
		{"ApplyXrayConfig", "testdata/opamp.d/opamp-aws-xray-config.yaml", false, "", nil},
		{"ApplyKenesisConfig", "testdata/opamp.d/opamp-aws-kenesis-config.yaml", false, "", nil},
		{"ApplyCarbonExporterConfig", "testdata/opamp.d/opamp-carbon-exporter-config.yaml", false, "", nil},
		{"ApplyDebugExporterConfig", "testdata/opamp.d/opamp-debug-exporter-config.yaml", false, "", nil},
		{"ApplyFileExporterConfig", "testdata/opamp.d/opamp-file-exporter-config.yaml", false, "", nil},
		{"ApplyKafkaExporterConfig", "testdata/opamp.d/opamp-kafka-exporter-config.yaml", false, "", nil},
		{"ApplyLoadbalancingExporterConfig", "testdata/opamp.d/opamp-loadbalancing-exporter-config.yaml", false, "", nil},
		{"ApplyCumulativetodeltaProcessorConfig", "testdata/opamp.d/opamp-cumulativetodelta-processor-config.yaml", false, "", nil},
		{"ApplyDeltatorateProcessorConfig", "testdata/opamp.d/opamp-deltatorate-processor-config.yaml", false, "", nil},
		{"ApplyMetricsgenerationProcessorConfig", "testdata/opamp.d/opamp-metricsgeneration-processor-config.yaml", false, "", nil},
		{"ApplyGroupbyattrsProcessorConfig", "testdata/opamp.d/opamp-groupbyattrs-processor-config.yaml", false, "", nil},
		{"ApplyGroupbytraceProcessorConfig", "testdata/opamp.d/opamp-groupbytrace-processor-config.yaml", false, "", nil},
		{"ApplyK8sattributesProcessorConfig", "testdata/opamp.d/opamp-k8sattributes-processor-config.yaml", false, "", nil},
		{"ApplyLogdedupProcessorConfig", "testdata/opamp.d/opamp-logdedup-processor-config.yaml", false, "", nil},
		{"ApplyLogstransformProcessorConfig", "testdata/opamp.d/opamp-logstransform-processor-config.yaml", false, "", nil},
		{"ApplyMetricstransformProcessorConfig", "testdata/opamp.d/opamp-metricstransform-processor-config.yaml", false, "", nil},
		{"ApplyProbabilisticsamplerProcessorConfig", "testdata/opamp.d/opamp-probabilisticsampler-processor-config.yaml", false, "", nil},
		{"ApplyActiveDirecotryDSConfig", "testdata/opamp.d/opamp-activedirectoryds-receiver-config.yaml", false, "", []string{"windows"}},
		{"ApplyAerospikeConfig", "testdata/opamp.d/opamp-aerospike-receiver-config.yaml", false, "", nil},
		{"ApplyAzureEventHubConfig", "testdata/opamp.d/opamp-azureeventhub-receiver-config.yaml", false, "", nil},
		{"ApplyBigipConfig", "testdata/opamp.d/opamp-bigip-receiver-config.yaml", false, "", nil},
		{"ApplyCarbonReceiverConfig", "testdata/opamp.d/opamp-carbon-receiver-config.yaml", false, "", nil},
		{"ApplyChronyConfig", "testdata/opamp.d/opamp-chrony-receiver-config.yaml", false, "", []string{"linux", "darwin"}},
		{"ApplyCloudlfareReceiverConfig", "testdata/opamp.d/opamp-cloudflare-receiver-config.yaml", false, "", nil},
		{"ApplyPrometheusExporterConfig", "testdata/opamp.d/opamp-prometheus-exporter-config.yaml", false, "", nil},
		{"ApplyOtlphttpConfig", "testdata/opamp.d/opamp-otlphttp-exporter-config.yaml", false, "", nil},
		{"ApplyECSObserverConfig", "testdata/opamp.d/opamp-ecsobserver-config.yaml", false, "", nil},
		{"ApplyRedactionProcessorConfig", "testdata/opamp.d/opamp-redaction-processor-config.yaml", false, "", nil},
		{"ApplyRemotetapProcessorConfig", "testdata/opamp.d/opamp-remotetap-processor-config.yaml", false, "", nil},
		{"ApplyGeoipProcessorConfig", "testdata/opamp.d/opamp-geoip-processor-config.yaml", false, "", nil},
		{"ApplySchemaProcessorConfig", "testdata/opamp.d/opamp-schema-processor-config.yaml", false, "", nil},
		{"ApplySpanProcessorConfig", "testdata/opamp.d/opamp-span-processor-config.yaml", false, "", nil},
		{"ApplyTailsamplingProcessorConfig", "testdata/opamp.d/opamp-tailsampling-processor-config.yaml", false, "", nil},
		{"ApplyCloudfoundryReceiverConfig", "testdata/opamp.d/opamp-cloudfoundry-receiver-config.yaml", false, "", nil},
		{"ApplyIisReceiverConfig", "testdata/opamp.d/opamp-iis-receiver-config.yaml", false, "", nil},
		{"ApplyHttpcheckReceiverConfig", "testdata/opamp.d/opamp-httpcheck-receiver-config.yaml", false, "", nil},
		{"ApplyAsapAuthExtensionConfig", "testdata/opamp.d/opamp-asapauth-extension-config.yaml", false, "", nil},
		{"ApplyBasicAuthExtensionConfig", "testdata/opamp.d/opamp-basicauth-extension-config.yaml", false, "", nil},
		{"ApplyBearerTokenAuthExtensionConfig", "testdata/opamp.d/opamp-bearertokenauth-extension-config.yaml", false, "", nil},
		{"ApplyDbStorageExtensionConfig", "testdata/opamp.d/opamp-dbstorage-extension-config.yaml", false, "", nil},
		{"ApplyDockerObserverExtensionConfig", "testdata/opamp.d/opamp-dockerobserver-extension-config.yaml", false, "", nil},
		{"ApplyHeadersSetterExtensionConfig", "testdata/opamp.d/opamp-headerssetter-extension-config.yaml", false, "", nil},
		{"ApplyHostObserverExtensionConfig", "testdata/opamp.d/opamp-hostobserver-extension-config.yaml", false, "", nil},
		{"ApplyHttpForwarderExtensionConfig", "testdata/opamp.d/opamp-httpforwarder-extension-config.yaml", false, "", nil},
		{"ApplyJaegerRemoteSamplingExtensionConfig", "testdata/opamp.d/opamp-jaegerremotesampling-extension-config.yaml", false, "", nil},
		{"ApplyK8sObserverExtensionConfig", "testdata/opamp.d/opamp-k8sobserver-extension-config.yaml", false, "", nil},
		{"ApplyOauth2ClientauthExtensionConfig", "testdata/opamp.d/opamp-oauth2clientauth-extension-config.yaml", false, "", nil},
		{"ApplyOidcAuthExtensionConfig", "testdata/opamp.d/opamp-oidcauth-extension-config.yaml", false, "", nil},
		{"ApplyPprofExtensionConfig", "testdata/opamp.d/opamp-pprof-extension-config.yaml", false, "", nil},
		{"ApplyZpagesExtensionConfig", "testdata/opamp.d/opamp-zpages-extension-config.yaml", false, "", nil},
		{"ApplyInfluxdbReceiverConfig", "testdata/opamp.d/opamp-influxdb-receiver-config.yaml", false, "", nil},
		{"ApplyJaegerReceiverConfig", "testdata/opamp.d/opamp-jaeger-receiver-config.yaml", false, "", nil},
		{"ApplyJournaldReceiverConfig", "testdata/opamp.d/opamp-journald-receiver-config.yaml", false, "", nil},
		{"ApplyK8sClusterReceiverConfig", "testdata/opamp.d/opamp-k8scluster-receiver-config.yaml", false, "", nil},
		{"ApplyK8sEventsReceiverConfig", "testdata/opamp.d/opamp-k8sevents-receiver-config.yaml", false, "", nil},
		{"ApplyK8sObjectsReceiverConfig", "testdata/opamp.d/opamp-k8sobjects-receiver-config.yaml", false, "", nil},
		{"ApplyKubeletStatsReceiverConfig", "testdata/opamp.d/opamp-kubeletstats-receiver-config.yaml", false, "", nil},
		{"ApplyLokiReceiverConfig", "testdata/opamp.d/opamp-loki-receiver-config.yaml", false, "", nil},
		{"ApplyMemcachedReceiverConfig", "testdata/opamp.d/opamp-memcached-receiver-config.yaml", false, "", nil},
		{"ApplyMongodbReceiverConfig", "testdata/opamp.d/opamp-mongodb-receiver-config.yaml", false, "", nil},
		{"ApplyMongodbAtlasReceiverConfig", "testdata/opamp.d/opamp-mongodbatlas-receiver-config.yaml", false, "", nil},
		{"ApplyNsxtReceiverConfig", "testdata/opamp.d/opamp-nsxt-receiver-config.yaml", false, "", nil},
		{"ApplyOracledbReceiverConfig", "testdata/opamp.d/opamp-oracledb-receiver-config.yaml", false, "", nil},
		{"ApplyOtlpJsonFileReceiverConfig", "testdata/opamp.d/opamp-otlpjsonfile-receiver-config.yaml", false, "", nil},
		{"ApplyPodmanReceiverConfig", "testdata/opamp.d/opamp-podman-receiver-config.yaml", false, "", nil},
		{"ApplySimplePrometheusReceiverConfig", "testdata/opamp.d/opamp-simpleprometheus-receiver-config.yaml", false, "", nil},
		{"ApplyPrometheusReceiverConfig", "testdata/opamp.d/opamp-prometheus-receiver-config.yaml", false, "", nil},
		{"ApplyPulsarReceiverConfig", "testdata/opamp.d/opamp-pulsar-receiver-config.yaml", false, "", nil},
		{"ApplyPurefaReceiverConfig", "testdata/opamp.d/opamp-purefa-receiver-config.yaml", false, "", nil},
		{"ApplyPurefbReceiverConfig", "testdata/opamp.d/opamp-purefb-receiver-config.yaml", false, "", nil},
		{"ApplyReceiverCreatorConfig", "testdata/opamp.d/opamp-receiver-creator-config.yaml", false, "", nil},
		{"ApplyRiakReceiverConfig", "testdata/opamp.d/opamp-riak-receiver-config.yaml", false, "", nil},
		{"ApplySaphanaReceiverConfig", "testdata/opamp.d/opamp-saphana-receiver-config.yaml", false, "", nil},
		{"ApplySignalFxReceiverConfig", "testdata/opamp.d/opamp-signalfx-receiver-config.yaml", false, "", nil},
		{"ApplySkyWalkingReceiverConfig", "testdata/opamp.d/opamp-skywalking-receiver-config.yaml", false, "", nil},
		{"ApplySnowflakeReceiverConfig", "testdata/opamp.d/opamp-snowflake-receiver-config.yaml", false, "", nil},
		{"ApplySolaceReceiverConfig", "testdata/opamp.d/opamp-solace-receiver-config.yaml", false, "", nil},
		{"ApplySplunkhecReceiverConfig", "testdata/opamp.d/opamp-splunkhec-receiver-config.yaml", false, "", nil},
		{"ApplySqlQueryReceiverConfig", "testdata/opamp.d/opamp-sqlquery-receiver-config.yaml", false, "", nil},
		{"ApplySqlServerReceiverConfig", "testdata/opamp.d/opamp-sqlserver-receiver-config.yaml", false, "", nil},
		{"ApplySshCheckReceiverConfig", "testdata/opamp.d/opamp-sshcheck-receiver-config.yaml", false, "", nil},
		{"ApplyStatsdReceiverConfig", "testdata/opamp.d/opamp-statsd-receiver-config.yaml", false, "", nil},
		{"ApplyTcplogReceiverConfig", "testdata/opamp.d/opamp-tcplog-receiver-config.yaml", false, "", nil},
		{"ApplyUdplogReceiverConfig", "testdata/opamp.d/opamp-udplog-receiver-config.yaml", false, "", nil},
		{"ApplyVcenterReceiverConfig", "testdata/opamp.d/opamp-vcenter-receiver-config.yaml", false, "", nil},
		{"ApplyWavefrontReceiverConfig", "testdata/opamp.d/opamp-wavefront-receiver-config.yaml", false, "", nil},
		{"ApplyZipkinReceiverConfig", "testdata/opamp.d/opamp-zipkin-receiver-config.yaml", false, "", nil},
		{"ApplyZookeeperReceiverConfig", "testdata/opamp.d/opamp-zookeeper-receiver-config.yaml", false, "", nil},
		{"ApplyEcstaskExtensionConfig", "testdata/opamp.d/opamp-ecstask-extension-config.yaml", false, "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !shouldRunOnCurrentOS(tt.os) {
				t.Skipf("Skipping test %s: not supported on %s (supported OS: %v)", tt.name, runtime.GOOS, tt.os)
				return
			}

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
