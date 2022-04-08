# Raw Kubernetes Events Receiver

Receiver for ingesting Kubernetes Events in their raw format, exactly as the Kubernetes API returns them. It intends
to return exactly the same output as the following [FluentD plugin].

Supported pipeline types: logs

> :construction: This receiver is in **ALPHA**. Configuration fields, behaviour and log data model are subject to change.

## Configuration

The following settings are optional:

- `namespaces` (default value is `[]`): Namespaces to collect events from. Empty array means all namespaces and is the
  default.
- `max_event_age` (default value is `1m`): Maximum age of collected events relative to receiver start time. The default
  setting of 1 minute means that we're going to collect events up to 1 minute backwards in time.
- `auth_type` (default value is `serviceAccount`): Authentication type for the Kubernetes client. Valid values are:
  `serviceAccount`, `kubeConfig`, `tls` and `none`.
- `consume_retry_delay` (default value is `500ms`): The retry delay for recoverable errors from the rest of the pipeline.
  Don't change this or the related setting below unless you know what you're doing.
- `consume_max_retries` (default value is `10`): The maximum number of retries for recoverable errors from the rest of
  the pipeline.

```yaml
receivers:
  raw_k8s_events:
    auth_type: serviceAccount
    namespaces: []
    max_event_age: 1m
    consume_max_retries: 10
    consume_retry_delay: 500ms
```

The full list of settings exposed for this receiver are documented in
[config.go](./config.go).

[FluentD plugin]: https://github.com/SumoLogic/sumologic-kubernetes-fluentd/tree/main/fluent-plugin-events
