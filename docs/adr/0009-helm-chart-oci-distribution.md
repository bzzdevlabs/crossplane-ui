# 0009. Helm chart distribution via OCI (and GitHub Release assets)

- Date: 2026-04-24
- Status: Accepted
- Deciders: @qgerard
- Tags: release, distribution, supply-chain

## Context

The CI workflow already packages the Helm chart on tag push
(`helm package deploy/helm/crossplane-ui`) and produces an
`actions/upload-artifact` entry. That artifact:

- expires after ~90 days (default GitHub retention);
- is visible only to users who navigate to the workflow run;
- is not reachable via `helm install`;
- is not part of any GitHub Release asset list, so semantic versioning
  buys nothing for discoverability.

Meanwhile the container images are pushed to `ghcr.io/bzzdevlabs/...`
with SLSA provenance and SBOM attestations. The chart has no equivalent
distribution surface. A user who wants to install a pinned version
today has to clone the repo at the tag — awkward for a Helm chart,
which is precisely the artifact designed to be installed without
cloning.

Helm 3.8+ treats OCI registries as first-class chart repositories
(`helm push oci://…`, `helm install oci://…`), and GHCR has supported
OCI Helm charts since 2023.

## Decision

We publish the chart through **two complementary channels**, both cut
by the tag-triggered `package-chart` job:

1. **OCI on GHCR** — the primary channel.
   `helm push dist/crossplane-ui-X.Y.Z.tgz oci://ghcr.io/bzzdevlabs/crossplane-ui/charts`
   Users install with:
   ```bash
   helm install crossplane-ui \
     oci://ghcr.io/bzzdevlabs/crossplane-ui/charts/crossplane-ui \
     --version X.Y.Z
   ```

2. **Tarball attached to the GitHub Release** — the browsable channel.
   `gh release upload vX.Y.Z dist/crossplane-ui-X.Y.Z.tgz` so users
   browsing releases on GitHub can download the chart directly without
   `helm` on their machine, and so the release page is a complete
   artifact inventory (images visible via "Packages", chart tarball
   attached, provenance attestations linked).

Both carry:

- **SLSA build provenance** via `actions/attest-build-provenance`;
- **Keyless cosign signature** on the OCI reference
  (`cosign sign ghcr.io/bzzdevlabs/crossplane-ui/charts/crossplane-ui:X.Y.Z`),
  using the workflow's OIDC identity (no long-lived keys).

Chart version stays in lock-step with the repo `vX.Y.Z` tag, bumped by
Release Please (ADR-0008). `appVersion` tracks the same value so the
chart's default image tags match.

## Consequences

Positive:

- `helm install oci://…` becomes the one-liner install path documented
  in the README, competitive with every other CNCF project's UX.
- Both Artifact Hub and Docker Hub index OCI Helm charts on GHCR —
  discovery comes for free once we publish.
- SBOM + provenance + signature form a verifiable supply chain for the
  chart, matching what we already ship for the images.
- No extra infrastructure: GHCR is already the image registry; one less
  system than GitHub Pages + `chart-releaser` or a ChartMuseum
  deployment.

Negative:

- OCI tags cannot contain `+`; if we ever use SemVer build metadata
  (`1.2.3+build4`) Helm silently rewrites to `1.2.3_build4` on push,
  and `cosign verify` must use the underscore form. Mitigated: Release
  Please emits clean `vX.Y.Z` tags without build metadata.
- `public` visibility on the `charts/crossplane-ui` GHCR package must
  be set manually once, the first time it is pushed.
- Consumers on Helm < 3.8 cannot use the OCI channel. The `.tgz`
  release asset covers them.

## Alternatives considered

- **GitHub Pages + `chart-releaser`**. Rejected. Adds a second
  distribution system with its own auth model, and `index.yaml` tends
  to grow unbounded. OCI is the direction Helm itself is pushing.
- **ChartMuseum or Harbor**. Rejected. Requires running a service we
  don't otherwise need. OCI on GHCR is free and ambient.
- **Only attach `.tgz` to GitHub Releases**. Rejected. Browsable but
  not installable; `helm install` still requires a manual download
  step.
- **Only push OCI, skip release assets**. Rejected. Anyone browsing
  the release page would see images and no chart, making it look
  incomplete; attaching the tarball has negligible cost.

## References

- <https://helm.sh/docs/topics/registries/>
- <https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry>
- <https://docs.sigstore.dev/cosign/signing/signing_with_blobs/>
- [ADR-0008 — Release Please for versioning and release automation](0008-release-please-for-versioning.md)
