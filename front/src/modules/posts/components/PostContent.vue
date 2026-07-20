<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import type { PostMedia } from '../types';
import PostMediaBlock from './PostMediaBlock.vue';

const router = useRouter();

function onBodyClick(event: MouseEvent): void {
  if (event.defaultPrevented || event.button !== 0) return;
  if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;

  const anchor = (event.target as HTMLElement | null)?.closest('a');
  if (!anchor || anchor.target) return;

  const href = anchor.getAttribute('href');
  // Same-origin only: "//host" is another site wearing a relative-looking coat.
  if (!href?.startsWith('/') || href.startsWith('//')) return;

  event.preventDefault();
  router.push(href);
}

const props = defineProps<{
  title: string;
  html: string;
  media?: PostMedia[];
  invertMedia?: boolean;
}>();

const MAX_LINES = { mobile: 1, desktop: 2 };

const titleEl = ref<HTMLHeadingElement | null>(null);
const shownTitle = ref(props.title);
const overflow = ref('');

onMounted(() => {
  const el = titleEl.value;
  if (!el) return;

  const maxLines = window.matchMedia('(min-width: 640px)').matches
    ? MAX_LINES.desktop
    : MAX_LINES.mobile;

  const probe = el.cloneNode() as HTMLElement;
  probe.style.position = 'absolute';
  probe.style.visibility = 'hidden';
  probe.style.display = 'block';
  probe.style.overflow = 'visible';
  probe.style.width = `${el.clientWidth}px`;
  el.parentElement?.appendChild(probe);

  const lines = (text: string): number => {
    probe.textContent = text;
    return Math.round(probe.scrollHeight / Number.parseFloat(getComputedStyle(probe).lineHeight));
  };

  try {
    if (lines(props.title) <= maxLines) return;

    const words = props.title.split(/\s+/);
    let lo = 1;
    let hi = words.length - 1;

    while (lo < hi) {
      const mid = Math.ceil((lo + hi) / 2);

      if (lines(`${words.slice(0, mid).join(' ')}…`) <= maxLines) {
        lo = mid;
      } else {
        hi = mid - 1;
      }
    }

    shownTitle.value = `${words.slice(0, lo).join(' ')}…`;
    overflow.value = `…${words.slice(lo).join(' ')}`;
  } finally {
    probe.remove();
  }
});
</script>

<template>
  <article class="text-base leading-normal text-gray-900">
    <h1
      ref="titleEl"
      class="mb-6 line-clamp-1 text-[28px] font-bold leading-[1.4] sm:line-clamp-2"
    >
      {{ shownTitle }}
    </h1>
    <PostMediaBlock v-if="media?.length && !invertMedia" :media="media" class="mb-8" />
    <div class="post-body prose max-w-none" @click="onBodyClick">
      <p v-if="overflow">{{ overflow }}</p>
      <div v-html="html" />
    </div>
    <PostMediaBlock v-if="media?.length && invertMedia" :media="media" class="mt-8" />
  </article>
</template>

<style scoped>
.post-body :deep(h3) {
  font-size: 24px;
  font-weight: 600;
  line-height: 1.4;
  padding-top: 12px;
  padding-bottom: 8px;
}
</style>
