const FALLBACK = 'http://localhost:8080';

export function apiBaseUrl(): string {
  if (import.meta.env.SSR) {
    return globalThis.__API_BASE__ ?? FALLBACK;
  }

  return import.meta.env.VITE_API_BASE_URL ?? FALLBACK;
}

/**
 * Public origin of the site, without a trailing slash — used for canonical
 * links. Unlike the API base it is a runtime value: the SSR server reads
 * PUBLIC_SITE_URL and hands it to the client via window.__PUBLIC_CONFIG__, so
 * changing the domain needs a container restart, not a rebuild. Empty when
 * unset, and callers must then omit the canonical link instead of emitting a
 * broken one.
 */
export function siteUrl(): string {
  if (import.meta.env.SSR) {
    return globalThis.__SITE_URL__ ?? '';
  }

  return window.__PUBLIC_CONFIG__?.siteUrl ?? '';
}
