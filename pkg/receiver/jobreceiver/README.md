# Monitoring Job Receiver

| Status        |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| ------------- | -----------                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| Stability     | [development]: logs                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |

This receiver makes it possible to collect telemetry data from sources that
do not instrument well. The monitoring job receiver executes a script or
executable at defined intervals, and propagates the output from that process
as log events. In addition, the monitoring job receiver simplifies the process
of downloading runtime assets necessary to run a particular monitoring job.

## Configuration

For each `monitoringjob` receiver defined in configuration, the specified
command will be run according to the schedule. The command _should_ start
a binary that writes to either stdout or stderr.

The following settings are required:

- `exec.command`: The command to run.
- `schedule.interval`: How often to run the monitoring job, specified as a decimal number followed by a unit suffix. e.g. `60s`, `45m`, `2h30m`

The following settings are optional:

- `exec.arguments`: List of string argumets to pass the command
- `exec.runtime_assets`: List of runtime assets required for the job
- `exec.timeout`:  How long to wait for the process to exit before attempting to make it exit.

### Example configuration

```yaml
monitoringjob:
  schedule:
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
```
