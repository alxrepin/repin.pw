import type { RouteRecordRaw } from 'vue-router';

export const postsRoutes: RouteRecordRaw[] = [
  {
    path: 'posts/:slug',
    name: 'post',
    component: () => import('./pages/PostPage.vue'),
  },
];
