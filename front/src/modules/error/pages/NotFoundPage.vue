<script setup lang="ts">
import { useHead } from '@unhead/vue';
import { RouterLink } from 'vue-router';
import { useChannel } from '@/modules/channel/composables/useChannel';
import { useHttpStatus } from '@/shared/composables/useAsyncData';
import { ErrorState, UiContainer } from '@/shared/ui';

const channel = await useChannel();

useHttpStatus().set(404);

useHead({ title: `Страница не найдена — ${channel.title}` });
</script>

<template>
  <UiContainer size="sm" class="pt-28">
    <ErrorState class="py-20" title="404" description="Такой страницы здесь нет">
      <RouterLink
        to="/"
        class="ui-button bg-brand px-4 py-3 text-sm text-white hover:bg-brand-hover"
      >
        На главную
      </RouterLink>
    </ErrorState>
  </UiContainer>
</template>
