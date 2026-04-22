# Architecture Decision Records

This directory collects ADRs following the lightweight format introduced by
Michael Nygard — see <https://adr.github.io>.

## How to add an ADR

1. Copy `0000-template.md` to `NNNN-short-title.md`, where `NNNN` is the next
   4-digit number.
2. Fill in the sections. Keep it short (a page or two).
3. Open a merge request. The ADR stays in `Proposed` status until merged; the
   merge itself transitions it to `Accepted`.
4. ADRs are **append-only**: if you change your mind later, write a new ADR
   that supersedes the old one (and update the old one's `Status` to
   `Superseded by NNNN`).

## Index

| #    | Title                                                                 | Status   |
| ---- | --------------------------------------------------------------------- | -------- |
| 0001 | [Record architecture decisions](0001-record-architecture-decisions.md) | Accepted |
| 0002 | [Go + Vue 3 monorepo](0002-go-and-vue3-monorepo.md)                    | Accepted |
| 0003 | [Kubernetes user impersonation for RBAC](0003-k8s-impersonation-rbac.md)| Accepted |
| 0004 | [Dex as the only authentication surface](0004-dex-as-auth-surface.md)  | Accepted |
| 0005 | [Crossplane v2 only](0005-crossplane-v2-only.md)                       | Accepted |
