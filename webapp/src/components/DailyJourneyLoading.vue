<template>
  <section class="journey-loading" :aria-busy="!error">
    <div class="journey-loading-art" aria-hidden="true"><LgLumi :size="210" contained :pose="error ? 'quizzical' : 'map'" /></div>
    <p class="eyebrow">{{ t('dailyJourney.title') }}</p>
    <h1>{{ t(error ? 'dailyJourney.loading.errorTitle' : 'dailyJourney.loading.title') }}</h1>
    <template v-if="!error">
      <p class="phase" role="status" aria-live="polite">{{ t('dailyJourney.loading.' + phase) }}</p>
      <ol class="loading-stages" :aria-label="t('dailyJourney.loading.stages')">
        <li :class="{ done: phase !== 'checking', current: phase === 'checking' }" :aria-current="phase === 'checking' ? 'step' : undefined">
          <span class="stage-dot" aria-hidden="true"><LgIcon v-if="phase !== 'checking'" name="check" :s="14" /><span v-else>1</span></span>
          {{ t('dailyJourney.loading.checkStep') }}
        </li>
        <li :class="{ current: phase !== 'checking' }" :aria-current="phase !== 'checking' ? 'step' : undefined">
          <span class="stage-dot" aria-hidden="true">2</span>{{ t(phase === 'restoring' ? 'dailyJourney.loading.restoreStep' : 'dailyJourney.loading.buildStep') }}
        </li>
      </ol>
      <p class="hint" aria-live="polite">{{ t(slow ? 'dailyJourney.loading.slow' : 'dailyJourney.loading.hint') }}</p>
    </template>
    <DailyJourneyError v-else :message="error" :show-back="false" @retry="$emit('retry')" />
  </section>
</template>
<script setup lang="ts">
import { onScopeDispose, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import LgLumi from './linglow/LgLumi.vue'
import LgIcon from './linglow/LgIcon.vue'
import DailyJourneyError from './DailyJourneyError.vue'
defineProps<{ phase: 'checking' | 'building' | 'restoring'; error?: string }>()
defineEmits<{ retry: [] }>()
const { t } = useI18n()
const slow = ref(false)
const timer = setTimeout(() => { slow.value = true }, 8000)
onScopeDispose(() => clearTimeout(timer))
</script>
<style scoped>
.journey-loading { width: min(100%, 640px); box-sizing: border-box; margin: 24px auto; padding: 32px clamp(20px, 5vw, 48px); border: 1px solid var(--border); border-radius: 28px; background: var(--card-bg); box-shadow: var(--shadow-soft); display: flex; flex-direction: column; align-items: center; text-align: center; color: var(--text); }
.journey-loading-art { display: grid; place-items: center; width: 230px; height: 200px; background: radial-gradient(ellipse, color-mix(in srgb, var(--salvia) 14%, transparent), transparent 70%); }
p { margin: 0; }.eyebrow { color: var(--subtext); font-size: .7rem; letter-spacing: .12em; text-transform: uppercase; margin-top: 16px; }
h1 { margin: 10px 0 12px; font-size: clamp(1.5rem, 5vw, 2rem); font-weight: 500; line-height: 1.25; }
.phase { min-height: 2.8em; line-height: 1.4; font-size: .95rem; }
.loading-stages { list-style: none; display: flex; gap: 24px; padding: 20px 0; margin: 10px 0 0; width: 100%; justify-content: center; border-top: 1px solid var(--border); }
li { display: flex; gap: 8px; align-items: center; font-size: .75rem; color: var(--subtext); text-align: left; flex: 1; max-width: 200px; }.current, .done { color: var(--text); }
.stage-dot { display: grid; place-items: center; flex-shrink: 0; width: 28px; height: 28px; border: 1px solid var(--border); border-radius: 50%; font-size: .7rem; }
.current .stage-dot { border-color: var(--salvia); background: color-mix(in srgb, var(--salvia) 18%, transparent); animation: breathe 2s ease-in-out infinite; }.done .stage-dot { background: var(--salvia); color: var(--crema); border-color: transparent; }
.hint { min-height: 3em; max-width: 380px; font-size: .8rem; line-height: 1.6; color: var(--subtext); }
.journey-error { width: 100%; box-sizing: border-box; margin: 16px 0 0; text-align: left; }
@keyframes breathe { 50% { box-shadow: 0 0 0 5px color-mix(in srgb, var(--salvia) 12%, transparent); } }
@media (max-width: 540px) { .journey-loading { margin: 12px auto; padding-top: 20px; }.journey-loading-art { height: 160px; }.journey-loading-art :deep(img) { width: 180px !important; }.loading-stages { gap: 12px; } }
@media (prefers-reduced-motion: reduce) { .current .stage-dot { animation: none; } }
</style>
