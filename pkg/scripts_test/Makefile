include ../Makefile.Common

ifneq ($(OS),windows)
	GOTESTPREFIX ?= sudo env PATH="${PATH}"
endif

.PHONY: test
test:
	$(GOTESTPREFIX) $(GOTEST) ./...
