import { type InjectionKey, reactive } from 'vue';

export interface StateManager {
  data: Record<string, unknown>;
  status: { code: number; location?: string };
  resolve<T>(key: string, loader: () => Promise<T>): Promise<T>;
  serialize(): string;
}

export const STATE_KEY: InjectionKey<StateManager> = Symbol('ssr-state');

export function createStateManager(initial: Record<string, unknown> = {}): StateManager {
  const data = reactive<Record<string, unknown>>({ ...initial });
  const pending = new Map<string, Promise<unknown>>();

  return {
    data,
    status: { code: 200 } as StateManager['status'],

    resolve<T>(key: string, loader: () => Promise<T>): Promise<T> {
      if (key in data) return Promise.resolve(data[key] as T);

      const existing = pending.get(key);
      if (existing) return existing as Promise<T>;

      const promise = loader().then(value => {
        data[key] = value;
        pending.delete(key);
        return value;
      });

      pending.set(key, promise);

      return promise;
    },

    serialize() {
      return JSON.stringify(data).replace(/</g, '\\u003C');
    },
  };
}
