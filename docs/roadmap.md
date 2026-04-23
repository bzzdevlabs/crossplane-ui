# Roadmap

The project is organised in incremental milestones. Each milestone produces a
working slice end-to-end; early ones focus on foundations and plumbing, later
ones add Crossplane-specific functionality.

| Milestone | Scope                                                                                                 | Exit criteria                                      |
| --------- | ----------------------------------------------------------------------------------------------------- | -------------------------------------------------- |
| ~~M1~~ ✓  | Repository scaffolding, tooling, licensing, docs skeleton, CI skeleton.                               | `task lint test` green on a fresh clone.           |
| ~~M2~~ ✓  | Gateway: Kubernetes client, OIDC middleware, impersonation, real Prometheus metrics.                  | `GET /api/v1/namespaces` serves live cluster data. |
| ~~M3~~ ✓  | Auth: `User` / `Group` CRDs, controller-runtime reconciler, bootstrap admin, bcrypt, Dex config sync. | `kubectl get users` works after `helm install`.    |
| ~~M4~~ ✓  | Helm polish + Dex wiring end-to-end (password DB connector from Auth).                                | `helm install` → login as admin succeeds.          |
| ~~M5~~ ✓  | UI: OIDC PKCE login flow, Rancher-inspired shell (nav, topbar), session store, home namespaces list. | Browser login end-to-end; matches Rancher look.    |
| ~~M6~~ ✓  | Home dashboard: card/tile view of all Compositions/XRs/MRs with aggregated ready/sync status.         | Dashboard renders live status from kube.           |
| ~~M7~~ ✓  | Detail view Rancher-like: monaco YAML editor, apply (server-side), delete, create-from-template.     | Full CRUD on Crossplane resources via UI.          |
| ~~M8~~ ✓  | Configuration (form) views for Composition, XRD, Provider, Function, ProviderConfig.                  | Typed forms for the 5 most common kinds.           |
| **M9**    | Embed Vue SPA into gateway (`//go:embed`) + multi-stage Dockerfile. One image ships API + UI.         | `docker run gateway` serves the UI and /api/v1/*.  |
| **M10**   | SSO: Dex connectors UI (LDAP / SAML / GitHub / Google), docs, sample `values.yaml` per provider.      | SSO login end-to-end against test IdP.             |
| **M11**   | Playwright e2e, `helm unittest`, `semantic-release`, chart publishing, v1.0.0 GA.                     | GA artefacts released by CI.                       |

## Beyond v1

Ideas for post-GA work, not scheduled:

- Multi control-plane support (kubeconfig registry + context switcher).
- Composition graph visualisation (d3 / vis-network).
- OTEL tracing across gateway/auth/Dex.
- Rancher Extensions packaging (so the UI can be embedded into Rancher Manager).
- CLI companion (`crossplane-ui login/apply/diff`).
- Audit log viewer (Kubernetes audit events cross-referenced with UI actions).
