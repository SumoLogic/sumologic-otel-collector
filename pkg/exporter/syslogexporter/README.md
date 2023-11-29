# Syslog Exporter

**Stability level**: Deprecated

This exporter is deprecated in favor of the [syslog exporter][syslog_exporter_contrib] that lives in the [OpenTelemetry Collector Contrib][contrib_repo] repository.
The functionality is the same but the configuration is slightly different.

To migrate, rename the following keys in configuration for `syslogexporter`:

- rename `protocol` property to `network`
- rename `format` property to `protocol`

For example, given the following configuration:

```yaml
  syslog:
    protocol: tcp
    port: 514
    endpoint: 127.0.0.1
    format: rfc5424
    tls:
      ca_file: ca.pem
      cert_file: cert.pem
      key_file: key.pem
```

change it to:

```yaml
  syslog:
    network: tcp
    port: 514
    endpoint: 127.0.0.1
    protocol: rfc5424
    tls:
      ca_file: ca.pem
      cert_file: cert.pem
      key_file:  key.pem
```

[syslog_exporter_contrib]: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/exporter/syslogexporter/README.md
[contrib_repo]: https://github.com/open-telemetry/opentelemetry-collector-contrib/

## About The Exporter

The syslog exporter supports sending messages to a remote syslog server.

- This exporter can forward syslog messages to a third party [syslog server][rsyslog] using [RFC5424][RFC5424] and [RFC3164][RFC3164].
- It also supports sending syslog messages to the [Cloud Syslog Source][CloudSyslogSource] configured on a Sumo Logic hosted collector
  using the [RFC5424][RFC5424] format, token required by [Cloud Syslog Source][CloudSyslogSource] can be added using [Logs Transform Processor][logstransform],
  please see [example configuration][configWithToken].
- It is recommended that this syslog exporter be used with the [syslog_parser][syslog_parser] configured in the receiver.
  This ensures that all the syslog message headers are populated with the expected values.
- Not using the `syslog_parser` will result in the syslog message being populated with default header values.

## Configuration

**The following are a few configuration options available to forward syslog messages**:

- `endpoint` - (default = `host.domain.com`) syslog endpoint (FQDN or IP address)
- `protocol` - (default = `tcp`) tcp/udp
- `port` - (default = `514`) A syslog port
- `format` - (default = `rfc5424`) rfc5424/rfc3164
  - `rfc5424` - Expects the syslog messages to be rfc5424 compliant
  - `rfc3164` - Expects the syslog messages to be rfc3164 compliant
- `tls` - configuration for TLS/mTLS
  - `insecure` (default = `false`) whether to enable client transport security, by default, TLS is enabled.
  - `cert_file` - Path to the TLS cert to use for TLS required connections. Should only be used if `insecure` is set to `false`.
  - `key_file` - Path to the TLS key to use for TLS required connections. Should only be used if `insecure` is set to `false`.
  - `ca_file` - Path to the CA cert. For a client this verifies the server certificate. For a server this verifies client certificates. If empty uses system root CA. Should only be used if `insecure` is set to `false`.
  - `insecure_skip_verify` -  (default = `false`) whether to skip verifying the certificate or not.
  - `min_version` (default = `1.2`) Minimum acceptable TLS version
  - `max_version` (default = `""` handled by [crypto/tls][cryptoTLS] - currently TLS 1.3) Maximum acceptable TLS version.
  - `reload_interval` - Specifies the duration after which the certificate will be reloaded. If not set, it will never be reloaded.

Please refer to the yaml below to configure the syslog exporter:

```yaml
extensions:
  file_storage/syslog:
    directory: .
    timeout: 10s

exporters:
  syslog:
    protocol: tcp
    port: 6514 # 514 (UDP)
    endpoint: 127.0.0.1 # FQDN or IP address
    tls:
      ca_file: certs/servercert.pem
      cert_file: certs/cert.pem
      key_file: certs/key.pem
    format: rfc5424 # rfc5424 or rfc3164


    # for below described queueing and retry related configuration please refer to:
    # https://github.com/open-telemetry/opentelemetry-collector/blob/main/exporter/exporterhelper/README.md#configuration
    retry_on_failure:
      # default = true
      enabled: true
      # time to wait after the first failure before retrying;
      # ignored if enabled is false, default = 5s
      initial_interval: 10s
      # is the upper bound on backoff; ignored if enabled is false, default = 30s
      max_interval: 40s
      # is the maximum amount of time spent trying to send a batch;
      # ignored if enabled is false, default = 120s
      max_elapsed_time: 150s

    sending_queue:
      # default = false
      enabled: true
      # number of consumers that dequeue batches; ignored if enabled is false,
      # default = 10
      num_consumers: 20
      # when set, enables persistence and uses the component specified as a storage extension for the persistent queue
      # make sure to configure and add a `file_storage` extension in `service.extensions`.
      # default = None
      storage: file_storage/syslog
      # maximum number of batches kept in memory before data;
      # ignored if enabled is false, default = 5000
      #
      # user should calculate this as num_seconds * requests_per_second where:
      # num_seconds is the number of seconds to buffer in case of a backend outage,
      # requests_per_second is the average number of requests per seconds.
      queue_size: 10000
receivers:
  filelog:
    start_at: beginning
    include:
    - /other/path/**/*.txt
    operators:
      - type: syslog_parser
        protocol: rfc5424 # the format used here must match the syslog exporter

service:
  telemetry:
      logs:
        level: "info"
  extensions:
    - file_storage/syslog
  pipelines:
    logs:
      receivers:
        - filelog
      exporters:
        - syslog
```

[rsyslog]: https://www.rsyslog.com/
[RFC5424]: https://www.rfc-editor.org/rfc/rfc5424
[RFC3164]: https://www.rfc-editor.org/rfc/rfc3164
[CloudSyslogSource]: https://help.sumologic.com/docs/send-data/hosted-collectors/cloud-syslog-source/
[logstransform]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/logstransformprocessor
[configWithToken]: ./examples/config_with_token.yaml
[syslog_parser]: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/pkg/stanza/docs/operators/syslog_parser.md
[cryptoTLS]: https://github.com/golang/go/blob/518889b35cb07f3e71963f2ccfc0f96ee26a51ce/src/crypto/tls/common.go#L706-L709
