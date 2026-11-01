# Supported versions and security maintenance

| Branch | Go module | CEL module | Support |
| --- | --- | --- | --- |
| main | github.com/flanksource/gomplate/v4 | cel.dev/cel-go | Active development |
| 3.x | github.com/flanksource/gomplate/v3 | github.com/google/cel-go | Security fixes only |

The v3 maintenance line starts at v3.24.90. There is no announced end-of-support date. Security patches preserve the v3 API, expression behavior, and CEL import path. Each patch must identify its advisory and use the lowest compatible fixed dependency version. Ordinary feature and dependency refreshes belong on main. Security fixes should be applied to every affected supported branch.

Renovate processes both branches. Routine updates are disabled on 3.x, while vulnerability fixes require review. OSV checks direct dependencies; daily Go vulnerability and production web dependency scans supplement them, including transitive dependencies. A scheduled scan runs from main and explicitly checks out both supported branches. Failed scans require investigation; scanner errors are not passing results.

Releases use v4.x tags from main and v3.x patch tags from 3.x. Existing downstream v3 dependency-update PR automation runs only for v3 releases. Consumers must migrate explicitly to v4.

## Migrating to v4

The initial security maintenance patch updates Go to 1.26.10 (including GO-2026-6809), x/net to 0.62.0 (GO-2026-6860 and GO-2026-6866), OpenTelemetry to 1.42.0 (GO-2026-5158), and their required dependency graph. This includes fixed x/text and x/mod versions. The web lockfile resolves DOMPurify 3.4.17, addressing the five advisories reported for 3.4.13. CEL remains at 0.31.0 on 3.x.

Change Go imports from github.com/flanksource/gomplate/v3 to github.com/flanksource/gomplate/v4. When supplying CEL environment options or using AST, evaluation, or value types, change CEL imports from github.com/google/cel-go to cel.dev/cel-go. Do not mix the old and new CEL types.

The initial v4 migration must use a Conventional Commit with a BREAKING CHANGE footer (or the existing MAJOR RELEASE footer) so semantic-release selects v4.0.0. The release analyzer and notes generator recognize both. Local preparation does not publish a release; publishing happens only after the reviewed migration is merged to main.

## Verify release configuration without publishing

Run from the repository root:

```sh
npm install --prefix .tmp/release-check --package-lock=false --no-audit --no-fund @semantic-release/commit-analyzer@13.0.1 yaml@2.8.1 semver@7.7.2
node hack/verify-release.mjs
go test ./internal/releasepolicy
```

This checks synthetic breaking and security-fix commits against the actual commit analyzer, parses the workflow YAML, and checks the v3 downstream-update guard. It does not create commits, tags, releases, or external PRs.
