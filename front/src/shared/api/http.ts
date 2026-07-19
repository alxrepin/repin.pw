import { apiBaseUrl } from '@/shared/config/env';

export interface HttpError extends Error {
  status: number;
  payload?: unknown;
}

export type HttpQuery = Record<string, string | number | boolean | undefined>;

export interface HttpOptions {
  query?: HttpQuery;
}

function buildUrl(path: string, query?: HttpQuery): string {
  const url = new URL(`${apiBaseUrl()}${path}`);

  if (query) {
    for (const [key, value] of Object.entries(query)) {
      if (value !== undefined) url.searchParams.set(key, String(value));
    }
  }

  return url.toString();
}

async function toError(response: Response): Promise<HttpError> {
  const error = new Error(`HTTP ${response.status}`) as HttpError;
  error.status = response.status;
  error.payload = await response.json().catch(() => undefined);

  return error;
}

export async function http<T>(path: string, options: HttpOptions = {}): Promise<T> {
  const response = await fetch(buildUrl(path, options.query));

  if (!response.ok) throw await toError(response);
  if (response.status === 204) return undefined as T;

  return (await response.json()) as T;
}
