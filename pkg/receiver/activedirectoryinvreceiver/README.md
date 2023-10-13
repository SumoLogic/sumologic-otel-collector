# Windows Active Directory inventory Receiver

**Stability level**: Alpha

A Windows Active Directory Inventory receiver collects inventory data from Active Directory. This includes information such as computer names, user names, email addresses, and location information

Supported pipeline types: logs

## Configuration

```yaml
receivers:
  activedirectoryinv:
    # Base DN
    base_dn: "CN=Users,DC=exampledomain,DC=com"

    # User attributes
    attributes: [name, mail, department, manager, memberOf]

    # The polling interval.
    # default = 60
    poll_interval: 60
```

The full list of settings exposed for this receiver are documented in
[config.go](./config.go).

Example configuration:

```yaml
receivers:
  ## All my example logs
  activedirectoryinv:
    base_dn: "CN=Users,DC=exampledomain,DC=com"
    attributes: [name, mail, department, manager, memberOf]
    poll_interval: 60

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
