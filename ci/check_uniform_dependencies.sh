#!/usr/bin/env bash

if [[ $(uname) == "Darwin" ]]; then
readonly GREP=ggrep
else
readonly GREP=grep
fi

function check(){
  local package
  readonly package="${1}"

  local VERSIONS
  VERSIONS="$( find . -name go.mod -print0 | \
    xargs -0 "${GREP}" -E --no-filename "${package} v" | \
    "${GREP}" -v module | sed 's| // indirect||' )"

  local VERSIONS_UNIQ
  VERSIONS_UNIQ="$( echo "${VERSIONS}" | sort | uniq )"

  local VERSIONS_COUNT
  VERSIONS_COUNT="$( echo "${VERSIONS_UNIQ}" | wc -l | awk '{$1=$1;print}' )"

  if [[ "${VERSIONS_COUNT}" != "1" ]]; then
    echo "There's more than one version of ${package} that this repo depends on"
    echo
    find . -name go.mod -print0 | \
      xargs -0 "${GREP}" -E "${package} v" | \
      "${GREP}" -v module
    exit 1
  else
    echo "OK: you only rely on \"${VERSIONS_UNIQ#replace }\""
  fi
}

check "go.opentelemetry.io/collector"
