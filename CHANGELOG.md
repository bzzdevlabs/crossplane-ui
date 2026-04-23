# [1.1.0](https://github.com/bzzdevlabs/crossplane-ui/compare/v1.0.0...v1.1.0) (2026-04-23)


### Features

* **gateway:** embed Vue SPA via //go:embed ([c1e9574](https://github.com/bzzdevlabs/crossplane-ui/commit/c1e95749cf7cef554216ea53cc731ef7faa2bb02))

# 1.0.0 (2026-04-23)


### Features

* **auth:** add bcrypt password helper ([4823223](https://github.com/bzzdevlabs/crossplane-ui/commit/4823223c4be8b696d80a9d08f68846ea4a0e9619))
* **auth:** initial local-user service scaffold ([a2e7411](https://github.com/bzzdevlabs/crossplane-ui/commit/a2e74119951b41e6e9e7e5b610a686c48e9e2806))
* **auth:** reconcile User and Group CRs, sync Dex, seed admin ([8fd3d62](https://github.com/bzzdevlabs/crossplane-ui/commit/8fd3d62e377278a45c8ed57ca161de33260fc35a))
* **chart:** initial Helm chart with embedded Dex dependency ([d6dd09d](https://github.com/bzzdevlabs/crossplane-ui/commit/d6dd09d126965812e8760bce7edf68766d30b14c))
* **chart:** ship User, Group and Dex storage CRDs ([1b1497b](https://github.com/bzzdevlabs/crossplane-ui/commit/1b1497bd269a6d2f380ebc658b48bbcfd5f51356))
* **chart:** wire auth RBAC, OIDC client Secret and Dex OAuth2Client ([52be90d](https://github.com/bzzdevlabs/crossplane-ui/commit/52be90dec81f93ab0dbcbcf0e1fd2308d02890b4))
* **compose:** local k3d dev stack with Dex, auth, gateway, UI ([27f0239](https://github.com/bzzdevlabs/crossplane-ui/commit/27f02390ed84895b9628db706d87c02bcea3645e))
* **gateway:** add Crossplane aggregated summary and generic resource CRUD ([6325c84](https://github.com/bzzdevlabs/crossplane-ui/commit/6325c8452bccdd4efa441c9f10d25172e98821d7))
* **gateway:** expose /api/v1/config for UI OIDC bootstrap ([3e10912](https://github.com/bzzdevlabs/crossplane-ui/commit/3e10912308732949fc85ad9796601ed20bd0f7dc))
* **gateway:** initial service with kube impersonation, OIDC and metrics ([fcb8017](https://github.com/bzzdevlabs/crossplane-ui/commit/fcb8017fc30a4cd8af0b24b8f28add91a33cd559))
* **pkg:** add User and Group API types for the auth service ([35a759c](https://github.com/bzzdevlabs/crossplane-ui/commit/35a759ce4bfe76107604dd98abcb95b4e5ab9326))
* **ui:** add Rancher-like resource detail and create views ([bd1faa8](https://github.com/bzzdevlabs/crossplane-ui/commit/bd1faa89bdcbc3d7a0726fecd9af5210c9be7b24))
* **ui:** app shell with product nav, kind registry and resource templates ([c778a1b](https://github.com/bzzdevlabs/crossplane-ui/commit/c778a1b6274ff9e4eb3f048f6260af7fbb12b203))
* **ui:** initial Vue 3 app with FR/EN i18n and auth store ([f7e371a](https://github.com/bzzdevlabs/crossplane-ui/commit/f7e371ad364eb6d505b374263c93c2d800c90050))
* **ui:** OIDC PKCE login with session store and silent renew ([79074de](https://github.com/bzzdevlabs/crossplane-ui/commit/79074de439cf6d61c81deb4664e1113dd2a4fbe9))
* **ui:** Rancher-inspired shell with side nav and namespaces tile ([b85561d](https://github.com/bzzdevlabs/crossplane-ui/commit/b85561de6f29c7a637308d4da1bfb809851383c0))
* **ui:** typed forms for Composition, XRD, Provider, Function and ProviderConfig ([b91d7fb](https://github.com/bzzdevlabs/crossplane-ui/commit/b91d7fb6b01943a945f8569650a8bd85b257a37a))

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

[Unreleased]: https://github.com/bzzdevlabs/crossplane-ui/compare/main...HEAD
