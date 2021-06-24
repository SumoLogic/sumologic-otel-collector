# Sumo Logic Extension

**This extension is experimental and may receive breaking changes at any time.**

This extension is to be used as part of Sumo Logic collector in conjuction with
[`sumologicexporter`][sumologicexporter] in order to export telemetry data to
[Sumo Logic][sumologic].

It manages:

* authentication (passing the provided credentials to `sumologicexporter`
  when configured as extension in the same service)
* registration (storing the registration info locally after successful registration
  for later use)
* heartbeats

[sumologicexporter]: ../../exporter/sumologicexporter/
[sumologic]: https://www.sumologic.com/

## Implementation

It implements [`HTTPClientAuthenticator`][httpclientauthenticator]
and can be used as an authenticator for the
[`configauth.Authentication`][configauth_authentication] option for HTTP clients.

[httpclientauthenticator]: https://github.com/open-telemetry/opentelemetry-collector/blob/2e84285efc665798d76773b9901727e8836e9d8f/config/configauth/clientauth.go#L34-L39
[configauth_authentication]: https://github.com/open-telemetry/opentelemetry-collector/blob/3f5c7180c51ed67a6f54158ede5e523822e9659e/config/configauth/configauth.go#L29-L33

## Configuration

* `access_id`: (required) access ID for Sumo Logic service, see
  [help][credentials_help] for more details
* `access_key`: (required) access key for Sumo Logic service, see
  [help][credentials_help] for more details
* `collector_name`: (required) name that will be used for locally stored
  registration info from previous runs or for registration in case those are not found
* `collector_description`: collector description that will be used for registration
* `collector_category`: collector category that will be used for registration
* `collector_fields`: a map of key value pairs that will be used as collector
  fields that will be used for registration.
  For more information on this subject please visit[this help document][fields_help]
* `api_base_url`: base URL that will be used for creating API requests
  (default: `https://collectors.sumologic.com`)
* `heartbeat_interval`: interval that will be used for sending heartbeats
  (default: `15s`)
* `collector_credentials_path`: path where registration info will be stored after
  successful collector registration (default: `$HOME/.sumologic-otel-collector`)
* `clobber`: defines whether to delete any existing collector with the same name
  and create a new one upon registration (default: `false`)
* `ephemeral`: defines whether the collector will be deleted after 12 hours
	of inactivity (default: `false`)
* `time_zone`: defines the time zone of the collector. For a list of all possible
  values, refer to the `TZ` column in
  https://en.wikipedia.org/wiki/List_of_tz_database_time_zones#List

[credentials_help]: https://help.sumologic.com/Manage/Security/Access-Keys
[fields_help]: https://help.sumologic.com/Manage/Fields

## Example Config

```yaml
extensions:
  sumologic:
    access_id: aaa
    access_key: bbbbbbbbbbbbbbbbbbbbbb
    collector_name: my_collector

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

## Storing credentials

When collector is starting first time Sumo Logic extension is using `access_key` and `access_id` to register collector with API.
After registration extension gets collector credentials which are used to authenticate instance of collector.
Credentials are stored on filesystem in `collector_credentials_path` which is by default set to `$HOME/.sumologic-otel-collector`.
Name of file that contains credentials is created by hashing combination of `collector_name`, `access_id` and `access_key`.
This mechanism allows to keep a state of collector (whether is registered or not).
When collector is restarting it checks if state file exists in `collector_credentials_path`.
If user want to register collector on the same machine and use another state file, need add the collector name to otc config.
