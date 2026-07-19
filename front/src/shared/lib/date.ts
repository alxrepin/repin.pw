const MONTH_MS = 30 * 24 * 60 * 60 * 1000;

const relative = new Intl.RelativeTimeFormat('ru-RU', { numeric: 'always' });
const absolute = new Intl.DateTimeFormat('ru-RU', {
  day: 'numeric',
  month: 'long',
  year: 'numeric',
});

export function formatPostDate(iso: string): string {
  const date = new Date(iso);
  const diffMs = Date.now() - date.getTime();

  if (diffMs < 0 || diffMs >= MONTH_MS) return absolute.format(date);

  const minutes = Math.floor(diffMs / 60_000);
  if (minutes < 1) return 'только что';
  if (minutes < 60) return relative.format(-minutes, 'minute');

  const hours = Math.floor(minutes / 60);
  if (hours < 24) return relative.format(-hours, 'hour');

  const days = Math.floor(hours / 24);
  if (days < 7) return relative.format(-days, 'day');

  return relative.format(-Math.floor(days / 7), 'week');
}
