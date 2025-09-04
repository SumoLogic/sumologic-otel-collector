module github.com/open-telemetry/opentelemetry-collector-contrib/extension/opampextension

go 1.24.0

require (
	github.com/google/uuid v1.6.0
	github.com/knadh/koanf/parsers/yaml v1.1.0
	github.com/knadh/koanf/providers/rawbytes v1.0.0
	github.com/knadh/koanf/v2 v2.2.2
	github.com/oklog/ulid/v2 v2.1.1
	github.com/open-telemetry/opamp-go v0.22.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/awskinesisexporter v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/awss3exporter v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/carbonexporter v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/fileexporter v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/kafkaexporter v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/loadbalancingexporter v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusexporter v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sumologicexporter v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/syslogexporter v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/awsproxy v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/observer/ecsobserver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/pprofextension v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/storage/filestorage v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/sumologicextension v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/cumulativetodeltaprocessor v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/deltatorateprocessor v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/groupbyattrsprocessor v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/groupbytraceprocessor v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/logdedupprocessor v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/logstransformprocessor v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricsgenerationprocessor v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/probabilisticsamplerprocessor v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/activedirectorydsreceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/aerospikereceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/apachereceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awscloudwatchreceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awscontainerinsightreceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awsecscontainermetricsreceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awsfirehosereceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awsxrayreceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/azureeventhubreceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/bigipreceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/carbonreceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/chronyreceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/cloudflarereceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/collectdreceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/couchdbreceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/datadogreceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/dockerstatsreceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/elasticsearchreceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/expvarreceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filestatsreceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/flinkmetricsreceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/fluentforwardreceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/googlecloudpubsubreceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/googlecloudspannerreceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/haproxyreceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkametricsreceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkareceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mysqlreceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/nginxreceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/postgresqlreceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/rabbitmqreceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/redisreceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/syslogreceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowseventlogreceiver v0.134.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowsperfcountersreceiver v0.134.0
	github.com/stretchr/testify v1.11.1
	go.opentelemetry.io/collector/component v1.40.0
	go.opentelemetry.io/collector/component/componenttest v0.134.0
	go.opentelemetry.io/collector/config/configauth v0.134.0
	go.opentelemetry.io/collector/config/confighttp v0.134.0
	go.opentelemetry.io/collector/config/configoptional v0.134.0
	go.opentelemetry.io/collector/confmap v1.40.0
	go.opentelemetry.io/collector/confmap/provider/envprovider v1.40.0
	go.opentelemetry.io/collector/confmap/provider/fileprovider v1.40.0
	go.opentelemetry.io/collector/confmap/provider/yamlprovider v1.40.0
	go.opentelemetry.io/collector/exporter v0.134.0
	go.opentelemetry.io/collector/exporter/debugexporter v0.134.0
	go.opentelemetry.io/collector/exporter/nopexporter v0.134.0
	go.opentelemetry.io/collector/exporter/otlpexporter v0.134.0
	go.opentelemetry.io/collector/exporter/otlphttpexporter v0.134.0
	go.opentelemetry.io/collector/extension v1.40.0
	go.opentelemetry.io/collector/extension/extensiontest v0.134.0
	go.opentelemetry.io/collector/otelcol v0.134.0
	go.opentelemetry.io/collector/pdata v1.40.0
	go.opentelemetry.io/collector/processor v1.40.0
	go.opentelemetry.io/collector/processor/batchprocessor v0.134.0
	go.opentelemetry.io/collector/processor/memorylimiterprocessor v0.134.0
	go.opentelemetry.io/collector/receiver v1.40.0
	go.opentelemetry.io/collector/receiver/nopreceiver v0.134.0
	go.opentelemetry.io/collector/receiver/otlpreceiver v0.134.0
	go.opentelemetry.io/otel v1.38.0
	go.uber.org/multierr v1.11.0
	go.uber.org/zap v1.27.0
)

require (
	cel.dev/expr v0.24.0 // indirect
	cloud.google.com/go v0.121.2 // indirect
	cloud.google.com/go/auth v0.16.5 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.8.0 // indirect
	cloud.google.com/go/iam v1.5.2 // indirect
	cloud.google.com/go/logging v1.13.0 // indirect
	cloud.google.com/go/longrunning v0.6.7 // indirect
	cloud.google.com/go/monitoring v1.24.2 // indirect
	cloud.google.com/go/pubsub v1.49.0 // indirect
	cloud.google.com/go/spanner v1.83.0 // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/Azure/azure-amqp-common-go/v4 v4.2.0 // indirect
	github.com/Azure/azure-event-hubs-go/v3 v3.6.2 // indirect
	github.com/Azure/azure-sdk-for-go v68.0.0+incompatible // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.18.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.10.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.11.1 // indirect
	github.com/Azure/go-amqp v1.0.2 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.28 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.21 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/to v0.4.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.3.1 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.4.2 // indirect
	github.com/DataDog/agent-payload/v5 v5.0.163 // indirect
	github.com/DataDog/datadog-agent/comp/core/tagger/origindetection v0.69.3 // indirect
	github.com/DataDog/datadog-agent/pkg/obfuscate v0.69.3 // indirect
	github.com/DataDog/datadog-agent/pkg/proto v0.69.3 // indirect
	github.com/DataDog/datadog-agent/pkg/remoteconfig/state v0.67.0 // indirect
	github.com/DataDog/datadog-agent/pkg/trace v0.71.0-devel.0.20250820164444-fcef12608466 // indirect
	github.com/DataDog/datadog-agent/pkg/util/hostname/validate v0.69.3 // indirect
	github.com/DataDog/datadog-agent/pkg/util/log v0.69.3 // indirect
	github.com/DataDog/datadog-agent/pkg/util/scrubber v0.69.3 // indirect
	github.com/DataDog/datadog-agent/pkg/version v0.69.3 // indirect
	github.com/DataDog/datadog-api-client-go/v2 v2.44.0 // indirect
	github.com/DataDog/datadog-go/v5 v5.6.0 // indirect
	github.com/DataDog/go-sqllexer v0.1.6 // indirect
	github.com/DataDog/go-tuf v1.1.0-0.5.2 // indirect
	github.com/DataDog/opentelemetry-mapping-go/pkg/otlp/attributes v0.31.0 // indirect
	github.com/DataDog/opentelemetry-mapping-go/pkg/otlp/metrics v0.29.1 // indirect
	github.com/DataDog/opentelemetry-mapping-go/pkg/quantile v0.29.1 // indirect
	github.com/DataDog/sketches-go v1.4.7 // indirect
	github.com/DataDog/zstd v1.5.6 // indirect
	github.com/GoogleCloudPlatform/grpc-gcp-go/grpcgcp v1.5.3 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.29.0 // indirect
	github.com/IBM/sarama v1.46.0 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/Showmax/go-fqdn v1.0.0 // indirect
	github.com/aerospike/aerospike-client-go/v8 v8.2.2 // indirect
	github.com/alecthomas/participle/v2 v2.1.4 // indirect
	github.com/alecthomas/units v0.0.0-20240927000941-0f3dac36c52b // indirect
	github.com/antchfx/xmlquery v1.4.4 // indirect
	github.com/antchfx/xpath v1.3.5 // indirect
	github.com/apache/thrift v0.22.0 // indirect
	github.com/armon/go-metrics v0.4.1 // indirect
	github.com/aws/aws-msk-iam-sasl-signer-go v1.0.4 // indirect
	github.com/aws/aws-sdk-go v1.55.7 // indirect
	github.com/aws/aws-sdk-go-v2 v1.37.0 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.0 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.30.1 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.18.1 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.0 // indirect
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.18.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs v1.54.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.237.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/ecs v1.61.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.8.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.19.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/kinesis v1.36.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.85.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/servicediscovery v1.36.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.26.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.31.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.35.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/xray v1.32.0 // indirect
	github.com/aws/smithy-go v1.22.5 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/bmatcuk/doublestar/v4 v4.9.1 // indirect
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cihub/seelog v0.0.0-20170130134532-f561c5e57575 // indirect
	github.com/cncf/xds/go v0.0.0-20250501225837-2ac532fd4443 // indirect
	github.com/containerd/containerd/api v1.9.0 // indirect
	github.com/containerd/errdefs v1.0.0 // indirect
	github.com/containerd/errdefs/pkg v0.3.0 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containerd/ttrpc v1.2.7 // indirect
	github.com/containerd/typeurl/v2 v2.2.3 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/cyphar/filepath-securejoin v0.4.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/devigned/tab v0.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/docker v28.3.3+incompatible // indirect
	github.com/docker/go-connections v0.6.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/eapache/go-resiliency v1.7.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20230731223053-c322873962e3 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/ebitengine/purego v0.8.4 // indirect
	github.com/elastic/go-grok v0.3.1 // indirect
	github.com/elastic/lunes v0.1.0 // indirect
	github.com/emicklei/go-restful/v3 v3.11.0 // indirect
	github.com/envoyproxy/go-control-plane/envoy v1.32.4 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.2.1 // indirect
	github.com/euank/go-kmsg-parser v2.0.0+incompatible // indirect
	github.com/expr-lang/expr v1.17.6 // indirect
	github.com/facebook/time v0.0.0-20240510113249-fa89cc575891 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/foxboron/go-tpm-keyfiles v0.0.0-20250323135004-b31fac66206e // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/fxamacker/cbor/v2 v2.7.0 // indirect
	github.com/go-jose/go-jose/v4 v4.1.1 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/go-sql-driver/mysql v1.9.3 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/cadvisor v0.53.0 // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/go-tpm v0.9.5 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.6 // indirect
	github.com/googleapis/gax-go/v2 v2.15.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/grafana/regexp v0.0.0-20240518133315-a468a5bfb3bc // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.1 // indirect
	github.com/hashicorp/consul/api v1.32.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-hclog v1.6.3 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/hashicorp/serf v0.10.1 // indirect
	github.com/iancoleman/strcase v0.3.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/itchyny/timefmt-go v0.1.6 // indirect
	github.com/jaegertracing/jaeger-idl v0.6.0 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.4 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jellydator/ttlcache/v3 v3.4.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jonboulle/clockwork v0.5.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/karrick/godirwalk v1.17.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/knadh/koanf/providers/confmap v1.0.0 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/leodido/go-syslog/v4 v4.2.0 // indirect
	github.com/leodido/ragel-machinery v0.0.0-20190525184631-5f46317e436b // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/lightstep/go-expohisto v1.0.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20250317134145-8bc96cf8fc35 // indirect
	github.com/magefile/mage v1.15.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/michel-laterman/proxy-connect-dialer-go v0.1.0 // indirect
	github.com/mistifyio/go-zfs v2.1.2-0.20190413222219-f784269be439+incompatible // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/hashstructure/v2 v2.0.2 // indirect
	github.com/mitchellh/mapstructure v1.5.1-0.20231216201459-8508981c8b6c // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/sys/mountinfo v0.7.2 // indirect
	github.com/moby/sys/userns v0.1.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/mostynb/go-grpc-compression v1.2.3 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mwitkow/go-conntrack v0.0.0-20190716064945-2f068394615f // indirect
	github.com/nginx/nginx-prometheus-exporter v1.4.1 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/encoding v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/encoding/awscloudwatchmetricstreamsencodingextension v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/awsutil v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/containerinsight v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/ecsutil v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/k8s v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/metrics v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/proxy v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/xray v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/collectd v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/common v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/datadog v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/docker v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/exp/metrics v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/filter v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/gopsutilenv v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/kafka v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/kubelet v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/metadataproviders v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/pdatautil v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/sharedcomponent v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/sqlquery v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchperresourceattr v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchpersignal v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/core/xidutils v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/datadog v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/experimentalmetricmetadata v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/kafka/configkafka v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/kafka/topic v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatautil v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/resourcetotelemetry v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sampling v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/azure v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/azurelogs v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/jaeger v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/prometheus v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/zipkin v0.134.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/winperfcounters v0.134.0 // indirect
	github.com/opencontainers/cgroups v0.0.2 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/opencontainers/runtime-spec v1.2.1 // indirect
	github.com/openshift/api v3.9.0+incompatible // indirect
	github.com/openshift/client-go v0.0.0-20241203091221-452dfb8fa071 // indirect
	github.com/openzipkin/zipkin-go v0.4.3 // indirect
	github.com/outcaste-io/ristretto v0.2.3 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/philhofer/fwd v1.2.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/prometheus/client_golang v1.23.0 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.65.0 // indirect
	github.com/prometheus/otlptranslator v0.0.0-20250717125610-8549f4ab4f8f // indirect
	github.com/prometheus/procfs v0.17.0 // indirect
	github.com/prometheus/prometheus v0.304.3-0.20250703114031-419d436a447a // indirect
	github.com/prometheus/sigv4 v0.2.0 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20250401214520-65e299d6c5c9 // indirect
	github.com/redis/go-redis/v9 v9.12.1 // indirect
	github.com/relvacode/iso8601 v1.6.0 // indirect
	github.com/rs/cors v1.11.1 // indirect
	github.com/secure-systems-lab/go-securesystemslib v0.9.0 // indirect
	github.com/shirou/gopsutil/v4 v4.25.8-0.20250809033336-ffcdc2b7662f // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spf13/cobra v1.9.1 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/spiffe/go-spiffe/v2 v2.5.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/tilinna/clock v1.1.0 // indirect
	github.com/tinylib/msgp v1.4.0 // indirect
	github.com/tklauser/go-sysconf v0.3.15 // indirect
	github.com/tklauser/numcpus v0.10.0 // indirect
	github.com/twmb/franz-go v1.19.5 // indirect
	github.com/twmb/franz-go/pkg/kmsg v1.11.2 // indirect
	github.com/twmb/franz-go/pkg/sasl/kerberos v1.1.0 // indirect
	github.com/twmb/franz-go/plugin/kzap v1.1.2 // indirect
	github.com/twmb/murmur3 v1.1.8 // indirect
	github.com/ua-parser/uap-go v0.0.0-20240611065828-3a4781585db6 // indirect
	github.com/valyala/fastjson v1.6.4 // indirect
	github.com/wadey/gocovmerge v0.0.0-20160331181800-b5bfa59ec0ad // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/yuin/gopher-lua v1.1.1 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	github.com/zeebo/errs v1.4.0 // indirect
	go.etcd.io/bbolt v1.4.3 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/collector v0.134.0 // indirect
	go.opentelemetry.io/collector/client v1.40.0 // indirect
	go.opentelemetry.io/collector/component/componentstatus v0.134.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v1.40.0 // indirect
	go.opentelemetry.io/collector/config/configgrpc v0.134.0 // indirect
	go.opentelemetry.io/collector/config/configmiddleware v0.134.0 // indirect
	go.opentelemetry.io/collector/config/confignet v1.40.0 // indirect
	go.opentelemetry.io/collector/config/configopaque v1.40.0 // indirect
	go.opentelemetry.io/collector/config/configretry v1.40.0 // indirect
	go.opentelemetry.io/collector/config/configtelemetry v0.134.0 // indirect
	go.opentelemetry.io/collector/config/configtls v1.40.0 // indirect
	go.opentelemetry.io/collector/confmap/xconfmap v0.134.0 // indirect
	go.opentelemetry.io/collector/connector v0.134.0 // indirect
	go.opentelemetry.io/collector/connector/connectortest v0.134.0 // indirect
	go.opentelemetry.io/collector/connector/xconnector v0.134.0 // indirect
	go.opentelemetry.io/collector/consumer v1.40.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.134.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror/xconsumererror v0.134.0 // indirect
	go.opentelemetry.io/collector/consumer/consumertest v0.134.0 // indirect
	go.opentelemetry.io/collector/consumer/xconsumer v0.134.0 // indirect
	go.opentelemetry.io/collector/exporter/exporterhelper v0.134.0 // indirect
	go.opentelemetry.io/collector/exporter/exporterhelper/xexporterhelper v0.134.0 // indirect
	go.opentelemetry.io/collector/exporter/exportertest v0.134.0 // indirect
	go.opentelemetry.io/collector/exporter/xexporter v0.134.0 // indirect
	go.opentelemetry.io/collector/extension/extensionauth v1.40.0 // indirect
	go.opentelemetry.io/collector/extension/extensioncapabilities v0.134.0 // indirect
	go.opentelemetry.io/collector/extension/extensionmiddleware v0.134.0 // indirect
	go.opentelemetry.io/collector/extension/xextension v0.134.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.40.0 // indirect
	go.opentelemetry.io/collector/filter v0.134.0 // indirect
	go.opentelemetry.io/collector/internal/fanoutconsumer v0.134.0 // indirect
	go.opentelemetry.io/collector/internal/memorylimiter v0.134.0 // indirect
	go.opentelemetry.io/collector/internal/sharedcomponent v0.134.0 // indirect
	go.opentelemetry.io/collector/internal/telemetry v0.134.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.134.0 // indirect
	go.opentelemetry.io/collector/pdata/testdata v0.134.0 // indirect
	go.opentelemetry.io/collector/pdata/xpdata v0.134.0 // indirect
	go.opentelemetry.io/collector/pipeline v1.40.0 // indirect
	go.opentelemetry.io/collector/pipeline/xpipeline v0.134.0 // indirect
	go.opentelemetry.io/collector/processor/processorhelper v0.134.0 // indirect
	go.opentelemetry.io/collector/processor/processorhelper/xprocessorhelper v0.134.0 // indirect
	go.opentelemetry.io/collector/processor/processortest v0.134.0 // indirect
	go.opentelemetry.io/collector/processor/xprocessor v0.134.0 // indirect
	go.opentelemetry.io/collector/receiver/receiverhelper v0.134.0 // indirect
	go.opentelemetry.io/collector/receiver/receivertest v0.134.0 // indirect
	go.opentelemetry.io/collector/receiver/xreceiver v0.134.0 // indirect
	go.opentelemetry.io/collector/scraper v0.134.0 // indirect
	go.opentelemetry.io/collector/scraper/scraperhelper v0.134.0 // indirect
	go.opentelemetry.io/collector/semconv v0.128.1-0.20250610090210-188191247685 // indirect
	go.opentelemetry.io/collector/service v0.134.0 // indirect
	go.opentelemetry.io/collector/service/hostcapabilities v0.134.0 // indirect
	go.opentelemetry.io/contrib/bridges/otelzap v0.12.0 // indirect
	go.opentelemetry.io/contrib/detectors/gcp v1.36.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.62.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.62.0 // indirect
	go.opentelemetry.io/contrib/otelconf v0.17.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.37.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc v0.13.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp v0.13.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.37.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.37.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.37.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.37.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.37.0 // indirect
	go.opentelemetry.io/otel/exporters/prometheus v0.59.1 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutlog v0.13.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.37.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.37.0 // indirect
	go.opentelemetry.io/otel/log v0.13.0 // indirect
	go.opentelemetry.io/otel/metric v1.38.0 // indirect
	go.opentelemetry.io/otel/sdk v1.37.0 // indirect
	go.opentelemetry.io/otel/sdk/log v0.13.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.37.0 // indirect
	go.opentelemetry.io/otel/trace v1.38.0 // indirect
	go.opentelemetry.io/proto/otlp v1.7.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/crypto v0.41.0 // indirect
	golang.org/x/exp v0.0.0-20250606033433-dcc06ee1d476 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/oauth2 v0.30.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/term v0.34.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	golang.org/x/time v0.12.0 // indirect
	golang.org/x/tools v0.36.0 // indirect
	gonum.org/v1/gonum v0.16.0 // indirect
	google.golang.org/api v0.248.0 // indirect
	google.golang.org/genproto v0.0.0-20250603155806-513f23925822 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250818200422-3122310a409c // indirect
	google.golang.org/grpc v1.75.0 // indirect
	google.golang.org/protobuf v1.36.8 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gopkg.in/zorkian/go-datadog-api.v2 v2.30.0 // indirect
	k8s.io/api v0.32.3 // indirect
	k8s.io/apimachinery v0.32.3 // indirect
	k8s.io/client-go v0.32.3 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/kube-openapi v0.0.0-20241105132330-32ad38e42d3f // indirect
	k8s.io/utils v0.0.0-20250502105355-0f33e8f1c979 // indirect
	sigs.k8s.io/controller-runtime v0.20.4 // indirect
	sigs.k8s.io/json v0.0.0-20241010143419-9aa6b5e7a4b3 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.5.0 // indirect
	sigs.k8s.io/yaml v1.6.0 // indirect
)
