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

By default, the `type: 'event'` output handler is used. When configured with
output type event, command output will be buffered until the process exits,
at which time the receiver will emit a single structured event capturing the
command output and optionally metadata about the execution such as exit code
and duration. This is useful for true monitoring jobs, such as using a
monitoring plugin not supported natively by OTEL.

When configured with the `log_entries` output type, command output will be
emitted as distinct log entries line-by-line. This can be used to act as a
bridge to third party logs and other log events that may not be straightforward
to access.

### Runtime Asset Configuration

//todo(ck)

## Additional Features

### Time Parameters

Time parameters are specified as sets of decimal numbers followed by a
unit suffix. e.g. `60s`, `45m`, `2h30m40s`.

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
    timeout: 5s
  output:
    type: event
    include_command_name: false
    include_command_status: true
    include_command_duration: true
    max_body_size: '32kib'
```
