import { describe, it, expect, vi } from 'vitest'
import { effectScope, nextTick, ref } from 'vue'
import { useDailyJourney } from './useDailyJourney'
import { createFixtureTransport, createJourneyFixture } from '../dev/dailyJourneyFixture'
import type { DailyJourney, JourneyTransport } from '../components/dailyJourney/types'
const flush = async () => { await nextTick(); await Promise.resolve(); await Promise.resolve() }
function memoryStorage() {
  const data = new Map<string, string>()
  return { getItem: (key: string) => data.get(key) ?? null, setItem: (key: string, value: string) => { data.set(key, value) }, removeItem: (key: string) => { data.delete(key) } }
}
describe('dailyJourney controller', () => {
  it('restores confirmed feedback and cursor without starting another journey', async () => {
    const storage = memoryStorage()
    const scope = effectScope()
    const factory = () => createFixtureTransport('user:en', 'en_ru', 'full', storage)
    const controller = scope.run(() => useDailyJourney(ref('user:en'), factory))!
    await flush()
    expect(controller.journey.value).toBeNull()
    await controller.start()
    await controller.act('start')
    await controller.act('answer', { text: 'книга' })
    const saved = controller.journey.value
    scope.stop()
    const restoredScope = effectScope()
    const restored = restoredScope.run(() => useDailyJourney(ref('user:en'), factory))!
    await flush()
    expect(restored.journey.value).toEqual(saved)
    expect(restored.currentStep.value?.feedback).toBeTruthy()
    await restored.start()
    expect(restored.journey.value).toEqual(saved)
    restoredScope.stop()
  })

  it('retries a committed action with the same key after losing its response', async () => {
    const base = createFixtureTransport('retry', 'en_ru', 'short', memoryStorage())
    let fail = true
    const keys: string[] = []
    const scope = effectScope()
    const controller = scope.run(() => useDailyJourney(ref('retry'), () => ({ ...base, async act(...args) {
      keys.push(args[2].client_action_id)
      const result = await base.act(...args)
      if (fail) { fail = false; throw new Error('Connection lost') }
      return result
    } })))!
    await flush()
    await controller.start()
    await controller.act('start')
    expect(controller.hasPendingAction.value).toBe(true)
    await controller.retry()
    expect(keys[0]).toBe(keys[1])
    expect(controller.journey.value?.revision).toBe(1)
    expect(controller.hasPendingAction.value).toBe(false)
    scope.stop()
  })

  it('restores a pending action after reload and does not send a second answer', async () => {
    const storage = memoryStorage()
    const base = createFixtureTransport('reload', 'en_ru', 'short', storage)
    let loseResponse = true
    const keys: string[] = []
    const factory = () => ({ ...base, async act(...args: Parameters<JourneyTransport['act']>) {
      keys.push(args[2].client_action_id)
      const result = await base.act(...args)
      if (loseResponse) { loseResponse = false; throw new Error('Lost response') }
      return result
    } })
    const firstScope = effectScope()
    const first = firstScope.run(() => useDailyJourney(ref('user:1:en_ru'), factory, storage))!
    await flush()
    await first.start()
    await first.act('start')
    firstScope.stop()
    const secondScope = effectScope()
    const second = secondScope.run(() => useDailyJourney(ref('user:1:en_ru'), factory, storage))!
    await flush()
    expect(second.hasPendingAction.value).toBe(true)
    await second.act('answer', { text: 'must not send' })
    expect(keys).toHaveLength(1)
    await second.retry()
    expect(keys[0]).toBe(keys[1])
    expect(second.journey.value?.revision).toBe(1)
    expect(second.hasPendingAction.value).toBe(false)
    secondScope.stop()
  })

  it('ignores stale results and errors when changing user or course', async () => {
    let resolve!: (value: DailyJourney) => void
    const oldRequest = new Promise<DailyJourney>(done => { resolve = done })
    const scopeValue = ref('old')
    const scope = effectScope()
    const next = createJourneyFixture('es_ru', 'short')
    const transport: JourneyTransport = { get: async () => next, start: async () => next, act: async () => next }
    const controller = scope.run(() => useDailyJourney(scopeValue, value => value === 'old' ? { ...transport, get: () => oldRequest } : transport))!
    scopeValue.value = 'new'
    await flush()
    resolve(createJourneyFixture('en_ru', 'full'))
    await flush()
    expect(controller.journey.value?.courseCode).toBe('es_ru')
    expect(controller.busy.value).toBe(false)
    expect(controller.error.value).toBe('')
    scope.stop()
  })
  it('retains the same AI action through processing, provider failure and reload', async () => {
    const storage = memoryStorage()
    const journey = createJourneyFixture('en_ru', 'short')
    const sent: Array<{ client_action_id: string; payload: Record<string, unknown> }> = []
    let attempt = 0
    const transport: JourneyTransport = {
      get: async () => journey, start: async () => journey,
      act: async (_id, _step, request) => {
        sent.push({ client_action_id: request.client_action_id, payload: request.payload })
        attempt++
        if (attempt < 3) throw Object.assign(new Error(attempt === 1 ? 'AI processing' : 'AI provider failed'), { status: attempt === 1 ? 503 : 502 })
        return { ...journey, revision: journey.revision + 1 }
      },
    }
    const firstScope = effectScope()
    const first = firstScope.run(() => useDailyJourney(ref('user:ai:en_ru'), () => transport, storage))!
    await flush()
    await first.act('message', { text: 'My original message' })
    expect(first.hasPendingAction.value).toBe(true)
    firstScope.stop()
    const secondScope = effectScope()
    const second = secondScope.run(() => useDailyJourney(ref('user:ai:en_ru'), () => transport, storage))!
    await flush()
    await second.retry()
    expect(second.hasPendingAction.value).toBe(true)
    await second.act('message', { text: 'Must not become another provider request' })
    expect(sent).toHaveLength(2)
    await second.retry()
    expect(new Set(sent.map(item => item.client_action_id)).size).toBe(1)
    expect(sent.every(item => item.payload.text === 'My original message')).toBe(true)
    expect(second.hasPendingAction.value).toBe(false)
    expect(second.journey.value?.revision).toBe(journey.revision + 1)
    secondScope.stop()
  })

})

it('notifies overview only after confirmed mutations in the current scope', async () => {
 const confirmed = vi.fn()
 const scopeValue = ref('old')
 const lifetime = effectScope()
 const journey = createJourneyFixture('en_ru', 'short')
 let resolve!: (value: DailyJourney) => void
 const late = new Promise<DailyJourney>(done => { resolve = done })
 const transport: JourneyTransport = { get: async () => journey, start: async () => journey, act: async () => late }
 const controller = lifetime.run(() => useDailyJourney(scopeValue, () => transport, undefined, confirmed))!
 await flush()
 expect(confirmed).not.toHaveBeenCalled()
 await controller.start()
 expect(confirmed).toHaveBeenCalledOnce()
 confirmed.mockClear()
 const action = controller.act('start')
 scopeValue.value = 'new'
 await flush()
 resolve(journey)
 await action
 expect(confirmed).not.toHaveBeenCalled()
 await controller.reload()
 expect(confirmed).not.toHaveBeenCalled()
 lifetime.stop()
})
