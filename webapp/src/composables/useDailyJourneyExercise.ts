import { journeyMessage } from '../components/dailyJourney/messages'
import { computed, nextTick, onScopeDispose, ref, shallowRef, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { currentUserScope } from '../api/sessionScope'
import { emitAppDataEvent } from '../api/cacheInvalidation'
import { dailyJourneyClient } from '../api/dailyJourneyClient'
import { useAuth } from './useAuth'
import { useDailyJourney } from './useDailyJourney'
import { dailyJourneyExerciseRoute, dailyJourneyReturnRoute, parseDailyJourneyContext, verifyDailyJourneyContext } from './useDailyJourneyContext'
import type { JourneyActionType, JourneyMode } from '../components/dailyJourney/types'

// Existing exercise pages choose this source BEFORE any normal session calls.
// Invalid context and network failures never fall back to a standalone session.
export function useDailyJourneyExercise<T>(mode: JourneyMode) {
  const route = useRoute()
  const { isAuthenticated } = useAuth()
  const router = useRouter()
  const contextError = ref('')
  const context = computed(() => {
    try { return parseDailyJourneyContext(route.query) }
    catch (cause) { return null }
  })
  const isJourney = computed(() => route.query.journey_id !== undefined || route.query.journey_step_id !== undefined)
  const scope = computed(() => `${isAuthenticated.value ? currentUserScope() : 'anon'}:${context.value?.courseCode}:${context.value?.journeyId}:${context.value?.stepId}`)
  const data = shallowRef<T | null>(null)
  const loading = ref(false)
  let generation = 0
  let loadPromise: Promise<void> = Promise.resolve()
  const controller = useDailyJourney(scope, () => {
    const ctx = context.value
    return {
      get: async () => {
        if (!isJourney.value) return null
        if (!ctx) throw new Error(journeyMessage('invalidContext'))
        return dailyJourneyClient.get(ctx.courseCode, ctx.journeyId)
      },
      start: async () => { throw new Error('Запускайте задание с экрана маршрута.') },
      act: (id, step, action) => {
        if (!ctx) throw new Error(journeyMessage('invalidContext'))
        return dailyJourneyClient.act(ctx.courseCode, id, step, action)
      },
    }
  }, sessionStorage, journey => emitAppDataEvent('daily-journey-progressed', journey.courseCode))
  async function load() {
    const request = ++generation
    contextError.value = ''
    const ctx = context.value
    const journey = controller.journey.value
    if (!isJourney.value || !ctx || !journey) return
    loading.value = true
    try {
      const ownStep = journey.steps.find(step => step.id === ctx.stepId)
      if (journey.id === ctx.journeyId && journey.courseCode === ctx.courseCode && ownStep && ['completed', 'unavailable', 'skipped'].includes(ownStep.status)) return
      const current = controller.currentStep.value
      if (mode === 'words' && current?.mode === 'words' && current.id === ctx.stepId && journey.id === ctx.journeyId && journey.courseCode === ctx.courseCode && current.launch && ['Training', 'WordSets', 'WordSetDetail', 'WordSetStudy'].includes(String(route.name))) {
        if ((current.launch.routeName === 'Training' && route.name !== 'Training') || (current.launch.routeName === 'WordSets' && route.name === 'Training')) {
          await router.replace(dailyJourneyExerciseRoute(journey, current)); return
        }
      }

      const step = verifyDailyJourneyContext(ctx, journey, String(route.name || ''), route.params)
      if (step.mode !== mode) throw new Error('Неверный режим задания.')
      const view = await dailyJourneyClient.exercise<T>(ctx.courseCode, ctx.journeyId, ctx.stepId)
      if (request !== generation) return
      verifyDailyJourneyContext(ctx, view.journey, String(route.name || ''), route.params)
      if (view.mode !== mode || view.step_id !== ctx.stepId) throw new Error('Неверный режим задания.')
      if (view.journey.revision !== journey.revision) { controller.journey.value = view.journey; return }
      data.value = view.data
    } catch (cause) {
      if (request === generation) contextError.value = cause instanceof Error ? cause.message : 'Не удалось загрузить задание.'
    } finally { if (request === generation) loading.value = false }
  }
  watch(scope, () => { data.value = null; loading.value = false; generation++ })
  watch([scope, () => controller.journey.value?.revision], () => { loadPromise = load() })
  onScopeDispose(() => { generation++ })
  async function act(action: JourneyActionType, payload: Record<string, unknown> = {}) {
    contextError.value = ''
    await controller.act(action, payload)
    await waitForLoad()
  }
  async function back() {
    if (context.value) await router.push(dailyJourneyReturnRoute(context.value))
    else await router.push({ name: 'CityDailyRoute', query: typeof route.query.course_code === 'string' ? { course_code: route.query.course_code } : {} })
  }
  // A response can advance the journey revision and schedule another exercise
  // read. Wait for the latest read, not just the promise captured before it.
  async function waitForLoad() {
    let pendingLoad: Promise<void>
    do {
      await nextTick()
      if (controller.busy.value || loading.value) await new Promise<void>(resolve => {
        const stop = watch([controller.busy, loading], ([busy, pending]) => { if (!busy && !pending) { stop(); resolve() } })
      })
      pendingLoad = loadPromise
      await pendingLoad
      await nextTick()
    } while (controller.busy.value || loading.value || pendingLoad !== loadPromise)
  }
  async function retry() {
    contextError.value = ''
    const previousLoad = loadPromise
    await controller.retry()
    await nextTick()
    // A revision change already scheduled a read; otherwise retry the failed
    // exercise read explicitly. Never send both reads at the same time.
    if (loadPromise === previousLoad) loadPromise = load()
    await waitForLoad()
  }
  async function ready(): Promise<T> {
    await waitForLoad()
    if (controller.error.value || contextError.value || !data.value) throw new Error(controller.error.value || contextError.value || 'Не удалось загрузить задание маршрута.')
    return data.value
  }
  const finished = computed(() => !!context.value && !!controller.journey.value && controller.journey.value.steps.some(step => step.id === context.value!.stepId && ['completed', 'unavailable', 'skipped'].includes(step.status)))
  return {
    isJourney, context, data, finished, act, back, ready,
    draftKey: (identity: string) => {
      const ctx = context.value
      return isJourney.value && isAuthenticated.value && ctx ? `daily-journey:draft:${currentUserScope()}:${ctx.courseCode}:${ctx.journeyId}:${ctx.stepId}:${identity}` : ''
    },
    busy: computed(() => controller.busy.value || loading.value),
    error: computed(() => controller.error.value || contextError.value || (isJourney.value && !context.value ? journeyMessage('invalidContext') : controller.hasPendingAction.value && !controller.busy.value && !loading.value ? journeyMessage('pending') : '')),
    retry,
    hasPendingAction: controller.hasPendingAction,
  }
}
