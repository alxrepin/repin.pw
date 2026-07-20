import { metrikaId } from '@/shared/config/env';

const previous: { path: string | null } = {
  path: import.meta.env.SSR ? null : location.pathname + location.search + location.hash,
};

export function trackPageview(fullPath: string): void {
  if (import.meta.env.SSR) return;

  const from = previous.path;
  previous.path = fullPath;

  if (from === null || from === fullPath) return;

  const id = Number(metrikaId());
  if (!id) return;

  setTimeout(() => {
    window.ym?.(id, 'hit', fullPath, { referer: from });
  }, 0);
}
