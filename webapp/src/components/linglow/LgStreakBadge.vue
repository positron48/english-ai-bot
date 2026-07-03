<template>
  <div class="lg-streak">
    <span class="lg-streak-fire"><LgIcon name="flame" :s="16" /></span>
    <span class="lg-streak-n">{{ compact ? n : (t as any)('lg.streakDays', n, { n }) }}</span>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import LgIcon from './LgIcon.vue'

withDefaults(defineProps<{ n: number; compact?: boolean }>(), { compact: false })
const { t } = useI18n()
</script>

<style scoped>
.lg-streak {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 7px 14px 7px 10px;
  border-radius: 999px;
  background: var(--streak-bg);
  border: 1px solid var(--streak-border);
  font-weight: 700;
  color: var(--text);
  flex-shrink: 0;
}
.lg-streak-fire { font-size: 18px; line-height: 1; }
.lg-streak-n { font-size: 14px; }
</style>
