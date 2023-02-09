# Best Practices

## Using batch processor to batch data

[Batch processor][batchprocessor] can be used to transform processed data into batches greater than given size or from given time interval.
This helps better compress the data and reduce the number of requests sent by the exporters.

It is highly recommended to use this processor in every pipeline. It should be defined after [memory limiter processor][memorylimiterprocessor]
and any processors that drop the data, such as [filter processor][filterprocessor].

Besides setting the lower limit for batch size, it is also possible to set a maximal size for a batch.
We highly recommend to set that limit to avoid sudden increase of request sizes in case more data is being received temporarily.
The value we recommend to set is `2 * send_batch_size`.

Overall, we recommend the following default configuration for this processor:

```yaml
batch:
  send_batch_size: 1_024
  timeout: 1s
  send_batch_max_size: 2_048 ## = 2 * 1024
```

**NOTE**: when using [Sumo Logic exporter][sumologicexporter] and sending data that is **not** in otlp format,
you can explicitly limit size of the requests in bytes using config option `max_request_body_size`.

[batchprocessor]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.71.0/processor/batchprocessor
[memorylimiterprocessor]: https://github.com/open-telemetry/opentelemetry-collector/tree/v0.71.0/processor/memorylimiterprocessor
[filterprocessor]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/filterprocessor
[sumologicexporter]: ../pkg/exporter/sumologicexporter
