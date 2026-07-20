<script setup lang="ts">
import { useHead } from '@unhead/vue';
import { RouterLink, useRoute } from 'vue-router';
import { useChannel } from '@/modules/channel/composables/useChannel';
import { useAsyncData, useHttpStatus } from '@/shared/composables/useAsyncData';
import { siteUrl } from '@/shared/config/env';
import { jsonLd } from '@/shared/config/seo';
import { formatPostDate } from '@/shared/lib/date';
import { ErrorState, UiContainer } from '@/shared/ui';
import { fetchPostRequest } from '../api/posts';
import PostCard from '../components/PostCard.vue';
import PostContent from '../components/PostContent.vue';

const route = useRoute();
const slug = String(route.params.slug);

const channel = await useChannel();
const post = await useAsyncData(`post:${slug}`, () => fetchPostRequest(slug));

if (!post) {
  useHttpStatus().set(404);
} else if (post.url !== slug && import.meta.env.SSR) {
  useHttpStatus().redirect(`/posts/${post.url}`);
}

const canonicalSlug = post?.url ?? slug;

const publishedAt = post ? formatPostDate(post.createdAt) : '';

const pageTitle = post?.seoTitle ?? post?.title ?? '';

const description = post?.seoDescription ?? post?.text.replace(/<[^>]*>/g, '').slice(0, 160) ?? '';

const cover = post?.media.find(m => m.type === 'photo');
const ogImage = cover?.url ?? channel.avatar ?? '';

const origin = siteUrl();
const canonicalUrl = origin ? `${origin}/posts/${canonicalSlug}` : '';
const canonicalLink = canonicalUrl ? [{ rel: 'canonical', href: canonicalUrl }] : [];

const schema = post
  ? jsonLd({
      '@context': 'https://schema.org',
      '@type': 'BlogPosting',
      headline: pageTitle,
      description,
      url: canonicalUrl || undefined,
      mainEntityOfPage: canonicalUrl || undefined,
      image: ogImage || undefined,
      datePublished: post.createdAt,
      dateModified: post.updatedAt ?? post.createdAt,
      inLanguage: 'ru',
      keywords: post.seoKeywords,
      author: { '@type': 'Person', name: channel.title, url: channel.url },
      publisher: { '@type': 'Person', name: channel.title, url: origin || channel.url },
    })
  : '';

useHead(
  post
    ? {
        title: `${pageTitle} — ${channel.title}`,
        meta: [
          { name: 'description', content: description },
          { name: 'keywords', content: post.seoKeywords ?? '' },
          { property: 'og:type', content: 'article' },
          { property: 'og:site_name', content: channel.title },
          { property: 'og:locale', content: 'ru_RU' },
          { property: 'og:title', content: pageTitle },
          { property: 'og:description', content: description },
          { property: 'og:url', content: canonicalUrl },
          { property: 'og:image', content: ogImage },
          { property: 'og:image:width', content: cover?.width ? String(cover.width) : '' },
          { property: 'og:image:height', content: cover?.height ? String(cover.height) : '' },
          { property: 'article:published_time', content: post.createdAt },
          { property: 'article:modified_time', content: post.updatedAt ?? post.createdAt },
          { name: 'twitter:card', content: 'summary_large_image' },
        ],
        link: canonicalLink,
        script: [{ type: 'application/ld+json', innerHTML: schema }],
      }
    : { title: `Пост не найден — ${channel.title}` },
);
</script>

<template>
  <UiContainer size="sm" class="pt-28">
    <template v-if="post">
      <div class="mb-3 flex items-center justify-between text-gray-500">
        <RouterLink
          to="/"
          class="inline-flex items-center gap-2 transition-colors hover:text-gray-900"
        >
          <span class="i-mdi-arrow-left" />
          Назад
        </RouterLink>
        <time :datetime="post.createdAt">{{ publishedAt }}</time>
      </div>

      <PostContent
        :title="post.title"
        :html="post.text"
        :media="post.media"
        :invert-media="post.invertMedia"
      />

      <section v-if="post.next" class="mt-12">
        <h2 class="mb-8 text-2xl font-bold">Следующий пост</h2>
        <PostCard :post="post.next" />
      </section>

      <section v-if="post.prev" class="mt-12">
        <h2 class="mb-8 text-2xl font-bold">Предыдущий пост</h2>
        <PostCard :post="post.prev" />
      </section>
    </template>

    <ErrorState v-else class="py-20" title="404" description="Пост не найден">
      <RouterLink
        to="/"
        class="ui-button bg-brand px-4 py-3 text-sm text-white hover:bg-brand-hover"
      >
        На главную
      </RouterLink>
    </ErrorState>
  </UiContainer>
</template>
