# 0002. Go + Vue 3 in a single monorepo

- Date: 2026-04-22
- Status: Accepted
- Deciders: @qgerard
- Tags: language, layout

## Context

The project needs a backend that talks to the Kubernetes API and a frontend
that feels at home next to Rancher Manager's UI. We also want easy local
development, a single CI pipeline and a single release surface.

## Decision

- Backend services are written in **Go 1.26**. The maintainer has a decade of
  Go experience; the Kubernetes / Crossplane ecosystem is Go-native
  (`client-go`, `controller-runtime`, `crossplane-runtime`), so we get
  first-class libraries at zero cost.
- Frontend is **Vue 3.5 + TypeScript + Vite + Pinia + vue-i18n**. Rancher
  Dashboard uses the same stack and publishes the `@rancher/components`
  package, which we consume to inherit the Rancher look & feel.
- Everything lives in a **single Git monorepo** (this one), with a Go
  workspace (`go.work`) and a pnpm project for the UI.

## Consequences

Positive:

- Cross-service refactors are a single MR.
- CI runs matrix builds in one pipeline.
- Consistent governance (one `LICENSE`, `CHANGELOG`, `CONTRIBUTING`).
- Straightforward `task bootstrap` developer experience.

Negative:

- Contributors need both Go and Node toolchains installed.
- Very large repos become slower to clone over time; we will revisit
  if the repo grows beyond a reasonable size.

## Alternatives considered

- **Separate repos per service**. Rejected: splits history, complicates
  releases, multiplies boilerplate (governance × N).
- **React instead of Vue**. Rejected because reusing `@rancher/components`
  is a major time saver for achieving a Rancher-like UI.
- **Rust backend**. Rejected because the Kubernetes Rust ecosystem is less
  mature than Go's; cost-benefit does not justify the switch for an admin
  UI.

## References

- [`@rancher/components`](https://www.npmjs.com/package/@rancher/components)
- [Rancher Dashboard source](https://github.com/rancher/dashboard)
