# Sumo Logic Extension

**Stability level**: Deprecated

This extension is deprecated in favor of the [Sumo Logic extension][sumologic_extension_docs] that lives in the [OpenTelemetry Collector Contrib][contrib_repo] repository.

This extension is to be used as part of Sumo Logic collector in conjuction with
[`sumologicexporter`][sumologicexporter] in order to export telemetry data to
[Sumo Logic][sumologic].

It manages:

- authentication (passing the provided credentials to `sumologicexporter`
  when configured as extension in the same service)
- registration (storing the registration info locally after successful registration
  for later use)
- heartbeats

[sumologicexporter]: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/exporter/sumologicexporter/README.md
[sumologic]: https://www.sumologic.com/
[sumologic_extension_docs]: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/extension/sumologicextension/README.md
[contrib_repo]: https://github.com/open-telemetry/opentelemetry-collector-contrib/

## Implementation

It implements [`HTTPClientAuthenticator`][httpclientauthenticator]
and can be used as an authenticator for the
[`configauth.Authentication`][configauth_authentication] option for HTTP clients.

[httpclientauthenticator]: https://github.com/open-telemetry/opentelemetry-collector/blob/2e84285efc665798d76773b9901727e8836e9d8f/config/configauth/clientauth.go#L34-L39
[configauth_authentication]: https://github.com/open-telemetry/opentelemetry-collector/blob/3f5c7180c51ed67a6f54158ede5e523822e9659e/config/configauth/configauth.go#L29-L33

## Configuration

- `installation_token`: (required) collector installation token for the Sumo Logic service, see
  [help][credentials_help] for more details
- `collector_name`: name that will be used for registration; by default the hostname is used. In the event of a conflict, a timestamp will be appended to the name. See [documentation][clobber] for more information.
- `collector_description`: collector description that will be used for registration
- `collector_category`: collector category that will be used for registration
- `collector_fields`: a map of key value pairs that will be used as collector
  fields that will be used for registration.
  For more information on this subject please visit [this help document][fields_help]
- `discover_collector_tags`: defines whether to auto-discover collector metadata
  tags (for local services, e.g. mysql) (default: `true`)

  **NOTE**: collector metadata tag auto-discovery is an alpha feature.
- `api_base_url`: base API URL that will be used for creating API requests,
  see [API URLs](#api-urls) details
  (default: `https://open-collectors.sumologic.com`)
- `heartbeat_interval`: interval that will be used for sending heartbeats
  (default: `15s`)
- `collector_credentials_directory`: directory where state files with registration
  info will be stored after successful collector registration
  (default: `$HOME/.sumologic-otel-collector`)
- `clobber`: defines whether to delete any existing collector with the same name. See [documentation][clobber] for more information.
- `force_registration`: defines whether to force registration every time the
  collector starts.
  This will cause the collector to not look at the locally stored credentials
  and to always reach out to API to register itself. (default: `false`)

  **NOTE**: if clobber is unset (default) then setting this to true will create
  a new collector (with new unique name) on Sumo UI on every collector start
  and create a new one upon registration.
- `ephemeral`: defines whether the collector will be deleted after 12 hours
  of inactivity (default: `false`)
- `time_zone`: defines the time zone of the collector, for example "America/Los_Angeles".
  For a list of all possible values, refer to the `TZ identifier` column in
  https://en.wikipedia.org/wiki/List_of_tz_database_time_zones#List
- `backoff`: defines backoff mechanism for retry in case of failed registration.
  [Exponential algorithm](https://pkg.go.dev/github.com/cenkalti/backoff/v4#ExponentialBackOff) is being used.
  - `initial_interval` - initial interval of backoff (default: `500ms`)
  - `max_interval` - maximum interval of backoff (default: `1m`)
  - `max_elapsed_time` - time after which registration fails definitely (default: `15m`)
  - `sticky_session_enabled` - default value is `false`

[credentials_help]: https://help.sumologic.com/docs/manage/security/installation-tokens
[fields_help]: https://help.sumologic.com/docs/manage/fields
[clobber]: https://help.sumologic.com/docs/send-data/installed-collectors/collector-installation-reference/force-collectors-name-clobber/

## Example Config

```yaml
extensions:
  sumologic:
    installation_token: <token>
    collector_name: my_collector
    time_zone: CET

receivers:
  hostmetrics:
    collection_interval: 30s
    scrapers:
      load:

processors:

exporters:
  sumologic:
    auth:
      authenticator: sumologic # Specify the name of the authenticator extension

service:
  extensions: [sumologic]
  pipelines:
    metrics:
      receivers: [hostmetrics]
      processors: []
      exporters: [sumologic]
```

## API URLs

When integrating the extension with different Sumo Logic deployment that the
default one (i.e. `https://open-collectors.sumologic.com`) one needs to specify
the base API URL in the configuration (via `api_base_url` option) in order to
specify against which URL the agent will be authenticating against.

Here is a list of valid values for this configuration option:

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

## Storing credentials

When collector is starting for the first time, Sumo Logic extension is using the `installation_token`
to register the collector with API.
Upon registration, the extension gets collector credentials which are used to authenticate the collector
when sending request to API (heartbeats, sending data etc).

Credentials are stored on local filesystem to be reused when collector gets restarted (to prevent re-registration).
The path that's used to store the credentials files is configured via `collector_credentials_directory` which is by default
set to `$HOME/.sumologic-otel-collector`.

Name of that file that contains the credentials is created in the following manner:

```go
filename := hash(collector_name, installation_token, api_base_url)
```

This mechanism allows to keep the state of the collector (whether it is registered or not).
When collector is restarting it checks if the state file exists in `collector_credentials_directory`.

If one would like to register another collector on the same machine then `collector_name` configuration property
has to be specified in order to register the collector under that specific name which will be used to create
a separate state file.

### Running the collector as systemd service

Systemd services are often run as users without a home directory,
so if the collector is run as such service, the credentials might not be stored properly. One should either make sure that the home directory exists for the user
or change the store location to another directory.
