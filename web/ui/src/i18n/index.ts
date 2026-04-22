import { createI18n } from 'vue-i18n';

import en from './locales/en.json';
import fr from './locales/fr.json';

export type Locale = 'en' | 'fr';
export type MessageSchema = typeof en;

const STORAGE_KEY = 'crossplane-ui.locale';

function resolveInitialLocale(): Locale {
  const saved = typeof window !== 'undefined' ? window.localStorage.getItem(STORAGE_KEY) : null;
  if (saved === 'en' || saved === 'fr') return saved;
  if (typeof navigator !== 'undefined' && navigator.language.startsWith('fr')) return 'fr';
  return 'en';
}

export const i18n = createI18n({
  legacy: false,
  locale: resolveInitialLocale(),
  fallbackLocale: 'en',
  missingWarn: false,
  fallbackWarn: false,
  messages: { en, fr },
});

export function setLocale(locale: Locale): void {
  // i18n.global.locale is a WritableComputedRef<string> in non-legacy mode.
  (i18n.global.locale as unknown as { value: Locale }).value = locale;
  if (typeof window !== 'undefined') {
    window.localStorage.setItem(STORAGE_KEY, locale);
    document.documentElement.lang = locale;
  }
}
