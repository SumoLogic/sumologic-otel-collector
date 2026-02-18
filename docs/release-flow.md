# Release Flow

Complete release process for Sumo Logic OpenTelemetry Collector.

## Table of Contents

- [Steps](#steps)
  - [Step 1: Renovate Bot](#step-1-renovate-bot)
  - [Step 2: Merge Dependency PR](#step-2-merge-dependency-pr)
  - [Step 3: Prepare Release PR](#step-3-prepare-release-pr)
  - [Step 4: Find Build Number](#step-4-find-build-number)
  - [Step 5: CI to RC Promotion](#step-5-ci-to-rc-promotion)
  - [Step 6: QE Sign-off](#step-6-qe-sign-off)
  - [Step 7: Release Orchestrator](#step-7-release-orchestrator)
- [References](#references)

## Steps

### Step 1: Renovate Bot

1. Run Renovate bot ([renovate-scheduler.yml](../.github/workflows/renovate-scheduler.yml))
2. Wait for PR: `chore(deps): bump opentelemetry-collector`

**Sample PR**: [#1944][pr_1944]

### Step 2: Merge Dependency PR

1. Check CI status
2. **If CI passes**: Merge
3. **If CI fails**: Resolve upstream issues first, then merge

### Step 3: Prepare Release PR

1. Create Python virtual environment:

   ```bash
   python3 -m venv venv
   source venv/bin/activate
   ```

2. Install dependencies:

   ```bash
   pip install towncrier==23.6.0
   make install-prettier
   ```

3. Update changelog:

   ```bash
   make update-changelog VERSION=0.144.0-sumo-0
   ```

4. Create and merge PR:

   ```bash
   git checkout -b prepare-release-0.144.0-sumo-0
   git commit -m "chore: prepare release 0.144.0-sumo-0"
   git push origin prepare-release-0.144.0-sumo-0
   ```

**Sample PR**: [#1965][pr_1965]

### Step 4: Find Build Number

1. Go to [Dev Builds](../.github/workflows/dev_builds.yml)
2. Open the workflow run and check the **"Trigger Remote Workflow"** step
3. Get the workflow ID from this step (e.g., `11672946742`)

4. Search for this workflow ID in packaging repo to get build number:

   **CLI:**

   ```bash
   WORKFLOW_ID="11672946742"
   gh run list -R sumologic/sumologic-otel-collector-packaging \
     -s success -w build_packages.yml -L 200 -b main \
     --json displayTitle,number,url \
     -q ".[] | select(.displayTitle | contains(\"${WORKFLOW_ID}\"))"
   ```

   Output will show the build number (e.g., `"number": 1790`)

   **Manual:**
   - Go to [Packaging Build][packaging_build_workflow]
   - Search for your workflow ID
   - Note build number (e.g., `#1790`)

5. Collector version format: `0.144.0-1790` (VERSION-BUILD_NUMBER)

**Reference**: [release.md](./release.md#determine-the-workflow-run-id-to-release)

### Step 5: CI to RC Promotion

1. Go to [sumologic-otel-collector-packaging][packaging_repo]
2. Actions → **[CI-to-RC Promotion][ci_to_rc_workflow]** workflow
3. Enter:
   - **Version**: `0.144.0-1790` (full version with build number)
4. Workflow will automatically:
   - Promote packaging artifacts (ci → rc)
   - Promote container images (ci → rc)

**Reference**: [CI to RC Promotion Guide][ci_to_rc_guide]

### Step 6: QE Sign-off

1. Notify QE team
2. Wait for testing (1-3 days)
3. Get formal approval

### Step 7: Release Orchestrator

1. Go to [sumologic-otel-collector-packaging][packaging_repo]
2. Actions → **[Release Orchestrator][release_orchestrator_workflow]** workflow
3. Enter:
   - **Package version**: `0.144.0-1790` (full version with build number)
4. Workflow will automatically:
   - Promote packages (rc → stable)
   - Create draft releases for collector, packaging, and containers

## **Publish Draft Releases (⚠️ IMPORTANT ORDER):**

**The orchestrator only creates draft releases. You must publish them manually in this exact order:**

1. **FIRST**: Publish [Collector Release][collector_releases]
   - Edit draft release
   - Add upstream OTel core and contrib versions
   - Copy changelog from [CHANGELOG.md](../CHANGELOG.md)
   - Publish release
   - Wait for [post-release workflow](../.github/workflows/post-release.yml) to complete (creates package tags)

2. **SECOND**: Publish [Packaging Release][packaging_releases]
   - Review draft release
   - Publish release

3. **THIRD**: Publish [Containers Release][containers_releases]
   - Review draft release
   - Publish release

**Reference**: [Release Orchestrator Guide][release_orchestrator_guide]

## References

- [release.md](./release.md) - Manual release process
- [CONTRIBUTING.md](../CONTRIBUTING.md) - Changelog process
- [Packaging Repo][packaging_repo] - Release workflows
- [Container Repo][containers_repo] - Container releases

[pr_1944]: https://github.com/SumoLogic/sumologic-otel-collector/pull/1944
[pr_1965]: https://github.com/SumoLogic/sumologic-otel-collector/pull/1965
[collector_releases]: https://github.com/SumoLogic/sumologic-otel-collector/releases
[packaging_repo]: https://github.com/SumoLogic/sumologic-otel-collector-packaging
[packaging_build_workflow]: https://github.com/SumoLogic/sumologic-otel-collector-packaging/actions/workflows/build_packages.yml
[packaging_releases]: https://github.com/SumoLogic/sumologic-otel-collector-packaging/releases
[ci_to_rc_workflow]: https://github.com/SumoLogic/sumologic-otel-collector-packaging/actions/workflows/ci-to-rc-promotion.yml
[ci_to_rc_guide]: https://github.com/SumoLogic/sumologic-otel-collector-packaging/blob/main/docs/ci-to-rc-promotion.md
[release_orchestrator_workflow]: https://github.com/SumoLogic/sumologic-otel-collector-packaging/actions/workflows/release-orchestrator.yml
[release_orchestrator_guide]: https://github.com/SumoLogic/sumologic-otel-collector-packaging/blob/main/docs/release-orchestrator.md
[containers_repo]: https://github.com/SumoLogic/sumologic-otel-collector-containers
[containers_releases]: https://github.com/SumoLogic/sumologic-otel-collector-containers/releases
