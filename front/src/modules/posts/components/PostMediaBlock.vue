<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import type { PostMedia } from '../types';
import MediaLightbox from './MediaLightbox.vue';

const props = defineProps<{ media: PostMedia[]; alt?: string }>();

const mounted = ref(false);
onMounted(() => {
  mounted.value = true;
});

const galleryTypes = new Set<PostMedia['type']>(['photo', 'video', 'gif']);

const gallery = computed(() => props.media.filter(m => galleryTypes.has(m.type)));
const rest = computed(() => props.media.filter(m => !galleryTypes.has(m.type)));

const isGrid = computed(() => gallery.value.length > 1);

function cellClass(index: number): string {
  const spansRow = gallery.value.length % 2 === 1 && index === gallery.value.length - 1;
  return spansRow ? 'col-span-2 aspect-[2/1]' : 'aspect-square';
}

function wrapClass(index: number): string {
  return isGrid.value ? cellClass(index) : 'w-full';
}

const fitClass = computed(() => (isGrid.value ? 'h-full w-full object-cover' : 'w-full'));

const viewable = computed(() =>
  gallery.value.filter(m => m.type === 'photo' || m.type === 'video'),
);

function imageAlt(index: number): string {
  if (!props.alt) return '';

  return gallery.value.length > 1 ? `${props.alt} — фото ${index + 1}` : props.alt;
}
const lightboxIndex = ref<number | null>(null);

function openLightbox(item: PostMedia): void {
  const i = viewable.value.findIndex(m => m.id === item.id);
  if (i >= 0) lightboxIndex.value = i;
}

function formatSize(bytes?: number): string {
  if (!bytes) return '';
  if (bytes < 1024 * 1024) return `${Math.round(bytes / 1024)} КБ`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} МБ`;
}
</script>

<template>
  <div class="space-y-4">
    <div v-if="gallery.length" :class="isGrid ? 'grid grid-cols-2 gap-2' : ''">
      <template v-for="(item, index) in gallery" :key="item.id">
        <button
          v-if="item.type === 'video'"
          type="button"
          aria-label="Смотреть видео"
          class="group relative block overflow-hidden rounded-3xl bg-black"
          :class="wrapClass(index)"
          @click="openLightbox(item)"
        >
          <video
            :src="item.url"
            muted
            playsinline
            preload="metadata"
            class="ui-zoom rounded-3xl"
            :class="fitClass"
          />
          <span class="pointer-events-none absolute inset-0 grid place-items-center">
            <span
              class="grid h-14 w-14 place-items-center rounded-full bg-black/60 text-white shadow-[0_2px_12px_rgba(0,0,0,0.55)] ring-2 ring-white/85 backdrop-blur-sm transition group-hover:scale-105 group-hover:bg-black/75"
            >
              <span class="i-mdi-play translate-x-px text-3xl drop-shadow-[0_1px_2px_rgba(0,0,0,0.6)]" />
            </span>
          </span>
        </button>

        <div
          v-else-if="item.type === 'gif'"
          class="group overflow-hidden rounded-3xl"
          :class="wrapClass(index)"
        >
          <video
            :src="item.url"
            autoplay
            muted
            loop
            playsinline
            class="ui-zoom rounded-3xl"
            :class="fitClass"
          />
        </div>

        <button
          v-else
          type="button"
          aria-label="Открыть изображение"
          class="group block cursor-zoom-in overflow-hidden rounded-3xl border border-black/5"
          :class="wrapClass(index)"
          @click="openLightbox(item)"
        >
          <img
            :src="item.url"
            :width="item.width"
            :height="item.height"
            :alt="imageAlt(index)"
            class="ui-zoom rounded-3xl"
            :class="fitClass"
          />
        </button>
      </template>
    </div>

    <template v-for="item in rest" :key="item.id">
      <video
        v-if="item.type === 'video_note'"
        :src="item.url"
        controls
        playsinline
        class="mx-auto aspect-square w-64 rounded-full object-cover"
      />

      <div v-else-if="item.type === 'voice'" class="flex items-center gap-3">
        <span class="i-mdi-microphone shrink-0 text-xl text-gray-400" />
        <audio :src="item.url" controls class="w-full" />
      </div>

      <div v-else-if="item.type === 'audio'">
        <audio :src="item.url" controls class="w-full" />
        <p v-if="item.fileName" class="mt-1 text-sm text-gray-400">{{ item.fileName }}</p>
      </div>

      <a
        v-else
        :href="item.url"
        download
        class="flex items-center gap-3 rounded-3xl border border-gray-100 p-4 transition-shadow hover:shadow-md"
      >
        <span class="i-mdi-file-document-outline shrink-0 text-2xl text-gray-400" />
        <span class="min-w-0">
          <span class="block truncate font-medium">{{ item.fileName ?? 'Файл' }}</span>
          <span v-if="item.size" class="block text-sm text-gray-400">{{ formatSize(item.size) }}</span>
        </span>
      </a>
    </template>

    <Teleport v-if="mounted" to="body">
      <MediaLightbox
        v-if="lightboxIndex !== null"
        :items="viewable"
        :index="lightboxIndex"
        @close="lightboxIndex = null"
      />
    </Teleport>
  </div>
</template>
