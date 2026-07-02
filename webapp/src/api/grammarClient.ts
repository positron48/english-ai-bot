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
  getStoredChapters,
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
  const meta = await getOfflineMeta(activeCourseCode())
  if (!meta) throw new OfflineGrammarUnavailableError()
  return meta
}

async function offlineFallback<T>(online: () => Promise<T>, offline: () => Promise<T>): Promise<T> {
  if (isBrowserOffline()) return offline()
  try {
    return await online()
  } catch (error) {
    if (!isNetworkError(error)) throw error
    return offline()
  }
}

const categoriesCacheTTL = 30_000
let categoriesCache: { data: { categories: any[] }; expiresAt: number } | null = null
let categoriesRequest: Promise<{ categories: any[] }> | null = null

// Current course code — set via setGrammarCourse() when user switches courses.
let _grammarCourseCode = ''

/** Notify grammarClient of course change. Clears in-memory cache so next
 *  API call fetches data for the new course. */
export function setGrammarCourse(courseCode: string): void {
  if (_grammarCourseCode !== courseCode) {
    _grammarCourseCode = courseCode
    clearCategoriesCache()
  }
}

function grammarCourseParam(): string {
  return _grammarCourseCode ? `?course_code=${encodeURIComponent(_grammarCourseCode)}` : ''
}

function activeCourseCode(): string {
  return _grammarCourseCode
}

/** Current course code, as set via setGrammarCourse(). Used by other clients
 *  (e.g. TTS, reading audio) that need to request resources for the active course. */
export function getGrammarCourseCode(): string {
  return _grammarCourseCode
}

function clearCategoriesCache() {
  categoriesCache = null
  categoriesRequest = null
}

async function fetchOnlineCategories(): Promise<{ categories: any[] }> {
  const now = Date.now()
  if (categoriesCache && categoriesCache.expiresAt > now) return clone(categoriesCache.data)
  if (categoriesRequest) return clone(await categoriesRequest)
  categoriesRequest = apiClient.request(`/api/learning/grammar/categories${grammarCourseParam()}`)
    .then((data: any) => {
      const normalized = { categories: data.categories || [] }
      categoriesCache = { data: normalized, expiresAt: Date.now() + categoriesCacheTTL }
      return normalized
    })
    .finally(() => {
      categoriesRequest = null
    })
  return clone(await categoriesRequest)
}

function computeChapterAccess(meta: OfflineGrammarMeta, chapterID: string): boolean {
  for (const section of meta.sections) {
    const index = section.chapters.findIndex((chapter) => chapter.chapter_id === chapterID)
    if (index < 0) continue
    const chapter = section.chapters[index]
    if (chapter.passed || chapter.can_access) return true
    if (index === 0) return computeSectionAccess(meta, section.section_id)
    return !!section.chapters[index - 1]?.passed
  }
  return false
}

function computeSectionAccess(meta: OfflineGrammarMeta, sectionID: string): boolean {
  const index = meta.sections.findIndex((section) => section.section_id === sectionID)
  if (index < 0) return false
  const section = meta.sections[index]
  if (section.can_access || section.opened_by_placement) return true
  if (index === 0) return true
  const prev = meta.sections[index - 1]
  if ((prev.category_test_score || 0) >= 50) return true
  return prev.chapters.length > 0 && prev.chapters.every((chapter) => chapter.passed)
}

function isChapterStudied(meta: OfflineGrammarMeta, chapterID: string): boolean {
  for (const section of meta.sections) {
    const chapter = section.chapters.find((item) => item.chapter_id === chapterID)
    if (!chapter) continue
    if (chapter.passed) return true
    if (section.opened_by_placement) return true
    return false
  }
  return false
}

async function filterTrainingQuestionsByStudiedChapters(questions: any[]): Promise<any[]> {
  const meta = await getOfflineMeta(activeCourseCode())
  if (!meta) return questions
  return questions.filter((q) => {
    const chapterID = q?.chapter_id
    return typeof chapterID === 'string' && chapterID.length > 0 && isChapterStudied(meta, chapterID)
  })
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
  await setOfflineMeta(meta, activeCourseCode())
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
    const chapterPayload = await getStoredChapter(chapterID, activeCourseCode())
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
    course_code: activeCourseCode(),
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

function chapterIDFromPayload(payload: any): string {
  return payload?.chapter?.id || payload?.chapter?.chapter_id || payload?.chapter_id || ''
}

function trainingQuestionsFromChapterPayload(payload: any): any[] {
  const chapterID = chapterIDFromPayload(payload)
  if (!chapterID) return []
  const questions = payload?.chapter?.question_bank?.questions || []
  if (!Array.isArray(questions)) return []
  return questions
    .filter((question: any) => question?.id)
    .map((question: any) => ({
      ...question,
      id: `${chapterID}:${question.id}`,
      _offline_original_question_id: question.id,
      chapter_id: chapterID,
    }))
}

async function getOfflineTrainingQuestionPool(): Promise<any[]> {
  const trainingQuestions = await getTrainingQuestions(activeCourseCode())
  const pool = trainingQuestions.length > 0
    ? trainingQuestions
    : (await getStoredChapters(activeCourseCode())).flatMap(trainingQuestionsFromChapterPayload)
  return filterTrainingQuestionsByStudiedChapters(pool)
}

async function queueOfflineTrainingAttempt(questionID: string, answer: any, result: any): Promise<void> {
  const id = typeof crypto !== 'undefined' && 'randomUUID' in crypto
    ? crypto.randomUUID()
    : `offline-training-${Date.now()}-${Math.random().toString(16).slice(2)}`
  await enqueueTrainingAttempt({
    client_attempt_id: id,
    course_code: activeCourseCode(),
    question_id: questionID,
    answer,
    created_at: new Date().toISOString(),
    result,
  })
}

export const grammarClient = {
  async getOfflineStatus(): Promise<OfflineStatus> {
    const meta = await getOfflineMeta(activeCourseCode())
    const downloadedChapters = await countStoredChapters(activeCourseCode())
    const pendingAttempts = (await queueCount(activeCourseCode())) + (await trainingQueueCount(activeCourseCode()))
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

  async getOfflineDebugState(): Promise<any> {
    const meta = await getOfflineMeta(activeCourseCode())
    const downloadedChapters = await countStoredChapters(activeCourseCode())
    const storedChapters = await getStoredChapters(activeCourseCode())
    const trainingQuestions = await getTrainingQuestions(activeCourseCode())
    const fallbackTrainingQuestions = storedChapters.flatMap(trainingQuestionsFromChapterPayload)
    const pendingAttempts = (await queueCount(activeCourseCode())) + (await trainingQueueCount(activeCourseCode()))
    const sections = meta?.sections || []
    const accessibleSections = sections
      .filter((section) => section.chapters.length > 0 || computeSectionAccess(meta!, section.section_id))
      .map((section) => section.section_id)
    return {
      href: typeof location !== 'undefined' ? location.href : '',
      navigatorOnLine: typeof navigator !== 'undefined' ? navigator.onLine : null,
      serviceWorkerControlled: typeof navigator !== 'undefined' ? !!navigator.serviceWorker?.controller : null,
      displayModeStandalone: typeof window !== 'undefined' ? window.matchMedia?.('(display-mode: standalone)').matches : null,
      referrer: typeof document !== 'undefined' ? document.referrer : '',
      hasMeta: !!meta,
      bundleID: meta?.bundle_id || '',
      versionHash: meta?.version_hash || '',
      downloadedChapters,
      totalChapters: meta?.total_chapters || 0,
      sectionCount: sections.length,
      accessibleSectionCount: accessibleSections.length,
      firstAccessibleSection: accessibleSections[0] || '',
      firstSection: sections[0]?.section_id || '',
      storedChapterSample: storedChapters.slice(0, 3).map(chapterIDFromPayload),
      trainingQuestions: trainingQuestions.length,
      fallbackTrainingQuestions: fallbackTrainingQuestions.length,
      pendingAttempts,
    }
  },

  async preload(onProgress?: (done: number, total: number) => void): Promise<OfflineStatus> {
    const manifest = await apiClient.request<OfflineGrammarMeta>(`/api/learning/grammar/offline/manifest${grammarCourseParam()}`)
    const meta = { ...manifest, downloaded_at: new Date().toISOString() }
    await setOfflineMeta(meta, activeCourseCode())
    const chapters = meta.sections.flatMap((section) => section.chapters)
    let done = 0
    onProgress?.(done, chapters.length)
    for (const chapter of chapters) {
      const payload = await apiClient.request<any>(chapter.download_url)
      await setStoredChapter(chapter.chapter_id, payload, activeCourseCode())
      done++
      onProgress?.(done, chapters.length)
    }
    const trainingURL = (manifest as any)?.training_pack?.download_url
    if (trainingURL) {
      const trainingPack = await apiClient.request<any>(trainingURL)
      await setTrainingQuestions(trainingPack?.questions || [], activeCourseCode())
    }
    return this.getOfflineStatus()
  },

  async clear(): Promise<void> {
    await clearOfflineGrammar(activeCourseCode())
  },

  async syncQueuedAttempts(): Promise<number> {
    if (isBrowserOffline()) return 0
    const attempts = await getQueuedAttempts(activeCourseCode())
    let synced = 0
    if (attempts.length > 0) {
      let response: any
      try {
        response = await apiClient.request('/api/learning/grammar/offline/sync-attempts', {
          method: 'POST',
          body: { attempts } as any,
        })
      } catch (error) {
        if (isNetworkError(error)) return synced
        throw error
      }
      for (const item of response.results || []) {
        if (item.synced && item.client_attempt_id) {
          await deleteQueuedAttempt(item.client_attempt_id)
          synced++
        }
      }
      if (synced > 0) clearCategoriesCache()
    }
    const trainingAttempts = await getQueuedTrainingAttempts(activeCourseCode())
    if (trainingAttempts.length > 0) {
      let trainingResponse: any
      try {
        trainingResponse = await apiClient.request('/api/learning/grammar/offline/sync-training-attempts', {
          method: 'POST',
          body: { attempts: trainingAttempts } as any,
        })
      } catch (error) {
        if (isNetworkError(error)) return synced
        throw error
      }
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
    return offlineFallback(
      () => fetchOnlineCategories(),
      async () => {
        const meta = await requireMeta()
        return { categories: meta.sections.map(({ chapters, ...section }) => ({
          ...section,
          can_access: computeSectionAccess(meta, section.section_id),
        })) }
      },
    )
  },

  async getStatistics(): Promise<any> {
    return offlineFallback(
      () => apiClient.request(`/api/learning/grammar/statistics${grammarCourseParam()}`),
      async () => buildStatistics(await requireMeta()),
    )
  },

  async getContinueChapter(): Promise<{ chapter: {
    chapter_id: string
    title: string
    title_translations?: Record<string, string>
    section_id?: string
    url: string
  } | null }> {
    return offlineFallback(
      () => apiClient.request(`/api/learning/grammar/continue-chapter${grammarCourseParam()}`),
      async () => {
        const meta = await requireMeta()
        let frontier: {
          chapter_id: string
          title: string
          title_translations?: Record<string, string>
          section_id: string
          url: string
        } | null = null
        for (const section of meta.sections) {
          if (!computeSectionAccess(meta, section.section_id)) break
          const placementOpened = !!section.opened_by_placement
          for (let i = 0; i < section.chapters.length; i++) {
            const chapter = section.chapters[i]
            const canAccess = computeChapterAccess(meta, chapter.chapter_id)
            const passed = !!chapter.passed
            if (!canAccess && !passed) break
            const item = {
              chapter_id: chapter.chapter_id,
              title: chapter.title,
              title_translations: chapter.title_translations,
              section_id: section.section_id,
              url: `/learning/grammar/chapter/${chapter.chapter_id}`,
            }
            frontier = item
            if (canAccess && !passed && !(placementOpened && !(chapter.best_score > 0))) {
              return { chapter: item }
            }
          }
        }
        return { chapter: frontier }
      },
    )
  },

  async getTrainingAvailability(): Promise<any> {
    return offlineFallback(
      () => apiClient.request(`/api/learning/grammar/training/availability${grammarCourseParam()}`),
      async () => {
        const questions = await getOfflineTrainingQuestionPool()
        const blocks = new Set(questions.map((q: any) => q?.theory_block_id).filter(Boolean))
        return { grammar_training: { available: questions.length > 0, offline: true, question_count: questions.length, theory_block_count: blocks.size, due_theory_block_count: blocks.size } }
      },
    )
  },

  async startTrainingSession(limit = 20): Promise<any> {
    return offlineFallback(
      () => apiClient.request(`/api/learning/grammar/training/session/start${grammarCourseParam()}`, {
        method: 'POST',
        body: { limit } as any,
      }),
      async () => {
        const questions = await getOfflineTrainingQuestionPool()
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
    )
  },

  async submitTrainingAnswer(questionID: string, answer: any): Promise<any> {
    if (!isBrowserOffline()) {
      try {
        return await apiClient.request(`/api/learning/grammar/training/session/answer${grammarCourseParam()}`, {
          method: 'POST',
          body: { question_id: questionID, answer } as any,
        })
      } catch (error) {
        if (!isNetworkError(error)) throw error
      }
    }
    const questions = await getOfflineTrainingQuestionPool()
    const question = questions.find((q: any) => q?.id === questionID || q?._offline_original_question_id === questionID)
    if (!question) throw new OfflineGrammarUnavailableError('Training question is not available offline')
    const correct = compareAnswers(answer, question.correct_answer, question.type)
    const result = {
      correct,
      correct_answer: question.correct_answer,
      explanation: question.explanation,
      offline: true,
      queued: true,
    }
    await queueOfflineTrainingAttempt(question._offline_original_question_id || questionID, answer, result)
    return result
  },

  async getChapters(sectionID: string): Promise<{ chapters: any[] }> {
    return offlineFallback(
      () => apiClient.request(`/api/learning/grammar/categories/${sectionID}/chapters${grammarCourseParam()}`),
      async () => {
        const meta = await requireMeta()
        const section = meta.sections.find((item) => item.section_id === sectionID)
        if (!section) throw new Error('Section not found')
        return { chapters: section.chapters.map((chapter) => ({ ...chapter, can_access: computeChapterAccess(meta, chapter.chapter_id) })) }
      },
    )
  },

  async getChapter(chapterID: string): Promise<any> {
    return offlineFallback(
      () => apiClient.request(`/api/learning/grammar/chapters/${chapterID}${grammarCourseParam()}`),
      async () => {
        const payload = await getStoredChapter(chapterID, activeCourseCode())
        if (!payload) throw new OfflineGrammarUnavailableError('Chapter is not available offline')
        if (!computeChapterAccess(await requireMeta(), chapterID)) throw new OfflineGrammarUnavailableError('Chapter is locked offline')
        return payload
      },
    )
  },

  async getChapterForTheory(chapterID: string): Promise<any> {
    return offlineFallback(
      () => apiClient.request(`/api/learning/grammar/chapters/${chapterID}${grammarCourseParam()}`),
      async () => {
        const payload = await getStoredChapter(chapterID, activeCourseCode())
        if (!payload) throw new OfflineGrammarUnavailableError('Chapter is not available offline')
        const meta = await requireMeta()
        let sectionMeta: any = payload.section
        if (!sectionMeta) {
          for (const section of meta.sections) {
            if (section.chapters.some((chapter) => chapter.chapter_id === chapterID)) {
              sectionMeta = {
                section_id: section.section_id,
                title: section.title,
                title_translations: section.title_translations,
                level: section.level,
              }
              break
            }
          }
        }
        return { ...payload, section: sectionMeta }
      },
    )
  },

  async canAccessChapter(chapterID: string): Promise<{ can_access: boolean }> {
    return offlineFallback(
      () => apiClient.request(`/api/learning/grammar/chapters/${chapterID}/access${grammarCourseParam()}`),
      async () => ({ can_access: computeChapterAccess(await requireMeta(), chapterID) }),
    )
  },

  async canAccessSection(sectionID: string): Promise<{ can_access: boolean }> {
    return offlineFallback(
      () => apiClient.request(`/api/learning/grammar/categories/${sectionID}/access${grammarCourseParam()}`),
      async () => ({ can_access: computeSectionAccess(await requireMeta(), sectionID) }),
    )
  },

  async getChapterTest(chapterID: string): Promise<{ questions: any[]; total: number }> {
    return offlineFallback(
      () => apiClient.request(`/api/learning/grammar/chapters/${chapterID}/test${grammarCourseParam()}`),
      async () => {
        const payload = await getStoredChapter(chapterID, activeCourseCode())
        if (!payload) throw new OfflineGrammarUnavailableError('Chapter test is not available offline')
        if (!computeChapterAccess(await requireMeta(), chapterID)) throw new OfflineGrammarUnavailableError('Chapter test is locked offline')
        const num = Number(payload?.chapter?.chapter_test?.num_questions || 10)
        const selected = shuffle(chapterPoolQuestions(payload)).slice(0, num).map(sanitizeQuestionForTest)
        return { questions: selected, total: selected.length }
      },
    )
  },

  async getCategoryTest(sectionID: string): Promise<{ questions: any[]; total: number }> {
    return offlineFallback(
      () => apiClient.request(`/api/learning/grammar/categories/${sectionID}/test${grammarCourseParam()}`),
      async () => {
        const meta = await requireMeta()
        const section = meta.sections.find((item) => item.section_id === sectionID)
        if (!section) throw new Error('Section not found')
        if (!computeSectionAccess(meta, sectionID)) throw new OfflineGrammarUnavailableError('Category test is locked offline')
        if (!section.chapters.every((chapter) => chapter.passed)) throw new OfflineGrammarUnavailableError('Complete all chapters to unlock this category test offline')
        const selected: any[] = []
        const seen = new Set<string>()
        const perChapter: Array<{ chapterID: string; questions: any[] }> = []
        for (const chapter of section.chapters) {
          const payload = await getStoredChapter(chapter.chapter_id, activeCourseCode())
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
    )
  },

  async getPlacementTest(): Promise<{ questions: any[]; total: number }> {
    return offlineFallback(
      () => apiClient.request('/api/learning/grammar/placement-test'),
      async () => {
        const meta = await requireMeta()
        const candidates: any[] = []
        for (const section of meta.sections) {
          for (const chapter of section.chapters) {
            const payload = await getStoredChapter(chapter.chapter_id, activeCourseCode())
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
    )
  },

  async submitTest(scope: 'chapter' | 'category', scopeID: string, answers: any[]): Promise<any> {
    if (!isBrowserOffline()) {
      try {
        const result = await apiClient.request(`/api/learning/grammar/tests/submit${grammarCourseParam()}`, {
          method: 'POST',
          body: { scope, scope_id: scopeID, answers } as any,
        })
        clearCategoriesCache()
        return result
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
      try {
        const result = await apiClient.request(`/api/learning/grammar/placement-test/submit${grammarCourseParam()}`, {
          method: 'POST',
          body: answersMap as any,
        })
        clearCategoriesCache()
        return result
      } catch (error) {
        if (!isNetworkError(error)) throw error
      }
    }
    const meta = await requireMeta()
    const results: any[] = []
    let correct = 0
    let total = 0
    for (const [compoundID, answer] of Object.entries(answersMap)) {
      const [chapterID, questionID] = compoundID.includes(':') ? compoundID.split(/:(.*)/s).filter(Boolean) : ['', compoundID]
      if (!chapterID || !questionID) continue
      const payload = await getStoredChapter(chapterID, activeCourseCode())
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
        if (openedSections.includes(section.section_id)) {
          section.can_access = true
          section.opened_by_placement = true
          for (const chapter of section.chapters) chapter.can_access = true
        }
      }
      await setOfflineMeta(meta, activeCourseCode())
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
    return offlineFallback(
      () => apiClient.request(`/api/learning/grammar/chapters/${chapterID}/next${grammarCourseParam()}`),
      async () => {
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
    )
  },
}
