<template>
  <div class="lg-lumi-fact">
    <LgLumi :size="lumiSize" />
    <div class="lg-lumi-fact-body">
      <div class="lg-lumi-fact-head">
        <span class="lg-lumi-fact-star">✦</span>
        <span class="lg-lumi-fact-title">{{ t('lg.lumiKnows') }}</span>
      </div>
      <p class="lg-lumi-fact-text">{{ fact }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import LgLumi from './LgLumi.vue'

const props = withDefaults(defineProps<{ lumiSize?: number }>(), { lumiSize: 52 })
void props

const { t, tm } = useI18n()

// Day-based rotating fact, list lives in locale files (lg.lumiFacts)
const fact = computed(() => {
  const facts = tm('lg.lumiFacts') as string[]
  if (!Array.isArray(facts) || facts.length === 0) return ''
  return facts[Math.floor(Date.now() / 86400000) % facts.length]
})
</script>

<style scoped>
.lg-lumi-fact {
  display: flex;
  gap: 12px;
  align-items: center;
  background: var(--card-bg);
  border: 1px solid var(--border);
  border-radius: 16px;
  padding: 14px 16px;
}
.lg-lumi-fact-body { flex: 1; }
.lg-lumi-fact-head {
  display: flex;
  gap: 4px;
  align-items: center;
  margin-bottom: 4px;
}
.lg-lumi-fact-star { font-size: 11px; color: var(--dorado); }
.lg-lumi-fact-title { font-size: 12px; font-weight: 600; color: var(--dorado); }
.lg-lumi-fact-text {
  margin: 0;
  font-size: 13px;
  line-height: 1.5;
  color: var(--text);
}
</style>
