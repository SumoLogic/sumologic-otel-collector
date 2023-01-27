# Syslog Exporter

Export logs to a syslog server

## Build

Build a binary using the [otelcolbuilder](../../../otelcolbuilder)

## Run the binary

### Configuration

```yaml
exporters:
  syslog:
receivers:
  filelog:
    include:
    - /tmp/test.log
service:
  pipelines:
    logs:
      receivers: 
        - filelog
      exporters:
        - syslog
```

### Execute

```yaml
./otelcol-sumo --config config.yaml
```

Running the above command sends messages from the file `/tmp/test.log` to a syslog server running on the localhost at `127.0.0.1:514`
