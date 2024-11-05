# Releasing

- [How to release](#how-to-release)
  - [Check end-to-end tests](#check-end-to-end-tests)
  - [Determine the Workflow Run ID to release](#determine-the-workflow-run-id-to-release)
    - [Find the package build number](#find-the-package-build-number)
    - [Find the collector workflow run](#find-the-collector-workflow-run)
  - [Trigger the release](#trigger-the-release)
  - [Publish GitHub release](#publish-github-release)
  - [Release packages](#release-packages)

## How to release

### Check end-to-end tests

Check if the Sumo internal e2e tests are passing.

### Determine the Workflow Run ID to release

We can begin the process of creating a release once QE has given a thumbs up for
a given package version. We can determine the Workflow Run ID to use for a
release using the following steps:

#### Find the package build number

Each package has a build number and it's included in the package version &
filename. For example, if the package version that QE validates is 0.108.0-1790
then the build number is 1790.

#### Find the collector workflow run

Each package uses binaries built from this repository. We can find the workflow
used to build the binaries by tracing our way back from the package build number.

The build number corresponds directly to the GitHub Run Number for a packaging
workflow run in GitHub Actions. Unfortunately, there does not currently appear to
be a way to reference a workflow run using the run number. Instead, we can use
one of two methods to find the workflow run:

#### Option 1 - Use the `gh` cli tool to find the workflow

```shell
BUILD_NUMBER="1790"
gh run list -R sumologic/sumologic-otel-collector-packaging -s success \
-w build_packages.yml -L 200 -b main --json displayTitle,number,url \
-q ".[] | select(.number == ${BUILD_NUMBER})"
```

This will output a number of fields, for example:

```json
{
  "displayTitle": "Build for Remote Workflow: 11672946742, Version: 0.108.0-sumo-1\n",
  "number": 1790,
  "url": "https://github.com/SumoLogic/sumologic-otel-collector-packaging/actions/runs/11673248730"
}
```

We need the number to the right of `Build for Remote Workflow`. This number is
the ID of the workflow run that built the binaries used in the package.

#### Option 2 - Search the GitHub website manually

Manually search for the run number on the
[Build packages workflow][build_workflow] page. Search for the build number
(e.g. 1790) until you find the corresponding workflow.

![Finding the packaging workflow run][release_0]

Once you've found the packaging workflow run, we need the number to the right of
`Build for Remote Workflow`. This number is
the ID of the workflow run that built the binaries used in the package.

![Finding the collector workflow ID][release_1]

### Trigger the release

Now that we have the Workflow Run ID we can trigger a release. There are two
methods of doing this.

#### Option 1 - Use the `gh` cli tool to trigger the release

A release can be triggered by using the following command:

```shell
# Be sure to replace this with workflow ID from the previous step
WORKFLOW_ID=11672946742
gh workflow run releases.yml -R sumologic/sumologic-otel-collector -f workflow_id=${WORKFLOW_ID}
```

#### Option 2 - Use the GitHub website to trigger the release

Navigate to the [Publish release][releases_workflow] workflow in GitHub Actions.
Find and click the `Run workflow` button on the right-hand side of the page.
Fill in the Workflow Run ID from the previous step and then click the green
`Run workflow` button.

![Triggering a release][release_2]

### Publish GitHub release

The GitHub release is created as draft by the
[releases](../.github/workflows/releases.yml) GitHub Action.

After the release draft is created, go to [GitHub releases](https://github.com/SumoLogic/sumologic-otel-collector/releases),
edit the release draft and fill in missing information:

- Specify versions for upstream OT core and contrib releases
- Copy and paste the Changelog entry for this release from [CHANGELOG.md][changelog]

After verifying that the release text and all links are good, publish the release.

### Release packages

The docs for triggering a package release can be found in the
[Packaging Release docs][release_packaging] in the
[sumologic-otel-collector-packaging][packaging_repo] repository.

[changelog]: ../CHANGELOG.md
[build_workflow]: https://github.com/SumoLogic/sumologic-otel-collector-packaging/actions/workflows/build_packages.yml?query=branch%3Amain
[releases_workflow]: https://github.com/SumoLogic/sumologic-otel-collector/actions/workflows/releases.yml
[release_packaging]: https://github.com/SumoLogic/sumologic-otel-collector-packaging/blob/main/docs/release.md
[packaging_repo]: https://github.com/SumoLogic/sumologic-otel-collector-packaging
[release_0]: ../images/release_0.png
[release_1]: ../images/release_1.png
[release_2]: ../images/release_2.png
