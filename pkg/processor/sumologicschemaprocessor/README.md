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
  sumologicschema:
    # Defines whether the `cloud.namespace` resource attribute should be added.
    # default = true
    add_cloud_namespace: {true,false}
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
