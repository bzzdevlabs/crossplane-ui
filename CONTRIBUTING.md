# Contributing to crossplane-ui

First off, thanks for considering a contribution. This document describes how
to set up a development environment, the coding conventions we follow, and how
to get changes merged.

By contributing to this project you agree to abide by our
[Code of Conduct](CODE_OF_CONDUCT.md) and to license your contributions under
the [Apache License, Version 2.0](LICENSE).

## Development environment

Prerequisites:

- [Go](https://go.dev/) 1.26 or newer
- [Node.js](https://nodejs.org/) 22 or newer
- [pnpm](https://pnpm.io/) 10 or newer (`corepack enable` will install it)
- [Task](https://taskfile.dev/) 3 or newer (the repo's task runner)
- [Helm](https://helm.sh/) 4 or newer
- A local Kubernetes cluster with Crossplane v2 ([kind](https://kind.sigs.k8s.io/)
  is the easiest choice)
- [pre-commit](https://pre-commit.com/) to run formatters on every commit

Bootstrap the repo:

```bash
task bootstrap   # Go tools + pnpm install + pre-commit install
```

Common tasks:

```bash
task lint        # golangci-lint + ESLint + stylelint + helm lint + yamllint
task test        # go test + vitest
task build       # cross-compile Go binaries and build the Vue app
task dev         # run gateway + auth + Vite dev server with hot reload
task images      # build container images locally (docker/podman)
```

See [`Taskfile.yml`](Taskfile.yml) for the full list.

## Coding conventions

### Go

- Target **Go 1.26**. Use the standard library wherever possible.
- Format with `gofumpt` and `goimports` (enforced by `golangci-lint`).
- Follow the [Effective Go](https://go.dev/doc/effective_go) and
  [CodeReviewComments](https://github.com/golang/go/wiki/CodeReviewComments)
  guidelines.
- Log with [`log/slog`](https://pkg.go.dev/log/slog) and the JSON handler.
- Return wrapped errors (`fmt.Errorf("...: %w", err)`) — never log-and-return.
- Use `context.Context` as the first argument of any function that performs I/O.
- Public packages live under `pkg/`; service-private code lives under
  `services/<svc>/internal/`.

### TypeScript / Vue

- Use `<script setup lang="ts">` single-file components with the Composition API.
- Strict mode is enabled in `tsconfig.json` — no `any`.
- State lives in [Pinia](https://pinia.vuejs.org/) stores, never on `window`.
- UI strings go through [`vue-i18n`](https://vue-i18n.intlify.dev/) (FR + EN).
- Styling uses scoped styles and the [`@rancher/components`](https://www.npmjs.com/package/@rancher/components)
  design tokens whenever a primitive is available.

### Commit messages

We follow [Conventional Commits 1.0.0](https://www.conventionalcommits.org/).
`commitlint` enforces this via the pre-commit hook. Examples:

```
feat(gateway): add OIDC middleware
fix(ui): correct status color for warning state
docs(adr): record decision to adopt impersonation
chore(deps): bump go to 1.26.2
```

Breaking changes MUST use the `!` marker and a `BREAKING CHANGE:` footer.

Releases and the changelog are produced automatically from the commit history
by `semantic-release` on the `main` branch.

## Submitting changes

1. Fork the repository and create a topic branch from `main`.
2. Make your changes; keep commits small and focused.
3. Run `task lint test` locally — CI mirrors these checks.
4. Open a Pull Request targeting `main`. Fill in the PR template.
5. At least one maintainer review and a green CI pipeline are required before
   a merge. We use squash-and-merge to keep history linear.

## Reporting bugs and requesting features

Please open a GitHub issue using the relevant template. Security vulnerabilities
must be reported privately — see [SECURITY.md](SECURITY.md).
