<template>
  <article class="journey-card">
    <div class="hero-art" aria-hidden="true"><div class="art-orbit" /><LgLumi :size="210" contained :pose="journey?.status === 'completed' ? 'cup' : 'map'" /></div>
    <div class="copy">
      <span class="kicker">{{ t('dailyJourney.card.kicker') }}</span>
      <h2>{{ t('dailyJourney.card.title') }}</h2>
      <p>{{ description }}</p>
      <div v-if="journey" class="dots" aria-hidden="true"><span v-for="step in journey.steps" :key="step.id" :class="step.status" /></div>
      <button type="button" :disabled="busy" @click="$emit('open')">{{ label }} <span aria-hidden="true">→</span></button>
    </div>
  </article>
</template>
<script setup lang="ts">
import { computed } from 'vue'
import { journeyStepTitle } from './dailyJourney/labels'
import { useI18n } from 'vue-i18n'
import LgLumi from './linglow/LgLumi.vue'
import { type JourneyCardData } from './dailyJourney/types'
const { t } = useI18n()
const props = defineProps<{ journey: JourneyCardData | null; busy?: boolean }>()
defineEmits<{ open: [] }>()
const currentStep = computed(() => props.journey?.steps.find(step => step.status === 'pending' || step.status === 'active'))
const description = computed(() => {
  if (props.journey?.status === 'completed') return t('dailyJourney.card.completed')
  if (props.journey?.status === 'unavailable') return t('dailyJourney.card.unavailable')
  return currentStep.value ? journeyStepTitle(currentStep.value, t) : t('dailyJourney.card.description')
})
const label = computed(() => {
  const j = props.journey
  if (!j) return t('dailyJourney.card.start')
  if (j.status === 'completed') return t('dailyJourney.card.results')
  if (j.status === 'unavailable') return t('dailyJourney.card.view')
  const index = j.steps.findIndex(s => s.id === currentStep.value?.id)
  return t('dailyJourney.card.resume', { position: index + 1, total: j.steps.length })
})
</script>
<style scoped>
.journey-card { position: relative; isolation: isolate; display: flex; flex-direction: row-reverse; align-items: center; gap: 16px; padding: 32px 36px; border: 1px solid var(--border-green); border-radius: 24px; background: radial-gradient(ellipse at 90% 10%, color-mix(in srgb, var(--hoja) 15%, transparent), transparent 65%), var(--card-bg); box-shadow: var(--shadow-soft); }
.hero-art { position: relative; flex: 0 0 210px; display: grid; place-items: center; }.hero-art :deep(img) { position: relative; z-index: 1; }.art-orbit { position: absolute; width: 170px; height: 170px; border: 1px dashed color-mix(in srgb, var(--hoja) 35%, transparent); border-radius: 50%; background: color-mix(in srgb, var(--hoja) 7%, transparent); }
.copy { flex: 1; min-width: 0; }.kicker { font-size: .65rem; letter-spacing: .14em; font-weight: 700; color: var(--subtext); }h2 { margin: 10px 0 12px; font-size: clamp(1.6rem, 3.5vw, 2.25rem); line-height: 1.15; font-weight: 500; }p { margin-bottom: 20px; font-size: .88rem; line-height: 1.6; color: var(--subtext); }
button { display: inline-flex; align-items: center; gap: 18px; padding: 13px 18px; border: 1px solid var(--btn-border); border-radius: 12px; background: var(--btn-gradient); box-shadow: var(--btn-shadow); color: var(--crema); font: inherit; font-size: .85rem; font-weight: 600; text-align: left; cursor: pointer; }button span { font-size: 1.2rem; }button:disabled { opacity: .6; cursor: wait; }
.dots { display: flex; gap: 6px; margin-bottom: 20px; }.dots span { height: 4px; width: 24px; border-radius: 5px; background: var(--border); }.dots .completed { background: var(--hoja); }.dots .skipped { background: var(--subtext); }
@media (max-width: 540px) { .journey-card { padding: 24px; gap: 0; }.hero-art { flex-basis: 108px; margin-right: -10px; }.hero-art :deep(img) { width: 140px !important; }.art-orbit { width: 100px; height: 100px; }.copy { position: relative; z-index: 2; }h2 { font-size: 1.6rem; }.kicker { font-size: .55rem; }button { font-size: .75rem; padding: 12px; gap: 8px; } }
</style>
