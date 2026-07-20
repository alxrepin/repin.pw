<script setup lang="ts">
import { useHead } from '@unhead/vue';
import { computed, ref, watch, watchEffect } from 'vue';
import { useRoute } from 'vue-router';
import { useChannel } from '@/modules/channel/composables/useChannel';
import { fetchPostsRequest } from '@/modules/posts/api/posts';
import PostCard from '@/modules/posts/components/PostCard.vue';
import PostCardSkeleton from '@/modules/posts/components/PostCardSkeleton.vue';
import PostsPagination from '@/modules/posts/components/PostsPagination.vue';
import type { PostList } from '@/modules/posts/types';
import { useDataLoader, useHttpStatus } from '@/shared/composables/useAsyncData';
import { siteUrl } from '@/shared/config/env';
import { channelDescription, jsonLd } from '@/shared/config/seo';
import { UiButton, UiContainer } from '@/shared/ui';

const POSTS_PER_PAGE = 6;

const route = useRoute();
const page = computed(() => Math.max(1, Number.parseInt(String(route.query.page ?? '1'), 10) || 1));

const channel = await useChannel();

const loadData = useDataLoader();
const loadPosts = (n: number) =>
  loadData(`posts:${n}:${POSTS_PER_PAGE}`, () => fetchPostsRequest(n, POSTS_PER_PAGE));

const posts = ref<PostList>(await loadPosts(page.value));

const isPaginating = ref(false);

watch(page, async current => {
  isPaginating.value = true;
  try {
    const next = await loadPosts(current);

    if (page.value === current) posts.value = next;
  } finally {
    if (page.value === current) isPaginating.value = false;
  }
});

const totalPages = computed(() => Math.max(1, Math.ceil(posts.value.total / POSTS_PER_PAGE)));

const httpStatus = useHttpStatus();
watchEffect(() => {
  if (page.value > totalPages.value) httpStatus.set(404);
});

const subscribers = channel.subscriptions.toLocaleString('ru-RU');
const description = channelDescription(channel);

const title = computed(() =>
  page.value > 1 ? `${channel.title} — страница ${page.value}` : channel.title,
);

const origin = siteUrl();

const canonicalUrl = computed(() => {
  if (!origin) return '';

  return page.value > 1 ? `${origin}/?page=${page.value}` : `${origin}/`;
});

const canonicalLink = computed(() =>
  canonicalUrl.value ? [{ rel: 'canonical', href: canonicalUrl.value }] : [],
);

const schema = jsonLd({
  '@context': 'https://schema.org',
  '@type': 'Blog',
  name: channel.title,
  description,
  url: origin ? `${origin}/` : undefined,
  inLanguage: 'ru',
  author: { '@type': 'Person', name: channel.title, url: channel.url },
  sameAs: [channel.url],
});

useHead({
  title,
  meta: [
    { name: 'description', content: description },
    { property: 'og:type', content: 'website' },
    { property: 'og:site_name', content: channel.title },
    { property: 'og:locale', content: 'ru_RU' },
    { property: 'og:title', content: title },
    { property: 'og:description', content: description },
    { property: 'og:url', content: canonicalUrl },
    { property: 'og:image', content: channel.avatar ?? '' },
    { name: 'twitter:card', content: 'summary_large_image' },
  ],
  link: canonicalLink,
  script: [{ type: 'application/ld+json', innerHTML: schema }],
});
</script>

<template>
  <UiContainer class="pt-10">
    <section id="hero" class="mb-22 flex flex-col items-center pt-12 text-center">
      <img
        v-if="channel.avatar"
        :src="channel.avatar"
        :alt="channel.title"
        class="mb-5 h-50 w-50 rounded-full object-cover"
      />
      <h1 class="pb-2.5 text-[28px] font-bold">{{ channel.title }}</h1>
      <p v-if="description" class="w-full pb-5 leading-relaxed text-gray-500 sm:w-2/3">
        {{ description }}
      </p>
      <div class="mb-5 inline-flex items-center gap-1.5 rounded-xl bg-gray-100 px-3 py-1 text-sm text-gray-700">
        <span class="i-mdi-account-multiple" />
        {{ subscribers }} подписчиков
      </div>
      <UiButton tag="a" :href="channel.url" size="md">
        <span class="i-mdi-telegram" />
        Подписаться
      </UiButton>
    </section>

    <section>
      <h2 id="posts" class="mb-8 text-2xl font-bold">
        Посты <span class="text-gray-400">{{ posts.total }}</span>
      </h2>
      <div id="posts-grid" class="grid grid-cols-1 gap-5 lg:grid-cols-2">
        <template v-if="isPaginating">
          <PostCardSkeleton v-for="n in POSTS_PER_PAGE" :key="`skeleton-${n}`" />
        </template>
        <PostCard v-for="post in posts.items" v-else :key="post.id" :post="post" />
      </div>

      <p v-if="!isPaginating && !posts.items.length" class="py-16 text-center text-gray-400">
        Здесь пока пусто
      </p>

      <PostsPagination :page="page" :total-pages="totalPages" />
    </section>
  </UiContainer>
</template>
