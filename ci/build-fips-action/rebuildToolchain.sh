#!/usr/bin/env bash

# builds and archives two gcc toolchains: aarch64-linux-musl and x86_64-linux-musl.
# Saves them in a directory named by the host architecture.

DOCKER="${DOCKER_CMD:-docker}"

"$DOCKER" build --build-arg PARALLEL="$(nproc)" --target archive -t toolchain-archive .;
"$DOCKER" run -v "$PWD":/save toolchain-archive:latest sh -c 'cp -rf $(apk --print-arch) /save'
