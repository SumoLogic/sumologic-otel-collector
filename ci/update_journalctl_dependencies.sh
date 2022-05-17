#!/usr/bin/env bash

set -eo pipefail

if [[ -z "${DEBIAN_VERSION}" ]]; then
    echo "No DEBIAN_VERSION passed in, exiting"
    exit -1
fi

# use gsed for macos
if [[ "$(go env GOOS)" == 'darwin' ]]; then
  readonly SED=gsed
else
  readonly SED=sed
fi

# prepare command to get list of dependencies
# crucial part is `ldd /bin/journalctl | grep -oP "\/.*? "`
readonly COMMAND='apt update && apt install -y systemd && for i in $(ldd /bin/journalctl | grep -oP "\/.*? " | sort); do echo -n "COPY --from=systemd ${i} ${i}\n"; done'
# run command in docker and extract only lines we are interested in
readonly COPY_LINES_WITH_NEW_LINE="$(docker run --rm -it "debian:${DEBIAN_VERSION}" bash -c "${COMMAND}" | grep -P "^COPY")"
# remove last newline
readonly COPY_LINES="$(echo "${COPY_LINES_WITH_NEW_LINE}" | sed 's/\\n$//g')"

# update Dockerfiles
for file in "Dockerfile" "Dockerfile_dev" "Dockerfile_local"; do
  # update debian version
  "${SED}" -i "s/FROM debian:.* as systemd/FROM debian:${DEBIAN_VERSION} as systemd/" "${file}"
  # remove old entries
  "${SED}" -i '/^COPY --from=systemd.*$/d' "${file}"
  # add new entries after `# journaldreceiver dependencies` line
  "${SED}" -i "/# journaldreceiver dependencies/a\\COPY --from=systemd /bin/journalctl /bin/journalctl\n${COPY_LINES}" "${file}"
done
