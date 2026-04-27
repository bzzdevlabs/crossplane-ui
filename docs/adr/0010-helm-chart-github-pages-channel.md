# 0010. Add GitHub Pages as a third Helm chart distribution channel

- Date: 2026-04-27
- Status: Accepted
- Deciders: @qgerard
- Tags: release, distribution

## Context

ADR-0009 chose two channels for the Helm chart: OCI on GHCR (primary)
and a tarball attached to the GitHub Release. Once we shipped the first
release this way, two friction points surfaced:

1. **GHCR UI is image-shaped**. The package page renders `docker pull
   ghcr.io/.../charts/crossplane-ui:X.Y.Z` regardless of OCI media type.
   The chart is technically pullable via `helm pull oci://…`, but the UI
   leads users away from the Helm CLI. ADR-0009 dismissed this as a
   GitHub UX limitation; in practice it is the first thing every visitor
   sees on the package page, so the cost is non-zero.
2. **No `helm repo add` workflow**. OCI installs require the full
   `oci://…` URL with `--version` every time. The traditional
   `helm repo add <name> <url>` + `helm search repo` + `helm install
   <repo>/<chart>` flow — which most CNCF projects (cert-manager,
   ingress-nginx, kube-prometheus-stack, …) ship as their primary
   install path — is not available.

ADR-0009 rejected GitHub Pages because of "second auth model" and
"`index.yaml` grows unbounded" concerns. Both were overstated:

- `gh-pages` is just a branch in the same repo; auth is the same
  `GITHUB_TOKEN` we already use everywhere else.
- `index.yaml` grows by one entry per release. A few KB per year.

## Decision

We add **GitHub Pages** as a third distribution channel, alongside OCI
and Release tarball.

- The `.tgz` continues to live **only** on the GitHub Release (single
  source of truth). The `gh-pages` branch carries **only `index.yaml`**,
  whose entry URLs point back at the release download path
  (`https://github.com/bzzdevlabs/crossplane-ui/releases/download/vX.Y.Z/crossplane-ui-X.Y.Z.tgz`).
- The `package-chart` job, after `gh release upload`, refreshes
  `index.yaml` via `helm repo index --merge` in a worktree on
  `gh-pages`, then commits and pushes.
- Pages is configured in repo settings to serve `gh-pages` / root.
  Public URL: `https://bzzdevlabs.github.io/crossplane-ui`.
- ADR-0009 stays in effect for OCI + tarball; this ADR adds a channel,
  it does not remove any.

User-facing install paths after this change:

```bash
# Channel A — Helm repo (recommended discovery path)
helm repo add crossplane-ui https://bzzdevlabs.github.io/crossplane-ui
helm repo update
helm install crossplane-ui crossplane-ui/crossplane-ui --version X.Y.Z

# Channel B — OCI (one-liner, no `helm repo add`)
helm install crossplane-ui \
  oci://ghcr.io/bzzdevlabs/crossplane-ui/charts/crossplane-ui --version X.Y.Z

# Channel C — Manual tarball (offline / air-gapped)
gh release download vX.Y.Z -p 'crossplane-ui-*.tgz'
helm install crossplane-ui ./crossplane-ui-X.Y.Z.tgz
```

## Consequences

Positive:

- The standard CNCF install UX (`helm repo add` + `helm search` +
  `helm install`) becomes available; new users land on a familiar path.
- Artifact Hub indexes Pages-hosted Helm repos out of the box —
  submitting `https://bzzdevlabs.github.io/crossplane-ui` once gets the
  chart listed.
- No `.tgz` duplication: the binary lives only on the Release; Pages
  carries the index.

Negative:

- One extra workflow step (write to `gh-pages`) and one extra branch in
  the repo. Both cheap.
- If the `gh-pages` worktree push races with another in-flight job
  modifying the same branch, the second push needs a rebase. The
  workflow is single-jobbed for releases, so the practical risk is
  near-zero.

## Alternatives considered

- **`helm/chart-releaser-action`**. Rejected. It wants to manage its
  own GitHub Releases (`<chart>-<version>`), which would collide with
  Release Please's `vX.Y.Z` releases. Manual `helm repo index --merge`
  is fewer moving parts and gives us full control of URLs.
- **Drop OCI, keep only Pages**. Rejected. OCI install is the
  direction Helm and most CNCF projects are headed; removing it would
  trade a forward-looking channel for a backward-compatible one.
- **Drop Pages, fix UX with chart-only annotations on GHCR**. Rejected.
  Annotations like `org.opencontainers.image.title` do not change the
  GHCR UI; the rendering is hard-coded to "Container image".

## References

- [ADR-0009 — Helm chart distribution via OCI](0009-helm-chart-oci-distribution.md)
- <https://helm.sh/docs/topics/chart_repository/>
- <https://docs.github.com/en/pages/getting-started-with-github-pages>
