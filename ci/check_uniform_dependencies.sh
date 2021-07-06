#!/usr/bin/env bash

if [[ $(uname) == "Darwin" ]]; then
readonly GREP=ggrep
else
readonly GREP=grep
fi

CORE_VERSIONS="$( find . -name go.mod -print0 | \
  xargs -0 "${GREP}" --no-filename github.com/SumoLogic/opentelemetry-collector | \
  "${GREP}" -v module )"
CORE_VERSIONS_UNIQ="$( echo "${CORE_VERSIONS}" | sort | uniq )"
CORE_VERSIONS_COUNT="$( echo "${CORE_VERSIONS_UNIQ}" | wc -l | awk '{$1=$1;print}' )"

if [[ "${CORE_VERSIONS_COUNT}" != "1" ]]; then
  echo "There's more than one version of github.com/SumoLogic/opentelemetry-collector that this repo depends on"
  echo
  find . -name go.mod -print0 | \
    xargs -0 "${GREP}" github.com/SumoLogic/opentelemetry-collector | \
    "${GREP}" -v module
  exit 1
else
  echo "OK: you only rely on \"${CORE_VERSIONS_UNIQ#replace }\""
fi
