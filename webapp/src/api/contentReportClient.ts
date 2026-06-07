import { apiClient } from './client'
import {
  ContentReportSourceType,
  QueuedContentReport,
  contentReportQueueCount,
  deleteQueuedContentReport,
  enqueueContentReport,
  getQueuedContentReports,
} from './contentReportOfflineStore'

const isBrowserOffline = () => typeof navigator !== 'undefined' && navigator.onLine === false
const isNetworkError = (error: any) =>
  error?.isNetworkError || error?.name === 'TypeError' || String(error?.message || '').includes('Failed to fetch')

function createID(prefix: string): string {
  if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) return crypto.randomUUID()
  return `${prefix}-${Date.now()}-${Math.random().toString(16).slice(2)}`
}

export interface SubmitContentReportInput {
  sourceType: ContentReportSourceType
  reportCategory: string
  comment: string
  payload?: Record<string, unknown>
  word?: string
  direction?: string
  wordCardID?: number
  trainingCardID?: number
  userCardID?: number
  wordCategory?: string
  grammarChapterID?: string
  theoryBlockID?: string
  grammarQuestionID?: string
}

function endpointForSource(sourceType: ContentReportSourceType): string {
  switch (sourceType) {
    case 'word_training':
      return '/api/training/report'
    case 'grammar_training':
      return '/api/learning/grammar/training/report'
    case 'grammar_chapter':
      return '/api/learning/grammar/chapter/report'
    case 'grammar_test':
      return '/api/learning/grammar/test/report'
    default:
      return '/api/content-reports/offline/sync-reports'
  }
}

function buildOnlineBody(input: SubmitContentReportInput, clientReportID: string): Record<string, unknown> {
  const payload = input.payload || {}
  switch (input.sourceType) {
    case 'word_training':
      return {
        user_card_id: input.userCardID || 0,
        word: input.word || '',
        direction: input.direction || '',
        word_card_id: input.wordCardID,
        training_card_id: input.trainingCardID,
        word_category: input.wordCategory || '',
        report_category: input.reportCategory,
        comment: input.comment,
        client_report_id: clientReportID,
        extra: payload,
      }
    case 'grammar_training':
      return {
        question_id: input.grammarQuestionID || payload.question_id || '',
        chapter_id: input.grammarChapterID || payload.chapter_id || '',
        theory_block_id: input.theoryBlockID || payload.theory_block_id || '',
        report_category: input.reportCategory,
        comment: input.comment,
        client_report_id: clientReportID,
        question_data: payload.question_snapshot || payload.question_data || payload,
      }
    case 'grammar_chapter':
      return {
        chapter_id: input.grammarChapterID || payload.chapter_id || '',
        theory_block_id: input.theoryBlockID || payload.theory_block_id || '',
        report_category: input.reportCategory,
        comment: input.comment,
        client_report_id: clientReportID,
        content_snapshot: payload.content_snapshot || payload,
      }
    case 'grammar_test':
      return {
        question_id: input.grammarQuestionID || payload.question_id || '',
        chapter_id: input.grammarChapterID || payload.chapter_id || '',
        scope: payload.scope || '',
        scope_id: payload.scope_id || '',
        report_category: input.reportCategory,
        comment: input.comment,
        client_report_id: clientReportID,
        question_data: payload.question_snapshot || payload.question_data || payload,
      }
    default:
      return { client_report_id: clientReportID }
  }
}

function toQueuedReport(input: SubmitContentReportInput, clientReportID: string): QueuedContentReport {
  return {
    client_report_id: clientReportID,
    source_type: input.sourceType,
    report_category: input.reportCategory,
    comment: input.comment,
    created_at: new Date().toISOString(),
    word: input.word,
    direction: input.direction,
    word_card_id: input.wordCardID,
    training_card_id: input.trainingCardID,
    user_card_id: input.userCardID,
    word_category: input.wordCategory,
    grammar_chapter_id: input.grammarChapterID,
    theory_block_id: input.theoryBlockID,
    grammar_question_id: input.grammarQuestionID,
    payload: input.payload || {},
  }
}

export const contentReportClient = {
  async getPendingCount(): Promise<number> {
    return contentReportQueueCount()
  },

  async submit(input: SubmitContentReportInput): Promise<{ queued: boolean }> {
    const clientReportID = createID('offline-content-report')
    const body = buildOnlineBody(input, clientReportID)
    const endpoint = endpointForSource(input.sourceType)

    if (!isBrowserOffline()) {
      try {
        await apiClient.request(endpoint, {
          method: 'POST',
          body: JSON.stringify(body),
        })
        return { queued: false }
      } catch (error) {
        if (!isNetworkError(error)) throw error
      }
    }

    await enqueueContentReport(toQueuedReport(input, clientReportID))
    return { queued: true }
  },

  async syncQueuedReports(): Promise<number> {
    if (isBrowserOffline()) return 0
    const reports = await getQueuedContentReports()
    if (reports.length === 0) return 0

    let response: any
    try {
      response = await apiClient.request('/api/content-reports/offline/sync-reports', {
        method: 'POST',
        body: JSON.stringify({
          reports: reports.map((item) => ({
            client_report_id: item.client_report_id,
            source_type: item.source_type,
            report_category: item.report_category,
            comment: item.comment,
            word: item.word,
            direction: item.direction,
            word_card_id: item.word_card_id,
            training_card_id: item.training_card_id,
            user_card_id: item.user_card_id,
            word_category: item.word_category,
            grammar_chapter_id: item.grammar_chapter_id,
            theory_block_id: item.theory_block_id,
            grammar_question_id: item.grammar_question_id,
            payload: item.payload,
          })),
        }),
      })
    } catch (error) {
      if (isNetworkError(error)) return 0
      throw error
    }

    let synced = 0
    for (const item of response.results || []) {
      if (item.synced && item.client_report_id) {
        await deleteQueuedContentReport(item.client_report_id)
        synced++
      }
    }
    return synced
  },
}
