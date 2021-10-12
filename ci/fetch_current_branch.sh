#!/usr/bin/env bash

readonly ROOT_DIR="$(dirname "$(dirname "${0}")")"
# shellcheck disable=SC1090
source "${ROOT_DIR}"/ci/_build_functions.sh

fetch_current_branch
