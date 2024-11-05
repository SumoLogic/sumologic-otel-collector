# Updating journalctl

The [journaldreceiver] has `journalctl` as a dependency. It is included in our
Docker images. In order to update it, please update debian version in each of
the Dockerfile files. This can be done using the command below.

Please see available [debian versions][debian_versions].

```bash
export DEBIAN_VERSION=11.3
make update-journalctl
```

[debian_versions]: https://hub.docker.com/_/debian/?tab=description
[journaldreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.75.0/receiver/journaldreceiver
