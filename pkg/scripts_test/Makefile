include ../Makefile.Common

ifneq ($(OS),windows)
	GOTESTPREFIX ?= sudo env PATH="${PATH}"
endif

GOTEST=go test
GOTESTBINARY=sumologic_scripts_tests.test

# We build the test binary separately to avoid downloading modules as root
.PHONY: test
test:
	$(GOTEST) -c
	$(GOTESTPREFIX) ./$(GOTESTBINARY) -test.v
