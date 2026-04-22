# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

This file is maintained automatically by
[semantic-release](https://semantic-release.gitbook.io/) via the commit history
(see [CONTRIBUTING.md](CONTRIBUTING.md) for the commit conventions).

## [Unreleased]

### Added

- **Milestone M2 — Kubernetes integration & authentication plumbing**
  in the `gateway`:
  - `internal/metrics` exposes a Prometheus registry with Go runtime,
    process, HTTP and Kubernetes API collectors.
  - `internal/middleware` adds request-ID, access-log, panic-recovery,
    CORS and Prometheus instrumentation, composed via a tiny `Chain`
    helper.
  - `internal/oidc` verifies Dex ID tokens with a split
    discovery / expected-issuer configuration (needed when the browser
    and the gateway see Dex under different hostnames). Ships a
    `DevPassthrough` fallback when `OIDC_ISSUER_URL` is empty.
  - `internal/kube` bundles `LoadConfig` (in-cluster or kubeconfig) and
    a per-request impersonation `ClientFactory`.
  - `internal/api` serves the first domain endpoint,
    `GET /api/v1/namespaces`, listing namespaces via impersonation.
- Local dev stack (compose) now threads OIDC URLs via env and defaults
  to dev pass-through for curl-friendly smoke testing.
- Gateway-side unit tests for every new package, including an httptest
  apiserver verifying impersonation headers, a fake OIDC verifier, and
  a fake `kubernetes.Interface`-based test for the namespaces handler.

### Added — milestone M1 (Foundations)

- Initial repository scaffolding:
  - Apache 2.0 license and governance files (`README`, `CONTRIBUTING`,
    `SECURITY`, `CODE_OF_CONDUCT`, `NOTICE`).
  - Developer tooling: EditorConfig, `.gitignore`, `.gitattributes`,
    VS Code workspace settings, `golangci-lint`, ESLint / Prettier,
    Taskfile, pre-commit and commitlint.
  - Go workspace with two services (`gateway`, `auth`).
  - Vue 3 + Vite + Pinia + vue-i18n UI skeleton.
  - Helm chart skeleton depending on the upstream Dex chart.
  - Architecture Decision Record template (ADR-0001).
  - GitLab CI skeleton with `lint`, `test`, `build`, `package` stages.
- Local dev stack under `compose/` (k3d + Dex + auth + gateway + UI)
  orchestrated by `task dev:up` / `dev:down` / `dev:reset`.

[Unreleased]: https://gitlab.telespazio-digital-factory.fr/icdo/tpzf/crossplane-ui/-/compare/main...HEAD
