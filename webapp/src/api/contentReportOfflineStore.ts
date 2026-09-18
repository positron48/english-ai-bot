import { createOfflineStore } from './offlineStore'

export type ContentReportSourceType =
  | 'word_training'
  | 'grammar_training'
  | 'grammar_chapter'
  | 'grammar_test'
  | 'reading_text'

export interface QueuedContentReport {
  client_report_id: string
  source_type: ContentReportSourceType
  report_category: string
  comment: string
  created_at: string
  word?: string
  direction?: string
  word_card_id?: number
  training_card_id?: number
  user_card_id?: number
  word_category?: string
  grammar_chapter_id?: string
  theory_block_id?: string
  grammar_question_id?: string
  reading_text_id?: string
  reading_category_id?: string
  payload: Record<string, unknown>
}

type StoreName = 'meta' | 'queue'

const withStore = createOfflineStore<StoreName>('contentReportOfflineStore', { meta: undefined, queue: { keyPath: 'client_report_id' } })

export async function enqueueContentReport(report: QueuedContentReport): Promise<void> {
  await withStore('queue', 'readwrite', (store) => store.put(report))
}

export async function getQueuedContentReports(): Promise<QueuedContentReport[]> {
  return await withStore<QueuedContentReport[]>('queue', 'readonly', store => store.getAll()) || []
}

export async function deleteQueuedContentReport(clientReportID: string): Promise<void> {
  await withStore('queue', 'readwrite', (store) => store.delete(clientReportID))
}

export async function contentReportQueueCount(): Promise<number> {
  const items = await getQueuedContentReports()
  return items.length
}
