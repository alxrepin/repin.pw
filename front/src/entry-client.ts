import '@unocss/reset/tailwind.css';
import 'virtual:uno.css';
import '@/app/styles/index.scss';

import { createHead } from '@unhead/vue/client';
import { createApp } from '@/app/main';

const head = createHead();
const initialState = window.__INITIAL_STATE__ ?? {};

const { app, router } = createApp(head, initialState);

router.isReady().then(() => {
  app.mount('#app');
  hideBootLoader();
});

function hideBootLoader() {
  const loader = document.getElementById('app-loader');
  if (!loader) return;

  loader.classList.add('is-hidden');
  loader.addEventListener('transitionend', () => loader.remove(), { once: true });

  setTimeout(() => loader.remove(), 600);
}
