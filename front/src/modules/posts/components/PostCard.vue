<script setup lang="ts">
import { RouterLink } from 'vue-router';
import { formatPostDate } from '@/shared/lib/date';
import type { PostSnippet } from '../types';

const props = defineProps<{ post: PostSnippet }>();

const publishedAt = formatPostDate(props.post.createdAt);
</script>

<template>
  <RouterLink
    :to="`/posts/${post.url}`"
    class="ui-card flex min-h-60 flex-col transition-shadow hover:shadow-[0_0_20px_rgba(0,0,0,0.12)] md:h-60 md:max-h-120 md:flex-row"
  >
    <div v-if="post.cover" class="m-1 shrink-0 overflow-hidden rounded-3xl md:hidden">
      <img :src="post.cover.url" alt="" loading="lazy" class="h-40 w-full object-cover" />
    </div>

    <div
      class="flex min-w-0 flex-1 flex-col justify-between p-5"
      :class="post.cover ? 'md:w-[70%]' : 'w-full'"
    >
      <div>
        <h2 class="line-clamp-2 text-xl font-semibold text-gray-900">{{ post.title }}</h2>
        <p class="line-clamp-5 pt-2 text-sm leading-relaxed text-gray-600">{{ post.excerpt }}</p>
      </div>
      <time class="pt-3 text-sm text-gray-500">{{ publishedAt }}</time>
    </div>

    <div v-if="post.cover" class="m-1 hidden w-[30%] shrink-0 overflow-hidden rounded-3xl md:block">
      <img :src="post.cover.url" alt="" loading="lazy" class="h-full w-full object-cover" />
    </div>
  </RouterLink>
</template>
