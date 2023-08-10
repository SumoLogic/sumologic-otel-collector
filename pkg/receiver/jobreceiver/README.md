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
| interval | required | [Time](#time-parameters) between executions
| attributes | | A map of key value pairs to add as log attributes
| resource | | A map of key value pairs to add as resource attributes
| encoding | utf-8 | The encoding of the output produced by the command
| multiline | | A `multiline` configuration block. See details [below](#multiline-configuration)

### Execution Configuration

| Configuration | Default | Description
| ------------ | ------------ | ------------
| command | required | The `command` to run. Should start a binary that writes to stdout and/or stderr
| arguments | | A list of string arguments to pass the command
| timeout | | [Time](#time-parameters) to wait for the process to exit before attempting to make it exit
| runtime_assets | | A list of `runtime_assets` required for the monitoring job

### Multiline Configuration

By default the monitoringjob receiver will split command output into log events
using a newline delimiter. When desired, the `multiline` configuration block
must contain either the `line_start_pattern` or the `line_end_pattern` option.
These are [re2] regex patterns that match either the beginning of a log event
or the end of a log entry.

To capture the entirety of command output as a single log event use `line_end_pattern: "\z"`

[re2]: https://github.com/google/re2/wiki/Syntax

### Runtime Asset Configuration

//todo(ck)

## Additional Features

### Time Parameters

Time parameters are specified as sets of decimal numbers followed by a
unit suffix. e.g. `60s`, `45m`, `2h30m40s`.

### Log Attributes

The monitorinjob receiver includes several attributes in its output.

- `command.name`: name of the command executed
- `command.output.stream`: Either `stdout` or `stderr`

## Example configuration

```yaml
monitoringjob:
  interval: 1h
  exec:
    command: check_ntp_time
    runtime_assets:
        - name: monitoring-plugins
          url: https://assets.bonsai.sensu.io/a7cfc70d3aa81ffd13ed3a7e55f2438c3c7e8f8e/monitoring-plugins-ubuntu2004_2.7.0_linux_amd64.tar.gz
          sha512: 550adf669715d7b97bbcaadc2a310d508844cfc5a3e5571ee0290436f99143fe7af0927c7f1667e2f43a598c2969fb734d1297adf5b258e571a42c884f90271c
          path: /opt/monitoring-plugins
    arguments:
        - "-H"
        - time.nist.gov
        - "--warn"
        - 0.5
        - "--critical"
        - 1.0
    timeout: 8s
  multiline:
    line_end_pattern: "\z"
```
