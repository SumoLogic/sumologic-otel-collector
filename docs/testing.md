# Testing

## Running Tracing E2E tests

We currently have some legacy E2E tests ported from [our OT fork][ot_fork],
which serve as a means of verifying feature parity for tracing as we migrate it
to this distribution. The tests are run by CircleCI on demand for `main` and
release branches, and are defined in the [tests][tracing_tests].

In order to run the tests, go to the [CircleCI page][circleci], choose the branch
you want, and manually approve the workflow to run. Note that you need commiter
rights in this repository to run the tests.

![Approving the workflow in CircleCI][circleci_approve]

[circleci]: https://app.circleci.com/pipelines/github/SumoLogic/sumologic-otel-collector
[circleci_approve]: ../images/circleci_approve_workflow.png
[ot_fork]: https://github.com/SumoLogic/opentelemetry-collector-contrib
[tracing_tests]: ../.circleci/config.yml
