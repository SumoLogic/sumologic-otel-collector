#!/usr/bin/env bash

if ! markdown-link-check --help >/dev/null 2>&1 ; then
    echo "markdown-link-check not found, please install it with 'npm install -g markdown-link-check'"
    exit 1
fi

# Get all markdown files
FILES=$(find . -type f -name '*.md')
readonly FILES

RET_VAL=0

for file in ${FILES}; do
    if ! markdown-link-check --progress --config .markdown_link_check.json "${file}"; then
        RET_VAL=1
    fi
done

set -x
exit "${RET_VAL}"
