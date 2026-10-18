import { beforeEach, afterEach, describe, expect, it, vi } from 'vitest'
import { effectScope, nextTick, ref } from 'vue'
import { useDailyJourneyChat } from './useDailyJourneyChat'
const { source, ordinary, emit } = vi.hoisted(() => ({ source: vi.fn(), ordinary: { startConversationSession: vi.fn(), postConversationMessage: vi.fn(), startPictureQuestSession: vi.fn(), postPictureQuestMessage: vi.fn() }, emit: vi.fn() }))
vi.mock('./useDailyJourneyExercise', () => ({ useDailyJourneyExercise: source }))
vi.mock('../api/courseClient', () => ({ courseClient: ordinary }))
vi.mock('../api/cacheInvalidation', () => ({ emitAppDataEvent: emit }))
vi.mock('../api/sessionScope', () => ({ currentUserScope: () => 'user-a' }))
vi.mock('./useAuth', () => ({ useAuth: () => ({ isAuthenticated: ref(true) }) }))
let daily: any, state: any
let scope = effectScope()
beforeEach(() => { localStorage.setItem('locale','ru');
 vi.clearAllMocks(); sessionStorage.clear();scope = effectScope()
 state = { session_id: 12, messages: [], tasks: [], turn_count: 0, max_turns: 2, status: 'open', quest_passed: false, budget_exhausted: false, attempt_finished: false }
 daily = { isJourney: ref(true), context: ref({ courseCode: 'es_ru', journeyId: 'j1', stepId: 's1' }), data: ref(state), error: ref(''), hasPendingAction: ref(false), ready: vi.fn(async () => state), act: vi.fn(), retry: vi.fn(), back: vi.fn() }
 source.mockReturnValue(daily)
})
afterEach(() => scope.stop())
describe('daily chat source on the ordinary chat pages', () => {
 it('restores confirmed messages, saves the scoped draft and never starts a standalone chat', async () => {
  const chat = scope.run(() => useDailyJourneyChat('picture'))!
  chat.draft.value = 'Mi descripción';await nextTick()
  expect(sessionStorage.getItem('daily-journey:chat-draft:user-a:es_ru:j1:s1')).toBe('Mi descripción')
  expect(await chat.session()).toEqual(state)
  expect(ordinary.startPictureQuestSession).not.toHaveBeenCalled()
  await expect(chat.send(13, 'Hola')).rejects.toThrow('изменилось')
  expect(daily.act).not.toHaveBeenCalled()
  await chat.send(12, 'Hola')
  expect(daily.act).toHaveBeenCalledExactlyOnceWith('message', { text: 'Hola' })
  expect(emit).toHaveBeenCalledWith('picture-quest-progressed', 'es_ru')
  expect(chat.draft.value).toBe('')
 })
 it('retries the original pending action and keeps a new message from replacing it', async () => {
  const chat = scope.run(() => useDailyJourneyChat('conversation'))!
  daily.hasPendingAction.value = true;chat.draft.value = 'Original'
  await expect(chat.send(12, 'Different')).rejects.toThrow('предыдущий')
  await chat.retry()
  expect(daily.retry).toHaveBeenCalledOnce()
  expect(daily.act).not.toHaveBeenCalled()
  expect(chat.draft.value).toBe('')
  expect(ordinary.postConversationMessage).not.toHaveBeenCalled()
 })
 it('pauses on ordinary exit and acknowledges only a finished attempt', async () => {
  const chat = scope.run(() => useDailyJourneyChat('conversation'))!
  await chat.back();expect(daily.act).not.toHaveBeenCalled()
  state.attempt_finished = true
  await chat.back();expect(daily.act).toHaveBeenCalledExactlyOnceWith('ack_feedback')
  expect(daily.back).toHaveBeenCalledTimes(2)
 })
 it('retains drafts on provider failure and read-only refresh', async () => {
  const chat = scope.run(() => useDailyJourneyChat('picture'))!
  chat.draft.value = 'Draft'
  await chat.retry();expect(chat.draft.value).toBe('Draft')
  daily.act.mockImplementation(() => { daily.error.value = 'Provider unavailable' })
  await expect(chat.send(12, 'Draft')).rejects.toThrow('Provider unavailable')
  expect(chat.draft.value).toBe('Draft')
  expect(ordinary.postPictureQuestMessage).not.toHaveBeenCalled()
 })
})
