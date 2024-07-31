#!/usr/bin/env bash

ROOT_DIR="$(dirname "$(dirname "${0}")")"
readonly ROOT_DIR
# shellcheck disable=SC1090
source "${ROOT_DIR}"/ci/_build_functions.sh

fetch_current_branch
