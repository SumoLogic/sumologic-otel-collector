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
