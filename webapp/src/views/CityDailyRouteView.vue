<template>
  <DailyJourneyRouteView v-if="enabled && !checking && !probeError" :key="courseCode + String(route.query.journey_id || 'today')" :course-code="courseCode" :initial-journey="initialJourney" />
  <div v-else-if="probeError" role="alert" class="journey-load-error"><p>{{ probeError }}</p><button @click="probe">{{ t('common.retry') }}</button></div>
  <LgLoader v-else-if="checking" />
  <LegacyDailyRouteView v-else />
</template>
<script setup lang="ts">
import { computed, onScopeDispose, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import type { DailyJourney } from '../components/dailyJourney/types'
import LgLoader from '../components/linglow/LgLoader.vue'
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
const probeError = ref('')
let generation = 0
async function probe() {
  const request = ++generation
  const user = currentUserScope()
  enabled.value = false
  initialJourney.value = null
  probeError.value = ''
  if (!isAuthenticated.value || !courseCode.value) return
  checking.value = true
  try {
    if (!navigator.onLine) throw new Error(t('dailyJourney.messages.offline'))
    const state = await dailyJourneyClient.current(courseCode.value)
    if (request !== generation || user !== currentUserScope()) return
    if (state.enabled) {
      const id = typeof route.query.journey_id === 'string' ? route.query.journey_id : ''
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
<style scoped>
.journey-load-error { padding: 24px; border: 1px solid var(--border); border-radius: 16px; background: var(--card-bg); }.journey-load-error button { margin-top: 16px; }
</style>
