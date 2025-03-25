module github.com/open-telemetry/opentelemetry-collector-contrib/extension/opampextension

go 1.22.7
toolchain go1.24.1

require (
	github.com/google/uuid v1.6.0
	github.com/knadh/koanf/parsers/yaml v0.1.0
	github.com/knadh/koanf/providers/rawbytes v0.1.0
	github.com/knadh/koanf/v2 v2.1.2
	github.com/oklog/ulid/v2 v2.1.0
	github.com/open-telemetry/opamp-go v0.14.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sumologicexporter v0.118.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/syslogexporter v0.118.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension v0.118.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/pprofextension v0.118.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/storage/filestorage v0.118.0
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/sumologicextension v0.118.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor v0.118.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor v0.118.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor v0.118.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor v0.118.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor v0.118.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/apachereceiver v0.118.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/dockerstatsreceiver v0.118.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/elasticsearchreceiver v0.118.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver v0.118.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver v0.118.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkametricsreceiver v0.118.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkareceiver v0.118.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mysqlreceiver v0.118.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/nginxreceiver v0.118.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/postgresqlreceiver v0.118.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/rabbitmqreceiver v0.118.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/redisreceiver v0.118.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/syslogreceiver v0.118.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowseventlogreceiver v0.118.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowsperfcountersreceiver v0.118.0
	github.com/stretchr/testify v1.10.0
	go.opentelemetry.io/collector/component v0.118.0
	go.opentelemetry.io/collector/component/componenttest v0.118.0
	go.opentelemetry.io/collector/config/configauth v0.118.0
	go.opentelemetry.io/collector/config/confighttp v0.118.0
	go.opentelemetry.io/collector/confmap v1.24.0
	go.opentelemetry.io/collector/confmap/provider/envprovider v1.22.0
	go.opentelemetry.io/collector/confmap/provider/fileprovider v1.22.0
	go.opentelemetry.io/collector/confmap/provider/yamlprovider v1.22.0
	go.opentelemetry.io/collector/exporter v0.118.0
	go.opentelemetry.io/collector/exporter/nopexporter v0.118.0
	go.opentelemetry.io/collector/exporter/otlpexporter v0.118.0
	go.opentelemetry.io/collector/extension v0.118.0
	go.opentelemetry.io/collector/extension/extensiontest v0.118.0
	go.opentelemetry.io/collector/otelcol v0.118.0
	go.opentelemetry.io/collector/pdata v1.24.0
	go.opentelemetry.io/collector/processor v0.118.0
	go.opentelemetry.io/collector/processor/batchprocessor v0.118.0
	go.opentelemetry.io/collector/processor/memorylimiterprocessor v0.118.0
	go.opentelemetry.io/collector/receiver v0.118.0
	go.opentelemetry.io/collector/receiver/nopreceiver v0.118.0
	go.opentelemetry.io/collector/receiver/otlpreceiver v0.118.0
	go.opentelemetry.io/collector/semconv v0.118.0
	go.uber.org/multierr v1.11.0
	go.uber.org/zap v1.27.0
)

require (
	cloud.google.com/go/compute/metadata v0.6.0 // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.25.0 // indirect
	github.com/IBM/sarama v1.45.0 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/Showmax/go-fqdn v1.0.0 // indirect
	github.com/alecthomas/participle/v2 v2.1.1 // indirect
	github.com/antchfx/xmlquery v1.4.3 // indirect
	github.com/antchfx/xpath v1.3.3 // indirect
	github.com/apache/thrift v0.21.0 // indirect
	github.com/armon/go-metrics v0.4.1 // indirect
	github.com/aws/aws-msk-iam-sasl-signer-go v1.0.0 // indirect
	github.com/aws/aws-sdk-go v1.55.6 // indirect
	github.com/aws/aws-sdk-go-v2 v1.33.0 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.29.1 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.54 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.24 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.28 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.28 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.200.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.24.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.28.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.33.9 // indirect
	github.com/aws/smithy-go v1.22.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bmatcuk/doublestar/v4 v4.8.0 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/docker v27.5.0+incompatible // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/eapache/go-resiliency v1.7.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20230731223053-c322873962e3 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/ebitengine/purego v0.8.1 // indirect
	github.com/elastic/go-grok v0.3.1 // indirect
	github.com/elastic/lunes v0.1.0 // indirect
	github.com/emicklei/go-restful/v3 v3.11.0 // indirect
	github.com/expr-lang/expr v1.16.9 // indirect
	github.com/fatih/color v1.16.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/fxamacker/cbor/v2 v2.7.0 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.22.4 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/goccy/go-json v0.10.4 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/gorilla/websocket v1.5.1 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.24.0 // indirect
	github.com/hashicorp/consul/api v1.31.0 // indirect
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
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jaegertracing/jaeger v1.65.0 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.4 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jonboulle/clockwork v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/jpillora/backoff v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/knadh/koanf/maps v0.1.1 // indirect
	github.com/knadh/koanf/providers/confmap v0.1.0 // indirect
	github.com/leodido/go-syslog/v4 v4.2.0 // indirect
	github.com/leodido/ragel-machinery v0.0.0-20190525184631-5f46317e436b // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/lufia/plan9stats v0.0.0-20220913051719-115f729f3c8c // indirect
	github.com/magefile/mage v1.15.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.1-0.20231216201459-8508981c8b6c // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mostynb/go-grpc-compression v1.2.3 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nginxinc/nginx-prometheus-exporter v0.11.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/kafkaexporter v0.118.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/aws/ecsutil v0.118.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/common v0.118.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal v0.118.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/docker v0.118.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/filter v0.118.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig v0.118.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/kafka v0.118.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/metadataproviders v0.118.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/pdatautil v0.118.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchpersignal v0.118.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/experimentalmetricmetadata v0.118.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/kafka/topic v0.118.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl v0.118.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatautil v0.118.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza v0.118.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/azure v0.118.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/jaeger v0.118.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/zipkin v0.118.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/winperfcounters v0.118.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0 // indirect
	github.com/openshift/api v3.9.0+incompatible // indirect
	github.com/openshift/client-go v0.0.0-20210521082421-73d9475a9142 // indirect
	github.com/openzipkin/zipkin-go v0.4.3 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20220216144756-c35f1ee13d7c // indirect
	github.com/prometheus-community/windows_exporter v0.27.2 // indirect
	github.com/prometheus/client_golang v1.20.5 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.61.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20201227073835-cf1acfcdf475 // indirect
	github.com/redis/go-redis/v9 v9.7.0 // indirect
	github.com/relvacode/iso8601 v1.6.0 // indirect
	github.com/rs/cors v1.11.1 // indirect
	github.com/shirou/gopsutil/v4 v4.24.12 // indirect
	github.com/spf13/cobra v1.8.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/ua-parser/uap-go v0.0.0-20240611065828-3a4781585db6 // indirect
	github.com/valyala/fastjson v1.6.4 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.etcd.io/bbolt v1.3.11 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/collector v0.118.0 // indirect
	go.opentelemetry.io/collector/client v1.24.0 // indirect
	go.opentelemetry.io/collector/component/componentstatus v0.118.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v1.24.0 // indirect
	go.opentelemetry.io/collector/config/configgrpc v0.118.0 // indirect
	go.opentelemetry.io/collector/config/confignet v1.24.0 // indirect
	go.opentelemetry.io/collector/config/configopaque v1.24.0 // indirect
	go.opentelemetry.io/collector/config/configretry v1.24.0 // indirect
	go.opentelemetry.io/collector/config/configtelemetry v0.118.0 // indirect
	go.opentelemetry.io/collector/config/configtls v1.24.0 // indirect
	go.opentelemetry.io/collector/connector v0.118.0 // indirect
	go.opentelemetry.io/collector/connector/connectortest v0.118.0 // indirect
	go.opentelemetry.io/collector/connector/xconnector v0.118.0 // indirect
	go.opentelemetry.io/collector/consumer v1.24.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.118.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror/xconsumererror v0.118.0 // indirect
	go.opentelemetry.io/collector/consumer/consumertest v0.118.0 // indirect
	go.opentelemetry.io/collector/consumer/xconsumer v0.118.0 // indirect
	go.opentelemetry.io/collector/exporter/exporterhelper/xexporterhelper v0.118.0 // indirect
	go.opentelemetry.io/collector/exporter/exportertest v0.118.0 // indirect
	go.opentelemetry.io/collector/exporter/xexporter v0.118.0 // indirect
	go.opentelemetry.io/collector/extension/auth v0.118.0 // indirect
	go.opentelemetry.io/collector/extension/extensioncapabilities v0.118.0 // indirect
	go.opentelemetry.io/collector/extension/xextension v0.118.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.24.0 // indirect
	go.opentelemetry.io/collector/filter v0.118.0 // indirect
	go.opentelemetry.io/collector/internal/fanoutconsumer v0.118.0 // indirect
	go.opentelemetry.io/collector/internal/memorylimiter v0.118.0 // indirect
	go.opentelemetry.io/collector/internal/sharedcomponent v0.118.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.118.0 // indirect
	go.opentelemetry.io/collector/pdata/testdata v0.118.0 // indirect
	go.opentelemetry.io/collector/pipeline v0.118.0 // indirect
	go.opentelemetry.io/collector/pipeline/xpipeline v0.118.0 // indirect
	go.opentelemetry.io/collector/processor/processorhelper/xprocessorhelper v0.118.0 // indirect
	go.opentelemetry.io/collector/processor/processortest v0.118.0 // indirect
	go.opentelemetry.io/collector/processor/xprocessor v0.118.0 // indirect
	go.opentelemetry.io/collector/receiver/receivertest v0.118.0 // indirect
	go.opentelemetry.io/collector/receiver/xreceiver v0.118.0 // indirect
	go.opentelemetry.io/collector/scraper v0.118.0 // indirect
	go.opentelemetry.io/collector/scraper/scraperhelper v0.118.0 // indirect
	go.opentelemetry.io/collector/service v0.118.0 // indirect
	go.opentelemetry.io/contrib/bridges/otelzap v0.6.0 // indirect
	go.opentelemetry.io/contrib/config v0.10.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.58.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.58.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.31.0 // indirect
	go.opentelemetry.io/otel v1.33.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp v0.7.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.32.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.32.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.33.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.33.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.33.0 // indirect
	go.opentelemetry.io/otel/exporters/prometheus v0.55.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutlog v0.7.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.32.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.33.0 // indirect
	go.opentelemetry.io/otel/log v0.8.0 // indirect
	go.opentelemetry.io/otel/metric v1.33.0 // indirect
	go.opentelemetry.io/otel/sdk v1.33.0 // indirect
	go.opentelemetry.io/otel/sdk/log v0.7.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.33.0 // indirect
	go.opentelemetry.io/otel/trace v1.33.0 // indirect
	go.opentelemetry.io/proto/otlp v1.4.0 // indirect
	golang.org/x/crypto v0.35.0 // indirect
	golang.org/x/exp v0.0.0-20240506185415-9bf2ced13842 // indirect
	golang.org/x/net v0.36.0 // indirect
	golang.org/x/oauth2 v0.24.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/term v0.29.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	golang.org/x/time v0.4.0 // indirect
	gonum.org/v1/gonum v0.15.1 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20241209162323-e6fa225c2576 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241216192217-9240e9c98484 // indirect
	google.golang.org/grpc v1.69.4 // indirect
	google.golang.org/protobuf v1.36.3 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/api v0.31.3 // indirect
	k8s.io/apimachinery v0.31.3 // indirect
	k8s.io/client-go v0.31.3 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/kube-openapi v0.0.0-20240228011516-70dd3763d340 // indirect
	k8s.io/utils v0.0.0-20240711033017-18e509b52bc8 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.1 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)
