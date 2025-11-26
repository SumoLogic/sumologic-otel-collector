# Releasing

- [How to release](#how-to-release)
  - [Check end-to-end tests](#check-end-to-end-tests)
  - [Find the package build number](#find-the-package-build-number)
  - [Trigger the release orchestrator](#trigger-the-release-orchestrator)
  - [Publish GitHub releases](#publish-github-releases)

## How to release

### Check end-to-end tests

Check if the Sumo internal e2e tests are passing.

### Find the package build number

We can begin the process of creating a release once QE has given a thumbs up for
a given package version. Each package has a build number and it's included in the
package version & filename. For example, if the package version that QE validates
is 0.108.0-1790 then the build number is **1790**.

This build number is all you need to trigger the release process!

### Trigger the release orchestrator

The [Drafting Release][orchestrator_workflow] workflow automates the entire release
process. It will automatically:

1. Find and validate all related workflow runs (collector, packaging, containers)
2. Create draft releases for all three repositories
3. Promote packaging release candidates to stable
4. Provide a summary with links to all releases

There are two methods to trigger the orchestrator:

#### Option 1 - Use the `gh` cli tool to trigger the release

Run the following command (replace `BUILD_NUMBER` with the build number from QE):

```shell
BUILD_NUMBER="1790"; \
gh workflow run Release_orchestrator.yml -R sumologic/sumologic-otel-collector \
-f package_build_number=${BUILD_NUMBER}
```

#### Option 2 - Use the GitHub website to trigger the release

Navigate to the [Drafting Release][orchestrator_workflow] workflow in GitHub Actions.
Find and click the `Run workflow` button on the right-hand side of the page.
Enter the package build number (e.g., 1790) and click the green `Run workflow` button.

The workflow will automatically discover the related workflow IDs and orchestrate
the entire release process across all three repositories.

### Publish GitHub releases

Once the orchestrator workflow completes successfully, it will provide a summary with
the status of all release operations and direct links to the draft releases.

The orchestrator creates draft releases for all three repositories:

1. **[Collector releases][collector_releases]** (this repository)
2. **[Packaging releases][packaging_releases]**
3. **[Container releases][containers_releases]**

#### Publishing order

⚠️ **IMPORTANT: Releases must be published in the following order:**

1. **Publish the [Collector Release][collector_releases] FIRST**
   - Edit the draft release and add the following information:
     - Specify versions for upstream OT core and contrib releases
     - Copy and paste the Changelog entry for this release from [CHANGELOG.md][changelog]
   - After verifying that the release text and all links are correct, publish the release
   - Publishing this release will automatically trigger the [post-release workflow][post_release_workflow]
     which creates the necessary package tags

2. **Verify the [post-release workflow][post_release_workflow] completed successfully**
   - Ensure the workflow created the required package tags
   - These tags are needed for the packaging and containers releases

3. **Publish the [Packaging Release][packaging_releases]**
   - Review the draft release and publish it

4. **Publish the [Containers Release][containers_releases]**
   - Review the draft release and publish it

The orchestrator workflow summary will provide direct links to all releases and
their current status.

[changelog]: ../CHANGELOG.md
[orchestrator_workflow]: https://github.com/SumoLogic/sumologic-otel-collector/actions/workflows/Release_orchestrator.yml
[post_release_workflow]: https://github.com/SumoLogic/sumologic-otel-collector/actions/workflows/post-release.yml
[collector_releases]: https://github.com/SumoLogic/sumologic-otel-collector/releases
[packaging_releases]: https://github.com/SumoLogic/sumologic-otel-collector-packaging/releases
[containers_releases]: https://github.com/SumoLogic/sumologic-otel-collector-containers/releases
