import { apiClient } from './client'

export interface PlacementChoice { id: string; text: string }
export interface PlacementQuestion {
  id: string
  context: string
  instruction: string
  prompt: string
  choices: PlacementChoice[]
}
export interface PlacementSkill {
  id: string
  level: string
  title: string
  description: string
  chapter_ids: string[]
  section_id: string
}
export interface PlacementReview extends PlacementQuestion {
  level: string
  skill_title: string
  answer: string
  correct_answer: string
  correct: boolean
  explanation: string
  chapter_ids: string[]
}
export interface PlacementResult {
  level: string
  upper_level: string
  estimated: boolean
  correct: number
  total: number
  profile: { level: string; correct: number; total: number; status: 'secure' | 'borderline' | 'limited' }[]
  review: PlacementReview[]
  recommended_skills: PlacementSkill[]
  opened_sections: string[]
}
export interface PlacementSession {
  id: string
  course_code: string
  status: 'active' | 'completed' | 'abandoned'
  questions: PlacementQuestion[]
  answers: Record<string, string>
  base_closed: boolean
  clarifying: boolean
  available_chapter_ids?: string[]
  result?: PlacementResult
}

const base = '/api/learning/placement/sessions'
const sessionURL = (id: string, course: string, action = '') =>
  `${base}/${encodeURIComponent(id)}${action}?course_code=${encodeURIComponent(course)}`
const post = <T>(url: string, body: object = {}) => apiClient.request<T>(url, {
  method: 'POST', body: JSON.stringify(body),
})

// Placement always needs its own server session. There is no chapter-bank or
// offline grading fallback: only an unsent choice is retained on the device.
export const placementClient = {
  start: (course: string, key: string, newAttempt = false) => post<PlacementSession>(base, {
    course_code: course, idempotency_key: key, new_attempt: newAttempt,
  }),
  get: (id: string, course: string) => apiClient.request<PlacementSession>(sessionURL(id, course)),
  answer: (id: string, course: string, questionID: string, answer: string) =>
    post<PlacementSession>(sessionURL(id, course, '/answers'), { question_id: questionID, answer }),
  finish: (id: string, course: string) => post<PlacementSession>(sessionURL(id, course, '/finish')),
}
