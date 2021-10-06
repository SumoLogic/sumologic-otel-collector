.PHONY: test
test:
	go test -count 1 -parallel 1 -tags enable_unstable ./...

# Don't lint the code generated with opentelemetry-collector-builder as it
# produces code that's not gofmt-ed:
# 
# $ golangci-lint run --skip-files main.go .
# components.go:1: File is not `gofmt`-ed with `-s` (gofmt)
# $ golangci-lint run --skip-files components.go .
# main.go:1: File is not `gofmt`-ed with `-s` (gofmt)
.PHONY: lint
lint:
	golangci-lint run --skip-files '(components|main).go' .
