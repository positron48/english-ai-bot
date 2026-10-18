<template>
  <main class="daily-journey-view">
    <header><RouterLink :to="{ name: 'City', query: { course_code: courseCode } }">← {{ t('city.backToCity') }}</RouterLink><h1>{{ t('dailyJourney.title') }}</h1></header>
    <DailyJourneyShell :journey="journey" :busy="busy" :error="navigationError || error" :has-pending-action="hasPendingAction" @start="start" @retry="retry" @reload="reload" @action="act" @open-step="openStep" />
  </main>
</template>
<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuth } from '../composables/useAuth'
import { currentUserScope } from '../api/sessionScope'
import { emitAppDataEvent } from '../api/cacheInvalidation'
import { createDailyJourneyTransport, dailyJourneyClient } from '../api/dailyJourneyClient'
import { useDailyJourney } from '../composables/useDailyJourney'
import { dailyJourneyExerciseRoute } from '../composables/useDailyJourneyContext'
import DailyJourneyShell from '../components/DailyJourneyShell.vue'
import type { JourneyStep } from '../components/dailyJourney/types'
const props = defineProps<{ courseCode: string }>()
const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const { isAuthenticated } = useAuth()
const navigationError = ref('')
const scope = computed(() => `${isAuthenticated.value ? currentUserScope() : 'anon'}:${props.courseCode}:${typeof route.query.journey_id === 'string' ? route.query.journey_id : 'today'}`)
const { journey, busy, error, currentStep, hasPendingAction, start, retry, reload, act } = useDailyJourney(scope, () => {
  const course = props.courseCode
  const id = typeof route.query.journey_id === 'string' ? route.query.journey_id : ''
  const transport = createDailyJourneyTransport(course)
  return { ...transport, get: () => id ? dailyJourneyClient.get(course, id) : transport.get() }
}, sessionStorage, journey => emitAppDataEvent('daily-journey-progressed', journey.courseCode))
async function openStep(step: JourneyStep) {
  navigationError.value = ''
  const requestScope = scope.value
  // Resuming also revalidates access and retrieves the bound subject session.
  await act('start')
  if (requestScope !== scope.value || error.value || hasPendingAction.value || !journey.value || currentStep.value?.id !== step.id) return
  try { await router.push(dailyJourneyExerciseRoute(journey.value, currentStep.value)) }
  catch (cause) { navigationError.value = cause instanceof Error ? cause.message : t('common.error') }
}
</script>
<style scoped>
.daily-journey-view { max-width: 900px; margin: 0 auto; display: grid; gap: 24px; padding-bottom: 32px; }header a { color: var(--subtext); text-decoration: none; font-size: .8rem; }h1 { margin-top: 12px; font-size: 2.2rem; font-weight: 500; }
</style>
