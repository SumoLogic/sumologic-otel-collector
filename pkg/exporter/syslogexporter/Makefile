ifeq ($(OS),Windows_NT)
	OS=windows
endif

GOTEST=go test -count 1
GORUN = go run 
LINT=golangci-lint

.PHONY: test
test:
	$(GOTEST) .

.PHONY: run
run:
	$(GORUN) examples/syslogsend.go

.PHONY: lint
lint:
	$(LINT) run

.PHONY: mod-download-all
mod-download-all:
	go mod download all && go mod tidy -compat=1.18