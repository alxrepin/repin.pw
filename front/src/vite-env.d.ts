/// <reference types="vite/client" />

declare global {
  var __API_BASE__: string | undefined;
  var __SITE_URL__: string | undefined;

  interface Window {
    __INITIAL_STATE__?: Record<string, unknown>;
    __PUBLIC_CONFIG__?: { siteUrl?: string };
  }
}

export {};
