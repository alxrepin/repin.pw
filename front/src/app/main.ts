import { createSSRApp, type Plugin, type App as VueApp } from 'vue';
import type { Router } from 'vue-router';
import App from '@/App.vue';
import { createAppRouter } from '@/app/router';
import { createStateManager, STATE_KEY, type StateManager } from '@/shared/ssr/state';

export interface AppContext {
  app: VueApp;
  router: Router;
  state: StateManager;
}

export function createApp(head: Plugin, initialState: Record<string, unknown> = {}): AppContext {
  const app = createSSRApp(App);
  const router = createAppRouter();
  const state = createStateManager(initialState);

  app.use(router);
  app.use(head);
  app.provide(STATE_KEY, state);

  return { app, router, state };
}
