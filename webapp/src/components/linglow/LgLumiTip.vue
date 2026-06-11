<template>
  <div class="lg-lumi-tip">
    <LgLumi :size="52" :pose="pose" />
    <div>
      <div class="lg-lumi-tip-head">
        <span class="lg-lumi-tip-title">{{ title || t('lg.lumiTip') }}</span>
        <span>✨</span>
      </div>
      <p class="lg-lumi-tip-text"><slot>{{ text }}</slot></p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import LgLumi from './LgLumi.vue'

withDefaults(defineProps<{
  text?: string
  title?: string
  pose?: 'default' | 'book' | 'pencil' | 'map'
}>(), { pose: 'default' })
const { t } = useI18n()
</script>

<style scoped>
.lg-lumi-tip {
  display: flex;
  gap: 12px;
  align-items: center;
  background: var(--card-bg);
  border: 1px solid var(--border);
  border-radius: 16px;
  padding: 14px 16px;
}
.lg-lumi-tip-head {
  display: flex;
  gap: 5px;
  align-items: center;
  margin-bottom: 3px;
}
.lg-lumi-tip-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text);
}
.lg-lumi-tip-text {
  margin: 0;
  font-size: 13px;
  line-height: 1.5;
  color: var(--subtext);
}
</style>
