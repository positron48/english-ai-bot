import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises } from '@vue/test-utils'
import { effectScope, reactive, ref } from 'vue'
import { useDailyJourneyExercise } from './useDailyJourneyExercise'
import { createJourneyFixture } from '../dev/dailyJourneyFixture'
const { client, push, replace } = vi.hoisted(() => ({ client: { get: vi.fn(), act: vi.fn(), exercise: vi.fn() }, push: vi.fn(), replace: vi.fn() }))
const route = reactive({ name: 'ReadingText', params: { textId: 'read-a' } as Record<string, string>, query: {} as Record<string, string> })
const authenticated = ref(true)
vi.mock('vue-router', () => ({ useRoute: () => route, useRouter: () => ({ push, replace }) }))
vi.mock('./useAuth', () => ({ useAuth: () => ({ isAuthenticated: authenticated }) }))
vi.mock('../api/sessionScope', () => ({ currentUserScope: () => 'user:1' }))
vi.mock('../api/dailyJourneyClient', () => ({ dailyJourneyClient: client }))
function readingJourney(course = 'en_ru') {
 const j = createJourneyFixture(course, 'short')
 j.steps[0].status = 'completed'; j.steps[1].status = 'active'
 j.steps[1].launch = { routeName: 'ReadingText', params: { textId: 'read-a' } }
 return j
}
beforeEach(() => {
 localStorage.setItem('locale','ru'); vi.clearAllMocks(); sessionStorage.clear(); authenticated.value = true
 route.name = 'ReadingText'; route.params = { textId: 'read-a' }
})
describe('dailyJourney confirmed exercise state', () => {
 it('waits for the server exercise after an action and finishes without treating completion as a context error', async () => {
  const j = readingJourney()
  route.query = { course_code: j.courseCode, journey_id: j.id, journey_step_id: j.steps[1].id }
  client.get.mockResolvedValue(j)
  client.exercise.mockImplementation(async () => ({ journey: j, step_id: j.steps[1].id, mode: 'reading', data: { confirmedRevision: j.revision } }))
  client.act.mockImplementation(async () => { j.revision++; return { ...j } })
  const scope = effectScope(); const daily = scope.run(() => useDailyJourneyExercise<{ confirmedRevision: number }>('reading'))!
  await flushPromises()
  expect(await daily.ready()).toEqual({ confirmedRevision: 0 })
  await daily.act('answer', { index: 0, answer: 'false' })
  expect(daily.data.value?.confirmedRevision).toBe(1)
  client.act.mockImplementationOnce(async () => { j.revision++; j.steps[1].status = 'completed'; j.status = 'completed'; return { ...j } })
  await daily.act('ack_feedback')
  expect(daily.finished.value).toBe(true)
  expect(daily.error.value).toBe('')
  await daily.back()
  expect(push).toHaveBeenCalledWith({ name: 'CityDailyRoute', query: { course_code: j.courseCode, journey_id: j.id } })
  scope.stop()
 })
 it('rejects a partial context before any exercise request', async () => {
  route.query = { journey_id: 'missing-step', course_code: 'en_ru' }
  const scope = effectScope(); const daily = scope.run(() => useDailyJourneyExercise('reading'))!
  await flushPromises()
  expect(daily.isJourney.value).toBe(true)
  expect(daily.error.value).toContain('Некорректный контекст')
  expect(client.exercise).not.toHaveBeenCalled()
  expect(client.get).not.toHaveBeenCalled()
  scope.stop()
 })
 it('moves from the ordinary word-set screens to Training when the server finishes the addition phase', async () => {
  const j = createJourneyFixture('en_ru', 'short'); j.steps[0].status = 'active'; j.steps[0].launch = { routeName: 'Training' }
  route.name = 'WordSetStudy'; route.params = { setId: '123' }
  route.query = { course_code: j.courseCode, journey_id: j.id, journey_step_id: j.steps[0].id }
  client.get.mockResolvedValue(j)
  const scope = effectScope(); scope.run(() => useDailyJourneyExercise('words'))
  await flushPromises()
  expect(replace).toHaveBeenCalledWith({ name: 'Training', params: {}, query: { course_code: j.courseCode, journey_id: j.id, journey_step_id: j.steps[0].id } })
  expect(client.exercise).not.toHaveBeenCalled()
  scope.stop()
 })
})

it('shows loading rather than a pending-action error while an answer is in flight', async () => {
 const j = readingJourney()
 route.query = { course_code: j.courseCode, journey_id: j.id, journey_step_id: j.steps[1].id }
 client.get.mockResolvedValue(j)
 client.exercise.mockResolvedValue({ journey: j, step_id: j.steps[1].id, mode: 'reading', data: { answer: true } })
 let reject!: (cause: Error) => void
 client.act.mockImplementationOnce(() => new Promise((_, fail) => { reject = fail }))
 const scope = effectScope()
 const daily = scope.run(() => useDailyJourneyExercise('reading'))!
 await flushPromises()
 const request = daily.act('answer', { answer: 'true' })
 await flushPromises()
 expect(daily.hasPendingAction.value).toBe(true)
 expect(daily.busy.value).toBe(true)
 expect(daily.error.value).toBe('')
 reject(new Error('Network failure'))
 await request
 expect(daily.busy.value).toBe(false)
 expect(daily.hasPendingAction.value).toBe(true)
 expect(daily.error.value).toBe('Network failure')
 scope.stop()
})
