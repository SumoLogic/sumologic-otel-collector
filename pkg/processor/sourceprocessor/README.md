# Source Processor

The `sourceprocessor` adds `_source` and other tags related to Sumo Logic metadata taxonomy.

It leverages data tagged by `k8sprocessor` and must be after it in the processing chain.
It has certain expectations on the label names used by `k8sprocessor` which might be configured below.

## Config

- `collector` (default = ``): name of the collector, put in `_collector` tag
- `source` (default = `traces`): name of the source, put in `_source` tag
- `source_name` (default = `%{namespace}.%{pod}.%{container}`): `_sourceName` template
- `source_category` (default = `%{namespace}/%{pod_name}`): `_sourceCategory` template
- `source_category_prefix` (default = `kubernetes/`): prefix added before each `_sourceCategory` value
- `source_category_replace_dash` (default = `/`): character which all dashes (`-`) are being replaced to

### Filtering section

**NOTE**: The filtering is done on the resource level attributes.

- `exclude` (default = `{}`): a mapping of field names to exclusion regexes
  for those particular fields. Whenever a value under particular field matches
  a corresponding regex, the processed entry is dropped.

  **NOTE**:

  When systemd related filtering is taking place (`exclude` contains
  an entry for `_SYSTEMD_UNIT`) then whenever the processed record contains
  `_HOSTNAME` attribute it will be added to the resulting record under `host` key.

### Keys section

The following keys must match resource attributes.
In most cases the keys should be the same like in [k8sprocessor](../k8sprocessor/README.md#extract-section) config:

- `annotation_prefix` (default = `k8s.pod.annotation.`): prefix which allows to find given annotation;
it is used for including/excluding pods, among other attributes
- `pod_template_hash_key` (default = `k8s.pod.label.pod-template-hash`): attribute where pod template
hash is found (used for `pod` extraction)
- `namespace_key` (default = `k8s.namespace.name`): attribute where namespace name is found
- `pod_key` (default = `k8s.pod.name`): attribute where pod full name is found
- `container_key` (default = `k8s.container.name`): attribute where container name is found
- `source_host_key` (default = `k8s.pod.hostname`): attribute where source host is found

The following key is going to be created:

- `pod_name_key` (default = `k8s.pod.pod_name`): attribute where name portion of the pod is stored
during enrichment. Please consider following examples:

  - for a daemonset pod `dset-otelcol-sumo-xa314` it's going to be `dset-otelcol-sumo`
  - for a deployment pod `dep-otelcol-sumo-75675f5861-qasd2` it's going to be `dep-otelcol-sumo`
  - for a statefulset pod `st-otelcol-sumo-0` it's going to be `st-otelcol-sumo`

### Name translation and template keys

The key names provided as `namespace`, `pod`, `pod_name`, `container` in templates for `source_category`
or `source_name`are replaced with the key name provided in `namespace_key`, `pod_key`,
`pod_name_key`, `container_key` respectively.

For example, when default template for `source_category` is being used (`%{namespace}/%{pod_name}`) and
`namespace_key=k8s.namespace.name`, the resource has attributes:

```yaml
k8s.namespace.name: my-namespace
pod_name: some-name
```

Then the `_source_category` will contain: `my-namespace/some-name`

### Example config

```yaml
processors:
  source:
    collector: "mycollector"
    source_name: "%{namespace}.%{pod}.%{container}"
    source_category: "%{namespace}/%{pod_name}"
    source_category_prefix: "kubernetes/"
    source_category_replace_dash: "/"
    exclude:
      namespace: "kube-system"
      pod: "custom-pod-.*"
```
