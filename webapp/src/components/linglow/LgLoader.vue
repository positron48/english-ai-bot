<template>
  <div class="lg-loader" :class="{ 'lg-loader--overlay': overlay }">
    <span class="lg-loader-spinner" :style="{ width: size + 'px', height: size + 'px' }" />
    <span v-if="label" class="lg-loader-label">{{ label }}</span>
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{ label?: string; size?: number; overlay?: boolean }>(), {
  label: '',
  size: 28,
  overlay: false,
})
</script>

<style scoped>
.lg-loader {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 40px 20px;
  color: var(--subtext);
}
.lg-loader--overlay {
  position: absolute;
  inset: 0;
  z-index: 30;
  background: var(--bg);
}
.lg-loader-spinner {
  display: block;
  border-radius: 50%;
  border: 3px solid var(--border);
  border-top-color: var(--salvia);
  animation: lg-loader-spin 0.8s linear infinite;
}
.lg-loader-label {
  font-size: 13px;
}
@keyframes lg-loader-spin {
  to { transform: rotate(360deg); }
}
</style>
