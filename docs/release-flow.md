# Release Flow

Complete release process for Sumo Logic OpenTelemetry Collector.

## Before You Start

1. Start a thread in **#pd-ot-collector-release** channel
2. Upload the release checklist document
   - Use this [sample release checklist](https://docs.google.com/document/d/17GPloLV18vJAQ5p9UPSV-KUFQnKEhrsqzF-eXGhAnuE/edit?tab=t.0) as template
3. Track all release activities in this thread

## Steps

### Step 1: Renovate Bot

1. Run Renovate bot ([renovate-scheduler.yml](../.github/workflows/renovate-scheduler.yml))
2. Wait for PR: `chore(deps): bump opentelemetry-collector`

**Sample PR**: [#1944](https://github.com/SumoLogic/sumologic-otel-collector/pull/1944)

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

**Sample PR**: [#1965](https://github.com/SumoLogic/sumologic-otel-collector/pull/1965)

### Step 4: Find Build Number

1. Go to [Dev Builds](https://github.com/SumoLogic/sumologic-otel-collector/actions/workflows/dev_builds.yml)
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
   - Go to [Packaging Build](https://github.com/SumoLogic/sumologic-otel-collector-packaging/actions/workflows/build_packages.yml)
   - Search for your workflow ID
   - Note build number (e.g., `#1790`)

6. Collector version format: `0.144.0-1790` (VERSION-BUILD_NUMBER)

**Reference**: [release.md](./release.md#determine-the-workflow-run-id-to-release)

### Step 5: CI to RC Promotion

1. Go to [sumologic-otel-collector-packaging](https://github.com/SumoLogic/sumologic-otel-collector-packaging)
2. Actions → **[CI-to-RC Promotion](https://github.com/SumoLogic/sumologic-otel-collector-packaging/actions/workflows/ci-to-rc-promotion.yml)** workflow
3. Enter:
   - **Version**: `0.144.0-1790` (full version with build number)
4. Workflow will automatically:
   - Promote packaging artifacts (ci → rc)
   - Promote container images (ci → rc)

**Reference**: [CI to RC Promotion Guide](https://github.com/SumoLogic/sumologic-otel-collector-packaging/blob/main/docs/ci-to-rc-promotion.md)

### Step 6: QE Sign-off

1. Notify QE team
2. Wait for testing (1-3 days)
3. Get formal approval

### Step 7: Release Orchestrator

1. Go to [sumologic-otel-collector-packaging](https://github.com/SumoLogic/sumologic-otel-collector-packaging)
2. Actions → **[Release Orchestrator](https://github.com/SumoLogic/sumologic-otel-collector-packaging/actions/workflows/release-orchestrator.yml)** workflow
3. Enter:
   - **Package version**: `0.144.0-1790` (full version with build number)
4. Workflow will automatically:
   - Promote packages (rc → stable)
   - Create draft releases for collector, packaging, and containers

## **Publish Draft Releases (⚠️ IMPORTANT ORDER):**

**The orchestrator only creates draft releases. You must publish them manually in this exact order:**

1. **FIRST**: Publish [Collector Release](https://github.com/SumoLogic/sumologic-otel-collector/releases)
   - Edit draft release
   - Add upstream OTel core and contrib versions
   - Copy changelog from [CHANGELOG.md](../CHANGELOG.md)
   - Publish release
   - Wait for [post-release workflow](https://github.com/SumoLogic/sumologic-otel-collector/actions/workflows/post-release.yml) to complete (creates package tags)

2. **SECOND**: Publish [Packaging Release](https://github.com/SumoLogic/sumologic-otel-collector-packaging/releases)
   - Review draft release
   - Publish release

3. **THIRD**: Publish [Containers Release](https://github.com/SumoLogic/sumologic-otel-collector-containers/releases)
   - Review draft release
   - Publish release

**Reference**: [Release Orchestrator Guide](https://github.com/SumoLogic/sumologic-otel-collector-packaging/blob/main/docs/release-orchestrator.md)


## References

- [release.md](./release.md) - Manual release process
- [CONTRIBUTING.md](../CONTRIBUTING.md) - Changelog process  
- [Packaging Repo](https://github.com/SumoLogic/sumologic-otel-collector-packaging) - Release workflows
- [Container Repo](https://github.com/SumoLogic/sumologic-otel-collector-containers) - Container releases
