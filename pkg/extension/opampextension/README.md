# OpAMP Agent Extension

**Stability level**: Alpha

This extension implements an [`OpAMP agent`][opamp_spec] for remote collector
configuration management. This extension needs to be used in conjuction with the
[`sumologicextension`][sumologicextension] in order to authenticate with the
[Sumo Logic][sumologic] OpAMP server.

It manages:

- authentication (using `sumologicextension` to retreive credentials)
- registration (sends an initial OpAMP agent-to-server message)
- reporting (responds to OpAMP server requests with an agent status, e.g. the
  collector's effective configuration)
- local configuration (writes to a local OpenTelemetry YAML configuration file
  for a provider (i.e. glob) to read)
- collector configuration reloads (SIGHUP reloads on local configuration changes)

[opamp_spec]: https://github.com/open-telemetry/opamp-spec/blob/main/specification.md#opamp-open-agent-management-protocol
[sumologicextension]: ../sumologicextension/
[sumologic]: https://www.sumologic.com/

## Configuration

- `instance_uid`: a ULID formatted as a 26 character string in canonical
  representation. Auto-generated on start if missing. Setting this ensures the
  instance UID remains constant across process restarts.
- `endpoint`: (required) the OpAMP server secure websocket URL.
- `remote_configuration_directory`: (required) the directory used to store
  configuration received from the OpAMP server. This directory must coincide
  with a configuration provider (e.g. glob) for the configuration to be loaded
  by the collector.

## Example Config

```yaml
extensions:
  sumologic:
    installation_token: <token>
    api_base_url: <api_endpoint_url>
  opamp:
    endpoint: <wss_endpoint_url>
    remote_configuration_directory: /etc/otelcol-sumo/opamp.d
```

## API URLs

When integrating the extension with a different Sumo Logic deployment than the
default one (i.e. `https://open-collectors.sumologic.com`), one needs to specify
the Sumo Logic extension (`sumologic`) base API URL (`api_base_url`) and the
OpAMP extension (`opamp`) secure websocket endpoint (`endpoint`).

Here is a list of valid values for the Sumo Logic `api_base_url` configuration
option:

|  Deployment   | API base URL                                |
|:-------------:|---------------------------------------------|
| default/`US1` | `https://open-collectors.sumologic.com`     |
|     `US2`     | `https://open-collectors.us2.sumologic.com` |
|     `AU`      | `https://open-collectors.au.sumologic.com`  |
|     `DE`      | `https://open-collectors.de.sumologic.com`  |
|     `EU`      | `https://open-collectors.eu.sumologic.com`  |
|     `JP`      | `https://open-collectors.jp.sumologic.com`  |
|     `CA`      | `https://open-collectors.ca.sumologic.com`  |
|     `IN`      | `https://open-collectors.in.sumologic.com`  |

Here is a list of valid values for the OpAMP `endpoint** configuration option:

**Note:** These endpoints are not yet available.

|  Deployment   | API base URL                                        |
|:-------------:|-----------------------------------------------------|
| default/`US1` | `wss://opamp-collectors.sumologic.com/v1/opamp`     |
|     `US2`     | `wss://opamp-collectors.us2.sumologic.com/v1/opamp` |
|     `AU`      | `wss://opamp-collectors.au.sumologic.com/v1/opamp`  |
|     `DE`      | `wss://opamp-collectors.de.sumologic.com/v1/opamp`  |
|     `EU`      | `wss://opamp-collectors.eu.sumologic.com/v1/opamp`  |
|     `JP`      | `wss://opamp-collectors.jp.sumologic.com/v1/opamp`  |
|     `CA`      | `wss://opamp-collectors.ca.sumologic.com/v1/opamp`  |
|     `IN`      | `wss://opamp-collectors.in.sumologic.com/v1/opamp`  |

## Storing local configuration

When the OpAMP extension receives a remote configuration from the OpAMP server,
it persists each received YAML configuration to a local file in the
`remote_configuration_directory`. The existing contents of the
`remote_configuration_directory` are removed before doing so. A configuration
provider must be used in order to load the stored configuration, for example:
`--config "glob:/etc/otelcol-sumo/opamp.d/*"`.
