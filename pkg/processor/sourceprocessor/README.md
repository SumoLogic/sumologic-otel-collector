# Source Processor

**Stability level**: Beta

The `sourceprocessor` adds `_sourceName` and other tags related to Sumo Logic metadata taxonomy.

It is recommended to use `k8sprocessor` to provide attributes used in default values.

## Configuration

```yaml
processors:
  source:
    # Name of the collector, put in `_collector` tag.
    # default: ""
    collector: <collector>

    # Template for source host, put in `_sourceHost` tag.
    # default: "%{k8s.pod.hostname}"
    source_host: <source_host>

    # Template for source name, put in `_sourceName` tag.
    # default: "%{k8s.namespace.name}.%{k8s.pod.name}.%{k8s.container.name}"
    source_name: <source_name>

    # Template for source category, put in `_sourceCategory` tag.
    # default: "%{k8s.namespace.name}/%{k8s.pod.pod_name}"
    source_category: <source_category>
    # Template added before each `_sourceCategory` value.
    # default: "kubernetes/"
    source_category_prefix: <source_category_prefix>
    # Character which all dashes ("-") in source category value are being replaced to.
    # default: "/"
    source_category_replace_dash: <source_category_replace_dash>

    # A mapping of resource attribute names to exclusion regexes for the attribute values.
    # Whenever a value under a particular attribute matches the corresponding regex,
    # the processed record is dropped.
    # default: {}
    exclude:
      <attribute_key_1>: <attribute_value_regex_1>
      <attribute_key_2>: <attribute_value_regex_2>

    # The processor assumes that pod annotations will be present as resource attributes,
    # one attribute per annotation, and that these attributes have a common prefix.
    # This setting controls the prefix.
    # default: "k8s.pod.annotation."
    annotation_prefix: <annotation_prefix>

    # The processor assumes that namespace annotations will be present as resource attributes,
    # one attribute per annotation, and that these attributes have a common prefix.
    # This setting controls the prefix.
    # default: "k8s.namespace.annotation."
    namespace_annotation_prefix: <namespace_annotation_prefix>

    # Name of the attribute that contains the full name of the pod.
    # default: "k8s.pod.name"
    pod_key: <pod_key>

    # Name of the attribute that will contain the deuniquified name of the pod.
    # Here are some examples of deuniquified pod names:
    # - for a daemonset pod `dset-otelcol-sumo-xa314` it's going to be `dset-otelcol-sumo`
    # - for a deployment pod `dep-otelcol-sumo-75675f5861-qasd2` it's going to be `dep-otelcol-sumo`
    # - for a statefulset pod `st-otelcol-sumo-0` it's going to be `st-otelcol-sumo`
    # default: "k8s.pod.pod_name"
    pod_name_key: <pod_name_key>

    # Name of the attribute that contains pod's template hash. It is used for pod name extraction.
    # default: "k8s.pod.label.pod-template-hash"
    pod_template_hash_key: <pod_template_hash_key>

    # See "Container-level pod annotations" section below
    container_annotations:
      # Specifies whether container-level annotations are enabled.
      # default: false
      enabled: {true, false}
      # Name of the attribute that contains the container name.
      # default: "k8s.container.name"
      container_name_key: <container_name_key>
      # List of prefixes for container-level pod annotations.
      # default: ["sumologic.com/"]
      prefixes:
      - <prefix_1>
      - <prefix_2>
```

## Source templates

You can specify a template with an attribute for `source_category`, `source_host`, `source_name`, using `%{attr_name}`.

For example, when there is an attribute `my_attr`: `my_value`, `metrics/%{my_attr}`
will be expanded to `metrics/my_value`.

If an attribute is not found, it is replaced with `undefined`.
For example, `%{existing_attr}/%{nonexistent_attr}` becomes `value-of-existing-attr/undefined`.

### Name translation and template keys

For example, when default template for `source_category` is being used (`%{k8s.namespace.name}/%{k8s.pod.pod_name}`),
the resource has attributes:

```yaml
k8s.namespace.name: my-namespace
k8s.pod.pod_name: some-name
```

and the default values for `source_category_prefix` and `source_category_replace_dash` are used (`kubernetes/` and `/`),
then the `_sourceCategory` attribute will contain: `kubernetes/my/namespace/some/name`

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

## Pod and namespace annotations

The following [Kubernetes annotations][k8s_annotations_doc] can be used on pods or namespace:

- `sumologic.com/exclude` - records from a pod/namespace that has this annotation set to
  `true` are dropped,

  **NOTE**: this has precedence over `sumologic.com/include` if both are set at
  the same time for one pod/namespace.

- `sumologic.com/include` - records from a pod/namespace that has this annotation set to
  `true` are not checked against exclusion regexes from `exclude` processor settings

- `sumologic.com/sourceCategory` - overrides `source_category` config option
- `sumologic.com/sourceCategoryPrefix` - overrides `source_category_prefix` config option
- `sumologic.com/sourceCategoryReplaceDash` - overrides `source_category_replace_dash` config option
- `sumologic.com/sourceHost` - overrides `source_host` config option;
  the value of this annotation will be set as the value of the `_sourceHost` resource attribute
- `sumologic.com/sourceName` - overrides `source_name` config option;
  the value of this annotation will be set as the value of the `_sourceName` resource attribute

For the processor to use them, the annotations need to be available as resource
attributes, prefixed with the value defined in `keys.annotation_prefix` config option.
This can be achieved with the [Kubernetes processor](../k8sprocessor).

For example, if a resource has the `k8s.pod.annotation.sumologic.com/exclude`
attribute set to `true`, the resource will be dropped.

*Note:** Pod annotations take precedence over namespace annotations.

[k8s_annotations_doc]: https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/

### Container-level pod annotations

To make it possible to set different metadata on logs from different containers inside a pod,
it is possible to set pod annotations that are container-specific.

The following rules apply:

- Container-level annotations take precedence over other forms of setting the source category.
- No other transformations are applied to the source categories retrieved from
  container-level annotations, like adding source category prefix or replacing the dash.

Let's look at an example. Assuming this plugin is configured with the following properties:

```yaml
processors:
  source:
    container_annotations:
      enabled: true
      prefixes:
      - sumologic.com/
```

and assuming there's a pod running that has containers named `container-name-1` and `container-name-2` in it,
setting the following annotations on the pod:

- `sumologic.com/container-name-1.sourceCategory` with the value of `first_source-category`
- `sumologic.com/container-name-2.sourceCategory` with the value of `another/source-category`

will make the logs from `container-name-1` be tagged with source category `first_source-category`
and logs from `container-name-2` be tagged with source category `another/source-category`.

If there is more than one prefix defined in `container_annotations.prefixes`,
they are checked in the order they are defined in. If an annotation is found for one prefix,
the other prefixes are not checked.
