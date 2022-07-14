# Sumo Logic Schema Processor

The Sumo Logic Schema processor (config name: `sumologic_schema`)
modifies the metadata on logs, metrics and traces sent to [Sumo Logic][sumologic_webpage]
so that the Sumo Logic [apps][sumologic_apps] can make full use of the ingested data.

Supported pipeline types: logs, metrics, traces.

[sumologic_webpage]: https://www.sumologic.com
[sumologic_apps]: https://www.sumologic.com/applications/

## Configuration

```yaml
processors:
  sumologic_schema:
    # Defines whether the `cloud.namespace` resource attribute should be added.
    # default = true
    add_cloud_namespace: {true,false}

    # Defines whether attributes should be translated
    # from OpenTelemetry to Sumo Logic conventions;
    # see "Attribute translation" documentation chapter from this document.
    # default = true
    translate_attributes: {true,false}
```

## Features

### Adding `cloud.namespace` resource attribute

Some of the apps in Sumo Logic require the `cloud.namespace` attribute to be set
to better understand the data coming from AWS EC2, AWS ECS and AWS Elactic Beanstalk.
This attribute is similar to the standard OpenTelemetry attribute [`cloud.provider`][opentelemetry_cloud_provider_attribute].
In the future, the Sumo Logic apps might switch to the standard `cloud.provider` attribute.
Before this happens, the following mapping defines the relationship between `cloud.provider` and `cloud.namespace` values:

|   `cloud.platform`    | `cloud.namespace` |
|:---------------------:|:-----------------:|
|        aws_ec2        |      aws/ec2      |
|        aws_ecs        |        ecs        |
| aws_elastic_beanstalk | ElasticBeanstalk  |

When this processor's `add_cloud_namespace` setting is set to `true`,
the processor looks for the above mentioned `cloud.platform` resource attribute values
and if found, adds the corresponding `cloud.namespace` resource attribute.

If the `cloud.platform` resource attribute is not found or has a value that is not in the table, nothing is added.

[opentelemetry_cloud_provider_attribute]: https://github.com/open-telemetry/opentelemetry-specification/blob/v1.9.0/specification/resource/semantic_conventions/cloud.md

### Attribute translation

Attribute translation changes some of the attribute keys from OpenTelemetry convention to Sumo Logic convention.
For example, OpenTelemetry convention for the attribute containing Kubernetes pod name is `k8s.pod.name`,
but Sumo Logic expects it to be in attribute named `pod`.

If attribute with target name eg. `pod` already exists,
translation is not being done for corresponding attribute (`k8s.pod.name` in this example).

This feature is turned on by default.
To turn it off, set the `translate_attributes` configuration option to `false`.
Note that this may cause some of Sumo Logic apps, built-in dashboards to not work correctly.

**Note**: the attributes are **not** translated for traces.

Below is a list of all attribute keys that are being translated.

| OTC key name              | Sumo Logic key name |
|---------------------------|---------------------|
| `cloud.account.id`        | `AccountId`         |
| `cloud.availability_zone` | `AvailabilityZone`  |
| `cloud.platform`          | `aws_service`       |
| `cloud.region`            | `Region`            |
| `host.id`                 | `InstanceId`        |
| `host.name`               | `host`              |
| `host.type`               | `InstanceType`      |
| `k8s.cluster.name`        | `Cluster`           |
| `k8s.container.name`      | `container`         |
| `k8s.daemonset.name`      | `daemonset`         |
| `k8s.deployment.name`     | `deployment`        |
| `k8s.namespace.name`      | `namespace`         |
| `k8s.node.name`           | `node`              |
| `k8s.service.name`        | `service`           |
| `k8s.pod.hostname`        | `host`              |
| `k8s.pod.name`            | `pod`               |
| `k8s.pod.uid`             | `pod_id`            |
| `k8s.replicaset.name`     | `replicaset`        |
| `k8s.statefulset.name`    | `statefulset`       |
| `service.name`            | `service`           |
| `log.file.path_resolved`  | `_sourceName`       |
