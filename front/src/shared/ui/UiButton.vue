<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    tag?: 'button' | 'a';
    href?: string;
    variant?: 'primary' | 'ghost';
    size?: 'sm' | 'md';
  }>(),
  { tag: 'button', variant: 'primary', size: 'md' },
);

const variants: Record<string, string> = {
  primary: 'bg-brand text-white hover:bg-brand-hover',
  ghost: 'bg-transparent text-gray-800 hover:bg-gray-100',
};

const sizes: Record<string, string> = {
  sm: 'text-sm px-4 py-3',
  md: 'text-base px-4.5 py-3.5',
};
</script>

<template>
  <component
    :is="props.tag"
    :href="props.href"
    class="ui-button"
    :class="[variants[props.variant], sizes[props.size]]"
  >
    <slot />
  </component>
</template>
