# 0001. Record architecture decisions

- Date: 2026-04-22
- Status: Accepted
- Deciders: @qgerard
- Tags: process

## Context

We need a durable, reviewable record of the architectural choices that shape
crossplane-ui. Without it, new contributors re-litigate settled questions and
decisions drift when the code changes without documentation.

## Decision

We will adopt the [ADR format proposed by Michael Nygard](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)
(lightweight Markdown, numbered, append-only). ADRs live under
[`docs/adr/`](.), are prefixed with a 4-digit number, and are tracked in the
[README](README.md) of that directory.

Status transitions:

- `Proposed` → `Accepted` happens on merge.
- `Accepted` → `Superseded by NNNN` when a later ADR replaces it.
- ADRs are never rewritten — corrections come from a new ADR that references
  the previous one.

## Consequences

Positive:

- Future contributors find rationale in-repo, not in lost Slack threads.
- The ADR register doubles as a table of contents of the project's
  architecture.
- ADRs scale with the codebase without adding external tooling.

Negative:

- Small overhead per decision (one file).
- Discipline required to actually write them (we aim for one per significant
  architectural choice, not per commit).

## Alternatives considered

- **Confluence / Notion pages** — rejected because the docs leave the code,
  require auth, and rot when out of reach of MRs.
- **No formal record** — rejected because we have already accumulated
  non-trivial decisions (Go, Vue, Dex, impersonation) that we want to
  document.

## References

- <https://adr.github.io>
- <https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions>
