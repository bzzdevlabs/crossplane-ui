# Security policy

## Supported versions

Until we reach the `v1.0.0` milestone only the `main` branch is supported. After
`v1.0.0`, the two most recent minor releases will receive security fixes.

| Version | Status            |
| ------- | ----------------- |
| `main`  | Supported (alpha) |
| < 1.0   | Pre-release       |

## Reporting a vulnerability

**Please do not open a public issue for security reports.**

Report vulnerabilities privately by email to
[quentingerard8@gmail.com](mailto:quentingerard8@gmail.com) with
the subject prefix `[crossplane-ui][security]`. Encrypt the message with the
maintainer's PGP key (fingerprint advertised on the same email address's
Keybase / key server listing).

Please include:

- A description of the issue and its potential impact.
- Steps to reproduce (proof of concept, version, configuration).
- Your name and contact so that we can credit you.

We aim to acknowledge reports within **three business days** and to publish a
fix within **30 days** for high-severity issues, earlier when possible.

## Disclosure policy

- We practice coordinated disclosure.
- Once a fix is available and deployed on the supported versions, we publish
  a GitHub Security Advisory and update the [CHANGELOG](CHANGELOG.md).
- Reporters are credited unless they request otherwise.

## Scope

In scope:

- The `gateway` and `auth` services (code under `services/`).
- The Vue UI (`web/ui/`).
- The Helm chart (`deploy/helm/crossplane-ui/`) and its default values.
- Dependencies we vendor.

Out of scope (report upstream):

- Vulnerabilities in [Crossplane](https://github.com/crossplane/crossplane),
  [Dex](https://github.com/dexidp/dex), [Kubernetes](https://kubernetes.io/),
  or any other third-party component we merely depend on.

## Security hardening in the chart

The default Helm values enable, at a minimum:

- Non-root, read-only root filesystem containers with all capabilities dropped.
- `seccompProfile: RuntimeDefault` and `runAsNonRoot: true`.
- `NetworkPolicy`s restricting traffic to the strict minimum.
- Mandatory TLS on the public ingress (via `cert-manager` when available).
- `ServiceAccount`s with least-privilege `(Cluster)Role`s.

See [docs/security.md](docs/security.md) for the full threat model.
