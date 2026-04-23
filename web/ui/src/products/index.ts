import type { Component } from 'vue';
import type { RouteLocationRaw } from 'vue-router';
import { Boxes, House, Settings as SettingsIcon, Users as UsersIcon } from 'lucide-vue-next';

export interface NavItem {
  readonly labelKey: string;
  readonly to: RouteLocationRaw;
  readonly routeNames?: readonly string[];
}

export interface NavGroup {
  readonly labelKey?: string;
  readonly items: readonly NavItem[];
}

export interface Product {
  readonly id: 'home' | 'crossplane' | 'users' | 'settings';
  readonly labelKey: string;
  readonly subtitleKey?: string;
  readonly icon: Component;
  readonly badgeColor: string;
  readonly defaultRoute: RouteLocationRaw;
  readonly groups: readonly NavGroup[];
}

export const PRODUCTS: readonly Product[] = [
  {
    id: 'home',
    labelKey: 'products.home.label',
    subtitleKey: 'products.home.subtitle',
    icon: House,
    badgeColor: '#2e4fd9',
    defaultRoute: { name: 'home' },
    groups: [],
  },
  {
    id: 'crossplane',
    labelKey: 'products.crossplane.label',
    subtitleKey: 'products.crossplane.subtitle',
    icon: Boxes,
    badgeColor: '#6c5ce7',
    defaultRoute: { name: 'crossplane-dashboard' },
    groups: [
      {
        labelKey: 'nav.sections.overview',
        items: [
          {
            labelKey: 'products.crossplane.dashboard',
            to: { name: 'crossplane-dashboard' },
            routeNames: ['crossplane-dashboard'],
          },
        ],
      },
      {
        labelKey: 'nav.sections.crossplane',
        items: [
          {
            labelKey: 'nav.compositions',
            to: { name: 'resource-list', params: { resource: 'compositions' } },
            routeNames: ['resource-list', 'resource-detail', 'resource-create'],
          },
          {
            labelKey: 'nav.xrds',
            to: { name: 'resource-list', params: { resource: 'xrds' } },
          },
          {
            labelKey: 'nav.composites',
            to: { name: 'resource-list', params: { resource: 'composites' } },
          },
          {
            labelKey: 'nav.managed',
            to: { name: 'resource-list', params: { resource: 'managed' } },
          },
          {
            labelKey: 'nav.providers',
            to: { name: 'resource-list', params: { resource: 'providers' } },
          },
          {
            labelKey: 'nav.functions',
            to: { name: 'resource-list', params: { resource: 'functions' } },
          },
          {
            labelKey: 'nav.providerConfigs',
            to: { name: 'resource-list', params: { resource: 'providerconfigs' } },
          },
        ],
      },
    ],
  },
  {
    id: 'users',
    labelKey: 'products.users.label',
    subtitleKey: 'products.users.subtitle',
    icon: UsersIcon,
    badgeColor: '#0984e3',
    defaultRoute: { name: 'users' },
    groups: [
      {
        labelKey: 'nav.sections.administration',
        items: [{ labelKey: 'nav.users', to: { name: 'users' } }],
      },
    ],
  },
  {
    id: 'settings',
    labelKey: 'products.settings.label',
    subtitleKey: 'products.settings.subtitle',
    icon: SettingsIcon,
    badgeColor: '#636e72',
    defaultRoute: { name: 'settings' },
    groups: [
      {
        labelKey: 'nav.sections.configuration',
        items: [{ labelKey: 'nav.settings', to: { name: 'settings' } }],
      },
    ],
  },
];

export function productById(id: string): Product | undefined {
  return PRODUCTS.find((p) => p.id === id);
}

// Derive the active product from the current route name. Route names are
// prefixed by convention (`crossplane-*`, `users*`, `settings*`); the home
// route is the default.
export function productForRouteName(name: string | null | undefined): Product {
  if (!name) return PRODUCTS[0]!;
  if (name.startsWith('crossplane') || name.startsWith('resource-')) {
    return productById('crossplane')!;
  }
  if (name.startsWith('users')) return productById('users')!;
  if (name.startsWith('settings')) return productById('settings')!;
  return PRODUCTS[0]!;
}
