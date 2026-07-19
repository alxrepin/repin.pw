<script setup lang="ts">
const props = withDefaults(defineProps<{ as?: string; size?: 'sm' | 'md' | 'lg' }>(), {
  as: 'section',
  size: 'lg',
});

const maxWidth: Record<string, string> = {
  sm: 'max-w-[834px]',
  md: 'max-w-[1024px]',
  lg: 'max-w-[1280px]',
};
</script>

<template>
  <component :is="props.as" class="ui-container" :class="maxWidth[props.size]">
    <slot />
  </component>
</template>
