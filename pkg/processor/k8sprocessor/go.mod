module github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sprocessor

go 1.15

require (
	github.com/onsi/ginkgo v1.14.1 // indirect
	github.com/onsi/gomega v1.10.2 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/k8sconfig v0.35.0
	github.com/stretchr/testify v1.7.0
	go.opencensus.io v0.23.0
	go.opentelemetry.io/collector v0.36.0
	go.opentelemetry.io/collector/model v0.36.0
	go.uber.org/zap v1.19.1
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v0.22.1
)
