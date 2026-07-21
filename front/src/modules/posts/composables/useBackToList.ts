import { onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';

const STORAGE_KEY = 'posts:list-path';

function isListPath(path: unknown): path is string {
  return typeof path === 'string' && (path === '/' || path.startsWith('/?'));
}

export function rememberListPath(path: string): void {
  if (!isListPath(path)) return;

  try {
    sessionStorage.setItem(STORAGE_KEY, path);
  } catch {
  }
}

function readListPath(): string | null {
  try {
    const stored = sessionStorage.getItem(STORAGE_KEY);

    return isListPath(stored) ? stored : null;
  } catch {
    return null;
  }
}

export function useBackToList() {
  const router = useRouter();

  const target = ref('/');
  const canGoBack = ref(false);

  onMounted(() => {
    const previous = (window.history.state as { back?: unknown } | null)?.back;

    if (isListPath(previous)) {
      target.value = previous;
      canGoBack.value = true;

      return;
    }

    const stored = readListPath();

    if (stored) target.value = stored;
  });

  function goBack(event: MouseEvent): void {
    if (event.defaultPrevented) return;
    if (event.button !== 0) return;
    if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;

    event.preventDefault();

    if (canGoBack.value) router.back();
    else router.push(target.value);
  }

  return { target, goBack };
}
