# Monitoring Job Receiver

| Status        |                     |
| ------------- | -----------         |
| Stability     | [development]: logs |

This receiver makes it possible to collect telemetry data from sources that
do not instrument well. The monitoring job receiver executes a script or
executable at defined intervals, and propagates the output from that process
as log events. In addition, the monitoring job receiver simplifies the process
of downloading runtime assets necessary to run a particular monitoring job.

## Configuration

| Configuration | Default | Description
| ------------ | ------------ | ------------
| exec | required | A `exec` configuration block. See details [below](#execution-configuration)
| schedule | required | A `schedule` configuration block. See details [below](#schedule-configuration)
| output | | An `output` configuration block. See details [below](#output-configuration)

### Execution Configuration

| Configuration | Default | Description
| ------------ | ------------ | ------------
| command | required | The `command` to run. Should start a binary that writes to stdout and/or stderr
| arguments | | A list of string arguments to pass the command
| timeout | | [Time](#time-parameters) to wait for the process to exit before attempting to make it exit
| runtime_assets | | A list of `runtime_assets` required for the monitoring job

### Schedule Configuration

The scheduling configuration block currently only supports a single **required**
`interval` [Time](#time-parameters) parameter. Counting from collector startup, the command will be
scheduled every `interval`.

### Output Configuration

The monitoringjob receiver can handle output in two different ways.

By default, the `format: 'event'` output handler is used. When configured with
the event output format, command output will be buffered until the process exits,
at which time the receiver will emit a single structured event capturing the
command output and optionally metadata about the execution such as exit code
and duration. This is useful for true monitoring jobs, such as using a
monitoring plugin not supported natively by OTEL.

When configured with the `log_entries` output format, command output will be
emitted as distinct log entries line-by-line. This can be used to act as a
bridge to third party logs and other log events that may not be straightforward
to access.

**Output Configuration - event**

| Configuration | Default | Description
| ------------ | ------------ | ------------
| include_command_name | true | When set, includes the attribute `command.name` in the log event.
| include_command_status | true | When set, includes the attribute `command.status` in the log event.
| include_command_duration | true | When set, includes the attribute `command.duration` in the log event.
| max_body_size | | When set, restricts the length of command output to a specified [ByteSize](#bytesize-parameters).

**Output Configuration - log_entries**

| Configuration | Default | Description
| ------------ | ------------ | ------------
| include_command_name | true | When set, includes the attribute `command.name` in the log event.
| include_stream_name | true | When set, includes the attribute `command.stream.name` in the log event. Indicating `stdin` or `stderr` as the origin of the event.
| max_log_size | | When set, restricts the length of a log entry to a specified [ByteSize](#bytesize-parameters).
| encoding | utf-8 | Encoding to expect from the command output. Used to detect log entry boundaries.
| multiline | | Used to override the default newline delimited log entries.
| multiline.line_start_pattern | | Regex pattern for the beginning of a log entry. Mutualy exclusive with end pattern.
| multiline.line_end_pattern | | Regex pattern for the ending of a log entry. Mutualy exclusive with start pattern.

### Runtime Asset Configuration

//todo(ck)

## Additional Features

### Time Parameters

Time parameters are specified as sets of decimal numbers followed by a
unit suffix. e.g. `60s`, `45m`, `2h30m40s`.

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
    format: event
    event:
      include_command_name: false
      include_command_status: false
      include_command_duration: false
      max_body_size: '32kib'
```
