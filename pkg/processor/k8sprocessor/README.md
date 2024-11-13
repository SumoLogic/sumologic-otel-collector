# Kubernetes Processor

**Stability level**: Beta

The `k8sprocessor` automatically tags logs, metrics and traces with Kubernetes metadata
like pod name, namespace name etc.

It automatically discovers k8s resources (pods), extracts metadata from them and adds the extracted
metadata to the records. The processor uses the Kubernetes API to discover all pods running
in a cluster, keeps a record of their IP addresses and interesting metadata. Upon receiving records,
the processor tries to identify the pod that sent the record and matches
it with the in-memory data. If a match is found, the cached metadata is added to the record as attributes.

## Configuration

```yaml
processors:
  k8s_tagger:
    # Limit page size when fetching pods from the k8s API
    # default: 200
    limit: 300

    # List of exclusion rules. For now it's possible to specify
    # a list of pod name regexes who's records should not be enriched with metadata.
    # default: []
    exclude:
      pods:
      - name: <pod_name_regex>

    # See "Extracting metadata" documentation section below
    extract:
      # List of rules to extract pod annotations into attributes.
      # See the "Field extract config" documentation section below for details on how to use it.
      # By default, no pod annotations are extracted into attributes.
      # default: []
      annotations:
      - key: "*"
        tag_name: k8s.pod.annotation.%s

      # If a pod is associated with more than one service, delimiter will be used to join the service names.
      # default: ", "
      delimiter: <delimiter>

      # List of rules to extract pod labels into attributes.
      # See the "Field extract config" documentation section below for details on how to use it.
      # By default, no pod labels are extracted into attributes.
      # default: []
      labels:
      - key: "*"
        tag_name: k8s.pod.label.%s

      # List of pod metadata to extract into attributes.
      # See "Extracting metadata" documentation section below for details.
      # default: []
      metadata:
      - container.id
      - container.image
      - container.name
      - k8s.cronjob.name
      - k8s.daemonset.name
      - k8s.deployment.name
      - host.name
      - k8s.job.name
      - k8s.namespace.name
      - k8s.node.name
      - k8s.pod.uid
      - k8s.pod.name
      - k8s.replicaset.name
      - k8s.service.name
      - k8s.statefulset.name
      - k8s.pod.startTime

      # List of rules to extract namespace annotations into attributes.
      # See the "Field extract config" documentation section below for details on how to use it.
      # By default, no namespace annotations are extracted into attributes.
      # default: []
      namespace_annotations:
      - key: "*"
        tag_name: k8s.namespace.annotation.%s

      # List of rules to extract namespace labels into attributes.
      # See the "Field extract config" documentation section below for details on how to use it.
      # By default, no namespace labels are extracted into attributes.
      # default: []
      namespace_labels:
      - key: "*"
        tag_name: k8s.namespace.label.%s

      # Specifies the names of the attributes to put the extracted metadata in.
      # See "Extracting metadata" documentation section below for details.
      # For example, if `deploymentName` exists in the `extract.metadata` list,
      # the name of the deployment will be put by default in an attribute named `k8s.deployment.name`.
      # The following map defines the defaults.
      # To override any of the defaults, specify a different attribute name for a selected key.
      tags:
        containerID: k8s.container.id
        containerImage: k8s.container.image
        containerName: k8s.container.name
        cronJobName: k8s.cronjob.name
        daemonSetName: k8s.daemonset.name
        deploymentName: k8s.deployment.name
        hostName: k8s.pod.hostname
        jobName: k8s.job.name
        namespaceName: k8s.namespace.name
        nodeName: k8s.node.name
        podID: k8s.pod.uid
        podName: k8s.pod.name
        replicaSetName: k8s.replicaset.name
        serviceName: k8s.service.name
        statefulSetName: k8s.statefulset.name
        startTime: k8s.pod.startTime

    # See "Filter section" documentation section below for details.
    filter:
      # Filters pods by pod fields.
      # default: []
      fields:
      - key: <key>
        op: {equals, not-equals}
        value: <value>

      # Filters pods by pod labels.
      # default: []
      labels:
      - key: <key>
        op: {equals, not-equals}
        value: <value>

      # Filters all pods by the provided namespace. All other pods are ignored.
      # default: ""
      namespace: <namespace>

      # If specified, any pods not running on the specified node will be ignored by the tagger.
      # default: ""
      node: <node_name>

      # Like `node`, but extracts the node name from the environment variable.
      # default: ""
      node_from_env_var: <env_var>

    # When set to true, fields such as `daemonSetName`, `replicaSetName`, `service`, etc.
    # can be extracted, though it requires fetching additional data to traverse the `owner` relationship.
    # See the "Extract" section for more information on which tags require the flag to be enabled.
    # default: false
    owner_lookup_enabled: {true, false}

    # When set to true, only annotates resources with the pod IP
    # and does not try to extract any other metadata.
    # It does not need access to the K8S cluster API.
    # Agent/Collector must receive records directly from services
    # to be able to correctly detect the pod IPs.
    # default: false
    passthrough: {true, false}
```

### Extracting metadata

The `extract` configuration section allows to specify rules to extract metadata from k8s pod specs.

The `extract.metadata` section defines which metadata should be retrieved.
The `extract.tags` section defines the names of the attributes that the metadata will be put in.

#### Specifying metadata attributes

The attributes are specified using the [Otel Resource Semantic Conventions for Kubernetes][k8s_semconv].
The following attributes are supported:

- `container.id`
- `container.image`
- `container.name`
- `k8s.cronjob.name`
- `k8s.daemonset.name`
- `k8s.deployment.name`
- `host.name`
- `k8s.job.name`
- `k8s.namespace.name`
- `k8s.node.name`
- `k8s.pod.uid`
- `k8s.pod.name`
- `k8s.replicaset.name`
- `k8s.service.name`
- `k8s.statefulset.name`
- `k8s.pod.startTime`

Some of the metadata is only extracted when the `owner_lookup_enabled` property is set to `true`.
The attributes that require this are:

- `k8s.cronjob.name`
- `k8s.daemonset.name`
- `k8s.deployment.name`
- `k8s.job.name`
- `k8s.replicaset.name`
- `k8s.service.name`
- `k8s.statefulset.name`

It's also possible to use the following legacy attribute names, though they will be deprecated at some point in the future:

- `containerId`
- `containerImage`
- `containerName`
- `cronJobName`
- `daemonSetName`
- `deploymentName`
- `hostName`
- `jobName`
- `namespace`
- `nodeName`
- `podId`
- `podName`
- `replicaSetName`
- `serviceName`
- `statefulSetName`
- `startTime`

### Field Extract Config

Allows specifying an extraction rule to extract a value from exactly one field.

The field accepts a list of maps accepting three keys: `tag_name`, `key` and `regex`

- `tag_name`: represents the name of the tag that will be added to the record.
  When not specified a default tag name will be used of the format:
  `k8s.annotations.<annotation key>`.
  For example, if `tag_name` is not specified and the key is `git_sha`,
  then the record name will be `k8s.annotations.git_sha`

- `key`: represents the annotation name. This must exactly match an annotation name.
  To capture all keys, `"*"` can be used

- `regex`: is an optional field used to extract a sub-string from a complex field value.
  The supplied regular expression must contain one named parameter with the string "value"
  as the name.
  For example, if your pod spec contains the following annotation,
  `kubernetes.io/change-cause: 2019-08-28T18:34:33Z APP_NAME=my-app GIT_SHA=58a1e39 CI_BUILD=4120`
  and you'd like to extract the GIT_SHA and the CI_BUILD values as tags, then you must specify
  the following two extraction rules:

  ```yaml
  processors:
    k8s_tagger:
      extract:
        annotations:
          - tag_name: git.sha
            key: kubernetes.io/change-cause
            regex: GIT_SHA=(?P<value>\w+)
          - tag_name: ci.build
            key: kubernetes.io/change-cause
            regex: CI_BUILD=(?P<value>[\w]+)
  ```

  this will add the `git.sha` and `ci.build` tags to the records. It is also possible to generically fetch
  all keys and fill them into a template. To substitute the original name, use `%s`. For example:

  ```yaml
  processors:
    k8s_tagger:
      extract:
        annotations:
          - tag_name: k8s.annotation/%s
            key: "*"
  ```

### Filter section

FilterConfig section allows specifying filters to filter pods by labels, fields, namespaces, nodes, etc.

- `node` (default = ""): represents a k8s node or host.
  If specified, any pods not running on the specified node will be ignored by the tagger.
- `node_from_env_var` (default = ""): can be used to extract the node name
  from an environment variable.
  The value must be the name of the environment variable.
  This is useful when the node a Otel agent will run on cannot be predicted.
  In such cases, the Kubernetes downward API can be used to add the node name
  to each pod as an environment variable.
  K8s tagger can then read this value and filter pods by it.
  For example, node name can be passed to each agent with the downward API as follows

  ```yaml
   env:
     - name: K8S_NODE_NAME
           valueFrom:
             fieldRef:
               fieldPath: spec.nodeName
  ```

  Then the NodeFromEnv field can be set to `K8S_NODE_NAME` to filter all pods by the node that the agent
  is running on. More on downward API here:
  https://kubernetes.io/docs/tasks/inject-data-application/downward-api-volume-expose-pod-information/
- `namespace` (default = ""): filters all pods by the provided namespace. All other pods are ignored.
- `fields` (default = empty): a list of maps accepting three keys: `key`, `value`, `op`.
  Allows to filter pods by generic k8s fields. Only the following operations (`op`)
  are supported: `equals`, `not-equals`.
  For example, to match pods having `key1=value1` and `key2<>value2` condition
  met for fields, one can specify:

  ```yaml
    fields:
     - key: key1 # `op` defaults to "equals" when not specified
       value: value1
     - key: key2
       value: value2
       op: not-equals
  ```

- `labels` (default = empty): a list of maps accepting three keys: `key`, `value`, `op`.
   Allows to filter pods by generic k8s pod labels.
   Only the following operations (`op`) are supported: `equals`, `not-equals`,
  `exists`, `not-exists`.
  For example, to match pods where `label1` exists, one can specify:

  ```yaml
    fields:
     - key: label1
       op: exists
  ```

### Example config

```yaml
processors:
  k8s_tagger:
    passthrough: false
    owner_lookup_enabled: true # To enable fetching additional metadata using `owner` relationship
    extract:
      metadata:
        # extract the following well-known metadata fields
        - containerId
        - containerName
        - containerImage
        - daemonSetName
        - deploymentName
        - hostName
        - namespace
        - nodeName
        - podId
        - podName
        - replicaSetName
        - serviceName
        - startTime
        - statefulSetName
      tags:
        # It is possible to provide your custom key names for each of the extracted metadata fields,
        # e.g. to store podName as "pod_name" rather than the default "k8s.pod.name", use following:
        podName: pod_name

      annotations:
        # Extract all annotations using a template
        - tag_name: k8s.annotation.%s
          key: "*"
      labels:
        - tag_name: l1 # extracts value of label with key `label1` and inserts it as a tag with key `l1`
          key: label1
        - tag_name: l2 # extracts value of label with key `label1` with regexp and inserts it as a tag with key `l2`
          key: label2
          regex: field=(?P<value>.+)
      delimiter: "_"

    filter:
      namespace: ns2 # only look for pods running in ns2 namespace
      node: ip-111.us-west-2.compute.internal # only look for pods running on this node/host
      node_from_env_var: K8S_NODE # only look for pods running on the node/host specified by the K8S_NODE environment variable
      labels: # only consider pods that match the following labels
       - key: key1 # match pods that have a label `key1=value1`. `op` defaults to "equals" when not specified
         value: value1
       - key: key2 # ignore pods that have a label `key2=value2`.
         value: value2
         op: not-equals
      fields: # works the same way as labels but for fields instead (like annotations)
       - key: key1
         value: value1
       - key: key2
         value: value2
         op: not-equals

    exclude:
      # Configure a list of exclusion rules. For now it's possible to specify
      # a list of pod name regexes who's records should not be enriched with metadata.
      #
      # By default this list is empty.
      pods:
        - name: jaeger-agent
        - name: my-agent
```

## RBAC

TODO: mention the required RBAC rules.

## Deployment scenarios

The processor supports running both in agent and collector mode.

### As an agent

When running as an agent, the processor detects IP addresses of pods sending records to the agent and uses this
information to extract metadata from pods and add to records. When running as an agent, it is important to apply
a discovery filter so that the processor only discovers pods from the same host that it is running on. Not using
such a filter can result in unnecessary resource usage especially on very large clusters. Once the fitler is applied,
each processor will only query the k8s API for pods running on it's own node.

Node filter can be applied by setting the `filter.node` config option to the name of a k8s node. While this works
as expected, it cannot be used to automatically filter pods by the same node that the processor is running on in
most cases as it is not know before hand which node a pod will be scheduled on. Luckily, kubernetes has a solution
for this called the downward API. To automatically filter pods by the node the processor is running on, you'll need
to complete the following steps:

1. Use the downward API to inject the node name as an environment variable.
  Add the following snippet under the pod env section of the OpenTelemetry container.

    ```yaml
       env:
       - name: KUBE_NODE_NAME
         valueFrom:
          fieldRef:
          apiVersion: v1
          fieldPath: spec.nodeName
    ```

    This will inject a new environment variable to the OpenTelemetry container with the value as the
    name of the node the pod was scheduled to run on.

1. Set "filter.node_from_env_var" to the name of the environment variable holding the node name.

    ```yaml
       k8s_tagger:
         filter:
           node_from_env_var: KUBE_NODE_NAME # this should be same as the var name used in previous step
    ```

    This will restrict each OpenTelemetry agent to query pods running on the same node only dramatically reducing
    resource requirements for very large clusters.

### As a collector

The processor can be deployed both as an agent or as a collector.

When running as a collector, the processor cannot correctly detect the IP address of the pods generating
the records when it receives the records from an agent instead of receiving them directly from the pods. To
workaround this issue, agents deployed with the k8s_tagger processor can be configured to detect
the IP addresses and forward them along with the record resources. Collector can then match this IP address
with k8s pods and enrich the records with the metadata. In order to set this up, you'll need to complete the
following steps:

1. Setup agents in passthrough mode

    Configure the agents' k8s_tagger processors to run in passthrough mode.

    ```yaml
    # k8s_tagger config for agent
    k8s_tagger:
      passthrough: true
    ```

    This will ensure that the agents detect the IP address as add it as an attribute to all records.
    Agents will not make any k8s API calls, do any discovery of pods or extract any metadata.

1. Configure the collector as usual

    No special configuration changes are needed to be made on the collector. It'll automatically detect
    the IP address of records sent by the agents as well as directly by other services/pods.

## Caveats

There are some edge-cases and scenarios where k8s_tagger will not work properly.

### Host networking mode

The processor cannot correct identify pods running in the host network mode and
enriching records generated by such pods is not supported at the moment.

### As a sidecar

The processor does not support detecting containers from the same pods when running
as a sidecar. While this can be done, we think it is simpler to just use the kubernetes
downward API to inject environment variables into the pods and directly use their values
as tags.

[k8s_semconv]: https://opentelemetry.io/docs/specification/otel/resource/semantic_conventions/k8s/
