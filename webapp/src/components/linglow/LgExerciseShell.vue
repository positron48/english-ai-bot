<template>
  <div class="lg-ex-shell">
    <div class="lg-ex-header">
      <button class="lg-ex-close" type="button" @click="emit('close')">
        <LgIcon name="x" :s="16" c="var(--subtext)" />
      </button>
      <div class="lg-ex-progress">
        <div class="lg-ex-segments">
          <div
            v-for="i in segmentCount"
            :key="i"
            class="lg-ex-segment"
            :class="{
              'lg-ex-segment--done': i - 1 < doneSegments,
              'lg-ex-segment--current': i - 1 === doneSegments,
            }"
          />
        </div>
        <div class="lg-ex-counter">{{ t('lg.exerciseCounter', { current, total }) }}</div>
      </div>
      <div class="lg-ex-right"><slot name="right" /></div>
    </div>
    <div class="lg-ex-body">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import LgIcon from './LgIcon.vue'

const props = defineProps<{
  current: number
  total: number
}>()

const emit = defineEmits<{ (e: 'close'): void }>()
const { t } = useI18n()

const segmentCount = computed(() => Math.max(1, Math.min(props.total, 12)))
const doneSegments = computed(() => {
  if (props.total <= 12) return Math.max(0, props.current - 1)
  return Math.floor(((props.current - 1) / props.total) * segmentCount.value)
})
</script>

<style scoped>
.lg-ex-shell {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  min-height: 100dvh;
}
.lg-ex-header {
  padding: 14px 16px 10px;
  display: flex;
  align-items: center;
  gap: 10px;
}
.lg-ex-close {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: 1.5px solid var(--border);
  background: var(--card-bg);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  flex-shrink: 0;
}
.lg-ex-progress { flex: 1; }
.lg-ex-segments {
  display: flex;
  gap: 4px;
  align-items: center;
  margin-bottom: 5px;
}
.lg-ex-segment {
  flex: 1;
  height: 5px;
  border-radius: 3px;
  background: var(--progress-track);
  transition: background .3s;
}
.lg-ex-segment--done { background: var(--salvia); }
.lg-ex-segment--current { background: var(--hoja); }
.lg-ex-counter {
  font-size: 11px;
  color: var(--subtext);
  text-align: center;
}
.lg-ex-right { flex-shrink: 0; }
.lg-ex-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}
</style>
