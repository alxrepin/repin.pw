import fs from 'node:fs/promises';
import compression from 'compression';
import express from 'express';
import sirv from 'sirv';
import { renderInstantView, TELEGRAM_BOT_UA } from './server/instantview';

interface RenderResult {
  html: string;
  headTags: string;
  state: string;
  statusCode: number;
  redirect?: string;
}

const isProd = process.env.NODE_ENV === 'production';
const port = Number(process.env.PORT ?? 3000);

const apiBase = process.env.API_INTERNAL_URL ?? 'http://localhost:8080';
globalThis.__API_BASE__ = apiBase;

const siteUrl = (process.env.PUBLIC_SITE_URL ?? '').trim().replace(/\/+$/, '');
globalThis.__SITE_URL__ = siteUrl;

const metrikaId = readMetrikaId();

function readMetrikaId(): string {
  const raw = (process.env.YANDEX_METRIKA_ID ?? '').trim();
  if (!raw) return '';

  if (!/^\d+$/.test(raw)) {
    console.warn(`YANDEX_METRIKA_ID must be digits only, got ${JSON.stringify(raw)} — counter off`);
    return '';
  }

  return raw;
}

const metrikaTag = metrikaId
  ? `<script>
(function(m,e,t,r,i,k,a){m[i]=m[i]||function(){(m[i].a=m[i].a||[]).push(arguments)};
m[i].l=1*new Date();k=e.createElement(t),a=e.getElementsByTagName(t)[0],k.async=1,k.src=r;
a.parentNode.insertBefore(k,a)})(window,document,"script","https://mc.yandex.ru/metrika/tag.js","ym");
ym(${metrikaId},"init",{clickmap:true,trackLinks:true,accurateTrackBounce:true,webvisor:true});
</script>`
  : '';

const metrikaNoscript = metrikaId
  ? `<noscript><div><img src="https://mc.yandex.ru/watch/${metrikaId}" style="position:absolute;left:-9999px" alt="" /></div></noscript>`
  : '';

const publicConfig = JSON.stringify({
  siteUrl: globalThis.__SITE_URL__,
  metrikaId,
}).replace(/</g, '\\u003C');

type Render = (url: string) => Promise<RenderResult>;

const app = express();

// biome-ignore lint/suspicious/noExplicitAny: vite dev server has no static type here
let vite: any;

if (isProd) {
  app.use(compression());
  app.use('/assets', sirv('./dist/client/assets', { maxAge: 31536000, immutable: true }));
  app.use(sirv('./dist/client', { extensions: [], maxAge: 3600 }));
} else {
  const { createServer } = await import('vite');
  vite = await createServer({
    server: { middlewareMode: true },
    appType: 'custom',
  });
  app.use(vite.middlewares);
}

for (const path of ['/sitemap.xml', '/rss.xml', '/llms.txt', '/llms-full.txt']) {
  app.get(path, async (_req, res) => {
    try {
      const upstream = await fetch(`${globalThis.__API_BASE__}${path}`);
      const contentType = upstream.headers.get('content-type');

      res.status(upstream.status);
      if (contentType) res.set('Content-Type', contentType);
      res.set('Cache-Control', 'public, max-age=300');
      res.send(Buffer.from(await upstream.arrayBuffer()));
    } catch (error) {
      console.error(`proxy ${path} failed:`, error);
      res.status(502).end('Bad Gateway');
    }
  });
}

app.get('/favicon.ico', (_req, res) => {
  res.redirect(301, '/favicon/favicon.jpg');
});

app.get('/robots.txt', (_req, res) => {
  const lines = ['User-Agent: *', 'Disallow:'];

  if (globalThis.__SITE_URL__) {
    lines.push('', `Sitemap: ${globalThis.__SITE_URL__}/sitemap.xml`);
  }

  res.set('Cache-Control', 'public, max-age=3600');
  res.type('text/plain').send(`${lines.join('\n')}\n`);
});

app.get('/posts/:slug', async (req, res, next) => {
  const isTelegramBot = req.headers['user-agent'] === TELEGRAM_BOT_UA;
  const isPreview = req.query['tg-iv'] === '1';
  if (!isTelegramBot && !isPreview) return next();

  const html = await renderInstantView(apiBase, siteUrl, req.params.slug);
  if (!html) return next();

  res.status(200).set({ 'Content-Type': 'text/html' }).send(html);
});

app.use(async (req, res) => {
  try {
    const url = req.originalUrl;

    let template: string;
    let render: Render;

    if (isProd) {
      const entryServer = './dist/server/entry-server.js';
      template = await fs.readFile('./dist/client/index.html', 'utf-8');
      render = (await import(entryServer)).render;
    } else {
      template = await fs.readFile('./index.html', 'utf-8');
      template = await vite.transformIndexHtml(url, template);
      render = (await vite.ssrLoadModule('/src/entry-server.ts')).render;
    }

    const rendered = await render(url);

    if (rendered.redirect) {
      res.redirect(rendered.statusCode, rendered.redirect);
      return;
    }

    const html = template
      .replace('<!--app-analytics-->', metrikaTag)
      .replace('<!--app-analytics-noscript-->', metrikaNoscript)
      .replace('<!--app-head-->', rendered.headTags)
      .replace('<!--app-html-->', rendered.html)
      .replace(
        '<!--app-state-->',
        `<script>window.__PUBLIC_CONFIG__=${publicConfig};window.__INITIAL_STATE__=${rendered.state}</script>`,
      );

    res.status(rendered.statusCode).set({ 'Content-Type': 'text/html' }).send(html);
  } catch (error) {
    vite?.ssrFixStacktrace?.(error);
    console.error(error);
    res.status(500).end('Internal Server Error');
  }
});

app.listen(port, () => {
  console.log(`SSR server running at http://localhost:${port}`);
});
