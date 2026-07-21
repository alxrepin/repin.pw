import {
  createMemoryHistory,
  createRouter,
  createWebHistory,
  type RouteRecordRaw,
  type Router,
} from 'vue-router';

import { errorRoutes } from '@/modules/error/routes';
import { homeRoutes } from '@/modules/home/routes';
import { rememberListPath } from '@/modules/posts/composables/useBackToList';
import { postsRoutes } from '@/modules/posts/routes';

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('@/layouts/DefaultLayout.vue'),
    children: [...homeRoutes, ...postsRoutes, ...errorRoutes],
  },
];

export function createAppRouter(): Router {
  const router = createRouter({
    history: import.meta.env.SSR ? createMemoryHistory() : createWebHistory(),
    routes,
    scrollBehavior(to, from, savedPosition) {
      if (savedPosition) return savedPosition;

      if (to.path === '/' && from.path === '/' && to.query.page !== from.query.page) {
        const grid = document.querySelector('#posts-grid');

        if (grid) {
          const rect = grid.getBoundingClientRect();
          if (rect.top >= 0 && rect.bottom <= window.innerHeight) return false;
        }

        return { el: '#posts-grid', top: 88, behavior: 'smooth' };
      }

      return { top: 0 };
    },
  });

  if (!import.meta.env.SSR) {
    router.afterEach(to => {
      if (to.name === 'home') rememberListPath(to.fullPath);
    });
  }

  return router;
}
