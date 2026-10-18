<template>
  <section class="summary" aria-labelledby="journey-summary-title">
    <LgLumi :size="160" contained :pose="journey.status === 'completed' ? 'clapping' : 'quizzical'" />
    <h2 id="journey-summary-title" tabindex="-1">{{ journey.status === 'completed' ? t('dailyJourney.finished') : t('dailyJourney.unavailable') }}</h2>
    <p>{{ t('dailyJourney.completedStops', { done: completed, total: journey.steps.length }) }}</p>
    <ul><li v-for="step in journey.steps" :key="step.id"><strong>{{ journeyStepTitle(step, t) }}</strong><span>{{ journeyResultLabel(step, t) }}</span></li></ul>
  </section>
</template>
<script setup lang="ts">
import { journeyStepTitle } from './dailyJourney/labels'
import { useI18n } from 'vue-i18n'
import { computed } from 'vue'
import LgLumi from './linglow/LgLumi.vue'
import { journeyResultLabel } from './dailyJourney/resultLabel'
import type { DailyJourney } from './dailyJourney/types'
const { t } = useI18n()
const props = defineProps<{ journey: DailyJourney }>()
const completed = computed(() => props.journey.steps.filter(s => s.status === 'completed').length)
</script>
<style scoped>
.summary { display: flex; align-items: center; flex-direction: column; padding: 32px 20px; } h2 { margin-bottom: 0; } p { color: var(--text-secondary, #566773); }
ul { list-style: none; width: 100%; padding: 0; } li { display: flex; flex-wrap: wrap; justify-content: space-between; gap: 8px; border-top: 1px solid var(--border, #d4dce4); padding: 16px 0; } li span { color: var(--text-secondary, #566773); }
</style>
