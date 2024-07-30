include ../../Makefile.Common

# override test command for this package so the race detector doesn't run.
# the tests are flaky on Darwin, timing out after 10 minutes, and it's likely
# due to a concurrency bug interacting with the race detector in the test
# harness.
.PHONY: test
test:
	go test ./...
