<script setup lang="ts">
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import type { Channel } from '@/modules/channel/types';

defineProps<{ channel: Channel }>();

const route = useRoute();
const isHome = computed(() => route.name === 'home');

const year = new Date().getFullYear();
</script>

<template>
  <footer>
    <div
      class="ui-container flex flex-col items-center justify-between gap-2 py-6 text-sm text-gray-500 sm:flex-row"
      :class="isHome ? 'max-w-[1280px]' : 'max-w-[834px]'"
    >
      <span>© {{ year }} {{ channel.title }}</span>
      <a :href="channel.url" target="_blank" rel="noopener noreferrer" class="transition-colors hover:text-gray-900">
        @{{ channel.name }}
      </a>
    </div>
  </footer>
</template>
