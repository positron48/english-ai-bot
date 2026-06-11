<template>
  <div class="lg-progress-track" :style="{ height: h + 'px', borderRadius: h + 'px' }">
    <div
      class="lg-progress-fill"
      :style="{ width: clamped + '%', borderRadius: h + 'px', background: color || 'var(--salvia)' }"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  pct: number
  h?: number
  color?: string
}>(), { h: 4 })

const clamped = computed(() => Math.min(100, Math.max(0, props.pct)))
</script>

<style scoped>
.lg-progress-track {
  background: var(--progress-track);
  overflow: hidden;
}
.lg-progress-fill {
  height: 100%;
  transition: width .5s ease;
}
</style>
