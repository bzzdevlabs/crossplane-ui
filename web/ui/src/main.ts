import { createApp } from 'vue';
import { createPinia } from 'pinia';

import App from './App.vue';
import { router } from './router';
import { i18n } from './i18n';
import { useAuthStore } from '@/stores/auth';

import '@fontsource/lato/latin-400.css';
import '@fontsource/lato/latin-700.css';
import '@fontsource/lato/latin-900.css';
import './styles/main.css';

async function bootstrap(): Promise<void> {
  const app = createApp(App);

  app.use(createPinia());
  app.use(router);
  app.use(i18n);

  // Resolve the session (existing OIDC cookie, silent-renew hook, dev mode)
  // BEFORE the first navigation so guards don't bounce the user to /login.
  const auth = useAuthStore();
  try {
    await auth.initialise();
  } catch {
    // Boot errors (gateway down, bad OIDC config) are surfaced through the
    // login view; the router remains functional.
  }

  await router.isReady();
  app.mount('#app');
}

void bootstrap();
