# Monitoring Job Receiver

| Status    |               |
| --------- | ------------- |
| Stability | [alpha]: logs |

This receiver makes it possible to collect telemetry data from sources that
do not instrument well. The monitoring job receiver executes a script or
executable at defined intervals, and propagates the output from that process
as log events. In addition, the monitoring job receiver simplifies the process
of downloading runtime assets necessary to run a particular monitoring job.

## Configuration

| Configuration | Default  | Description                                                                    |
| ------------- | -------- | ------------------------------------------------------------------------------ |
| exec          | required | A `exec` configuration block. See details [below](#execution-configuration)    |
| schedule      | required | A `schedule` configuration block. See details [below](#schedule-configuration) |
| output        |          | An `output` configuration block. See details [below](#output-configuration)    |

### Execution Configuration

| Configuration  | Default  | Description                                                                                      |
| -------------- | -------- | ------------------------------------------------------------------------------------------------ |
| command        | required | The `command` to run. Should start a binary that writes to stdout and/or stderr                  |
| arguments      |          | A list of string arguments to pass the command                                                   |
| timeout        |          | [Time](#time-parameters) to wait for the process to exit before attempting to make it exit       |
| runtime_assets |          | A list of `runtime_assets` required for the monitoring job. See details [below](#runtime-assets) |

### Schedule Configuration

The scheduling configuration block currently only supports a single **required**
`interval` [Time](#time-parameters) parameter. Counting from collector startup, the command will be
scheduled every `interval`.

### Output Configuration

The monitoringjob receiver output is a configurable as a [Stanza][stanza]
pipeline. This allows complex operators to be chained to parse the command
output.

| Configuration                       | Default | Description                                                                                                                                                                                                       |
| ----------------------------------- | ------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `type`                              | `event` | The output handler type [see below](#output-handler-type). Valid values are `event` and `log_entries`.                                                                                                            |
| `event`                             | {}      | When `type` == `event` this block may be configured with additional properties as described [below](#output-handler-type).                                                                                        |
| `log_entries`                       | {}      | When `type` == `log_entries` this block may be configured with additional properties as described [below](#output-handler-type).                                                                                  |
| `operators`                         | []      | An array of [stanza][stanza] operators to act on the output.                                                                                                                                                      |
| `retry_on_failure.enabled`          | `false` | If `true`, the receiver will pause reading a file and attempt to resend the current batch of logs if it encounters an error from downstream components.                                                           |
| `retry_on_failure.initial_interval` | `1s`    | [Time](#time-parameters) to wait after the first failure before retrying.                                                                                                                                         |
| `retry_on_failure.max_interval`     | `30s`   | Upper bound on retry backoff [interval](#time-parameters). Once this value is reached the delay between consecutive retries will remain constant at the specified value.                                          |
| `retry_on_failure.max_elapsed_time` | `5m`    | Maximum amount of [time](#time-parameters) (including retries) spent trying to send a logs batch to a downstream consumer. Once this value is reached, the data is discarded. Retrying never stops if set to `0`. |
| `attributes`                        | {}      | A map of `key: value` pairs to add to the entry's attributes.                                                                                                                                                     |
| `resource`                          | {}      | A map of `key: value` pairs to add to the entry's resource.                                                                                                                                                       |

#### Output Handler Type

The monitoringjob receiver can feed output into the stanza pipeline in one of
two ways depending on the configured `type`.

By default, the `type: 'event'` output handler is used. When configured with
the event output type, command output will be buffered until the process exits,
at which time the receiver will emit a single structured event capturing the
command output and optionally metadata about the execution such as exit code
and duration. This is useful for true monitoring jobs, such as using a
monitoring plugin not supported natively by OTEL.

When configured with the `log_entries` output type, command output will be
emitted as distinct log entries line-by-line. This can be used to act as a
bridge to third party logs and other log events that may not be straightforward
to access.

[stanza]: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/pkg/stanza/docs/operators/README.md

#### Output Configuration - event

| Configuration            | Default | Description                                                                                       |
| ------------------------ | ------- | ------------------------------------------------------------------------------------------------- |
| include_command_name     | true    | When set, includes the attribute `command.name` in the log event.                                 |
| include_command_status   | true    | When set, includes the attribute `command.status` in the log event.                               |
| include_command_duration | true    | When set, includes the attribute `command.duration` in the log event.                             |
| max_body_size            |         | When set, restricts the length of command output to a specified [ByteSize](#bytesize-parameters). |

#### Output Configuration - log_entries

| Configuration                | Default | Description                                                                                                                         |
| ---------------------------- | ------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| include_command_name         | true    | When set, includes the attribute `command.name` in the log event.                                                                   |
| include_stream_name          | true    | When set, includes the attribute `command.stream.name` in the log event. Indicating `stdin` or `stderr` as the origin of the event. |
| max_log_size                 |         | When set, restricts the length of a log entry to a specified [ByteSize](#bytesize-parameters).                                      |
| encoding                     | utf-8   | Encoding to expect from the command output. Used to detect log entry boundaries.                                                    |
| multiline                    |         | Used to override the default newline delimited log entries.                                                                         |
| multiline.line_start_pattern |         | Regex pattern for the beginning of a log entry. Mutualy exclusive with end pattern.                                                 |
| multiline.line_end_pattern   |         | Regex pattern for the ending of a log entry. Mutualy exclusive with start pattern.                                                  |

### Runtime Assets

Runtime assets can be specified on a monitoring job to make it easier to
distribute monitoring instrumentation to collectors. When a runtime asset is
specified on a monitoring job receiver, the collector will download and install
the asset locally before the monitoring job is ran. This eliminates the need to
pre-install monitoring instrumentation alongside the collector.

Runtime assets are packaged according to the [sensu-go asset
specification][sensu-go-assets-spec], allowing users to take advantage of the
existing asset definitions in [Bonsai](https://bonsai.sensu.io/), sensu-go's
asset hub. Packaging an asset yourself is relatively straight-forward. Simply
serve a tar archive (optionally gzipped) containing a `bin` directory
containing one or multiple executables. The asset will be unpacked and the
monitoring job will be executed with `./bin` included in the `PATH` environment
variable, `./lib` included in `LD_LIBRARY_PATH` and `./include` in `CPATH`.

[sensu-go-assets-spec]: https://docs.sensu.io/sensu-go/latest/plugins/assets/#dynamic-runtime-asset-format-specification

| Configuration | Default  | Description                       |
| ------------- | -------- | --------------------------------- |
| name          | required | Name for the asset.               |
| url           | required | HTTP URL used to fetch the asset. |
| sha512        | required | SHA512 hash of the asset archive. |

## Additional Features

### Time Parameters

Time parameters are expressed as go duration strings as defined by
[time.ParseDuration](https://pkg.go.dev/time#ParseDuration).

Examples: `60s`, `45m`, `2h30m40s`.

### ByteSize Parameters

ByteSize parameters can be specified either as an integer number or as a string
starting with decimal numbers optionally followed by a common byte unit
prefixes. e.g. `16kb`, `32MiB`, `1GB`.

## Example configuration

```yaml
monitoringjob:
  schedule:
    interval: 1h
  exec:
    command: check_ntp_time
    arguments:
      - "-H"
      - time.nist.gov
    timeout: 8s
  output:
    type: event
    event:
      include_command_name: false
      include_command_status: true
      include_command_duration: true
      max_body_size: "32kib"
```
