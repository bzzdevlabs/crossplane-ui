# 0005. Target Crossplane v2 only

- Date: 2026-04-22
- Status: Accepted
- Deciders: @qgerard
- Tags: scope, compatibility

## Context

Crossplane v2 (latest v2.2 at the time of this writing) introduces
fundamental API changes compared to v1:

- Resources can be namespace-scoped.
- `Operations` replace imperative actions previously driven by `Function`s.
- Managed resources must be explicitly **activated** before reconciliation.
- Composition functions become the default composition strategy.

Crossplane v1 is in its deprecation window. Supporting both versions would
require a dual-mode backend (different typed Go clients, duplicate domain
logic) and double the testing surface.

## Decision

We target **Crossplane v2 only** (`>= 2.0`). The chart declares this via
`kubeVersion` + a `NOTES.txt` check; the gateway fails fast on startup if it
detects v1 CRDs.

## Consequences

Positive:

- Single, modern client code path.
- No cruft carrying v1-only quirks forward.
- Operations, activation and namespaced resources can be first-class in the
  UI.

Negative:

- Users on Crossplane v1 must upgrade before installing crossplane-ui.
  Mitigated by v1 already being deprecated upstream.

## Alternatives considered

- **Dual v1 + v2 support**. Rejected for the reasons above, and because v1
  is going away regardless.
- **v1 only**. Rejected: we'd be shipping a dead product.

## References

- <https://docs.crossplane.io/latest/whats-new/>
- <https://github.com/crossplane/crossplane/releases>
