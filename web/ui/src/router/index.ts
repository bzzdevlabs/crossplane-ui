import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router';

import { useAuthStore } from '@/stores/auth';

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/components/shell/AppShell.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'home',
        component: () => import('@/views/HomeView.vue'),
      },
      {
        path: 'crossplane',
        name: 'crossplane-dashboard',
        component: () => import('@/views/CrossplaneDashboardView.vue'),
      },
      {
        path: 'crossplane/_create',
        name: 'resource-create',
        component: () => import('@/views/ResourceCreateView.vue'),
      },
      {
        path: 'crossplane/:resource',
        name: 'resource-list',
        component: () => import('@/views/ResourceListView.vue'),
      },
      {
        path: 'crossplane/:resource/:name',
        name: 'resource-detail',
        component: () => import('@/views/ResourceDetailView.vue'),
      },
      {
        path: 'users',
        name: 'users',
        component: () => import('@/views/UsersView.vue'),
      },
      {
        path: 'users/_create',
        name: 'user-create',
        component: () => import('@/views/UserFormView.vue'),
      },
      {
        path: 'users/:name',
        name: 'user-detail',
        component: () => import('@/views/UserFormView.vue'),
      },
      {
        path: 'groups',
        name: 'groups',
        component: () => import('@/views/GroupsView.vue'),
      },
      {
        path: 'groups/_create',
        name: 'group-create',
        component: () => import('@/views/GroupFormView.vue'),
      },
      {
        path: 'groups/:name',
        name: 'group-detail',
        component: () => import('@/views/GroupFormView.vue'),
      },
      {
        path: 'connectors',
        name: 'connectors',
        component: () => import('@/views/ConnectorsView.vue'),
      },
      {
        path: 'connectors/_create',
        name: 'connector-create',
        component: () => import('@/views/ConnectorFormView.vue'),
      },
      {
        path: 'connectors/:id',
        name: 'connector-detail',
        component: () => import('@/views/ConnectorFormView.vue'),
      },
      {
        path: 'settings',
        name: 'settings',
        component: () => import('@/views/SettingsView.vue'),
      },
    ],
  },
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/LoginView.vue'),
    meta: { requiresAuth: false },
  },
  {
    path: '/auth/callback',
    name: 'auth-callback',
    component: () => import('@/views/AuthCallbackView.vue'),
    meta: { requiresAuth: false },
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: () => import('@/views/NotFoundView.vue'),
    meta: { requiresAuth: false },
  },
];

export const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
});

router.beforeEach(async (to) => {
  const auth = useAuthStore();
  if (!auth.ready) {
    await auth.initialise();
  }
  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: 'login', query: { redirect: to.fullPath } };
  }
  if (to.name === 'login' && auth.isAuthenticated) {
    return { name: 'home' };
  }
  return true;
});
