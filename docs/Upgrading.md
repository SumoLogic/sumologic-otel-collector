# Upgrading

## Upgrading to 0.49.0-sumo-0

### Several changes to receivers using [opentelemetry-log-collection][opentelemetry-log-collection]

The affected receivers are: `filelog`, `syslog`, `tcplog`, `udplog`, and `journald`.

[opentelemetry-log-collection] is a shared library used by the aforementioned receivers. This release contains several breaking
changes to its configuration syntax.
Please refer to the [official upgrade guide][opentelemetry-log-collection-upgrade-guide] for more information.

[opentelemetry-log-collection]: https://github.com/open-telemetry/opentelemetry-log-collection
[opentelemetry-log-collection-upgrade-guide]: https://github.com/open-telemetry/opentelemetry-log-collection/releases/tag/v0.29.0
