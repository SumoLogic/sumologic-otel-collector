module github.com/SumoLogic/sumologic-otel-collector/pkg/receiver/jobreceiver

go 1.24.0

toolchain go1.24.7

require (
	github.com/cenkalti/backoff/v4 v4.3.0
	github.com/mholt/archives v0.1.5
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/stanza v0.144.0
	github.com/stretchr/testify v1.11.1
	go.opentelemetry.io/collector/component v1.50.0
	go.opentelemetry.io/collector/component/componenttest v0.144.0
	go.opentelemetry.io/collector/confmap v1.50.0
	go.opentelemetry.io/collector/consumer/consumertest v0.144.0
	go.opentelemetry.io/collector/receiver v1.50.0
	go.opentelemetry.io/collector/receiver/receivertest v0.144.0
	go.uber.org/zap v1.27.1
	golang.org/x/text v0.31.0
	gopkg.in/h2non/filetype.v1 v1.0.5
)

require (
	github.com/STARRY-S/zip v0.2.3 // indirect
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/bodgit/plumbing v1.3.0 // indirect
	github.com/bodgit/sevenzip v1.6.1 // indirect
	github.com/bodgit/windows v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dsnet/compress v0.0.2-0.20230904184137-39efe44ab707 // indirect
	github.com/elastic/lunes v0.2.0 // indirect
	github.com/expr-lang/expr v1.17.7 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-version v1.8.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/klauspost/pgzip v1.2.6 // indirect
	github.com/knadh/koanf v1.5.0 // indirect
	github.com/knadh/koanf/v2 v2.3.0 // indirect
	github.com/leodido/go-syslog/v4 v4.3.0 // indirect
	github.com/leodido/ragel-machinery v0.0.0-20190525184631-5f46317e436b // indirect
	github.com/magefile/mage v1.15.0 // indirect
	github.com/mikelolasagasti/xz v1.0.1 // indirect
	github.com/minio/minlz v1.0.1 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/nwaples/rardecode/v2 v2.2.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal v0.144.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sorairolake/lzip-go v0.3.8 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/therootcompany/xz v1.0.1 // indirect
	github.com/ulikunitz/xz v0.5.15 // indirect
	github.com/valyala/fastjson v1.6.7 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/collector/consumer v1.50.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.144.0 // indirect
	go.opentelemetry.io/collector/consumer/xconsumer v0.144.0 // indirect
	go.opentelemetry.io/collector/extension v1.50.0 // indirect
	go.opentelemetry.io/collector/extension/xextension v0.144.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.50.0 // indirect
	go.opentelemetry.io/collector/internal/componentalias v0.144.0 // indirect
	go.opentelemetry.io/collector/pdata v1.50.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.144.0 // indirect
	go.opentelemetry.io/collector/pipeline v1.50.0 // indirect
	go.opentelemetry.io/collector/pipeline/xpipeline v0.144.0 // indirect
	go.opentelemetry.io/collector/receiver/receiverhelper v0.144.0 // indirect
	go.opentelemetry.io/collector/receiver/xreceiver v0.144.0 // indirect
	go.opentelemetry.io/otel v1.39.0 // indirect
	go.opentelemetry.io/otel/metric v1.39.0 // indirect
	go.opentelemetry.io/otel/sdk v1.39.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.39.0 // indirect
	go.opentelemetry.io/otel/trace v1.39.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	go4.org v0.0.0-20230225012048-214862532bf5 // indirect
	golang.org/x/sys v0.39.0 // indirect
	gonum.org/v1/gonum v0.17.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251222181119-0a764e51fe1b // indirect
	google.golang.org/grpc v1.78.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
