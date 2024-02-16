# Windows Active Directory Inventory Receiver

**Stability level**: Alpha

A Windows Active Directory Inventory receiver collects inventory data from Active Directory. This includes information such as computer names, user names, email addresses, and location information. This receiver is only supported on Windows servers.

Supported pipeline types: logs

## Configuration

```yaml
receivers:
  active_directory_inv:
    # Base DN
    # default = ""
    # Base DN is a required field and cannot remain empty (default)
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
  active_directory_inv:
    base_dn: "CN=Users,DC=exampledomain,DC=com"
    attributes: [name, mail, department, manager, memberOf]
    poll_interval: 24h

exporters:
  debug:
    verbosity: detailed

service:
  telemetry:
    logs:
      level: "debug"
  pipelines:
    logs/syslog source:
      receivers:
        - active_directory_inv
      exporters:
        - debug
```
