<script setup lang="ts">
import { ref } from 'vue';
import { NavigationProgress } from '@/shared/ui';

const navigating = ref(false);
</script>

<template>
  <NavigationProgress :active="navigating" />
  <!--
    Fallback-слот намеренно не задан: приложение всегда гидрируется поверх
    SSR-разметки, а до первой отрисовки экран держит #app-loader из index.html.
    Переходы между роутами Suspense отрабатывает, не снимая текущую страницу, —
    индикатором служит NavigationProgress по событию @pending.
  -->
  <Suspense @pending="navigating = true" @resolve="navigating = false">
    <RouterView />
  </Suspense>
</template>
