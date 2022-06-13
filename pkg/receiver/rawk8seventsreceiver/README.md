# Raw Kubernetes Events Receiver

Receiver for ingesting Kubernetes Events in their raw format, exactly as the Kubernetes API returns them.
It intends to return exactly the same output as the following [Fluentd plugin].

Supported pipeline types: logs

> :construction: This receiver is in **ALPHA**. Configuration fields, behaviour and log data model are subject to change.

## Configuration

```yaml
receivers:
  raw_k8s_events:
    # Namespaces to collect events from. Empty array means all namespaces and is the default.
    # default = []
    namespaces: []

    # Maximum age of collected events relative to receiver start time.
    # The default setting of 1 minute means that we're going to collect events up to 1 minute backwards in time.
    # default = 1m
    max_event_age: 1m

    # Authentication type for the Kubernetes client. Valid values are: `serviceAccount`, `kubeConfig`, `tls` and `none`.
    # default = `serviceAccount`
    auth_type: serviceAccount

    # The retry delay for recoverable errors from the rest of the pipeline.
    # Don't change this or the related setting below unless you know what you're doing.
    # default = 500ms
    consume_retry_delay: 500ms

    # The maximum number of retries for recoverable errors from the rest of the pipeline.
    # default = 20
    consume_max_retries: 20
```

The full list of settings exposed for this receiver are documented in
[config.go](./config.go).

[Fluentd plugin]: https://github.com/SumoLogic/sumologic-kubernetes-fluentd/tree/main/fluent-plugin-events
