# Raw Kubernetes Events Receiver

**Stability level**: Beta

Receiver for ingesting Kubernetes Events in their raw format, exactly as the Kubernetes API returns them.
It intends to return exactly the same output as the following [Fluentd plugin].

Supported pipeline types: logs

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

## Persistent Storage

If a storage extension is configured in the collector configuration's `service.extensions` property,
the `raw_k8s_events` receiver stores the latest resource version retrieved from the Kubernetes API server
so that after a restart, the receiver will continue retrieving events starting from that resource version
instead of using the `max_event_age` property.
This prevents the receiver from reporting duplicate events when the receiver is restarted in less than the `max_event_age` time.
On the other hand, this also allows the receiver to catch up on missed events in case it was not running for longer than `max_event_age` time.
Note that the default maximum age of events retained by the API Server is one hour - see the [--event-ttl][event_ttl] option of `kube-apiserver`.

Example configuration:

```yaml
extensions:
  file_storage:
    directory: .

receivers:
  raw_k8s_events:

service:
  extensions:
  - file_storage
  pipelines:
    logs:
      receivers:
      - raw_k8s_events
      exporters:
      - nop
```

[Fluentd plugin]: https://github.com/SumoLogic/sumologic-kubernetes-fluentd/tree/main/fluent-plugin-events
[event_ttl]: https://kubernetes.io/docs/reference/command-line-tools-reference/kube-apiserver/#options
