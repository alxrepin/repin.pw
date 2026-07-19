<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue';
import type { PostMedia } from '../types';

const props = defineProps<{ items: PostMedia[]; index: number }>();
const emit = defineEmits<{ close: [] }>();

const current = ref(props.index);
watch(
  () => props.index,
  n => {
    current.value = n;
  },
);

function close(): void {
  emit('close');
}

function step(delta: number): void {
  const total = props.items.length;
  current.value = (current.value + delta + total) % total;
}

function onKey(event: KeyboardEvent): void {
  if (event.key === 'Escape') close();
  else if (event.key === 'ArrowLeft') step(-1);
  else if (event.key === 'ArrowRight') step(1);
}

onMounted(() => {
  document.addEventListener('keydown', onKey);
  document.body.style.overflow = 'hidden';
});

onBeforeUnmount(() => {
  document.removeEventListener('keydown', onKey);
  document.body.style.overflow = '';
});
</script>

<template>
  <div
    class="lightbox fixed inset-0 z-50 flex items-center justify-center bg-black/90 p-4"
    role="dialog"
    aria-modal="true"
    aria-label="Просмотр медиа"
    @click="close"
  >
    <button
      type="button"
      aria-label="Закрыть"
      class="absolute right-4 top-4 flex h-11 w-11 items-center justify-center rounded-full bg-white/10 text-white transition hover:bg-white/20"
      @click.stop="close"
    >
      <span class="i-mdi-close text-2xl" />
    </button>

    <button
      v-if="items.length > 1"
      type="button"
      aria-label="Предыдущее"
      class="absolute left-4 flex h-11 w-11 items-center justify-center rounded-full bg-white/10 text-white transition hover:bg-white/20"
      @click.stop="step(-1)"
    >
      <span class="i-mdi-chevron-left text-3xl" />
    </button>

    <video
      v-if="items[current].type === 'video'"
      :key="current"
      :src="items[current].url"
      controls
      autoplay
      playsinline
      class="max-h-[90vh] max-w-[90vw] rounded-2xl"
      @click.stop
    />

    <img
      v-else
      :key="current"
      :src="items[current].url"
      :width="items[current].width"
      :height="items[current].height"
      alt=""
      class="max-h-[90vh] max-w-[90vw] rounded-2xl object-contain"
      @click.stop
    />

    <button
      v-if="items.length > 1"
      type="button"
      aria-label="Следующее"
      class="absolute right-4 flex h-11 w-11 items-center justify-center rounded-full bg-white/10 text-white transition hover:bg-white/20"
      @click.stop="step(1)"
    >
      <span class="i-mdi-chevron-right text-3xl" />
    </button>

    <span
      v-if="items.length > 1"
      class="absolute bottom-4 left-1/2 -translate-x-1/2 rounded-full bg-white/10 px-3 py-1 text-sm text-white"
    >
      {{ current + 1 }} / {{ items.length }}
    </span>
  </div>
</template>

<style scoped>
.lightbox {
  animation: lightbox-in 0.15s ease-out;
}

@keyframes lightbox-in {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

@media (prefers-reduced-motion: reduce) {
  .lightbox {
    animation: none;
  }
}
</style>
