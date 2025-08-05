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

# Override GOFLAGS for this package so the race detector doesn't run. The tests
# are flaky on Darwin, timing out after 10 minutes, and it's likely due to a
# concurrency bug interacting with the race detector in the test harness.
#GOFLAGS :=

include $(pkg_dir)/common.mk
