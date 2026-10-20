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
  chat.draft.value = 'Hola'
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

it.each(['picture', 'conversation'] as const)('%s displays the sent message immediately and reconciles confirmation without duplicates', async mode => {
 let finish!: () => void
 const pending = new Promise<void>(resolve => { finish = resolve })
 daily.act.mockImplementation(async () => { daily.hasPendingAction.value = true; await pending })
 const chat = scope.run(() => useDailyJourneyChat(mode))!
 chat.draft.value = 'Mi mensaje'
 const sending = chat.send(12, 'Mi mensaje')
 await nextTick(); await nextTick()
 expect(chat.draft.value).toBe('')
 expect(chat.messages.value).toEqual([{ role: 'user', content: 'Mi mensaje' }])
 // A refresh with the previous confirmed snapshot must not remove the pending bubble.
 daily.data.value = { ...state }
 await nextTick()
 expect(chat.messages.value).toHaveLength(1)
 chat.draft.value = 'Next draft'
 state = { ...state, messages: [{ role: 'user', content: 'Mi mensaje' }, { role: 'assistant', content: 'Respuesta' }] }
 daily.data.value = state
 await nextTick()
 expect(chat.messages.value).toHaveLength(2)
 daily.hasPendingAction.value = false
 finish(); await sending
 expect(chat.messages.value).toEqual(state.messages)
 expect(chat.draft.value).toBe('Next draft')
 expect(daily.act).toHaveBeenCalledOnce()
})

it('restores a failed message and retries its saved text without overwriting a new draft', async () => {
 const chat = scope.run(() => useDailyJourneyChat('picture'))!
 chat.draft.value = 'Original'
 daily.pendingAction = ref(null)
 daily.act.mockImplementation(async () => {
  daily.pendingAction.value = { action: 'message', payload: { text: 'Original' } }
  daily.hasPendingAction.value = true
  daily.error.value = 'Network failure'
 })
 await expect(chat.send(12, 'Original')).rejects.toThrow('Network failure')
 expect(chat.draft.value).toBe('Original')
 expect(chat.messages.value).toEqual([])
 chat.draft.value = 'New draft'
 let finish!: () => void
 const pending = new Promise<void>(resolve => { finish = resolve })
 daily.retry.mockImplementation(async () => {
  await pending
  state = { ...state, messages: [{ role: 'user', content: 'Original' }, { role: 'assistant', content: 'Reply' }] }
  daily.data.value = state; daily.hasPendingAction.value = false; daily.error.value = ''
 })
 const retrying = chat.retry()
 await nextTick()
 expect(chat.messages.value).toEqual([{ role: 'user', content: 'Original' }])
 expect(chat.draft.value).toBe('New draft')
 finish(); await retrying
 expect(chat.messages.value).toEqual(state.messages)
 expect(chat.draft.value).toBe('New draft')
 expect(daily.act).toHaveBeenCalledOnce()
})
