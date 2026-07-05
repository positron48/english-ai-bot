<template>
  <div class="lg-card-loading">
    <LgLumi pose="word-card" :size="size" />
    <p class="lg-card-loading-text">{{ message }}</p>
  </div>
</template>

<script setup lang="ts">
import LgLumi from './LgLumi.vue'

withDefaults(defineProps<{
  message: string
  size?: number
}>(), {
  size: 80,
})
</script>

<style scoped>
.lg-card-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 40px 24px;
  text-align: center;
}
.lg-card-loading-text {
  margin: 0;
  font-size: 15px;
  line-height: 1.5;
  color: var(--subtext);
}
</style>
