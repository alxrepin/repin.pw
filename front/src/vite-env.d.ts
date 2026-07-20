/// <reference types="vite/client" />

declare global {
  var __API_BASE__: string | undefined;
  var __SITE_URL__: string | undefined;

  interface YandexMetrika {
    (...args: unknown[]): void;
    a?: unknown[][];
    l?: number;
  }

  interface Window {
    __INITIAL_STATE__?: Record<string, unknown>;
    __PUBLIC_CONFIG__?: { siteUrl?: string; metrikaId?: string };
    ym?: YandexMetrika;
  }
}

export {};
