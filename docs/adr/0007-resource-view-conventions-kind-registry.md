# 0007. Resource view conventions & kind registry

- Date: 2026-04-23
- Status: Accepted
- Deciders: @qgerard
- Tags: ui, ux, resources, rancher-alignment

## Context

Since M6–M8 the UI exposes Crossplane resources through three generic views:

- [HomeView](../../web/ui/src/views/HomeView.vue) — one card grid per category
  (`composition`, `composite`, `managed`, `provider`, `function`).
- [ResourceDetailView](../../web/ui/src/views/ResourceDetailView.vue) — a
  single page with a Form/YAML toggle driven by
  [FormShell](../../web/ui/src/components/forms/FormShell.vue) and a
  kind-specific typed form from
  [components/forms/schemas.ts](../../web/ui/src/components/forms/schemas.ts).
- [ResourceCreateView](../../web/ui/src/views/ResourceCreateView.vue) — a
  template picker on top of the same form shell.

These views got us to parity on CRUD. They do not match the richer resource
experience that Rancher Manager offers and that operators expect:

- **List pages** are tables first (not card grids), with sortable columns,
  filter input, checkbox selection, row-level overflow menus, and bulk actions
  (Download YAML, Delete). Cards are a secondary view for dashboards.
- **Detail pages** are tabbed: `Details / Config / Events / Related Resources
  / Conditions / YAML`. The header shows a state pill next to the title, a
  meta row (namespace, age, provider, API version, …) and a right-aligned
  action cluster with an overflow (3-dots) menu.
- **Form pages** for complex resources (XRD, Composition) use a left sidebar
  of sub-sections (`Basics / Versions / Member Roles / Add-On Config / …`), a
  multi-column field grid, and a sticky footer with `Cancel / Edit as YAML /
  Save`.
- Every kind has its own column set, status rules, tabs, and form sections.

Our current architecture cannot express this per-kind variation: there is one
detail component, one create component, and the form-schema registry only
decides which typed form component renders inside FormShell. To support M9+
(Operations, Claims, Providers needing package details, Dex connectors) we
need a pluggable way to declare per-kind UX.

## Decision

We will introduce shared **view templates** for List / Detail / Form, and a
**`ResourceKindRegistry`** that maps each GVR to its view configuration.

### Templates

All resource pages compose three layout templates:

- `<ResourceListTemplate>`
  - Header: breadcrumb + title + count badge + primary action (`Create`) +
    filter input + view toggle (table / cards).
  - Body: `<DataTable>` with declared `columns` (key, label, sortable,
    formatter) and `rows`. Checkbox selection and row overflow menu built in.
  - Footer bulk-action bar (shown when rows are selected).
- `<ResourceDetailTemplate>`
  - Header: breadcrumb, title, `<StatusPill>`, meta row (slots), right-aligned
    actions (`Refresh / Delete / Apply` + overflow).
  - Body: `<Tabs>` with declared tabs — at minimum `Details`, `YAML`,
    `Events`, `Related`, `Conditions`; additional tabs declared per kind.
- `<ResourceFormTemplate>`
  - Optional left sub-nav (sections).
  - Body: slot for the typed form (multi-column `<FieldRow>` grid) or the
    YAML editor.
  - Sticky footer: `Cancel / Edit as YAML / Apply` with dirty + saving state.

### Kind registry

A new `src/resources/registry.ts` declares, per kind:

```ts
interface ResourceKind {
  id: string;                 // short slug, used in routes: 'compositions'
  label: MessageKey;          // i18n label
  pluralLabel: MessageKey;
  gvr: { group: string; version: string; resource: string };
  scope: 'Cluster' | 'Namespaced';
  icon?: string;

  // List view
  columns: Column[];
  defaultSort?: { column: string; order: 'asc' | 'desc' };

  // Detail view
  statusFor: (obj: KubeObject) => StatusPillVariant;
  tabs?: TabDescriptor[];     // extra tabs beyond the default five

  // Form view (reuses existing FormSchema from schemas.ts)
  form?: FormSchema;
  formSections?: FormSection[]; // enables left sub-nav
}
```

The existing `FORM_SCHEMAS` registry is absorbed into `ResourceKindRegistry`
(via `form` field). No schema logic changes; the move is purely organisational.

`HomeView` becomes an aggregated dashboard (totals, health tiles, recent
events) rather than the primary list.

### Status model

Kubernetes `conditions` vary per CRD. We define a small canonical variant set
(`ready`, `degraded`, `pending`, `errored`, `unknown`) and each kind supplies
a `statusFor(obj)` function mapping its shape to that set. `<StatusPill>`
renders a uniform outlined badge in the matching color. This replaces the
ad-hoc green/red badges currently inlined in HomeView.

## Consequences

Positive:

- One implementation of List / Detail / Form UX — enforced consistency,
  easier Playwright coverage (M10).
- Adding a new kind is "fill in the registry" rather than "write three views".
- Per-kind columns, tabs, status rules, and form sections all live in one
  place.
- The homepage can focus on dashboard value (totals, health, events) instead
  of duplicating list screens.

Negative / cost:

- Three new template components plus `<DataTable>`, `<StatusPill>`, `<Tabs>`,
  `<ActionMenu>`, `<StickyFooter>` primitives (ADR-0006 phase-2).
- Existing typed forms get rewrapped in the new form template; schemas keep
  their current shape.
- Additional per-kind code (one registry entry per kind) — acceptable and
  centralised.

## Alternatives considered

- **Keep one generic detail view and add conditional slots.** Rejected: does
  not model tabs cleanly; grows into a god-component; diverges sharply from
  Rancher.
- **Render everything from CRD OpenAPI schemas (fully generic).** Rejected as
  v1 goal: schema-driven forms work for simple shapes but give a poor UX on
  the kinds that matter most (Composition, XRD). Registry entries stay
  compatible with a future schema-driven fallback for unregistered kinds.
- **Per-kind views without a shared template (pure components).** Rejected:
  reinvents layout each time, diverges visually, multiplies test surface.

## References

- Rancher Manager UX reference screenshots, captured 2026-04-23.
- [ADR-0006. UI shell](0006-ui-shell-product-nav-namespace-scope.md) —
  templates live inside the contextual product area defined there.
- [components/forms/schemas.ts](../../web/ui/src/components/forms/schemas.ts)
  — current form registry, absorbed by the kind registry.
- [docs/roadmap.md](../roadmap.md) — M9+ adds Operations, Claims, Dex
  connectors, all of which benefit directly from this registry.
