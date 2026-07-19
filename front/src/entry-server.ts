import { createHead, renderSSRHead } from '@unhead/vue/server';
import { renderToString } from 'vue/server-renderer';
import { createApp } from '@/app/main';

export interface RenderResult {
  html: string;
  headTags: string;
  state: string;
  statusCode: number;
}

export async function render(url: string): Promise<RenderResult> {
  const head = createHead();
  const { app, router, state } = createApp(head);

  await router.push(url);
  await router.isReady();

  const html = await renderToString(app);
  const payload = await renderSSRHead(head);

  return {
    html,
    headTags: payload.headTags,
    state: state.serialize(),
    statusCode: state.status.code,
  };
}
