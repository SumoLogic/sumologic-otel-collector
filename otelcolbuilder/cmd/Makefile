# Workaround for loadbalancing exporter having an incorrectly formatted config tag and failing the test
FLAGS=-skip '^TestValidateConfigs$$'

GOTEST=go test -trimpath -parallel 1

.PHONY: test
test:
	$(GOTEST) ./...

.PHONY: test-v
test-v:
	$(GOTEST) -v ./...

.PHONY: test-junit
test-junit: $(eval SHELL:=/bin/bash)
	$(GOTEST) -v ./... | tee >(go-junit-report -set-exit-code > report.xml)

# Don't lint the code generated with opentelemetry-collector-builder as it
# produces code that's not gofmt-ed:
#
# $ golangci-lint run --skip-files main.go .
# components.go:1: File is not `gofmt`-ed with `-s` (gofmt)
# $ golangci-lint run --skip-files components.go .
# main.go:1: File is not `gofmt`-ed with `-s` (gofmt)
.PHONY: lint
lint:
	@echo -n ""
