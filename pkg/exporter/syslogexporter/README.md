# Syslog Exporter

Export/Forward syslog messages to a syslog server

**Stability level**: Experimental

This exporter can forward syslog messages to a third party [syslog server](https://www.rsyslog.com/)

This exporter/forwarder also supports sending syslog messages to the [Cloud Syslog](https://help.sumologic.com/docs/send-data/hosted-collectors/cloud-syslog-source/).

Configuration is specified via the yaml in the following structure:

```yaml
extensions:
  file_storage/syslog:
    directory: /tmp/otc
    timeout: 10s

exporters:
  syslog:
    protocol: tcp
    port: 514
    endpoint: 127.0.0.1
    ca_certificate: certs/servercert.pem
    format: rfc5424
    additional_structured_data: # only if messages are in RFC5424 format
    - tab=abc

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
        protocol: rfc5424

service:
  telemetry:
      logs:
        level: "debug"
  extensions:
    - file_storage/syslog
  pipelines:
    logs:
      receivers:
        - filelog
      exporters:
        - syslog
```

The following are a few configuration options available to forward syslog messages

- `endpoint` - syslog endpoint (FQDN or IP address)
- `protocol` - tcp/udp
- `port` - A syslog port
- `format` - rfc5424/rfc3164
  - `rfc5424` - Checks whether a syslog messages is compliant with RFC 5424
  - `rfc3164` - Checks whether a syslog messages is compliant with RFC 3164
- `additional_structured_data` - Additional [structured data](https://www.rfc-editor.org/rfc/rfc5424#page-15) to specify in the syslog message (Example: An authentication token)
- `ca_certificate` - A publicly verifiable server certificate (`note`: Self signed certificates are not supported in this version)
