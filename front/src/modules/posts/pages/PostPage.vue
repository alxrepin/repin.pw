<script setup lang="ts">
import { useHead } from '@unhead/vue';
import { RouterLink, useRoute } from 'vue-router';
import { useChannel } from '@/modules/channel/composables/useChannel';
import { useAsyncData, useHttpStatus } from '@/shared/composables/useAsyncData';
import { siteUrl } from '@/shared/config/env';
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

const description = post?.seoDescription ?? post?.text.replace(/<[^>]*>/g, '').slice(0, 160) ?? '';

const ogImage = post?.media.find(m => m.type === 'photo')?.url ?? channel.avatar ?? '';

const origin = siteUrl();
const canonicalLink = origin
  ? [{ rel: 'canonical', href: `${origin}/posts/${canonicalSlug}` }]
  : [];

useHead(
  post
    ? {
        title: `${post.title} — ${channel.title}`,
        meta: [
          { name: 'description', content: description },
          { property: 'og:type', content: 'article' },
          { property: 'og:title', content: post.title },
          { property: 'og:description', content: description },
          { property: 'og:image', content: ogImage },
          { name: 'twitter:card', content: 'summary_large_image' },
        ],
        link: canonicalLink,
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
        <time>{{ publishedAt }}</time>
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
