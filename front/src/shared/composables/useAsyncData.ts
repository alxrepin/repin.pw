import { inject } from 'vue';
import { STATE_KEY } from '@/shared/ssr/state';

export function useDataLoader(): <T>(key: string, loader: () => Promise<T>) => Promise<T> {
  const state = inject(STATE_KEY);
  if (!state) throw new Error('useDataLoader must be used within the app');

  return (key, loader) => state.resolve(key, loader);
}

export function useAsyncData<T>(key: string, loader: () => Promise<T>): Promise<T> {
  return useDataLoader()(key, loader);
}

export function useHttpStatus() {
  const state = inject(STATE_KEY);
  if (!state) throw new Error('useHttpStatus must be used within the app');

  return {
    set(code: number) {
      state.status.code = code;
    },

    redirect(location: string, code = 301) {
      state.status.code = code;
      state.status.location = location;
    },
  };
}
