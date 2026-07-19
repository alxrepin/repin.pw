import fs from 'node:fs/promises';
import compression from 'compression';
import express from 'express';
import sirv from 'sirv';

interface RenderResult {
  html: string;
  headTags: string;
  state: string;
  statusCode: number;
}

const isProd = process.env.NODE_ENV === 'production';
const port = Number(process.env.PORT ?? 3000);

globalThis.__API_BASE__ = process.env.API_INTERNAL_URL ?? 'http://localhost:8080';

// Public origin for canonical links. Normalised once here so neither the pages
// nor the deployer have to care about a trailing slash in the env value.
globalThis.__SITE_URL__ = (process.env.PUBLIC_SITE_URL ?? '').trim().replace(/\/+$/, '');

// Same escaping as the SSR state: keeps a "</script>" inside a value from
// closing the inline script tag.
const publicConfig = JSON.stringify({ siteUrl: globalThis.__SITE_URL__ }).replace(/</g, '\\u003C');

type Render = (url: string) => Promise<RenderResult>;

const app = express();

// biome-ignore lint/suspicious/noExplicitAny: vite dev server has no static type here
let vite: any;

if (isProd) {
  app.use(compression());
  app.use(sirv('./dist/client', { extensions: [] }));
} else {
  const { createServer } = await import('vite');
  vite = await createServer({
    server: { middlewareMode: true },
    appType: 'custom',
  });
  app.use(vite.middlewares);
}

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

    const html = template
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
