#################################################################################
# Path variables & stop implicit rules
#################################################################################

mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
current_dir := $(patsubst %/,%,$(dir $(mkfile_path)))
pkg_dir := $(abspath $(current_dir)/..)

Makefile:
	@true

$(pkg_dir)/common.mk:
	@true

#################################################################################
# Override variables & include makefile dependencies
#################################################################################

include $(pkg_dir)/common.mk

PROJECT_ROOT := $(shell git rev-parse --show-toplevel)
OTELCOL_BINARY := $(PROJECT_ROOT)/otelcolbuilder/cmd/otelcol-sumo

.PHONY: test-integration
test-integration:
	@if [ ! -f $(OTELCOL_BINARY) ]; then \
        echo "Binary not found, building..."; \
        $(MAKE) -C $(PROJECT_ROOT)/otelcolbuilder build; \
    fi
	go clean -testcache
	go test -tags=integration ./...
