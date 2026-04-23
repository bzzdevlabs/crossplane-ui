# 0006. UI shell: product-level nav + contextual sidebar + namespace scope

- Date: 2026-04-23
- Status: Accepted
- Deciders: @qgerard
- Tags: ui, ux, navigation, rancher-alignment

## Context

The UI shell delivered in M5 is a single flat sidebar with a hardcoded list of
sections (`Cluster`, `Crossplane`, `Administration`) plus a minimal topbar
(locale toggle, user display, logout). It served M5–M8 well while the feature
set was narrow (home, detail, create). It does not scale to the Crossplane v2
surface we target from M9 onwards (Dex connectors UI, auth-provider
configuration, per-namespace filtering, cluster management views).

We use Rancher Manager as our UX reference — it is the tool operators in our
target audience already know, and the look-and-feel alignment is stated in
M5's exit criteria and in the project README. Rancher Manager uses a two-level
information architecture that the current shell does not model:

1. A **global "product" switcher** (hamburger overlay) exposing top-level areas
   such as *Home*, *Explore Cluster*, *Cluster Management*, *Users &
   Authentication*, *Global Settings*.
2. A **contextual sidebar** that changes entirely depending on the active
   product (e.g. *Cluster Management* shows `Clusters / Cloud Credentials /
   Drivers / Pod Security Policies`; *Users & Authentication* shows `Users /
   Roles / Groups / Auth Provider`).
3. A **namespace/project picker** pinned in the topbar with grouped options
   (`All Namespaces`, `Only User Namespaces`, `Only System Namespaces`, `Only
   Namespaced Resources`, `Only Cluster Resources`, `Project: X`). The active
   selection scopes every list view in the current product.

Forcing everything into one flat sidebar would either bury administration
items, mix concerns (Crossplane resources with Dex connector config), or
require long scrolling nav columns.

## Decision

We will restructure the UI shell around three primitives:

### 1. Global product switcher

A hamburger overlay (left-edge icon in the topbar) opens a full-height panel
listing the top-level products. For v1 we ship:

- **Home** — aggregated dashboard (counts, health, recent events).
- **Crossplane** — Compositions, CompositeResourceDefinitions, Composites,
  Managed Resources, Providers, Functions, ProviderConfigs, Operations.
- **Users & Authentication** — Users, Groups, Dex connectors.
- **Settings** — cluster-scoped config, about.

Products are declared in a `products.ts` registry (id, label, icon, default
route, sidebar groups). The list is extensible — a new product is added by
appending one registry entry, not by editing the shell.

### 2. Contextual sidebar

The left sidebar renders the active product's groups/items only. Each product
defines its nav tree:

```ts
{
  id: 'crossplane',
  label: t('products.crossplane'),
  icon: 'cx',
  defaultRoute: { name: 'crossplane-dashboard' },
  groups: [
    { label: t('nav.crossplane'), items: [
      { label: t('nav.compositions'), to: { name: 'compositions-list' } },
      { label: t('nav.xrds'),         to: { name: 'xrds-list' } },
      // ...
    ]},
  ],
}
```

The sidebar header displays a colored product badge (letter/icon tile + label)
matching Rancher's cluster badges. A collapse toggle persists to
`localStorage`.

### 3. Topbar with namespace picker + action cluster

The topbar, from left to right:

- Hamburger button (opens product switcher).
- Active product badge + title.
- **Namespace picker** (centre-right) with grouped options listed above. The
  selection is stored in a new `useUiStore()` Pinia store and mirrored to the
  URL (`?namespace=`) so deep links work. List views filter by it; detail
  views keep the existing `?namespace=` param semantics.
- Action icons: search (command palette, stub for now), refresh, user avatar
  menu (name, email, locale toggle, sign out).

### 4. Routing

Routes gain a product segment. We move from:

```text
/                                  → home
/resources/:group/:version/:resource/:name
/resources/new
```

to:

```text
/                                  → home (aggregated)
/crossplane                        → crossplane dashboard
/crossplane/:resource              → list page for a kind
/crossplane/:resource/:name        → detail page (namespace via query)
/crossplane/:resource/_create      → create page
/users                             → users list
/settings                          → settings
```

The `:resource` segment is the plural REST resource name (`compositions`,
`xrds`, `providers`, …). `group` and `version` are looked up from the kind
registry (see ADR-0007) rather than embedded in the URL — shorter, stable
URLs that survive API bumps.

Existing bookmarked URLs are rare (pre-GA) and will not be redirected.

## Consequences

Positive:

- The shell scales to new products without sidebar congestion.
- Operators familiar with Rancher find their bearings instantly.
- Namespace scoping becomes a global concern, not a per-page prop: list views
  read from the store, filtering is consistent, URL-sharable.
- The product registry lets future milestones (Operations, Multi-cluster) add
  top-level areas with zero shell code change.

Negative / cost:

- Significant refactor: [DefaultLayout.vue](../../web/ui/src/layouts/DefaultLayout.vue)
  is replaced by a new `AppShell` that composes `ProductSwitcher`, `ProductNav`,
  `Topbar`. New components: `NamespacePicker`, `UserMenu`, `ProductBadge`.
- Router tree is rewritten; view files move under `views/products/<product>/`.
- New `ui` Pinia store for `activeProduct`, `sidebarCollapsed`, `namespace`,
  `productSwitcherOpen`.
- i18n keys reshuffle (nav keys get `nav.<product>.<item>` scoping).
- Pre-GA bookmarks break. Acceptable — no external users yet.

## Alternatives considered

- **Keep the flat sidebar, add grouping only.** Rejected: the Rancher reference
  was explicit in M5 and the screenshots gathered during planning; grouping
  alone does not solve context switching between Crossplane and admin areas,
  nor does it address the namespace-picker gap.
- **Top-nav horizontal tabs for products.** Rejected: Rancher does not use
  this; horizontal tabs do not scale once we add more than 5–6 products; the
  sidebar space is needed for contextual items regardless.
- **Server-driven navigation (RBAC-filtered items from the API).** Rejected
  *for now*: valuable but orthogonal and premature. The registry is a client
  primitive; an API-driven filter can wrap it later without breaking
  downstream code.

## References

- Rancher Manager UX reference screenshots, captured 2026-04-23.
- [ADR-0005. Crossplane v2 only](0005-crossplane-v2-only.md) — v2-specific
  resources (Operations, Claims, Activation) are among the products to model.
- [docs/roadmap.md](../roadmap.md) — M9+ scope includes auth-provider UI which
  drives the need for multi-product IA.
