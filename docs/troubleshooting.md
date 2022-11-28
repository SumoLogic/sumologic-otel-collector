# Troubleshooting

Make sure to read [OpenTelemetry Collector Troubleshooting][otc_troubleshooting] documentation for generic troubleshooting instructions.

See below for more information.

- [Accessing the collector's logs](#accessing-the-collectors-logs)
- [Accessing the collector's metrics](#accessing-the-collectors-metrics)
- [Accessing the collector's configuration](#accessing-the-collectors-configuration)
- [Collector registration failure](#collector-registration-failure)
- [Metrics port in use](#metrics-port-in-use)
- [Error in OTC logs: `Dropping data because sending_queue is full`](#error-in-otc-logs-dropping-data-because-sending_queue-is-full)

[otc_troubleshooting]: https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/troubleshooting.md

## Accessing the collector's logs

On systems with systemd, the logs are available in journald:

```sh
journalctl --unit otelcol-sumo
```

On systems without systemd, the logs are available in the console output of the running process.

## Accessing the collector's metrics

By default, the collector's own metrics are available in Prometheus format at `http://localhost:8888/metrics`.

To access them, use a tool like `curl` or just open the URL in a browser running on the same host as the collector.

```sh
curl http://localhost:8888/metrics
```

To modify the port, use the `service.telemetry.metrics.address` property:

```yaml
service:
  telemetry:
    metrics:
      address: ":8889"
```

## Accessing the collector's configuration

By default, the collector's configuration can be found in `/etc/otelcol-sumo/` directory.

## Collector registration failure

If you see a log containing `"token:invalid_token_format"` in the collector logs, similar to the following:

```console
2022-11-09T12:07:07.171+0100        warn        sumologicextension@v0.57.2-sumo-1/extension.go:423        Collector registration failed        {"kind": "extension", "name": "sumologic", "status_code": 401, "error_id": "DC0JU-XI3IY-Z703S", "errors": [{"code":"token:invalid_token_format","message":"The Sumo Logic credentials could not be verified."}]}
```

this means that the installation token used in the Sumo Logic extension's configuration is invalid.
Make sure the token was entered correctly.

## Metrics port in use

If you see a log containing `"listen tcp :8888: bind: address already in use"` in the collector logs, similar to the following:

```console
2022-11-17T11:22:17.733+0100    error   service/collector.go:156        Asynchronous error received, terminating process        {"error": "listen tcp :8888: bind: address already in use"}
```

this means that the `8888` port on the machine is busy.
It is possible that there is another collector already running on the host,
but it can also be any other process.
To find out what process is using the port, run the following command:

```sh
sudo lsof -i :8888
```

You can either stop the process using the port, or change the metrics port that your collector uses.
See [Accessing the collector's metrics](#accessing-the-collectors-metrics) section above.

## Error in OTC logs: `Dropping data because sending_queue is full`

If you see logs like the following in the output from OTC:

```console
otelcol_1   | 2022-02-11T12:10:06.755Z  warn    internal/persistent_storage.go:237      Maximum queue capacity reached{"kind": "exporter", "name": "sumologic", "queueName": "sumologic-metrics"}
otelcol_1   | 2022-02-11T12:10:06.755Z  error   exporterhelper/queued_retry.go:99       Dropping data because sending_queue is full. Try increasing queue_size.        {"kind": "exporter", "name": "sumologic", "dropped_items": 5}
otelcol_1   | 2022-02-11T12:24:42.437Z  warn    batchprocessor/batch_processor.go:185   Sender failed   {"kind": "processor", "name": "batch", "error": "sending_queue is full"}
```

This means that the `sumologicexporter` is not able to send data as quickly as it receives new data.
There may be a couple ways to fix this is, depending on the root cause.

If the problem is intermittent and caused by temporary spike in data volume,
increasing the queue size for the exporter with `sending_queue.queue_size` property
might be enough to accommodate temporary additional load.

If you see `429 Too Many Requests` HTTP response codes from Sumo in exporter logs,
this means the exporter is being throttled by Sumo backend.
In this case, you need to either decrease the volume of data sent,
or reach out to Sumo support to increase the quota.

If the exporter is not being throttled the best option might be
to increase the number of consumers sending data from queue to Sumo with `sending_queue.num_consumers`.

The `sumologicexporter` sends data to Sumo in batches.
If the batches are small, more requests need to be performed to send data.
If you have set the `max_request_body_size` setting to a low value, consider increasing it
to make batches bigger and in effect make sending more efficient.
