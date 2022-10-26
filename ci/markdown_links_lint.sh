#!/usr/bin/env bash

# The purpose of this script is to make sure all links pointing to this repository are relative.

if [[ $(uname) == "Darwin" ]]; then
readonly GREP=ggrep
else
readonly GREP=grep
fi

# Get all markdown files
FILES=$(find . -type f -name '*.md')
readonly FILES

RET_VAL=0

for file in ${FILES}; do
    # '\[[^\]]*\]\([^\)]*\)' - get all markdown links [*](*)
    # filter in only linked to this repository
    # filter out all links pointing to specific release, tag or commit
    # filter out links ended with /releases
    # filter out links to CI badges
    if ${GREP} -HnoP '\[[^\]]*\]\([^\)]*\)' "${file}" \
        | ${GREP} -i 'github\.com\/sumologic\/sumologic-otel-collector' \
        | ${GREP} -vP '(\/(blob|tree)\/(v\d+\.|[a-f0-9]{40}\/|release\-))' \
        | ${GREP} -vP '\/releases\)' \
        | ${GREP} -vP '\/badge.svg\)' \
        ; then
        # Set RET_VAL to 1 if grep was successful (found something)
        RET_VAL=1
    fi
done

exit "${RET_VAL}"
