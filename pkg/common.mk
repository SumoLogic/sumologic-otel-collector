#################################################################################
# Path variables & stop implicit rules
#################################################################################

mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
current_dir := $(patsubst %/,%,$(dir $(mkfile_path)))
parent_dir := $(abspath $(current_dir)/../)

common.mk:
	@true

$(parent_dir)/common.mk:
	@true

#################################################################################
# Override variables & include makefile dependencies
#################################################################################

# Include common.mk from the parent directory
include $(parent_dir)/common.mk

#################################################################################
# Variables
#################################################################################

GOFLAGS ?= -race
GOTEST := go test $(GOFLAGS)
LINT := staticcheck

#################################################################################
# Targets
#################################################################################

.PHONY: test
test:
	$(GOTEST) ./...

.PHONY: fmt
fmt:
	gofmt -w -s ./
	goimports -w  -local github.com/open-telemetry/opentelemetry-collector-contrib ./

.PHONY: lint
lint:
	$(LINT) .

.PHONY: mod-download-all
mod-download-all:
	go mod download all && go mod tidy
