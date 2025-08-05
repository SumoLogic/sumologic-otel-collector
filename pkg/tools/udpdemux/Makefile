#################################################################################
# Path variables & stop implicit rules
#################################################################################

mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
current_dir := $(patsubst %/,%,$(dir $(mkfile_path)))
parent_dir := $(abspath $(current_dir)/..)
pkg_dir := $(abspath $(parent_dir)/..)

Makefile:
	@true

$(pkg_dir)/common.mk:
	@true

#################################################################################
# Override variables & include makefile dependencies
#################################################################################

include $(pkg_dir)/common.mk

#################################################################################
# Targets
#################################################################################

.PHONY: build
build:
	go build -v -o udpdemux .

.PHONY: test
test: build
