<script setup lang="ts">
defineProps<{ title: string; description?: string }>();
</script>

<template>
  <div class="text-center text-gray-400">
    <p class="text-2xl font-semibold">{{ title }}</p>
    <p v-if="description" class="mt-2">{{ description }}</p>
    <div v-if="$slots.default" class="mt-6">
      <slot />
    </div>
  </div>
</template>
