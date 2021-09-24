# Source Processor

The `sourceprocessor` adds `_sourceName` and other tags related to Sumo Logic metadata taxonomy.

It is recommended to use `k8sprocessor` to provide attributes used in default values.

## Config

- `collector` (default = ``): name of the collector, put in `_collector` tag
- `source_name` (default = `%{k8s.namespace.name}.%{k8s.pod.name}.%{k8s.container.name}`): `_sourceName` template
- `source_category` (default = `%{k8s.namespace.name}/%{k8s.pod.pod_name}`): `_sourceCategory` template
- `source_category_prefix` (default = `kubernetes/`): prefix added before each `_sourceCategory` value
- `source_category_replace_dash` (default = `/`): character which all dashes (`-`) are being replaced to

### Filtering section

**NOTE**: The filtering is done on the resource level attributes.

- `exclude` (default = `{}`): a mapping of field names to exclusion regexes
  for those particular fields. Whenever a value under particular field matches
  a corresponding regex, the processed entry is dropped.

### Keys section

The following keys must match resource attributes.
In most cases the keys should be the same like in [k8sprocessor](../k8sprocessor/README.md#extract-section) config:

- `annotation_prefix` (default = `k8s.pod.annotation.`): prefix which allows to find given annotation;
it is used for including/excluding pods, among other attributes
- `pod_template_hash_key` (default = `k8s.pod.label.pod-template-hash`): attribute where pod template
hash is found (used for `pod` extraction)
- `pod_key` (default = `k8s.pod.name`): attribute where pod full name is found
- `source_host_key` (default = `k8s.pod.hostname`): attribute where source host is found

The following key is going to be created:

- `pod_name_key` (default = `k8s.pod.pod_name`): attribute where name portion of the pod is stored
during enrichment. Please consider following examples:

  - for a daemonset pod `dset-otelcol-sumo-xa314` it's going to be `dset-otelcol-sumo`
  - for a deployment pod `dep-otelcol-sumo-75675f5861-qasd2` it's going to be `dep-otelcol-sumo`
  - for a statefulset pod `st-otelcol-sumo-0` it's going to be `st-otelcol-sumo`

### Name translation and template keys

For example, when default template for `source_category` is being used (`%{k8s.namespace.name}/%{k8s.pod.pod_name}`),
the resource has attributes:

```yaml
k8s.namespace.name: my-namespace
k8s.pod.pod_name: some-name
```

Then the `_source_category` will contain: `my-namespace/some-name`

### Example config

```yaml
processors:
  source:
    collector: "mycollector"
    source_name: "%{k8s.namespace.name}.%{k8s.pod.name}.%{k8s.container.name}"
    source_category: "%{k8s.namespace.name}/%{k8s.pod.pod_name}"
    source_category_prefix: "kubernetes/"
    source_category_replace_dash: "/"
    exclude:
      namespace: "kube-system"
      pod: "custom-pod-.*"
```
