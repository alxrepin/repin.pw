<script setup lang="ts">
defineProps<{ active: boolean }>();
</script>

<template>
  <Transition name="np-fade">
    <div v-if="active" class="np" role="progressbar" aria-label="Загрузка страницы">
      <div class="np__bar" />
    </div>
  </Transition>
</template>

<style scoped>
.np {
  position: fixed;
  inset: 0 0 auto 0;
  z-index: 60;
  height: 3px;
  overflow: hidden;
  background: rgba(2, 23, 38, 0.08);
}

.np__bar {
  position: absolute;
  inset: 0;
  width: 40%;
  background: linear-gradient(90deg, transparent, #021726, transparent);
  animation: np-slide 1s ease-in-out infinite;
}

@keyframes np-slide {
  0% {
    transform: translateX(-120%);
  }
  100% {
    transform: translateX(360%);
  }
}

.np-fade-leave-active {
  transition: opacity 0.3s ease;
}
.np-fade-leave-to {
  opacity: 0;
}

@media (prefers-reduced-motion: reduce) {
  .np__bar {
    animation-duration: 2s;
  }
}
</style>
