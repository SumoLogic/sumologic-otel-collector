# Upstream Relation

This document describes what is purpose of this repository,
why do we provide another OpenTelemetry Collector distrubution and
how look like our versioning and breaking changes policies.

## Purpose of Sumo Logic Distribution for OpenTelemetry Collector

Sumo Logic Distribution for OpenTelemetry Collector is the OpenTelemetry Collector with addition of Sumo Logic specific changes.

**Our aim is to extend and not to replace the OpenTelemetry Collector.**

We work closely with the OpenTelemetry community in order to improve overall experience by:

- attending SIG meetings
- creating issues in upstream if we find a bug or have a feature request
- issuing pull requests in upstream with bug fixes and features

Releasing our own distribution allows us to include Sumo Logic platform specific features
and to provide better customer support:

- bypass the OpenTelemetry release schedule for critical bug fixes
- provide customer oriented features
- provide various installation methods
- quick response to customer needs
- use case oriented documentation

## Versioning policy

We are using upstream OpenTelemetry Collector version numbers as the base for the Sumo Logic Distribution for OpenTelemetry Collector version numbers.
This means that Sumo Logic Distribution for OpenTelemetry Collector `v0.47.0-sumo-0` is based on `v0.47.0`
of the OpenTelemetry Collector core and contrib packages.

In order to prevent confusion we are going to add Sumo Logic specific features
when updating the OpenTelemetry Collector version.
The OpenTelemetry Collector is released every two weeks and we are releasing up to week after the date.

In case of critical fixes in our code, we will provide version with the same base like OT Distro,
but changed suffix, e.g. `v0.47.0-sumo-1`.

## Breaking changes policy

As the OpenTelemetry Collector is constanly changing they can be some breaking changes between releases,
and we inherit them. In addition we follow the same policy, so every minor update can contain breaking changes.
Due to that every upgrade should be preceded with careful reading of release notes.
