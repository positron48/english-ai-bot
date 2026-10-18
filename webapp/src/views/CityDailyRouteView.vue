<template>
  <DailyJourneyRouteView v-if="enabled" :course-code="courseCode" />
  <div v-else-if="explicitJourney && probeError" role="alert" class="journey-load-error"><p>{{ probeError }}</p><button @click="probe">{{ t('common.retry') }}</button></div>
  <div v-else-if="checking && explicitJourney">{{ t('common.loading') }}</div>
  <LegacyDailyRouteView v-else />
</template>
<script setup lang="ts">
import { computed, onScopeDispose, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
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
const enabled = ref(false)
const checking = ref(false)
const probeError = ref('')
let generation = 0
async function probe() {
  const request = ++generation
  const user = currentUserScope()
  enabled.value = false
  probeError.value = ''
  if (!isAuthenticated.value || !courseCode.value) return
  checking.value = true
  try {
    if (!navigator.onLine) throw new Error(t('dailyJourney.messages.offline'))
    const state = await dailyJourneyClient.current(courseCode.value)
    if (request !== generation || user !== currentUserScope()) return
    enabled.value = state.enabled
    if (!state.enabled && explicitJourney.value) probeError.value = t('dailyJourney.messages.disabled')
  } catch (cause) {
    if (request === generation && user === currentUserScope()) probeError.value = cause instanceof Error ? cause.message : t('common.networkError')
  } finally {
    if (request === generation) checking.value = false
  }
}
watch([courseCode, isAuthenticated, explicitJourney], () => { void probe() }, { immediate: true })
onScopeDispose(() => { generation++ })
</script>
<style scoped>
.journey-load-error { padding: 24px; border: 1px solid var(--border); border-radius: 16px; background: var(--card-bg); }.journey-load-error button { margin-top: 16px; }
</style>
