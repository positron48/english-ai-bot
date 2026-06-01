import { apiClient } from './client'
import {
  OfflineGrammarMeta,
  QueuedGrammarAttempt,
  clearOfflineGrammar,
  countStoredChapters,
  deleteQueuedAttempt,
  deleteQueuedTrainingAttempt,
  enqueueTrainingAttempt,
  enqueueAttempt,
  getOfflineMeta,
  getQueuedAttempts,
  getQueuedTrainingAttempts,
  getStoredChapter,
  getTrainingQuestions,
  queueCount,
  setTrainingQuestions,
  setOfflineMeta,
  setStoredChapter,
  trainingQueueCount,
} from './grammarOfflineStore'

export class OfflineGrammarUnavailableError extends Error {
  constructor(message = 'Grammar is not preloaded for offline use') {
    super(message)
    this.name = 'OfflineGrammarUnavailableError'
  }
}

export interface OfflineStatus {
  ready: boolean
  downloading: boolean
  downloadedChapters: number
  totalChapters: number
  versionHash?: string
  downloadedAt?: string
  pendingAttempts: number
}

const isBrowserOffline = () => typeof navigator !== 'undefined' && navigator.onLine === false
const isNetworkError = (error: any) => error?.isNetworkError || error?.name === 'TypeError' || String(error?.message || '').includes('Failed to fetch')

function clone<T>(value: T): T {
  return JSON.parse(JSON.stringify(value))
}

function shuffle<T>(items: T[]): T[] {
  const arr = [...items]
  for (let i = arr.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1))
    ;[arr[i], arr[j]] = [arr[j], arr[i]]
  }
  return arr
}

function normalizeAnswer(value: any): any {
  if (typeof value === 'string') return value.trim().replace(/\s+/g, ' ').toLowerCase()
  return value
}

function normalizeTrueFalse(value: any): string | null {
  if (typeof value === 'boolean') return value ? 'true' : 'false'
  if (typeof value === 'string') {
    const lower = value.trim().toLowerCase()
    if (['true', 'да', 'yes', '1'].includes(lower)) return 'true'
    if (['false', 'нет', 'no', '0'].includes(lower)) return 'false'
  }
  return null
}

function compareAnswers(userAnswer: any, correctAnswer: any, questionType?: string): boolean {
  if (questionType === 'true_false') {
    const u = normalizeTrueFalse(userAnswer)
    const c = normalizeTrueFalse(correctAnswer)
    if (u !== null && c !== null) return u === c
  }
  if (typeof correctAnswer === 'string') return normalizeAnswer(userAnswer) === normalizeAnswer(correctAnswer)
  if (Array.isArray(correctAnswer)) {
    if (!Array.isArray(userAnswer) || userAnswer.length !== correctAnswer.length) return false
    const left = [...userAnswer].map(normalizeAnswer).sort()
    const right = [...correctAnswer].map(normalizeAnswer).sort()
    return left.every((value, index) => value === right[index])
  }
  return userAnswer === correctAnswer
}

function sanitizeQuestionForTest(question: any): any {
  const q = clone(question)
  if (q.type !== 'reorder') delete q.correct_answer
  return q
}

function questionsByID(chapterPayload: any): Map<string, any> {
  const questions = chapterPayload?.chapter?.question_bank?.questions || []
  return new Map(questions.map((q: any) => [q.id, q]))
}

function chapterPoolQuestions(chapterPayload: any): any[] {
  const questionMap = questionsByID(chapterPayload)
  const pool = chapterPayload?.chapter?.chapter_test?.pool_question_ids || []
  const ids = Array.isArray(pool) && pool.length > 0 ? pool : [...questionMap.keys()]
  return ids.map((id: string) => questionMap.get(id)).filter(Boolean)
}

async function requireMeta(): Promise<OfflineGrammarMeta> {
  const meta = await getOfflineMeta()
  if (!meta) throw new OfflineGrammarUnavailableError()
  return meta
}

function computeChapterAccess(meta: OfflineGrammarMeta, chapterID: string): boolean {
  for (const section of meta.sections) {
    const index = section.chapters.findIndex((chapter) => chapter.chapter_id === chapterID)
    if (index < 0) continue
    const chapter = section.chapters[index]
    if (chapter.passed || chapter.can_access || section.can_access) return true
    if (index === 0) return section.can_access
    return !!section.chapters[index - 1]?.passed
  }
  return false
}

function computeSectionAccess(meta: OfflineGrammarMeta, sectionID: string): boolean {
  const index = meta.sections.findIndex((section) => section.section_id === sectionID)
  if (index < 0) return false
  const section = meta.sections[index]
  if (section.can_access) return true
  if (index === 0) return true
  const prev = meta.sections[index - 1]
  return prev.chapters.length > 0 && prev.chapters.every((chapter) => chapter.passed) && (prev.category_test_score || 0) >= 50
}

async function updateLocalProgress(scope: 'chapter' | 'category', scopeID: string, result: any): Promise<void> {
  const meta = await requireMeta()
  const now = new Date().toISOString()
  if (scope === 'chapter') {
    for (const section of meta.sections) {
      const chapter = section.chapters.find((item) => item.chapter_id === scopeID)
      if (!chapter) continue
      chapter.best_score = Math.max(chapter.best_score || 0, result.score || 0)
      chapter.passed = chapter.passed || !!result.passed
      ;(chapter as any).last_attempt_at = now
      section.passed_chapters = section.chapters.filter((item) => item.passed).length
      section.progress_percentage = section.chapters.length > 0
        ? Math.floor(section.chapters.reduce((sum, item) => sum + (item.best_score || 0), 0) / section.chapters.length)
        : 0
    }
  } else {
    const section = meta.sections.find((item) => item.section_id === scopeID)
    if (section) section.category_test_score = Math.max(section.category_test_score || 0, result.score || 0)
  }
  await setOfflineMeta(meta)
}

function buildStatistics(meta: OfflineGrammarMeta) {
  const chapters = meta.sections.flatMap((section) => section.chapters)
  const passed = chapters.filter((chapter) => chapter.passed).length
  const total = chapters.length
  const scoreSum = chapters.reduce((sum, chapter) => sum + (chapter.best_score || 0), 0)
  return {
    confirmed_level: '',
    course_completion_pct: total > 0 ? Math.round((passed / total) * 100) : 0,
    whole_course_completion_pct: total > 0 ? Math.round((passed / total) * 100) : 0,
    average_test_score: total > 0 ? Math.round(scoreSum / total) : 0,
    passed_chapters: passed,
    total_chapters: total,
    total_chapters_in_course: total,
    hide_placement_test_button: false,
  }
}

async function gradeOfflineTest(scope: 'chapter' | 'category', scopeID: string, answers: Array<{ question_id: string; chapter_id?: string; answer: any }>) {
  const resultItems: any[] = []
  let correct = 0
  let total = 0
  for (const item of answers) {
    const chapterID = scope === 'chapter' ? scopeID : item.chapter_id
    if (!chapterID) continue
    const chapterPayload = await getStoredChapter(chapterID)
    const question = questionsByID(chapterPayload).get(item.question_id)
    if (!question) continue
    const isCorrect = item.answer !== null && item.answer !== undefined && compareAnswers(item.answer, question.correct_answer, question.type)
    if (isCorrect) correct++
    total++
    const resultItem: any = {
      question_id: item.question_id,
      prompt: question.prompt,
      correct: isCorrect,
      user_answer: item.answer,
      correct_answer: question.correct_answer,
      explanation: question.explanation,
    }
    if (scope === 'category' && chapterID) resultItem.chapter_id = chapterID
    resultItems.push(resultItem)
  }
  const score = total > 0 ? Math.floor((correct * 100) / total) : 0
  return { score, passed: score >= 50, correct, total, results: resultItems, offline: true, queued: true }
}

async function queueOfflineAttempt(scope: 'chapter' | 'category', scopeID: string, answers: any[], result: any): Promise<void> {
  const meta = await requireMeta()
  const id = typeof crypto !== 'undefined' && 'randomUUID' in crypto
    ? crypto.randomUUID()
    : `offline-${Date.now()}-${Math.random().toString(16).slice(2)}`
  const attempt: QueuedGrammarAttempt = {
    client_attempt_id: id,
    scope,
    scope_id: scopeID,
    answers,
    course_version: meta.version_hash,
    created_at: new Date().toISOString(),
    result,
  }
  await enqueueAttempt(attempt)
}

function sanitizeTrainingQuestion(question: any): any {
  const q = clone(question)
  delete q.correct_answer
  return q
}

async function queueOfflineTrainingAttempt(questionID: string, answer: any, result: any): Promise<void> {
  const id = typeof crypto !== 'undefined' && 'randomUUID' in crypto
    ? crypto.randomUUID()
    : `offline-training-${Date.now()}-${Math.random().toString(16).slice(2)}`
  await enqueueTrainingAttempt({
    client_attempt_id: id,
    question_id: questionID,
    answer,
    created_at: new Date().toISOString(),
    result,
  })
}

export const grammarClient = {
  async getOfflineStatus(): Promise<OfflineStatus> {
    const meta = await getOfflineMeta()
    const downloadedChapters = await countStoredChapters()
    const pendingAttempts = (await queueCount()) + (await trainingQueueCount())
    return {
      ready: !!meta && downloadedChapters >= (meta.total_chapters || 0),
      downloading: false,
      downloadedChapters,
      totalChapters: meta?.total_chapters || 0,
      versionHash: meta?.version_hash,
      downloadedAt: meta?.downloaded_at,
      pendingAttempts,
    }
  },

  async preload(onProgress?: (done: number, total: number) => void): Promise<OfflineStatus> {
    const manifest = await apiClient.request<OfflineGrammarMeta>('/api/learning/grammar/offline/manifest')
    const meta = { ...manifest, downloaded_at: new Date().toISOString() }
    await setOfflineMeta(meta)
    const chapters = meta.sections.flatMap((section) => section.chapters)
    let done = 0
    onProgress?.(done, chapters.length)
    for (const chapter of chapters) {
      const payload = await apiClient.request<any>(chapter.download_url)
      await setStoredChapter(chapter.chapter_id, payload)
      done++
      onProgress?.(done, chapters.length)
    }
    const trainingURL = (manifest as any)?.training_pack?.download_url
    if (trainingURL) {
      const trainingPack = await apiClient.request<any>(trainingURL)
      await setTrainingQuestions(trainingPack?.questions || [])
    }
    return this.getOfflineStatus()
  },

  async clear(): Promise<void> {
    await clearOfflineGrammar()
  },

  async syncQueuedAttempts(): Promise<number> {
    if (isBrowserOffline()) return 0
    const attempts = await getQueuedAttempts()
    let synced = 0
    if (attempts.length > 0) {
      const response: any = await apiClient.request('/api/learning/grammar/offline/sync-attempts', {
        method: 'POST',
        body: { attempts } as any,
      })
      for (const item of response.results || []) {
        if (item.synced && item.client_attempt_id) {
          await deleteQueuedAttempt(item.client_attempt_id)
          synced++
        }
      }
    }
    const trainingAttempts = await getQueuedTrainingAttempts()
    if (trainingAttempts.length > 0) {
      const trainingResponse: any = await apiClient.request('/api/learning/grammar/offline/sync-training-attempts', {
        method: 'POST',
        body: { attempts: trainingAttempts } as any,
      })
      for (const item of trainingResponse.results || []) {
        if (item.synced && item.client_attempt_id) {
          await deleteQueuedTrainingAttempt(item.client_attempt_id)
          synced++
        }
      }
    }
    return synced
  },

  async getCategories(): Promise<{ categories: any[] }> {
    if (!isBrowserOffline()) return apiClient.request('/api/learning/grammar/categories')
    const meta = await requireMeta()
    return { categories: meta.sections.map(({ chapters, ...section }) => ({
      ...section,
      can_access: computeSectionAccess(meta, section.section_id),
    })) }
  },

  async getStatistics(): Promise<any> {
    if (!isBrowserOffline()) return apiClient.request('/api/learning/grammar/statistics')
    return buildStatistics(await requireMeta())
  },

  async getTrainingAvailability(): Promise<any> {
    if (isBrowserOffline()) {
      const questions = await getTrainingQuestions()
      const blocks = new Set(questions.map((q: any) => q?.theory_block_id).filter(Boolean))
      return { grammar_training: { available: questions.length > 0, offline: true, question_count: questions.length, theory_block_count: blocks.size, due_theory_block_count: blocks.size } }
    }
    return apiClient.request('/api/learning/grammar/training/availability')
  },

  async startTrainingSession(limit = 20): Promise<any> {
    if (!isBrowserOffline()) {
      return apiClient.request('/api/learning/grammar/training/session/start', {
        method: 'POST',
        body: { limit } as any,
      })
    }
    const questions = await getTrainingQuestions()
    if (questions.length === 0) return { items: [] }
    const byBlock = new Map<string, any[]>()
    for (const q of questions) {
      const block = q?.theory_block_id || q?.id
      if (!byBlock.has(block)) byBlock.set(block, [])
      byBlock.get(block)!.push(q)
    }
    const selectedBlocks = shuffle([...byBlock.keys()]).slice(0, Math.max(1, Math.min(30, limit)))
    return {
      items: selectedBlocks.map((block) => {
        const qs = byBlock.get(block) || []
        return { question: sanitizeTrainingQuestion(qs[Math.floor(Math.random() * qs.length)]) }
      }),
      offline: true,
    }
  },

  async submitTrainingAnswer(questionID: string, answer: any): Promise<any> {
    if (!isBrowserOffline()) {
      try {
        return await apiClient.request('/api/learning/grammar/training/session/answer', {
          method: 'POST',
          body: { question_id: questionID, answer } as any,
        })
      } catch (error) {
        if (!isNetworkError(error)) throw error
      }
    }
    const questions = await getTrainingQuestions()
    const question = questions.find((q: any) => q?.id === questionID)
    if (!question) throw new OfflineGrammarUnavailableError('Training question is not available offline')
    const correct = compareAnswers(answer, question.correct_answer, question.type)
    const result = {
      correct,
      correct_answer: question.correct_answer,
      explanation: question.explanation,
      offline: true,
      queued: true,
    }
    await queueOfflineTrainingAttempt(questionID, answer, result)
    return result
  },

  async getChapters(sectionID: string): Promise<{ chapters: any[] }> {
    if (!isBrowserOffline()) return apiClient.request(`/api/learning/grammar/categories/${sectionID}/chapters`)
    const meta = await requireMeta()
    const section = meta.sections.find((item) => item.section_id === sectionID)
    if (!section) throw new Error('Section not found')
    return { chapters: section.chapters.map((chapter) => ({ ...chapter, can_access: computeChapterAccess(meta, chapter.chapter_id) })) }
  },

  async getChapter(chapterID: string): Promise<any> {
    if (!isBrowserOffline()) return apiClient.request(`/api/learning/grammar/chapters/${chapterID}`)
    const payload = await getStoredChapter(chapterID)
    if (!payload) throw new OfflineGrammarUnavailableError('Chapter is not available offline')
    return payload
  },

  async canAccessChapter(chapterID: string): Promise<{ can_access: boolean }> {
    if (!isBrowserOffline()) return apiClient.request(`/api/learning/grammar/chapters/${chapterID}/access`)
    return { can_access: computeChapterAccess(await requireMeta(), chapterID) }
  },

  async canAccessSection(sectionID: string): Promise<{ can_access: boolean }> {
    if (!isBrowserOffline()) return apiClient.request(`/api/learning/grammar/categories/${sectionID}/access`)
    return { can_access: computeSectionAccess(await requireMeta(), sectionID) }
  },

  async getChapterTest(chapterID: string): Promise<{ questions: any[]; total: number }> {
    if (!isBrowserOffline()) return apiClient.request(`/api/learning/grammar/chapters/${chapterID}/test`)
    const payload = await getStoredChapter(chapterID)
    if (!payload) throw new OfflineGrammarUnavailableError('Chapter test is not available offline')
    const num = Number(payload?.chapter?.chapter_test?.num_questions || 10)
    const selected = shuffle(chapterPoolQuestions(payload)).slice(0, num).map(sanitizeQuestionForTest)
    return { questions: selected, total: selected.length }
  },

  async getCategoryTest(sectionID: string): Promise<{ questions: any[]; total: number }> {
    if (!isBrowserOffline()) return apiClient.request(`/api/learning/grammar/categories/${sectionID}/test`)
    const meta = await requireMeta()
    const section = meta.sections.find((item) => item.section_id === sectionID)
    if (!section) throw new Error('Section not found')
    const selected: any[] = []
    const seen = new Set<string>()
    const perChapter: Array<{ chapterID: string; questions: any[] }> = []
    for (const chapter of section.chapters) {
      const payload = await getStoredChapter(chapter.chapter_id)
      const questions = shuffle(chapterPoolQuestions(payload))
      perChapter.push({ chapterID: chapter.chapter_id, questions })
      for (const question of questions.slice(0, 2)) {
        const key = `${chapter.chapter_id}:${question.id}`
        if (seen.has(key)) continue
        seen.add(key)
        selected.push(sanitizeQuestionForTest({ ...question, _category_test_chapter_id: chapter.chapter_id }))
      }
    }
    const remaining = shuffle(perChapter.flatMap((item) => item.questions.map((question) => ({ chapterID: item.chapterID, question }))))
    for (const item of remaining) {
      if (selected.length >= 20) break
      const key = `${item.chapterID}:${item.question.id}`
      if (seen.has(key)) continue
      seen.add(key)
      selected.push(sanitizeQuestionForTest({ ...item.question, _category_test_chapter_id: item.chapterID }))
    }
    return { questions: selected, total: selected.length }
  },

  async getPlacementTest(): Promise<{ questions: any[]; total: number }> {
    if (!isBrowserOffline()) return apiClient.request('/api/learning/grammar/placement-test')
    const meta = await requireMeta()
    const candidates: any[] = []
    for (const section of meta.sections) {
      for (const chapter of section.chapters) {
        const payload = await getStoredChapter(chapter.chapter_id)
        for (const question of chapterPoolQuestions(payload)) {
          candidates.push(sanitizeQuestionForTest({
            ...question,
            id: `${chapter.chapter_id}:${question.id}`,
            _offline_original_question_id: question.id,
            _category_test_chapter_id: chapter.chapter_id,
            placement_chapter_title: chapter.title,
            level: chapter.level || section.level,
          }))
        }
      }
    }
    const selected = shuffle(candidates).slice(0, 25)
    return { questions: selected, total: selected.length }
  },

  async submitTest(scope: 'chapter' | 'category', scopeID: string, answers: any[]): Promise<any> {
    if (!isBrowserOffline()) {
      try {
        return await apiClient.request('/api/learning/grammar/tests/submit', {
          method: 'POST',
          body: { scope, scope_id: scopeID, answers } as any,
        })
      } catch (error) {
        if (!isNetworkError(error)) throw error
      }
    }
    const result = await gradeOfflineTest(scope, scopeID, answers)
    await updateLocalProgress(scope, scopeID, result)
    await queueOfflineAttempt(scope, scopeID, answers, result)
    return result
  },

  async submitPlacementTest(answersMap: Record<string, any>): Promise<any> {
    if (!isBrowserOffline()) {
      return apiClient.request('/api/learning/grammar/placement-test/submit', {
        method: 'POST',
        body: answersMap as any,
      })
    }
    const meta = await requireMeta()
    const results: any[] = []
    let correct = 0
    let total = 0
    for (const [compoundID, answer] of Object.entries(answersMap)) {
      const [chapterID, questionID] = compoundID.includes(':') ? compoundID.split(/:(.*)/s).filter(Boolean) : ['', compoundID]
      if (!chapterID || !questionID) continue
      const payload = await getStoredChapter(chapterID)
      const question = questionsByID(payload).get(questionID)
      if (!question) continue
      const isCorrect = answer !== null && answer !== undefined && compareAnswers(answer, question.correct_answer, question.type)
      if (isCorrect) correct++
      total++
      results.push({
        question_id: compoundID,
        correct: isCorrect,
        user_answer: answer,
        correct_answer: question.correct_answer,
        explanation: question.explanation,
        level: payload?.chapter?.level,
        placement_chapter_title: payload?.title || payload?.chapter?.title,
      })
    }
    const score = total > 0 ? Math.floor((correct * 100) / total) : 0
    const sectionCount = score >= 80 ? meta.sections.length : score >= 50 ? Math.max(1, Math.ceil(meta.sections.length / 3)) : 0
    const openedSections = meta.sections.slice(0, sectionCount).map((section) => section.section_id)
    if (openedSections.length > 0) {
      for (const section of meta.sections) {
        section.can_access = openedSections.includes(section.section_id) || section.can_access
      }
      await setOfflineMeta(meta)
    }
    return {
      score,
      total_questions: total,
      correct,
      opened_sections: openedSections,
      level: score >= 80 ? 'B+' : score >= 50 ? 'A+' : 'A0',
      results,
      offline: true,
    }
  },

  async getNextChapter(chapterID: string): Promise<any> {
    if (!isBrowserOffline()) return apiClient.request(`/api/learning/grammar/chapters/${chapterID}/next`)
    const meta = await requireMeta()
    for (const section of meta.sections) {
      const index = section.chapters.findIndex((chapter) => chapter.chapter_id === chapterID)
      if (index < 0) continue
      const next = section.chapters[index + 1]
      return {
        section_id: section.section_id,
        is_last: !next,
        next_chapter_id: next?.chapter_id || '',
      }
    }
    throw new Error('Chapter not found')
  },
}
