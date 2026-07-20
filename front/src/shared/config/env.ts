const FALLBACK = 'http://localhost:8080';

export function apiBaseUrl(): string {
  if (import.meta.env.SSR) {
    return globalThis.__API_BASE__ ?? FALLBACK;
  }

  return import.meta.env.VITE_API_BASE_URL ?? FALLBACK;
}

export function siteUrl(): string {
  if (import.meta.env.SSR) {
    return globalThis.__SITE_URL__ ?? '';
  }

  return window.__PUBLIC_CONFIG__?.siteUrl ?? '';
}

export function metrikaId(): string {
  if (import.meta.env.SSR) return '';

  return window.__PUBLIC_CONFIG__?.metrikaId ?? '';
}
