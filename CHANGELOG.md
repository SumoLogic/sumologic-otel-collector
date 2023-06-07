# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- chore: update OT core to v0.79.0 [#1158]

[#1158]: https://github.com/SumoLogic/sumologic-otel-collector/pull/1154
[unreleased]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.78.0-sumo-0...main

## [v0.78.0-sumo-0]

### Released 2023-06-05

This release introduces the following breaking changes:

- fix(sumologicextension)!: check credentials dir at start [#1152] [#1153]

Set `force_registatrion: true` in the extension configuration if you don't want the credentials persisted at all.

### Added

- feat(receiver/filestats): add File Stats receiver [#1146]
- feat(receiver/sqlquery): add experimental logs support [#1144]
- feat(exporter/awss3): add AWS S3 exporter [#1149]

### Changed

- chore: update OT to v0.78.0 [#1142]

### Fixed

- fix: do not send html special characters as unicode [#1145]

[v0.78.0-sumo-0]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.77.0-sumo-0...v0.78.0-sumo-0
[#1142]: https://github.com/SumoLogic/sumologic-otel-collector/pull/1142
[#1144]: https://github.com/SumoLogic/sumologic-otel-collector/pull/1144
[#1145]: https://github.com/SumoLogic/sumologic-otel-collector/pull/1145
[#1146]: https://github.com/SumoLogic/sumologic-otel-collector/pull/1146
[#1149]: https://github.com/SumoLogic/sumologic-otel-collector/pull/1149
[#1152]: https://github.com/SumoLogic/sumologic-otel-collector/pull/1152
[#1153]: https://github.com/SumoLogic/sumologic-otel-collector/pull/1153

## [v0.77.0-sumo-0]

### Released 2023-05-24

This release introduces the following breaking changes:

- feat!: disable Prometheus metrics name normalization by default [#1138]

- feat(datadogprocessor)!: remove DataDog processor [#1135]

  It doesn't make much sense to include it if the DataDog exporter is not included.

See the [upgrade guide][upgrade_guide_v0.77] for more details.

### Added

- feat(k8sprocessor): support otel semantic convention in config [#1122]
- chore: upgrade OT core to 0.77.0 [#1125]

### Changed

- feat(sumologicextension): retry validation and exit in case of connection issues [#1134]

### Fixed

- fix(sumologicexporter): avoid allocations in compressor [#1118]

[v0.77.0-sumo-0]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.76.1-sumo-0...v0.77.0-sumo-0
[#1118]: https://github.com/SumoLogic/sumologic-otel-collector/pull/1118
[#1122]: https://github.com/SumoLogic/sumologic-otel-collector/pull/1122
[#1125]: https://github.com/SumoLogic/sumologic-otel-collector/pull/1125
[#1134]: https://github.com/SumoLogic/sumologic-otel-collector/pull/1134
[#1135]: https://github.com/SumoLogic/sumologic-otel-collector/pull/1135
[#1138]: https://github.com/SumoLogic/sumologic-otel-collector/pull/1138
[upgrade_guide_v0.77]: ./docs/upgrading.md#upgrading-to-v0770-sumo-0

## [0.76.1-sumo-0]

### Released 2023-28-04

### Changed

- chore: upgrade OT core to 0.76.1 [#1112]

[#1112]: https://github.com/SumoLogic/sumologic-otel-collector/pull/1112
[0.76.1-sumo-0]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.75.0-sumo-0...v0.76.1-sumo-0

## [v0.75.0-sumo-0]

### Released 2023-04-13

### Added

- feat: add Service Graph connector [#1102]

### Changed

- chore: upgrade OT core to 0.75.0 [#1094]

### Removed

- feat!: remove Dotnet Diagnostics receiver [#1103]

[#1094]: https://github.com/SumoLogic/sumologic-otel-collector/pull/1094
[#1102]: https://github.com/SumoLogic/sumologic-otel-collector/pull/1102
[#1103]: https://github.com/SumoLogic/sumologic-otel-collector/pull/1103
[v0.75.0-sumo-0]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.74.0-sumo-0...v0.75.0-sumo-0

## [v0.74.0-sumo-0]

### Released 2023-03-30

### Added

- feat(sumologicschemaprocessor): add translating docker stats resource attributes [#1081]
- chore: add new components [#1091]
  - cloudflarereceiver
  - lokireceiver
  - spanmetricsconnector

### Changed

- chore: upgrade OT core to 0.74.0 [#1089]

### Fixed

- fix(sumologicexporter): Prometheus histogram metric names [#1087]

[#1081]: https://github.com/SumoLogic/sumologic-otel-collector/pull/1081
[#1087]: https://github.com/SumoLogic/sumologic-otel-collector/pull/1087
[#1089]: https://github.com/SumoLogic/sumologic-otel-collector/pull/1089
[#1091]: https://github.com/SumoLogic/sumologic-otel-collector/pull/1091
[v0.74.0-sumo-0]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.73.0-sumo-1...v0.74.0-sumo-0

## [v0.73.0-sumo-1]

### Released 2023-03-13

This release introduces the following breaking changes:

- fix(sumologicextension)!: use fqdn before os.Hostname

See the [upgrade guide][upgrade_guide_v0.74] for more details.

### Added

- feat!(sumologicschemaprocessor): add translating docker stats metric names [#1055]

### Changed

- fix: fix carbon2 parser for telegrafreceiver [#1058]

### Fixed

- fix(scripts/install.ps1): treat app as not installed if otelcol-sumo.exe is missing [#1061]
- fix(syslogexporter): set default settings for sending_queue and retry_on_failure [#1056]

[#1058]: https://github.com/SumoLogic/sumologic-otel-collector/pull/1058
[#1055]: https://github.com/SumoLogic/sumologic-otel-collector/pull/1055
[#1061]: https://github.com/SumoLogic/sumologic-otel-collector/pull/1061
[#1056]: https://github.com/SumoLogic/sumologic-otel-collector/pull/1056
[upgrade_guide_v0.74]: ./docs/upgrading.md#upgrading-to-v0660-sumo-0
[v0.73.0-sumo-1]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.73.0-sumo-0...v0.73.0-sumo-1

## [v0.73.0-sumo-0]

### Released 2023-03-08

### Added

- feat(sumologicextension): enable updateCollectorMetadata feature gate by default [#1027]
  The original feature has been added in [#858].

### Changed

- chore: update OT core to v0.73.0 [#1048]

[#858]: https://github.com/SumoLogic/sumologic-otel-collector/pull/858
[#1027]: https://github.com/SumoLogic/sumologic-otel-collector/pull/1027
[#1048]: https://github.com/SumoLogic/sumologic-otel-collector/pull/1048
[v0.73.0-sumo-0]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.72.0-sumo-0...v0.73.0-sumo-0

## [v0.72.0-sumo-0]

### Released 2023-02-24

### Added

- feat(syslogexporter): add mTLS support [#980]

### Changed

- feat(sumologicextension): deprecate `install_token` [#969]
- feat(syslogexporter): remove adding additional structure data in syslog exporter [#975]
- feat(syslogexporter): change TLS configuration options to opentelemetry configtls [#983]
- chore: bump golang to 1.19 [#1011]
- chore: update OT core to v0.72.0 [#1013]
- chore: add new components [#1016]

  - datadogprocessor
  - servicegraphprocessor
  - awscloudwatchreceiver
  - azureeventhubreceiver
  - datadogreceiver
  - haproxyreceiver
  - iisreceiver
  - k8sobjectsreceiver
  - otlpjsonfilereceiver
  - purefareceiver
  - purefbreceiver
  - solacereceiver
  - sshcheckreceiver
  - forwardconnector
  - countconnector

- feat(sumologicexporter): use otlp url suffixes [#1015]

[#969]: https://github.com/SumoLogic/sumologic-otel-collector/pull/969
[#975]: https://github.com/SumoLogic/sumologic-otel-collector/pull/975
[#980]: https://github.com/SumoLogic/sumologic-otel-collector/pull/980
[#983]: https://github.com/SumoLogic/sumologic-otel-collector/pull/983
[#1011]: https://github.com/SumoLogic/sumologic-otel-collector/pull/1011
[#1013]: https://github.com/SumoLogic/sumologic-otel-collector/pull/1013
[#1016]: https://github.com/SumoLogic/sumologic-otel-collector/pull/1016
[#1015]: https://github.com/SumoLogic/sumologic-otel-collector/pull/1015
[v0.72.0-sumo-0]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.71.0-sumo-0...v0.72.0-sumo-0

## [v0.71.0-sumo-0]

### Released 2023-02-09

This release introduces the following breaking changes:

- feat(sumologicextension): use hostname as default collector name [#918]
- feat(script)!: be consistent with installation token naming [#941]

  We depracated `--skip-install-token` in favor of `--skip-installation-token`
  and `SUMOLOGIC_INSTALL_TOKEN` environmental variable in favor of `SUMOLOGIC_INSTALLATION_TOKEN`.
  Please update your configuration and automation scripts.

### Added

- feat(sourceprocessor): add debug logs for source category filler [#944]
- feat(snmpreceiver): add SNMP receiver to distro [#945]
- feat(syslogexporter): add syslog exporter [#936]
- feat(sourceprocessor): make container name attribute configurable [#950]

### Changed

- chore: update OT core to v0.71.0 [#958]

### Fixed

- fix(k8sprocessor): race condition when getting Pod data [#938]

[#918]: https://github.com/SumoLogic/sumologic-otel-collector/pull/918
[#938]: https://github.com/SumoLogic/sumologic-otel-collector/pull/938
[#944]: https://github.com/SumoLogic/sumologic-otel-collector/pull/944
[#945]: https://github.com/SumoLogic/sumologic-otel-collector/pull/945
[#936]: https://github.com/SumoLogic/sumologic-otel-collector/pull/936
[#950]: https://github.com/SumoLogic/sumologic-otel-collector/pull/950
[#958]: https://github.com/SumoLogic/sumologic-otel-collector/pull/958
[v0.71.0-sumo-0]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.70.0-sumo-0...v0.71.0-sumo-0

## [v0.70.0-sumo-2]

### Fixed

- fix release binary versions [#943]

[v0.70.0-sumo-2]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.70.0-sumo-1...v0.70.0-sumo-2
[#943]: https://github.com/SumoLogic/sumologic-otel-collector/pull/943

## [v0.70.0-sumo-1]

### Added

- feat: Collector metadata tag auto-discovery (local services, e.g. mysql) [#893]
- feat(extension/opamp): implemented an opamp agent for remote configuration [#885]
- feat: FIPS compliance [#902]
- fix(telegrafreceiver): make shutdown safe to call before start [#913]
- chore: update OT core to v0.70.0 [#915]

[v0.70.0-sumo-1]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.69.0-sumo-0...v0.70.0-sumo-1
[#893]: https://github.com/SumoLogic/sumologic-otel-collector/pull/893
[#885]: https://github.com/SumoLogic/sumologic-otel-collector/pull/885
[#902]: https://github.com/SumoLogic/sumologic-otel-collector/pull/902
[#913]: https://github.com/SumoLogic/sumologic-otel-collector/pull/913
[#915]: https://github.com/SumoLogic/sumologic-otel-collector/pull/915

## [v0.70.0-sumo-0]

This was a failed release. Use [v0.70.0-sumo-1] instead.

[v0.70.0-sumo-0]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.69.0-sumo-0...v0.70.0-sumo-0

## [v0.69.0-sumo-0]

### Added

- feat: Integrated collector with new metadata API [#858]
- chore: upgrade OT core to v0.69.0 [#891]

[v0.69.0-sumo-0]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.68.0-sumo-0...v0.69.0-sumo-0
[#858]: https://github.com/SumoLogic/sumologic-otel-collector/pull/858
[#891]: https://github.com/SumoLogic/sumologic-otel-collector/pull/891

## [v0.68.0-sumo-0]

### Added

- feat(sumologicschemaprocessor): add nesting processor [#877]
- feat(sumologicschemaprocessor): add allowlist and denylist to nesting processor [#880]
- feat(sumologicschemaprocessor) allow aggregating attributes with given name patterns [#871]
- feat(sumologicschemaprocessor): add squashing single values in nesting processor [#881]
- feat(sumologicschemaprocessor): report attributes as fields [#874]
- feat(extension/sumologic): mark install_token as opaque [#882]

[v0.68.0-sumo-0]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.67.0-sumo-0...v0.68.0-sumo-0
[#877]: https://github.com/SumoLogic/sumologic-otel-collector/pull/877
[#880]: https://github.com/SumoLogic/sumologic-otel-collector/pull/880
[#871]: https://github.com/SumoLogic/sumologic-otel-collector/pull/871
[#881]: https://github.com/SumoLogic/sumologic-otel-collector/pull/881
[#874]: https://github.com/SumoLogic/sumologic-otel-collector/pull/874
[#882]: https://github.com/SumoLogic/sumologic-otel-collector/pull/882

### Changed

- chore: upgrade OT core to v0.68.0

## [v0.67.0-sumo-0]

### Released 2022-12-19

### Added

- feature(packaging/msi): add conf.d dir, mv token/tags to common.yaml [869]
- feat(ci): build msi packages for dev & release jobs [#856]

### Changed

- chore: upgrade OT core to v0.67.0 [#867]

### Fixed

- fix(otelcolbuilder): use correct upstream modules [#864]

[v0.67.0-sumo-0]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.66.0-sumo-0...v0.67.0-sumo-0
[#864]: https://github.com/SumoLogic/sumologic-otel-collector/pull/864
[#867]: https://github.com/SumoLogic/sumologic-otel-collector/pull/867
[#869]: https://github.com/SumoLogic/sumologic-otel-collector/pull/869
[#856]: https://github.com/SumoLogic/sumologic-otel-collector/pull/856

## [v0.66.0-sumo-0]

### Released 2022-12-08

This release introduces the following breaking changes:

- `filelog` receiver: has been removed from sub-parsers ([upgrade guide][upgrade_guide_v0.66]) [#769]
- `sending_queue`: require explicit storage set ([upgrade guide][upgrade_guide_v0.66]) [#769]
- `apache` receiver: turn on feature gates for resource attributes ([upgrade guide][upgrade_guide_v0.66]) [#839]
- `elasticsearch` receiver: turn on feature gates for resource attributes ([upgrade guide][upgrade_guide_v0.66]) [#848]

### Added

- feat: add glob config provider [#713]
- feat(build): validate FIPS mode at build time and runtime [#693]
- feat(ci): add windows builds to dev & pr jobs [#762]
- feat(packaging/msi): add msi packaging [#852]

### Changed

- fix(sumologicexporter): do not crash if server returns unknown length response [#718]
- fix(k8sprocessor): fix metadata enrichment [#724]
- fix(k8sprocessor): keep pod's services information up to date [#710]
- chore(deps): bump golang from 1.18.4 to 1.19.2 [#745]
- chore(deps): bump go-boringcrypto to 1.18.7b7 [#746]
- feat(sourceprocessor): ensure that '\_collector' is set before other source headers [#824]
- chore(deps): upgrade Telegraf to 1.24.3-sumo-1 [#828]
- chore: upgrade OT core to v0.66.0 [#769] [#826] [#844] [#849]

### Removed

- feat(filterprocessor): drop custom changes ([upgrade guide][upgrade_guide_v0_55_0_expr_support]) [#709] [#714]
- feat(sumologicexporter): remove translating telegraf metric names ([upgrade guide][upgrade_guide_unreleased_moved_telegraf_translation]) [#678]
- feat(sumologicexporter): remove translating attributes ([upgrade guide][upgrade_guide_unreleased_moved_translation]) [#672]
- feat(sumologicexporter): remove setting source headers ([upgrade guide][upgrade_guide_v0_57_0_deprecate_source_templates]) [#686]

[v0.66.0-sumo-0]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.57.2-sumo-1...v0.66.0-sumo-0
[#672]: https://github.com/SumoLogic/sumologic-otel-collector/pull/672
[#678]: https://github.com/SumoLogic/sumologic-otel-collector/pull/678
[#686]: https://github.com/SumoLogic/sumologic-otel-collector/pull/686
[#709]: https://github.com/SumoLogic/sumologic-otel-collector/pull/709
[#710]: https://github.com/SumoLogic/sumologic-otel-collector/pull/710
[#714]: https://github.com/SumoLogic/sumologic-otel-collector/pull/714
[#713]: https://github.com/SumoLogic/sumologic-otel-collector/pull/713
[#718]: https://github.com/SumoLogic/sumologic-otel-collector/pull/718
[#724]: https://github.com/SumoLogic/sumologic-otel-collector/pull/724
[#745]: https://github.com/SumoLogic/sumologic-otel-collector/pull/745
[#746]: https://github.com/SumoLogic/sumologic-otel-collector/pull/746
[#762]: https://github.com/SumoLogic/sumologic-otel-collector/pull/762
[#769]: https://github.com/SumoLogic/sumologic-otel-collector/pull/769
[#693]: https://github.com/SumoLogic/sumologic-otel-collector/pull/693
[#824]: https://github.com/SumoLogic/sumologic-otel-collector/pull/824
[#828]: https://github.com/SumoLogic/sumologic-otel-collector/pull/828
[#826]: https://github.com/SumoLogic/sumologic-otel-collector/pull/826
[#844]: https://github.com/SumoLogic/sumologic-otel-collector/pull/844
[#849]: https://github.com/SumoLogic/sumologic-otel-collector/pull/849
[#839]: https://github.com/SumoLogic/sumologic-otel-collector/pull/839
[#848]: https://github.com/SumoLogic/sumologic-otel-collector/pull/848
[#852]: https://github.com/SumoLogic/sumologic-otel-collector/pull/852
[upgrade_guide_v0.66]: ./docs/upgrading.md#upgrading-to-v0660-sumo-0

## [v0.57.2-sumo-1]

### Released 2022-09-14

### Changed

- fix(k8sprocessor): fix metadata enrichment [#725]

[v0.57.2-sumo-1]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.57.2-sumo-0...v0.57.2-sumo-1
[#725]: https://github.com/SumoLogic/sumologic-otel-collector/pull/725

## [v0.57.2-sumo-0]

### Released 2022-08-12

This release deprecates the following features, which will be removed in `v0.60.0`:

- feat(sumologicexporter): deprecate source templates ([upgrade guide][upgrade_guide_v0_57_0_deprecate_source_templates])

### Changed

- feat(sumologicexporter): deprecate source templates ([upgrade guide][upgrade_guide_v0_57_0_deprecate_source_templates])
- feat: define stability levels for components [#701]
- chore: upgrade OpenTelemetry Core to v0.57.2 [#699]

[v0.57.2-sumo-0]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.56.0-sumo-0...v0.57.2-sumo-0
[upgrade_guide_v0_57_0_deprecate_source_templates]: ./docs/upgrading.md#sumologic-exporter-drop-support-for-source-headers
[#699]: https://github.com/SumoLogic/sumologic-otel-collector/pull/699/
[#701]: https://github.com/SumoLogic/sumologic-otel-collector/pull/701/

## [v0.56.0-sumo-0]

### Released 2022-07-22

This release deprecates the following features, which will be removed in `v0.59.0`:

- 'sumologic' exporter: translate attributes ([upgrade guide][upgrade_guide_unreleased_moved_translation])
- 'sumologic' exporter: translate Telegraf metric names ([upgrade guide][upgrade_guide_unreleased_moved_telegraf_translation])

### Added

- feat(sumologicschemaprocessor): add translating attributes
- feat: add aerospikereceiver [#674]
- feat(sumologicschemaprocessor): add translating Telegraf metric names

### Changed

- feat(sumologicexporter): deprecate translating attributes ([upgrade guide][upgrade_guide_unreleased_moved_translation])
- chore: upgrade OpenTelemetry Core to v0.56.0 [#674]
- feat(sumologicexporter): deprecate translating Telegraf metric names ([upgrade guide][upgrade_guide_unreleased_moved_telegraf_translation])

### Fixed

- fix(k8sprocessor): only apply the node filter to Pods [#668]

[v0.56.0-sumo-0]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.55.0-sumo-0...v0.56.0-sumo-0
[#668]: https://github.com/SumoLogic/sumologic-otel-collector/pull/668
[#674]: https://github.com/SumoLogic/sumologic-otel-collector/pull/674
[upgrade_guide_unreleased_moved_translation]: ./docs/upgrading.md#sumologic-exporter-drop-support-for-translating-attributes
[upgrade_guide_unreleased_moved_telegraf_translation]: ./docs/upgrading.md#sumologic-exporter-drop-support-for-translating-telegraf-metric-names

## [v0.55.0-sumo-0]

### Released 2022-07-13

This release deprecates the following change:

- `filter` processor: support for `expr` language ([upgrade guide][upgrade_guide_v0_55_0_expr_support])

### Added

- feat(cascadingfilter): use LRU Cache for storing sampling decisions [#654]
- feat(cascadingfilter): use limit for maximum volume of passed spans for which decisions were made earlier [#654]
- feat(cascadingfilter): store information on which policy filtered the trace in `sampling.filter` [#654]
- feat(cascadingfilter): store information about late span arrival in `sampling.late_arrival: true` [#654]
- feat(cascadingfilter): add `otelcol_count_late_spans` and `otelcol_count_decided_spans` metrics [#654]

### Changed

- feat(sumologicexporter): do not send empty OTLP requests [#660]
- feat(sumologicexporter): do not retry on '400 Bad Request' response [#661]
- chore: upgrade OpenTelemetry Core to v0.55.0 [#655]

### Fixed

- fix(sumologicexporter): translate Telegraf metrics with OTLP format [#659]

[v0.55.0-sumo-0]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.54.0-sumo-0...v0.55.0-sumo-0
[#654]: https://github.com/SumoLogic/sumologic-otel-collector/pull/654
[#659]: https://github.com/SumoLogic/sumologic-otel-collector/pull/659
[#660]: https://github.com/SumoLogic/sumologic-otel-collector/pull/660
[#661]: https://github.com/SumoLogic/sumologic-otel-collector/pull/661
[#655]: https://github.com/SumoLogic/sumologic-otel-collector/pull/655
[upgrade_guide_v0_55_0_expr_support]: ./docs/upgrading.md#filter-processor-drop-support-for-expr-language

## [v0.54.0-sumo-0]

### Released 2022-07-04

### Added

- feat(rawk8seventsreceiver): remember last processed resource version [#620]

### Changed

- chore: upgrade OT core to v0.54.0 [#637]
- ci: re-enable MacOS builds [#642]

[v0.54.0-sumo-0]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.53.0-sumo-0...v0.54.0-sumo-0
[#620]: https://github.com/SumoLogic/sumologic-otel-collector/pull/620
[#637]: https://github.com/SumoLogic/sumologic-otel-collector/pull/637
[#642]: https://github.com/SumoLogic/sumologic-otel-collector/pull/642

## [v0.53.0-sumo-0]

### Released 2022-06-28

:warning: Due to an infrastructure problem, this release lacks the MacOS binaries.
We hope to restore building the binaries for MacOS as soon as possible.

This release adds missing [receivers], [processors] and [extensions] from the OpenTelemetry Distribution.
It also includes `journald` binary required by [journaldreceiver]
and begin support of arm64 architecture for Darwin OS.

### Added

- feat: build arm64 binary for darwin [#611]
- feat: add missing core receivers, processors and extensions [#597], [#604], [#614]
- chore(ci): add pipelines to test and build on Go+BoringCrypto [#588]

### Fixed

- fix(cascadingfilter): fix leak memory leak on late arriving traces where decision was already made [#616]

### Changed

- chore(core): upgrade to v0.53.0 [#615]
- feat(journaldreceiver): add missing dependencies [#577]
- ci: disable MacOS builds while signing not possible [#628], [#629]

[v0.53.0-sumo-0]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.52.0-sumo-0...v0.53.0-sumo-0
[#597]: https://github.com/SumoLogic/sumologic-otel-collector/pull/597
[#577]: https://github.com/SumoLogic/sumologic-otel-collector/pull/577
[#604]: https://github.com/SumoLogic/sumologic-otel-collector/pull/604
[#611]: https://github.com/SumoLogic/sumologic-otel-collector/pull/611
[#616]: https://github.com/SumoLogic/sumologic-otel-collector/pull/616
[#615]: https://github.com/SumoLogic/sumologic-otel-collector/pull/615
[#614]: https://github.com/SumoLogic/sumologic-otel-collector/pull/614
[#628]: https://github.com/SumoLogic/sumologic-otel-collector/pull/628
[#629]: https://github.com/SumoLogic/sumologic-otel-collector/pull/629
[journaldreceiver]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/receiver/journaldreceiver#journald-receiver
[receivers]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/receiver
[processors]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/processor
[extensions]: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.53.0/extension

## [v0.52.0-sumo-1]

### Released 2022-06-14

### Fixed

- fix(cascadingfilter): fix leak memory leak on late arriving traces where decision was already made [#619]

[v0.52.0-sumo-1]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.52.0-sumo-0...v0.52.0-sumo-1
[#619]: https://github.com/SumoLogic/sumologic-otel-collector/pull/619
[#588]: https://github.com/SumoLogic/sumologic-otel-collector/pull/588

## [v0.52.0-sumo-0]

### Released 2022-06-02

See [Upgrade guide][upgrade_guide_v0_52_0] for the breaking changes in this version.

### Breaking changes

- feat(sumologicexporter)!: remove support for Carbon2 metrics format [#590][#590] ([upgrade guide][upgrade_guide_v0_52_0_metrics_support])
- feat(sumologicexporter)!: remove support for Graphite metrics format [#592][#592] ([upgrade guide][upgrade_guide_v0_52_0_metrics_support])

### Fixed

- fix(k8sprocessor): store only necessary Pod data [#593][#593]
- fix(filelogreceiver): fix changing fingerprint_size [#601]

### Changed

- chore(deps): update OT core to v0.52.0 [#600]

[v0.52.0-sumo-0]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.51.0-sumo-0...v0.52.0-sumo-0
[#590]: https://github.com/SumoLogic/sumologic-otel-collector/pull/590
[#592]: https://github.com/SumoLogic/sumologic-otel-collector/pull/592
[#593]: https://github.com/SumoLogic/sumologic-otel-collector/pull/593
[#600]: https://github.com/SumoLogic/sumologic-otel-collector/pull/600
[#601]: https://github.com/SumoLogic/sumologic-otel-collector/pull/601
[upgrade_guide_v0_52_0]: ./docs/upgrading.md#upgrading-to-v0520-sumo-0
[upgrade_guide_v0_52_0_metrics_support]: ./docs/upgrading.md#sumologic-exporter-removed-carbon2-and-graphite-metric-formats

## [v0.51.0-sumo-0]

### Released 2022-05-19

See [Upgrade guide][upgrade_guide_v0_51_0] for the breaking changes in this version.

### Breaking changes

- fix(k8sprocessor)!: remove `clusterName` metadata extraction option [#578] ([upgrade guide][upgrade_guide_v0_51_0_cluster_name])
- feat(sumologicexporter)!: attribute translation: change `file.path.resolved` to `log.file.path_resolved` [#579] ([upgrade guide][upgrade_guide_v0_51_0_attribute_translation])

### Added

- feat: enable rawk8seventsreceiver [#576]

### Changed

- chore(deps): update OT core to v0.51.0 [#580]
- chore(deps): update Telegraf to v1.22.0-sumo-4 [#580]

### Fixed

- fix: fix(cascadingfilterprocessor): do not attach sampling.rule attribute if trace accept rules are not specified [#575][#575]

[v0.51.0-sumo-0]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.50.0-sumo-0...v0.51.0-sumo-0
[upgrade_guide_v0_51_0]: ./docs/upgrading.md#upgrading-to-v0510-sumo-0
[upgrade_guide_v0_51_0_cluster_name]: ./docs/upgrading.md#k8s_tagger-processor-removed-clustername-metadata-extraction-option
[upgrade_guide_v0_51_0_attribute_translation]: ./docs/upgrading.md#sumologic-exporter-metadata-translation-changed-the-attribute-that-is-translated-to-_sourcename-from-filepathresolved-to-logfilepath_resolved
[#576]: https://github.com/SumoLogic/sumologic-otel-collector/pull/576
[#575]: https://github.com/SumoLogic/sumologic-otel-collector/pull/575
[#578]: https://github.com/SumoLogic/sumologic-otel-collector/pull/578
[#579]: https://github.com/SumoLogic/sumologic-otel-collector/pull/579
[#580]: https://github.com/SumoLogic/sumologic-otel-collector/pull/580

## [v0.50.0-sumo-0]

### Released 2022-04-29

Aside from upstream changes, this release only contains a performance fix to metrics batching in the Sumo Logic exporter.
The performance improvement is very substantial, so we recommend upgrading to this version immediately after `0.49.0-sumo-0`.

### Changed

- chore: update OT core to 0.50.0 [#562][#562]

### Fixed

- fix: fix(sumologicexporter): batch metrics if source headers match [#561][#561]

[v0.50.0-sumo-0]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.49.0-sumo-0...v0.50.0-sumo-0
[#561]: https://github.com/SumoLogic/sumologic-otel-collector/pull/561
[#562]: https://github.com/SumoLogic/sumologic-otel-collector/pull/562

## [v0.49.0-sumo-0]

### Released 2022-04-26

This release includes two breaking changes. One is an upstream change to the configuration syntax of several
log receivers, most notably the `filelog` receiver. The other changes how the Sumo Logic exporter determines
metadata based on the attributes of a OpenTelemetry record. Please consult the upgrade guides linked below
for more details.

### Breaking changes

- chore: bump OT core to v0.49.0 [#550][#550] ([upgrade guide][upgrade-guide-log-collection])
- fix!(sumologicexporter): send resource attributes as fields for non-otlp, removing metadata_attributes [#549][#549] ([upgrade guide][upgrade-guide-metadata])

### Changed

- docs: clarify status of sumologicextension [#553][#553]
- chore(deps): bump golang from 1.18 to 1.18.1 [#546][#546]
- chore: bump Telegraf to v1.22.0-sumo-3 [#557][#557]

### Fixed

- fix(cascadingfilterprocessor): prevent overriding metrics in cascading filter processor - add processor tag [#539][#539]

[v0.49.0-sumo-0]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.48.0-sumo-0...v0.49.0-sumo-0
[upgrade-guide-log-collection]: docs/upgrading.md#several-changes-to-receivers-using-opentelemetry-log-collection
[upgrade-guide-metadata]: docs/upgrading.md#sumo-logic-exporter-metadata-handling
[#546]: https://github.com/SumoLogic/sumologic-otel-collector/pull/546
[#550]: https://github.com/SumoLogic/sumologic-otel-collector/pull/550
[#553]: https://github.com/SumoLogic/sumologic-otel-collector/pull/553
[#539]: https://github.com/SumoLogic/sumologic-otel-collector/pull/539
[#557]: https://github.com/SumoLogic/sumologic-otel-collector/pull/557
[#549]: https://github.com/SumoLogic/sumologic-otel-collector/pull/549

## [v0.48.0-sumo-0]

### Released 2022-04-12

### Changed

- chore: bump OT core to v0.48.0 [#534][#534]
- chore(deps): bump alpine from 3.15.3 to 3.15.4 [#525][#525]

### Fixed

- fix(sumologicexporter): treat resource attributes as fields for otlp #536

### Other

- refactor(sumologicexporter): use golang.org/x/exp/slices for sorting fields [#519][#519]
- refactor(sumologicextension): use bytes slices and strings.Builder to decrease allocations [#530][#530]

[v0.48.0-sumo-0]: https://github.com/SumoLogic/sumologic-otel-collector/releases/tag/v0.48.0-sumo-0
[#519]: https://github.com/SumoLogic/sumologic-otel-collector/pull/519
[#525]: https://github.com/SumoLogic/sumologic-otel-collector/pull/525
[#530]: https://github.com/SumoLogic/sumologic-otel-collector/pull/530
[#534]: https://github.com/SumoLogic/sumologic-otel-collector/pull/534
[#536]: https://github.com/SumoLogic/sumologic-otel-collector/pull/536

## [v0.47.0-sumo-0]

### Released 2022-04-05

Welcome to the Sumo Logic Distribution for OpenTelemetry Collector!

With this release, we are officially out of beta status and in GA, as in General Availability. 🎉

This means the software is ready to be used by all customers (without signing up for beta program)
and is commercially fully supported by Sumo Logic in production environments.

Starting with this release, we are using upstream [OpenTelemetry Collector][otc] version numbers
as the base for the OT Distro version numbers.
This means that Sumo Logic Distribution for OpenTelemetry Collector `v0.47.0-sumo-0` is based on `v0.47.0`
of the OpenTelemetry Collector [core][otc_v0_47_0] and [contrib][contrib_v0_47_0] packages.

[otc]: https://github.com/open-telemetry/opentelemetry-collector
[otc_v0_47_0]: https://github.com/open-telemetry/opentelemetry-collector/releases/v0.47.0
[contrib_v0_47_0]: https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/v0.47.0

### Fixed

- fix(k8sprocessor): fix metadata dependencies by @astencel-sumo [#513]

### Other

- refactor(sumologicexporter): optimize fields stringification by @pmalek-sumo [#517]
- refactor(sumologicexporter): optimize compressor using sync.Pool by @pmalek-sumo [#518]

All changes: [v0.0.58-beta.0...v0.47.0-sumo-0]

[v0.47.0-sumo-0]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.0.58-beta.0...v0.47.0-sumo-0
[#513]: https://github.com/SumoLogic/sumologic-otel-collector/pull/513
[#517]: https://github.com/SumoLogic/sumologic-otel-collector/pull/517
[#518]: https://github.com/SumoLogic/sumologic-otel-collector/pull/518
[v0.0.58-beta.0...v0.47.0-sumo-0]: https://github.com/SumoLogic/sumologic-otel-collector/compare/v0.0.58-beta.0...v0.47.0-sumo-0
