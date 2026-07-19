<script setup lang="ts">
import { useRoute } from 'vue-router';
import { useChannel } from '@/modules/channel/composables/useChannel';
import type { Channel } from '@/modules/channel/types';
import { useHttpStatus } from '@/shared/composables/useAsyncData';
import { ErrorState } from '@/shared/ui';
import AppFooter from './components/AppFooter.vue';
import AppHeader from './components/AppHeader.vue';

const route = useRoute();
const httpStatus = useHttpStatus();

let channel: Channel | null = null;
try {
  channel = await useChannel();
} catch {
  httpStatus.set(503);
}
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <template v-if="channel">
      <AppHeader :channel="channel" />

      <main class="flex-1 pb-10">
        <RouterView v-slot="{ Component }">
          <Transition name="page" mode="out-in">
            <component :is="Component" :key="route.path" />
          </Transition>
        </RouterView>
      </main>

      <AppFooter :channel="channel" />
    </template>

    <main v-else class="flex flex-1 items-center justify-center px-4 py-20">
      <ErrorState
        title="Сайт временно недоступен"
        description="Не удалось загрузить данные канала. Попробуйте обновить страницу позже."
      />
    </main>
  </div>
</template>
