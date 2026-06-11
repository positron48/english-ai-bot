<template>
  <div class="lg-bubble"><slot>{{ text }}</slot></div>
</template>

<script setup lang="ts">
defineProps<{ text?: string }>()
</script>

<style scoped>
.lg-bubble {
  position: relative;
  background: var(--card-bg);
  border: 1.5px solid var(--border);
  border-radius: 14px;
  padding: 9px 13px;
  font-size: 13px;
  font-weight: 500;
  color: var(--text);
  line-height: 1.45;
  white-space: pre-line;
  max-width: 150px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
}
</style>
