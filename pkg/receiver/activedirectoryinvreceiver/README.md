# Windows Active Directory inventory Receiver

**Stability level**: Alpha

A Windows Active Directory Inventory receiver collects inventory data from Active Directory. This includes information such as computer names, user names, email addresses, and location information. This receiver is only supported on Windows servers.

Supported pipeline types: logs

## Configuration

```yaml
receivers:
  activedirectoryinv:
    # Base DN
    # default = ""
    base_dn: "CN=Users,DC=exampledomain,DC=com"

    # User attributes
    # default = [name, mail, department, manager, memberOf]
    attributes: [name, mail, department, manager, memberOf]

    # The polling interval.
    # default = 24h
    poll_interval: 24h
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
    poll_interval: 24h

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
