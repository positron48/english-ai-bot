<template>
  <div class="lg-header">
    <div class="lg-header-row">
      <button v-if="showBack" class="lg-header-btn" type="button" @click="emit('back')">
        <LgIcon name="chevron-left" :s="20" c="var(--text)" />
      </button>
      <div class="lg-header-title-wrap">
        <div class="lg-header-title">{{ title }}</div>
      </div>
      <div class="lg-header-right">
        <slot name="right">
          <LgStreakBadge v-if="streak !== undefined" :n="streak" />
        </slot>
      </div>
    </div>
    <div v-if="subtitle" class="lg-header-subtitle">{{ subtitle }}</div>
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  title: string
  subtitle?: string
  showBack?: boolean
  streak?: number
}>(), { showBack: false })

const emit = defineEmits<{ (e: 'back'): void }>()

import LgIcon from './LgIcon.vue'
import LgStreakBadge from './LgStreakBadge.vue'
</script>

<style scoped>
.lg-header { padding: 16px 16px 8px; }
.lg-header-row {
  display: flex;
  align-items: center;
  position: relative;
  min-height: 40px;
}
.lg-header-btn {
  width: 38px;
  height: 38px;
  border-radius: 50%;
  border: 1px solid var(--border);
  background: var(--card-bg);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  flex-shrink: 0;
  position: absolute;
  left: 0;
  z-index: 1;
}
.lg-header-title-wrap { flex: 1; text-align: center; }
.lg-header-title {
  font-family: 'Lora', serif;
  font-weight: 700;
  font-size: 20px;
  color: var(--text);
  line-height: 1.2;
}
.lg-header-right { position: absolute; right: 0; z-index: 1; }
.lg-header-subtitle {
  text-align: center;
  font-size: 12px;
  color: var(--subtext);
  margin-top: 4px;
}
</style>
