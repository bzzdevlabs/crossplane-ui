# web/ui

The Vue 3 frontend of **crossplane-ui**. Ships as a static bundle that is
embedded into the `gateway` Go binary at build time.

## Stack

- [Vue 3.5](https://vuejs.org/) + [TypeScript 5.7](https://www.typescriptlang.org/) (`<script setup>`)
- [Vite 7](https://vitejs.dev/) dev server and bundler
- [Pinia](https://pinia.vuejs.org/) for state
- [vue-router 4](https://router.vuejs.org/)
- [vue-i18n 11](https://vue-i18n.intlify.dev/) — FR + EN first-class
- [Vitest 3](https://vitest.dev/) for unit tests
- [ESLint 9](https://eslint.org/) (flat config) + [Prettier](https://prettier.io/) + [Stylelint](https://stylelint.io/)
- [`@rancher/components`](https://www.npmjs.com/package/@rancher/components)
  _(added in milestone M5 for the Rancher look & feel)_

## Scripts

```bash
pnpm install        # install dependencies
pnpm dev            # Vite dev server on :5173 (proxy /api → gateway)
pnpm build          # type-check + production bundle
pnpm test           # Vitest once
pnpm test:watch     # Vitest in watch mode
pnpm lint           # ESLint + Stylelint
pnpm format         # Prettier write
```

## Layout

```
web/ui/
├── index.html
├── vite.config.ts
├── vitest.config.ts
├── eslint.config.ts
├── tsconfig.json + tsconfig.{app,node,vitest}.json
├── env.d.ts
└── src/
    ├── main.ts            App bootstrap (Pinia, router, i18n)
    ├── App.vue
    ├── router/            Vue Router + navigation guards
    ├── stores/            Pinia stores
    ├── i18n/              vue-i18n + locales/{en,fr}.json
    ├── layouts/           Shared layouts
    ├── views/             Page-level components
    ├── components/        Shared presentational components (wired in M5+)
    ├── composables/       Reusable composition functions (wired in M5+)
    └── styles/            Global stylesheet
```

## Aliases

- `@/` → `src/`

## Environment variables

Variables are read from `import.meta.env.*` and must start with `VITE_`.

| Variable              | Purpose                                                                         |
| --------------------- | ------------------------------------------------------------------------------- |
| `VITE_APP_TITLE`      | Overrides the browser title (default: `crossplane-ui`).                         |
| `VITE_API_BASE_URL`   | Base URL of the gateway's REST API. Defaults to the current origin.             |
| `VITE_OIDC_ISSUER_URL`| Dex issuer URL used for the client-side OIDC redirect flow (wired in M4).       |
| `VITE_OIDC_CLIENT_ID` | Dex client ID used by the UI (wired in M4).                                     |
