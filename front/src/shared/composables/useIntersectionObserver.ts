import { onBeforeUnmount, onMounted, type Ref, ref, watch } from 'vue';

interface Options {
  enabled?: Ref<boolean>;
  threshold?: number;
}

export function useIntersectionObserver(selector: string, options: Options = {}) {
  const { enabled, threshold = 0 } = options;

  const isIntersecting = ref(true);

  let observer: IntersectionObserver | null = null;
  let retryFrame = 0;

  const stop = () => {
    observer?.disconnect();
    observer = null;
    cancelAnimationFrame(retryFrame);
  };

  const start = (retriesLeft = 30) => {
    stop();

    if (enabled && !enabled.value) {
      isIntersecting.value = true;
      return;
    }

    const target = document.querySelector(selector);
    if (!target) {
      if (retriesLeft > 0) {
        retryFrame = requestAnimationFrame(() => start(retriesLeft - 1));
      }

      return;
    }

    observer = new IntersectionObserver(
      ([entry]) => {
        isIntersecting.value = entry.isIntersecting;
      },
      { threshold },
    );

    observer.observe(target);
  };

  onMounted(() => start());
  onBeforeUnmount(stop);

  if (enabled) {
    watch(enabled, () => start());
  }

  return { isIntersecting };
}
