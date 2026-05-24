import { ref } from 'vue'
import { apiClient } from '../api/client'
import { recordingFileName } from '../utils/speechAudio'

export interface SpeakingAvailability {
  available: boolean
  subscription_tier: string
  can_access: boolean
  features: { speaking?: boolean }
  levels: string[]
}

export interface SpeakingCategory {
  category_id: string
  title: string
  level: string
  order: number
  task_count: number
}

export interface SpeakingTask {
  id: string
  category_id: string
  level: string
  type: string
  target_language: string
  title: string
  prompt_ru: string
  display_text?: string
  max_attempts: number
  order: number
}

export interface SpeakingEvaluation {
  attempt_no: number
  max_attempts: number
  can_advance: boolean
  understood_answer: string
  meaning_score: number
  grammar_score: number
  pronunciation_score: number
  fluency_score: number
  is_acceptable: boolean
  audio_quality: string
  short_feedback_ru: string
  better_version: string
  repeat_task: string
}

export interface SpeakingSession {
  id: number
  category_id: string
  status: string
  task_ids: string[]
  current_task_index: number
  current_task: SpeakingTask | null
  total_tasks: number
}

const availability = ref<SpeakingAvailability | null>(null)

export function useSpeaking() {
  async function loadAvailability(): Promise<SpeakingAvailability> {
    const data = await apiClient.request<SpeakingAvailability>('/api/learning/speaking/availability')
    availability.value = data
    return data
  }

  async function loadCategories(): Promise<SpeakingCategory[]> {
    const data = await apiClient.request<{ categories: SpeakingCategory[] }>('/api/learning/speaking/categories')
    return data.categories || []
  }

  async function startSession(categoryId: string): Promise<SpeakingSession> {
    return apiClient.request<SpeakingSession>('/api/learning/speaking/sessions', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ category_id: categoryId }),
    })
  }

  async function getSession(sessionId: number): Promise<SpeakingSession> {
    return apiClient.request<SpeakingSession>(`/api/learning/speaking/sessions/${sessionId}`)
  }

  async function submitAudio(
    sessionId: number,
    taskId: string,
    audio: Blob,
    mode: 'initial' | 'repair' = 'initial'
  ): Promise<SpeakingEvaluation> {
    const form = new FormData()
    form.append('task_id', taskId)
    form.append('mode', mode)
    form.append('audio', audio, recordingFileName(audio))
    return apiClient.request<SpeakingEvaluation>(`/api/learning/speaking/sessions/${sessionId}/submit`, {
      method: 'POST',
      body: form,
    })
  }

  async function nextTask(sessionId: number): Promise<SpeakingSession> {
    return apiClient.request<SpeakingSession>(`/api/learning/speaking/sessions/${sessionId}/next`, {
      method: 'POST',
    })
  }

  return {
    availability,
    loadAvailability,
    loadCategories,
    startSession,
    getSession,
    submitAudio,
    nextTask,
  }
}
