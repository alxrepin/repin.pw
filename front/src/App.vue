<script setup lang="ts">
import { ref } from 'vue';
import { useRoute } from 'vue-router';
import { trackPageview } from '@/shared/lib/metrika';
import { NavigationProgress } from '@/shared/ui';

const navigating = ref(false);
const route = useRoute();

function onResolve(): void {
  navigating.value = false;
  trackPageview(route.fullPath);
}
</script>

<template>
  <NavigationProgress :active="navigating" />
  <Suspense @pending="navigating = true" @resolve="onResolve">
    <RouterView />
  </Suspense>
</template>
