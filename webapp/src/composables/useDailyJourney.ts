import { computed, onScopeDispose, ref, shallowRef, watch, type Ref } from 'vue'
import { currentJourneyStep, journeySkipAvailable, type DailyJourney, type JourneyAction, type JourneyActionType, type JourneyTransport } from '../components/dailyJourney/types'

type PendingAction = { journeyId: string; stepId: string; request: JourneyAction }
type PendingStorage = Pick<Storage, 'getItem' | 'setItem' | 'removeItem'>

// scope must include account and course. Pending storage is optional for previews;
// real screens supply sessionStorage so a reload retains a lost response's key.
export function useDailyJourney(scope: Ref<string>, createTransport: (scope: string) => JourneyTransport, pendingStorage?: PendingStorage, onConfirmed?: (journey: DailyJourney) => void) {
  const journey = shallowRef<DailyJourney | null>(null)
  const busy = ref(false)
  const error = ref('')
  const pending = shallowRef<PendingAction | null>(null)
  let generation = 0
  let transport: JourneyTransport
  const pendingKey = () => `linglow:daily-journey-pending:v1:${scope.value}`

  function savePending(value: PendingAction | null) {
    pending.value = value
    try {
      if (value) pendingStorage?.setItem(pendingKey(), JSON.stringify(value))
      else pendingStorage?.removeItem(pendingKey())
    } catch { /* Keep in-memory retry usable when browser storage is unavailable. */ }
  }
  function restorePending(): PendingAction | null {
    try {
      const raw = pendingStorage?.getItem(pendingKey())
      if (!raw) return null
      const value = JSON.parse(raw) as PendingAction
      if (typeof value.journeyId !== 'string' || typeof value.stepId !== 'string' || typeof value.request?.client_action_id !== 'string' || !Number.isSafeInteger(value.request.expected_revision) || typeof value.request.action !== 'string' || !value.request.payload || typeof value.request.payload !== 'object') return null
      return value
    } catch { return null }
  }

  async function run(operation: () => Promise<DailyJourney | null>, clearPending = false, notify = false) {
    if (busy.value) return
    const requestGeneration = generation
    busy.value = true
    error.value = ''
    try {
      const result = await operation()
      if (generation !== requestGeneration) return
      journey.value = result
      if (clearPending) savePending(null)
      if (notify && result) onConfirmed?.(result)
    } catch (cause) {
      if (generation !== requestGeneration) return
      error.value = cause instanceof Error ? cause.message : 'Не удалось загрузить маршрут. Повторите попытку.'
      // A rejected answer was not committed; allow editing it. Network errors
      // and revision conflicts retain the request until retry/explicit reload.
      const status = typeof cause === 'object' && cause && 'status' in cause ? Number(cause.status) : 0
      if (status >= 400 && status < 500 && status !== 409) savePending(null)
    } finally {
      if (generation === requestGeneration) busy.value = false
    }
  }

  watch(scope, value => {
    generation++
    journey.value = null
    pending.value = restorePending()
    busy.value = false
    error.value = ''
    transport = createTransport(value)
    void run(() => transport.get())
  }, { immediate: true })
  onScopeDispose(() => { generation++ })

  const currentStep = computed(() => currentJourneyStep(journey.value))
  const canSkip = computed(() => !!currentStep.value && !currentStep.value.required && journeySkipAvailable(journey.value))

  async function sendPending() {
    const item = pending.value
    if (!item) return
    await run(() => transport.act(item.journeyId, item.stepId, item.request), true, true)
  }
  async function act(action: JourneyActionType, payload: Record<string, unknown> = {}) {
    const step = currentStep.value
    if (busy.value || pending.value || !journey.value || !step) return
    savePending({ journeyId: journey.value.id, stepId: step.id, request: {
      client_action_id: crypto.randomUUID(), expected_revision: journey.value.revision, action, payload,
    } })
    await sendPending()
  }

  return {
    journey, busy, error, currentStep, canSkip,
    hasPendingAction: computed(() => pending.value !== null),
    pendingAction: computed(() => pending.value?.request || null),
    start: () => pending.value ? sendPending() : run(() => transport.start(), false, true),
    reload: () => run(() => transport.get(), true),
    retry: () => pending.value ? sendPending() : run(() => transport.get()),
    act,
  }
}
