import { apiClient } from './client'
import type { DailyJourney, JourneyAction, JourneyTransport, JourneyMode } from '../components/dailyJourney/types'
export interface JourneyExerciseView<T> { journey: DailyJourney; step_id: string; mode: JourneyMode; data: T }
export interface ReadingJourneyExercise {
 document: { id: string; category_id: string; title: string; reading_passage: Record<string, unknown>; cover_hero_rel_path?: string }
 questions: any[]
 answers: Record<number, string>
 correct: Record<number, boolean>
 index: number
 quiz_open: boolean
 feedback: boolean
 done: boolean
 passed: boolean
}
const base = '/api/linglow/daily-journeys'
export const dailyJourneyClient = {
  current(courseCode: string) {
    return apiClient.request<{ enabled: boolean; journey: DailyJourney | null }>(`/api/linglow/daily-journey?course_code=${encodeURIComponent(courseCode)}`)
  },
  start(courseCode: string) {
    return apiClient.request<DailyJourney>(base, { method: 'POST', body: { course_code: courseCode } })
  },
  get(courseCode: string, id: string) {
    return apiClient.request<DailyJourney>(`${base}/${encodeURIComponent(id)}?course_code=${encodeURIComponent(courseCode)}`)
  },
  exercise<T>(courseCode: string, id: string, stepId: string) {
    return apiClient.request<JourneyExerciseView<T>>(`${base}/${encodeURIComponent(id)}/steps/${encodeURIComponent(stepId)}/exercise?course_code=${encodeURIComponent(courseCode)}`)
  },
  act(courseCode: string, id: string, stepId: string, action: JourneyAction) {
    return apiClient.request<DailyJourney>(`${base}/${encodeURIComponent(id)}/steps/${encodeURIComponent(stepId)}/actions?course_code=${encodeURIComponent(courseCode)}`, { method: 'POST', body: { ...action } })
  },
}
export function createDailyJourneyTransport(courseCode: string): JourneyTransport {
  return {
    get: async () => (await dailyJourneyClient.current(courseCode)).journey,
    start: () => dailyJourneyClient.start(courseCode),
    act: (id, stepId, action) => dailyJourneyClient.act(courseCode, id, stepId, action),
  }
}

export interface WordsJourneyExercise {
 phase: 'add' | 'train'
 add_target: number
 added_count: number
 addable: Record<number, number[]> | null
 card?: any
 feedback?: any
 options?: string[]
}
