module github.com/SumoLogic/sumologic-otel-collector/pkg/exporter/sumologicexporter

go 1.19

require (
	github.com/SumoLogic/sumologic-otel-collector/pkg/extension/sumologicextension v0.77.0-sumo-0
	github.com/klauspost/compress v1.16.5
	github.com/stretchr/testify v1.8.2
	go.opencensus.io v0.24.0
	go.opentelemetry.io/collector v0.78.2
	go.opentelemetry.io/collector/component v0.78.2
	go.opentelemetry.io/collector/consumer v0.78.2
	go.opentelemetry.io/collector/exporter v0.78.2
	go.opentelemetry.io/collector/pdata v1.0.0-rcv0012
	go.uber.org/multierr v1.11.0
	go.uber.org/zap v1.24.0
	golang.org/x/exp v0.0.0-20221004215720-b9f4876ce741
)

require (
	github.com/Showmax/go-fqdn v1.0.0 // indirect
	github.com/cenkalti/backoff/v4 v4.2.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-logr/logr v1.2.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/knadh/koanf v1.5.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-ps v1.0.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/rs/cors v1.9.0 // indirect
	github.com/shirou/gopsutil/v3 v3.23.4 // indirect
	github.com/shoenig/go-m1cpu v0.1.5 // indirect
	github.com/tklauser/go-sysconf v0.3.11 // indirect
	github.com/tklauser/numcpus v0.6.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	go.opentelemetry.io/collector/confmap v0.78.2 // indirect
	go.opentelemetry.io/collector/featuregate v1.0.0-rcv0012 // indirect
	go.opentelemetry.io/collector/receiver v0.78.2 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.41.1 // indirect
	go.opentelemetry.io/otel v1.15.1 // indirect
	go.opentelemetry.io/otel/metric v0.38.1 // indirect
	go.opentelemetry.io/otel/trace v1.15.1 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	golang.org/x/net v0.10.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	google.golang.org/genproto v0.0.0-20230410155749-daa745c078e1 // indirect
	google.golang.org/grpc v1.55.0 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/SumoLogic/sumologic-otel-collector/pkg/extension/sumologicextension => ./../../extension/sumologicextension
