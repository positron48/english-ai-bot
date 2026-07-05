<template>
  <div class="lg-card-loading">
    <LgLumi pose="word-card" :size="size" />
    <div class="lg-card-loading-copy">
      <span class="lg-card-loading-name">Lumi</span>
      <p class="lg-card-loading-text">{{ message }}</p>
    </div>
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
.lg-card-loading-copy {
  display: flex;
  flex-direction: column;
  gap: 4px;
  align-items: center;
}
.lg-card-loading-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--text);
}
.lg-card-loading-text {
  margin: 0;
  font-size: 15px;
  line-height: 1.5;
  color: var(--subtext);
}
</style>
