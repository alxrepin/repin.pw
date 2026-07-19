<script setup lang="ts">
import { computed } from 'vue';
import { useRouter } from 'vue-router';

const props = defineProps<{ page: number; totalPages: number }>();

const router = useRouter();

function href(n: number): string {
  return n === 1 ? '/' : `/?page=${n}`;
}

function navigate(n: number, event: MouseEvent) {
  if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey || event.button !== 0)
    return;

  event.preventDefault();
  router.push(n === 1 ? { path: '/' } : { path: '/', query: { page: n } });
}

const items = computed<(number | null)[]>(() => {
  const { page, totalPages } = props;

  if (totalPages <= 7) {
    return Array.from({ length: totalPages }, (_, i) => i + 1);
  }

  const pages = new Set<number>([1, totalPages]);
  for (let n = page - 1; n <= page + 1; n++) {
    if (n >= 1 && n <= totalPages) pages.add(n);
  }

  const sorted = [...pages].sort((a, b) => a - b);
  const result: (number | null)[] = [];

  for (const [i, n] of sorted.entries()) {
    if (i > 0 && n - (sorted[i - 1] as number) > 1) result.push(null);
    result.push(n);
  }

  return result;
});
</script>

<template>
  <nav
    v-if="totalPages > 1"
    aria-label="Страницы постов"
    class="mt-10 flex flex-wrap items-center justify-center gap-2"
  >
    <a
      v-if="page > 1"
      :href="href(page - 1)"
      aria-label="Предыдущая страница"
      class="ui-button bg-gray-100 px-3.5 py-2.5 text-sm text-gray-700 hover:bg-gray-200"
      @click="navigate(page - 1, $event)"
    >
      <span class="i-mdi-arrow-left" />
    </a>

    <template v-for="(item, i) in items" :key="i">
      <span v-if="item === null" class="px-1 text-gray-400">…</span>
      <a
        v-else
        :href="href(item)"
        :aria-current="item === page ? 'page' : undefined"
        class="ui-button px-3.5 py-2.5 text-sm"
        :class="item === page ? 'bg-brand text-white' : 'bg-gray-100 text-gray-700 hover:bg-gray-200'"
        @click="navigate(item, $event)"
      >
        {{ item }}
      </a>
    </template>

    <a
      v-if="page < totalPages"
      :href="href(page + 1)"
      aria-label="Следующая страница"
      class="ui-button bg-gray-100 px-3.5 py-2.5 text-sm text-gray-700 hover:bg-gray-200"
      @click="navigate(page + 1, $event)"
    >
      <span class="i-mdi-arrow-right" />
    </a>
  </nav>
</template>
