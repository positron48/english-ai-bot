<template>
  <div class="lg-ring" :style="{ width: size + 'px', height: size + 'px' }">
    <svg :width="size" :height="size" style="transform: rotate(-90deg)">
      <circle :cx="size / 2" :cy="size / 2" :r="r" stroke="var(--progress-track)" :stroke-width="stroke" fill="none" />
      <circle
        :cx="size / 2" :cy="size / 2" :r="r"
        stroke="var(--salvia)" :stroke-width="stroke" fill="none"
        :stroke-dasharray="circ" :stroke-dashoffset="offset"
        stroke-linecap="round"
      />
    </svg>
    <div class="lg-ring-content"><slot /></div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  val: number
  max: number
  size?: number
  stroke?: number
}>(), { size: 72, stroke: 6 })

const r = computed(() => (props.size - props.stroke) / 2)
const circ = computed(() => 2 * Math.PI * r.value)
const offset = computed(() => circ.value * (1 - Math.min(1, props.max > 0 ? props.val / props.max : 0)))
</script>

<style scoped>
.lg-ring { position: relative; flex-shrink: 0; }
.lg-ring-content {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
}
</style>
