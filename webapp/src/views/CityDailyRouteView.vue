<template>
  <DailyJourneyRouteView v-if="enabled && !checking && !probeError" :key="courseCode + String(route.query.journey_id || 'today')" :course-code="courseCode" :initial-journey="initialJourney" />
  <DailyJourneyLoading v-else-if="checking || probeError" :key="loadingAttempt" :phase="loadingPhase" :error="probeError" @retry="probe" />
  <LegacyDailyRouteView v-else />
</template>
<script setup lang="ts">
import { computed, onScopeDispose, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import type { DailyJourney } from '../components/dailyJourney/types'
import DailyJourneyLoading from '../components/DailyJourneyLoading.vue'
import { dailyJourneyClient } from '../api/dailyJourneyClient'
import { currentUserScope } from '../api/sessionScope'
import { useCourse } from '../composables/useCourse'
import { useAuth } from '../composables/useAuth'
import DailyJourneyRouteView from './DailyJourneyRouteView.vue'
import LegacyDailyRouteView from './LegacyDailyRouteView.vue'
const route = useRoute()
const { t } = useI18n()
const { currentCourseCode } = useCourse()
const { isAuthenticated } = useAuth()
const courseCode = computed(() => typeof route.query.course_code === 'string' ? route.query.course_code : currentCourseCode.value)
const explicitJourney = computed(() => typeof route.query.journey_id === 'string' && !!route.query.journey_id)
const initialJourney = ref<DailyJourney | null>(null)
const enabled = ref(false)
const checking = ref(false)
const loadingPhase = ref<'checking' | 'building' | 'restoring'>('checking')
const loadingAttempt = ref(0)
const probeError = ref('')
let generation = 0
async function probe() {
  const request = ++generation
  const user = currentUserScope()
  enabled.value = false
  initialJourney.value = null
  probeError.value = ''
  checking.value = false
  loadingPhase.value = 'checking'
  loadingAttempt.value++
  if (!isAuthenticated.value || !courseCode.value) return
  checking.value = true
  try {
    if (!navigator.onLine) throw new Error(t('dailyJourney.messages.offline'))
    const state = await dailyJourneyClient.current(courseCode.value)
    if (request !== generation || user !== currentUserScope()) return
    if (state.enabled) {
      const id = typeof route.query.journey_id === 'string' ? route.query.journey_id : ''
      loadingPhase.value = id && state.journey?.id !== id ? 'restoring' : state.journey ? 'restoring' : 'building'
      const journey = id && state.journey?.id !== id
        ? await dailyJourneyClient.get(courseCode.value, id)
        : state.journey || await dailyJourneyClient.start(courseCode.value)
      if (request !== generation || user !== currentUserScope()) return
      initialJourney.value = journey
    }
    enabled.value = state.enabled
    if (!state.enabled && explicitJourney.value) probeError.value = t('dailyJourney.messages.disabled')
  } catch (cause) {
    if (request === generation && user === currentUserScope()) probeError.value = cause instanceof Error ? cause.message : t('common.networkError')
  } finally {
    if (request === generation) checking.value = false
  }
}
watch([courseCode, isAuthenticated, () => route.query.journey_id], () => { void probe() }, { immediate: true })
onScopeDispose(() => { generation++ })
</script>
