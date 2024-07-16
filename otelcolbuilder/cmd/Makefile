# Workaround for loadbalancing exporter having an incorrectly formatted config tag and failing the test
FLAGS=-skip '^TestValidateConfigs$$'

.PHONY: test
test:
	go test $(FLAGS) -trimpath -parallel 1 ./...

.PHONY: test-v
test-v:
	@$(MAKE) test FLAGS=-v

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
