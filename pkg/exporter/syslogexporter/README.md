# Syslog Exporter

**Stability level**: Experimental

## About The Exporter

The syslog exporter supports sending messages to a remote syslog server.

- This exporter can forward syslog messages to a third party [syslog server](https://www.rsyslog.com/) using RFC5424 and RFC3164
- It also supports sending syslog messages to the [Cloud Syslog Source](https://help.sumologic.com/docs/send-data/hosted-collectors/cloud-syslog-source/) configured on a Sumo Logic hosted collector using the RFC5424 format, token required by Cloud Syslog Source can be added using logstransform processor, please see [example configuration](./examples/config_with_token.yaml)
- It is recommended that this syslog exporter be used with the [syslog_parser](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/pkg/stanza/docs/operators/syslog_parser.md) configured in the receiver. This ensures that all the syslog message headers are populated with the expected values
- Not using the `syslog_parser` will result in the syslog message being populated with default header values

`Note` - Syslog over UDP doesn't support certificate verification (ca_certificate).

## Configuration

**The following are a few configuration options available to forward syslog messages**:

- `endpoint` - syslog endpoint (FQDN or IP address)
- `protocol` - tcp/udp
- `port` - A syslog port
- `format` - rfc5424/rfc3164
  - `rfc5424` - Expects the syslog messages to be rfc5424 compliant
  - `rfc3164` - Expects the syslog messages to be rfc3164 compliant
- `ca_certificate` [tcp only] - A publicly verifiable server certificate (`note`: Self signed certificates are not supported in this version)

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
    ca_certificate: certs/servercert.pem # tcp only
    format: rfc5424 # RFC5424 or RFC3164

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
