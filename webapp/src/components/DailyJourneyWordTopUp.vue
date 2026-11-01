<template>
  <section class="journey-top-up" :aria-label="t('dailyJourney.topUp.title')" :aria-busy="busy">
    <div class="top-up-copy">
      <p class="eyebrow">{{ t('dailyJourney.title') }}</p>
      <h2>{{ t('dailyJourney.topUp.title') }}</h2>
      <p class="goal">{{ t('dailyJourney.topUp.remaining', { count: Math.max(0, total - done) }) }}</p>
    </div>
    <div class="top-up-art" aria-hidden="true"><LgLumi pose="word-card" :size="96" contained /></div>
    <div class="top-up-progress">
      <div class="progress-caption" aria-live="polite"><span>{{ t('dailyJourney.addWords', { done, total }) }}</span><span aria-hidden="true">{{ done }} / {{ total }}</span></div>
      <progress :value="done" :max="Math.max(1, total)" :aria-label="t('dailyJourney.addWords', { done, total })" />
    </div>
    <button type="button" class="top-up-back" @click="$emit('back')"><span aria-hidden="true">←</span>{{ t('dailyJourney.returnToRoute') }}</button>
  </section>
</template>
<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import LgLumi from './linglow/LgLumi.vue'
defineProps<{ done: number; total: number; busy?: boolean }>()
defineEmits<{ back: [] }>()
const { t } = useI18n()
</script>
<style scoped>
.journey-top-up { display: grid; grid-template-columns: minmax(0, 1fr) 96px; align-items: center; gap: 16px; margin: 16px 0 24px; padding: 22px 24px 12px; border: 1px solid var(--border); border-radius: 22px; background: var(--card-bg); color: var(--text); box-shadow: var(--shadow-soft); }
.eyebrow { margin: 0 0 8px; font-size: .65rem; letter-spacing: .1em; text-transform: uppercase; color: var(--subtext); }h2 { margin: 0; font-size: 1.25rem; font-weight: 500; line-height: 1.3; }.goal { margin: 8px 0 0; font-size: .85rem; line-height: 1.5; color: var(--subtext); }
.top-up-progress { grid-column: 1 / -1; }.progress-caption { display: flex; align-items: baseline; justify-content: space-between; gap: 16px; font-size: .75rem; color: var(--subtext); line-height: 1.4; }.progress-caption span:last-child { white-space: nowrap; font-variant-numeric: tabular-nums; color: var(--text); }
progress { display: block; width: 100%; height: 6px; margin-top: 10px; border: 0; border-radius: 6px; overflow: hidden; appearance: none; background: var(--border); accent-color: var(--salvia); }progress::-webkit-progress-bar { background: var(--border); }progress::-webkit-progress-value { background: var(--salvia); border-radius: 6px; }progress::-moz-progress-bar { background: var(--salvia); border-radius: 6px; }
.top-up-back { grid-column: 1 / -1; justify-self: start; display: inline-flex; gap: 8px; align-items: center; min-height: 44px; padding: 8px 0; border: 0; background: transparent; color: var(--subtext); font: inherit; font-size: .8rem; cursor: pointer; text-align: left; }.top-up-back:hover { color: var(--text); }.top-up-back:focus-visible { outline: 2px solid var(--salvia); outline-offset: 4px; border-radius: 4px; }
@media (max-width: 540px) { .journey-top-up { margin-top: 0; padding: 18px 18px 8px; grid-template-columns: minmax(0, 1fr) 64px; gap: 12px; }.top-up-art :deep(img) { width: 72px !important; }h2 { font-size: 1.1rem; } }
</style>
