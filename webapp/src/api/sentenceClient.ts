import { apiClient } from './client'

export interface SentenceGradeToken {
  text: string
  status: 'ok' | 'wrong' | 'insert'
  correction?: string
}

export interface SentenceGrade {
  error_count: number
  outcome: 'star' | 'passed' | 'failed'
  corrected_es: string
  tokens: SentenceGradeToken[]
  explanation?: string
}

export interface SentenceItem {
  id: number
  position: number
  attempted: boolean
  prompt_ru?: string
  outcome?: 'star' | 'passed' | 'failed'
  error_count?: number
  user_input?: string
  grading?: SentenceGrade
  total?: number
  attempted_count?: number
  remaining?: number
}

export interface SentenceTodaySummary {
  available: boolean
  set_id?: number
  status?: string
  generation_date?: string
  total?: number
  attempted?: number
  remaining?: number
  stars?: number
  passed?: number
  failed?: number
}

export interface SentenceSetState {
  status: string
  stars: number
  passed: number
  failed: number
  total: number
  attempted: number
}

export interface SentenceAnswerResult {
  grading: SentenceGrade
  outcome: 'star' | 'passed' | 'failed'
  error_count: number
  done: boolean
  set: SentenceSetState
}

function courseQuery(courseCode?: string): string {
  return courseCode ? `?course_code=${encodeURIComponent(courseCode)}` : ''
}

export const sentenceClient = {
  async today(courseCode?: string): Promise<SentenceTodaySummary> {
    return apiClient.request<SentenceTodaySummary>(`/api/sentence-training/today${courseQuery(courseCode)}`)
  },

  async start(courseCode?: string): Promise<{ set_id: number; status: string; items: SentenceItem[] }> {
    return apiClient.request(`/api/sentence-training/start${courseQuery(courseCode)}`, { method: 'POST' })
  },

  async current(courseCode?: string): Promise<SentenceItem & { done?: boolean }> {
    return apiClient.request(`/api/sentence-training/current${courseQuery(courseCode)}`)
  },

  async answer(itemId: number, userInput: string): Promise<SentenceAnswerResult> {
    return apiClient.request(`/api/sentence-training/answer`, {
      method: 'POST',
      body: JSON.stringify({ item_id: itemId, user_input: userInput }),
    })
  },
}
