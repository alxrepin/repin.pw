import type { RouteRecordRaw } from 'vue-router';

// Must stay last among the layout children: vue-router matches in order and
// this pattern swallows everything.
export const errorRoutes: RouteRecordRaw[] = [
  {
    path: ':pathMatch(.*)*',
    name: 'not-found',
    component: () => import('./pages/NotFoundPage.vue'),
  },
];
