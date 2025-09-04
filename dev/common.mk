#################################################################################
# Path variables
#################################################################################

mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
mkfile_dir := $(patsubst %/,%,$(dir $(mkfile_path)))

#################################################################################
# Command name variables
#################################################################################

CURL ?= curl
CUT ?= cut
DOCKER ?= docker
GIT ?= git
GREP ?= grep

#################################################################################
# OpenTelemetry Contrib variables
#################################################################################

CONTRIB_ORG ?= open-telemetry
CONTRIB_NAME ?= opentelemetry-collector-contrib
CONTRIB_REPO ?= https://github.com/$(CONTRIB_ORG)/$(CONTRIB_NAME)
CONTRIB_DIR ?= $(mkfile_dir)/contrib
CONTRIB_SENTINEL ?= $(CONTRIB_DIR)/.sentinel

#################################################################################
# Other variables
#################################################################################

PROMETHEUS_ENDPOINT ?= http://localhost:8889/metrics

DOCKER_COMPOSE_CMD ?= $(DOCKER) compose
METRICS_CMD ?= $(CURL) -sS "$(PROMETHEUS_ENDPOINT)"

#################################################################################
# Targets
#################################################################################

.PHONY: help
help:
	@echo "# Clones or updates the OpenTelemetry Collector Contrib repository"
	@echo make contrib
	@echo
	@echo "# Spin up the Docker Compose environment"
	@echo make up
	@echo
	@echo "# Spin down the Docker Compose environment"
	@echo make down
	@echo
	@echo "# Show Docker Compose logs"
	@echo make logs
	@echo
	@echo "# Show Prometheus metrics"
	@echo make metrics
	@echo
	@echo "# Show Prometheus metrics with value of zero"
	@echo make zero-metrics

# NOTE: This target will clone the contrib repository and create a sentinel
# file. It will only run if the sentinel file does not exist or has been
# modified.
$(CONTRIB_SENTINEL):
	@echo Cloning the contrib repository...
	@$(GIT) clone "$(CONTRIB_REPO)" "$(CONTRIB_DIR)" || \
	(echo "Please manually remove the contrib dir and try again." >&2 && exit 1)
	@touch "$(CONTRIB_SENTINEL)"

.PHONY: contrib
contrib: $(CONTRIB_SENTINEL)
contrib:
	@echo "Updating the contrib repository in: $(CONTRIB_DIR)"
	@cd "$(CONTRIB_DIR)" && $(GIT) fetch origin

.PHONY: up
up:
	@$(DOCKER_COMPOSE_CMD) up -d --build

.PHONY: down
down:
	@$(DOCKER_COMPOSE_CMD) down

.PHONY: downup
downup: down
downup: up


.PHONY: logs
logs:
	@$(DOCKER_COMPOSE_CMD) logs

.PHONY: logs-follow
logs-follow:
	@$(DOCKER_COMPOSE_CMD) logs -f

.PHONY: metrics
metrics:
	@$(METRICS_CMD)

.PHONY: zero-metrics
zero-metrics:
	@$(METRICS_CMD) | $(GREP) -v "^#" | $(GREP) -e " 0$$" | $(CUT) -f 1 -d '{'
