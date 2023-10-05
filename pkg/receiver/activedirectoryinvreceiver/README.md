# Windows Active Directory inventory Receiver

**Stability level**: Alpha

A Windows Active Directory Inventory receiver collects inventory data from Active Directory. This includes information such as computer names, user names, email addresses, and location information

Supported pipeline types: logs

## Configuration

```yaml
receivers:
  activedirectoryinv:
    # Common name
    # default: "test user"
    CN: "test user"

    # Organizational Unit
    # default: "test"
    OU: "test"

    # Password for authenticating with the AD server
    # default: "test"
    Password: "test"

    # Domain name
    # default: "exampledomain.com"
    DC: "exampledomain.com"

    # The fully qualified domain name (FQDN).
    # default: "examplehost"
    Host: "hostname.exampledomain.com"

    # The polling interval.
    # default = 60
    PollInterval: 60
```

The full list of settings exposed for this receiver are documented in
[config.go](./config.go).

Example configuration:

```yaml
receivers:
  ## All my example logs
  activedirectoryinvreceiver:
    cn: "test user"
    ou: "test"
    password: "Exampledomain@123"
    domain: "exampledomain"
    host: "EC2AMAZ.exampledomain.com"

exporters:
  logging:
    verbosity: detailed

service:
  telemetry:
    logs:
      level: "debug"
  pipelines:
    logs/syslog source:
      receivers:
      - activedirectoryinvreceiver
      exporters:
      - logging
```
