import { journeyMessage } from '../components/dailyJourney/messages'
import { computed, ref, watch } from 'vue'
import { courseClient } from '../api/courseClient'
import { emitAppDataEvent } from '../api/cacheInvalidation'
import { currentUserScope } from '../api/sessionScope'
import { useAuth } from './useAuth'
import { useDailyJourneyExercise } from './useDailyJourneyExercise'
export interface JourneyChatSession {
 session_id: number; messages: any[]; tasks: any[]; turn_count: number; max_turns: number
 status: string; quest_passed: boolean; budget_exhausted: boolean; attempt_finished: boolean
 [key: string]: any
}
export function useDailyJourneyChat(mode: 'picture' | 'conversation') {
 const daily = useDailyJourneyExercise<JourneyChatSession>(mode)
 const { isAuthenticated } = useAuth()
 const draft = ref('')
 const pending = ref<{ sessionId: number; text: string; index: number } | null>(null)
 const messages = computed(() => {
  const state = daily.data.value
  const confirmed = state?.messages || []
  const message = pending.value
  if (!message || state?.session_id !== message.sessionId) return confirmed
  const saved = confirmed[message.index]
  if (saved?.role === 'user' && saved.content === message.text) return confirmed
  return [...confirmed, { role: 'user', content: message.text }]
 })
 function beginMessage(state: JourneyChatSession | null, text: string) {
  if (state && text) pending.value = { sessionId: state.session_id, text, index: state.messages.length }
  if (draft.value.trim() === text.trim()) draft.value = ''
 }
 function restoreFailedMessage(text: string) {
  pending.value = null
  if (!draft.value) draft.value = text
 }
 const key = computed(() => {
  const ctx = daily.context.value
  return daily.isJourney.value && ctx && isAuthenticated.value ? `daily-journey:chat-draft:${currentUserScope()}:${ctx.courseCode}:${ctx.journeyId}:${ctx.stepId}` : ''
 })
 watch(key, value => { pending.value = null; try { draft.value = value ? sessionStorage.getItem(value) || '' : '' } catch { draft.value = '' } }, { immediate: true })
 watch(draft, value => { if (!key.value) return; try { if (value) sessionStorage.setItem(key.value, value); else sessionStorage.removeItem(key.value) } catch { /* storage is optional */ } })
 async function session() { return daily.ready() }
 async function send(sessionId: number, text: string) {
  const requestKey = key.value
  const course = daily.context.value?.courseCode
  const state = await daily.ready()
  if (requestKey !== key.value || sessionId !== state.session_id || state.attempt_finished || daily.hasPendingAction.value) throw new Error(journeyMessage('staleChat'))
  beginMessage(state, text)
  try {
   await daily.act('message', { text })
   if (daily.error.value) throw new Error(daily.error.value)
   const confirmed = await daily.ready()
   if (requestKey === key.value) {
    emitAppDataEvent(mode === 'picture' ? 'picture-quest-progressed' : 'conversation-progressed', course)
    pending.value = null
   }
   return confirmed
  } catch (error) { if (requestKey === key.value) restoreFailedMessage(text); throw error }
 }
 async function retry() {
  const requestKey = key.value
  const course = daily.context.value?.courseCode
  const resending = daily.hasPendingAction.value
  const action = daily.pendingAction?.value
  const text = action?.action === 'message' && typeof action.payload?.text === 'string' ? action.payload.text : draft.value
  if (resending) beginMessage(daily.data.value, text)
  try {
   await daily.retry()
   if (daily.error.value) throw new Error(daily.error.value)
   const confirmed = await daily.ready()
   if (requestKey === key.value) {
    if (resending) emitAppDataEvent(mode === 'picture' ? 'picture-quest-progressed' : 'conversation-progressed', course)
    pending.value = null
   }
   return confirmed
  } catch (error) { if (resending && requestKey === key.value) restoreFailedMessage(text); throw error }
 }

 async function back() {
  if (daily.data.value?.attempt_finished) {
   await daily.act('ack_feedback')
   if (daily.error.value) throw new Error(daily.error.value)
   draft.value = ''
  }
  await daily.back()
 }
 // A standalone screen keeps its usual transport; a journey never falls back.
 return { daily, draft, messages, session, send, retry, back, standalone: courseClient }
}
