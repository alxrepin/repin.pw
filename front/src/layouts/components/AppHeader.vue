<script setup lang="ts">
import { computed } from 'vue';
import { RouterLink, useRoute } from 'vue-router';
import type { Channel } from '@/modules/channel/types';
import { useIntersectionObserver } from '@/shared/composables/useIntersectionObserver';

defineProps<{ channel: Channel }>();

const route = useRoute();
const isHome = computed(() => route.name === 'home');

const { isIntersecting: heroVisible } = useIntersectionObserver('#hero', { enabled: isHome });

const visible = computed(() => !isHome.value || !heroVisible.value);
</script>

<template>
  <Transition name="header">
    <header v-if="visible" class="fixed inset-x-0 top-0 z-10 backdrop-blur-[5px]">
      <div
        class="ui-container flex items-center justify-between py-4"
        :class="isHome ? 'max-w-[1280px]' : 'max-w-[834px]'"
      >
        <RouterLink to="/" class="flex items-center gap-2.5">
          <img
            v-if="channel.avatar"
            :src="channel.avatar"
            :alt="channel.title"
            class="h-10 w-10 rounded-full object-cover"
          />
          <span class="text-xl font-bold">{{ channel.title }}</span>
        </RouterLink>

        <a
          :href="channel.url"
          target="_blank"
          rel="noopener noreferrer"
          class="ui-button bg-brand px-4 py-3 text-sm text-white hover:bg-brand-hover"
        >
          <span class="i-mdi-telegram" />
          Подписаться
        </a>
      </div>
    </header>
  </Transition>
</template>
