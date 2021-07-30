module github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sprocessor

go 1.15

require (
	github.com/hashicorp/go-msgpack v0.5.5 // indirect
	github.com/mattn/go-colorable v0.1.7 // indirect
	github.com/onsi/ginkgo v1.14.1 // indirect
	github.com/onsi/gomega v1.10.2 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig v0.27.0
	github.com/stretchr/testify v1.7.0
	go.opencensus.io v0.23.0
	go.opentelemetry.io/collector v0.30.1
	go.opentelemetry.io/collector/model v0.31.0
	go.uber.org/zap v1.18.1
	k8s.io/api v0.21.1
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v0.21.1
)

replace go.opentelemetry.io/collector => github.com/SumoLogic/opentelemetry-collector v0.30.1-sumo-1
