# Troubleshooting

Make sure to read [OpenTelemetry Collector Troubleshooting][otc_troubleshooting] documentation for generic troubleshooting instructions.

See below for some specific scenarios.

[otc_troubleshooting]: https://github.com/open-telemetry/opentelemetry-collector/blob/main/docs/troubleshooting.md

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
