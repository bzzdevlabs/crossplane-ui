# 0008. Release Please for versioning and release automation

- Date: 2026-04-24
- Status: Accepted
- Deciders: @qgerard
- Tags: release, ci, supply-chain

## Context

Until now the project used [`semantic-release`](https://semantic-release.gitbook.io/)
to derive the next version from Conventional Commits on `main`, push a
`v*` tag, and create a GitHub Release. That tag push then re-entered the
CI workflow and triggered `package-images`, `package-chart` and
`scan-docker-scout`.

This two-phase model has four concrete problems:

1. **Race window**. The GitHub Release is created during phase 1 (main
   push) and sits with zero assets until phase 2 (tag push) finishes
   building and pushing. Users refreshing the releases page see an empty
   release for several minutes.
2. **Divergent provenance**. Phase 1's workflow run creates the release
   metadata; phase 2's run produces the images and chart. SLSA
   attestations and release notes therefore point at different runs, and
   the two timestamps never line up.
3. **No human gate**. `semantic-release` commits and tags without
   review — handy for libraries, awkward for infrastructure tooling
   where release notes, Chart.yaml bumps and breaking-change callouts
   benefit from a pre-merge read-through.
4. **Ecosystem mismatch**. `semantic-release` is dominant in npm-land
   (≈2.5M weekly downloads) but rare in Go/k8s projects. Kubernetes,
   Crossplane, Argo, Flux, cert-manager, Prometheus and Grafana all
   release via manual tags or bot-opened PRs, not via
   `semantic-release`. Staying on it would lock us into a tooling
   pattern foreign to the ecosystem our users operate in.

## Decision

We adopt [**Release Please**](https://github.com/googleapis/release-please-action)
(Google, v5) as the single source of truth for versioning and release
orchestration.

**Release mode: unified single version.**
`crossplane-ui` ships as one Helm chart containing two Go services
(ADR-0002, ADR-0004). Consumers install one artifact; there is no use
case for `gateway` at v2.3.0 while `auth` is at v1.0.5. Release Please
is therefore configured with a single root package and a single
`vX.Y.Z` tag that governs:

- `ghcr.io/bzzdevlabs/crossplane-ui/gateway:vX.Y.Z`
- `ghcr.io/bzzdevlabs/crossplane-ui/auth:vX.Y.Z`
- `oci://ghcr.io/bzzdevlabs/crossplane-ui/charts/crossplane-ui:X.Y.Z`
  (see ADR-0009)

**Bumps**. Release Please owns `CHANGELOG.md`, `.release-please-manifest.json`
and — via `extra-files` — `deploy/helm/crossplane-ui/Chart.yaml`
(`version:` + `appVersion:`).

**Workflow shape**. A single workflow on `push: main` runs lint → test →
build → `release-please`. When Release Please reports
`release_created == true` in the same run, downstream jobs
(`package-images`, `package-chart`, `scan`) execute immediately using
that run's outputs. One run, one coherent release event.

**Human gate**. Release Please opens a "chore(release): X.Y.Z" PR rather
than committing straight to `main`. Merging the PR triggers the tag +
artifacts. Non-release commits leave `main` untouched.

## Consequences

Positive:

- Release exists with its assets attached within a single workflow run.
  No empty-release window.
- One SLSA provenance predicate covers the whole release event.
- Maintainer can edit the release PR body (add migration notes,
  highlight breaking changes) before the tag is cut.
- Aligns tooling with the Go/cloud-native ecosystem, reducing surprise
  for contributors coming from Crossplane / Kubernetes / Argo projects.
- `CHANGELOG.md` now lives in-repo (Release Please manages it), so
  offline readers and `git blame` can trace what shipped when.

Negative:

- Releases are no longer "zero-click". Someone has to merge the release
  PR. Acceptable tradeoff: that merge is the human gate we want.
- Release Please does not natively sign artifacts; signing stays in the
  downstream jobs (cosign keyless, SLSA provenance, SBOM).
- Migration cost: `.releaserc.json` is removed, existing `v1.0.0` and
  `v1.1.0` tags are preserved; the manifest seeds at `1.1.0`.

## Alternatives considered

- **Keep `semantic-release`, patch the race**. Rejected. The race is
  structural: the release is created before the tag push that builds
  the assets. Papering over it with cross-job `gh release edit` calls
  introduces its own race between parallel jobs, and the ecosystem
  mismatch remains.
- **GoReleaser + manual tags**. Rejected. GoReleaser is excellent for
  building, but ditching conventional-commit-driven versioning would
  mean either hand-writing changelogs or manually deciding each bump,
  regressing on automation.
- **Hybrid: `semantic-release` restricted to auto-tagging, all
  packaging tag-triggered**. Rejected. Still two runs, still a race on
  the release body, and now we maintain two release tools.
- **Per-component versioning** (e.g. `gateway-v2.3.0` +
  `auth-v1.0.5`). Rejected. Components ship together via one chart;
  separate versions would confuse users and force the chart to pin
  image tags that move independently from its own version.

## References

- <https://github.com/googleapis/release-please-action>
- <https://github.com/googleapis/release-please/blob/main/docs/manifest-releaser.md>
- [ADR-0002 — Go + Vue 3 monorepo](0002-go-and-vue3-monorepo.md)
- [ADR-0009 — Helm chart distribution via OCI](0009-helm-chart-oci-distribution.md)
