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
 const key = computed(() => {
  const ctx = daily.context.value
  return daily.isJourney.value && ctx && isAuthenticated.value ? `daily-journey:chat-draft:${currentUserScope()}:${ctx.courseCode}:${ctx.journeyId}:${ctx.stepId}` : ''
 })
 watch(key, value => { try { draft.value = value ? sessionStorage.getItem(value) || '' : '' } catch { draft.value = '' } }, { immediate: true })
 watch(draft, value => { if (!key.value) return; try { if (value) sessionStorage.setItem(key.value, value); else sessionStorage.removeItem(key.value) } catch { /* storage is optional */ } })
 async function session() { return daily.ready() }
 async function send(sessionId: number, text: string) {
  const state = await daily.ready()
  if (sessionId !== state.session_id || state.attempt_finished || daily.hasPendingAction.value) throw new Error(journeyMessage('staleChat'))
  await daily.act('message', { text })
  if (daily.error.value) throw new Error(daily.error.value)
  emitAppDataEvent(mode === 'picture' ? 'picture-quest-progressed' : 'conversation-progressed', daily.context.value?.courseCode)
  draft.value = ''
  return daily.ready()
 }
 async function retry() { const pending = daily.hasPendingAction.value; await daily.retry(); if (daily.error.value) throw new Error(daily.error.value); if (pending) { draft.value = ''; emitAppDataEvent(mode === 'picture' ? 'picture-quest-progressed' : 'conversation-progressed', daily.context.value?.courseCode) }; return daily.ready() }
 async function back() {
  if (daily.data.value?.attempt_finished) {
   await daily.act('ack_feedback')
   if (daily.error.value) throw new Error(daily.error.value)
   draft.value = ''
  }
  await daily.back()
 }
 // A standalone screen keeps its usual transport; a journey never falls back.
 return { daily, draft, session, send, retry, back, standalone: courseClient }
}
