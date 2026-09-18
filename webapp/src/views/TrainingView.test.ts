import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { flushPromises, shallowMount } from '@vue/test-utils'
import { ref } from 'vue'
import TrainingView from './TrainingView.vue'
import { wordTrainingClient } from '../api/wordTrainingClient'

vi.mock('../api/wordTrainingClient', () => ({ wordTrainingClient: {
  current: vi.fn(), prefetchNext: vi.fn(), reveal: vi.fn(), answer: vi.fn(),
  getDashboard: vi.fn(), getUpcoming: vi.fn(),
} }))
vi.mock('../api/client', () => ({ apiClient: {
  request: vi.fn().mockResolvedValue({}), setNetworkErrorCallback: vi.fn(), setNetworkSuccessCallback: vi.fn(),
} }))
vi.mock('vue-i18n', async importOriginal => ({ ...await importOriginal<typeof import('vue-i18n')>(), useI18n: () => ({ t: (key: string) => key, tm: () => [], locale: ref('ru') }) }))
vi.mock('../composables/useLocale', () => ({ useLocale: () => ({ currentLocale: ref('ru') }) }))
vi.mock('../composables/useLearningConfig', () => ({ useLearningConfig: () => ({ ensureLearningLoaded: vi.fn(), learning: ref({ target_lang: 'es' }) }) }))
vi.mock('../composables/useCourse', () => ({ useCourse: () => ({ currentCourseCode: ref('es_ru') }) }))
vi.mock('../composables/useSpanishVerbFormsPractice', () => ({ useSpanishVerbFormsPractice: () => ({ refreshVerbFormsPoolCount: vi.fn(), showSpanishVerbFormsTraining: ref(false) }) }))
vi.mock('../composables/useSettings', () => ({ useSettings: () => ({ setAutoplayPronunciation: vi.fn(), settings: ref({ soundsEnabled: false, vibrationEnabled: false, autoplayPronunciation: false }) }) }))
vi.mock('../composables/useAudio', () => ({ useAudio: () => ({ getWordPronunciationURL: vi.fn().mockResolvedValue(null) }) }))
vi.mock('../composables/useDialog', () => ({ showAlert: vi.fn() }))
vi.mock('chart.js', () => ({ Chart: class { static register() {} }, registerables: [] }))

const card = (index: number) => ({
  type: 'card', session_id: 10, card_index: index, total_cards: 3,
  user_card_id: index, question: `<strong>${['casa', 'дерево', 'libro'][index - 1]}</strong>`,
  direction: index === 2 ? 'ru_en' : 'en_ru', delay_ms: 1000,
})
const reveal = (index: number) => ({
  user_card_id: index, session_id: 10, card_index: index,
  options: index === 2 ? ['árbol', 'flor'] : ['дом', 'книга'],
})
function deferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>(r => { resolve = r })
  return { promise, resolve }
}
let wrapper: ReturnType<typeof shallowMount> | undefined
async function mountTraining() {
  wrapper = shallowMount(TrainingView)
  await flushPromises()
  return wrapper
}
beforeEach(() => {
  vi.useFakeTimers()
  vi.clearAllMocks()
  vi.mocked(wordTrainingClient.current).mockResolvedValue(card(1))
  vi.mocked(wordTrainingClient.prefetchNext).mockResolvedValue(card(2))
  vi.mocked(wordTrainingClient.reveal).mockResolvedValue(reveal(1))
  vi.mocked(wordTrainingClient.answer).mockResolvedValue({ is_correct: true, correct_answer: 'дом' })
  vi.mocked(wordTrainingClient.getDashboard).mockResolvedValue({ due_count: 3 })
  vi.mocked(wordTrainingClient.getUpcoming).mockResolvedValue({})
})
afterEach(() => { wrapper?.unmount(); wrapper = undefined; vi.clearAllTimers(); vi.useRealTimers() })

it('sends only one reveal while a request for the same card is pending', async () => {
  const pending = deferred<ReturnType<typeof reveal>>()
  vi.mocked(wordTrainingClient.reveal).mockReturnValue(pending.promise)
  const w = await mountTraining()
  const state = (w.vm as any).$.setupState
  void state.revealOptions()
  void state.revealOptions()
  await vi.advanceTimersByTimeAsync(1000)
  expect(wordTrainingClient.reveal).toHaveBeenCalledTimes(1)
  pending.resolve(reveal(1))
  await flushPromises()
  expect(w.findAll('.option-text').map(n => n.text())).toEqual(['дом', 'книга'])
})

it('recovers the current question when reveal belongs to another card', async () => {
  const w = await mountTraining()
  vi.mocked(wordTrainingClient.current).mockResolvedValue(card(2))
  vi.mocked(wordTrainingClient.reveal).mockResolvedValueOnce(reveal(2))
  await vi.advanceTimersByTimeAsync(1000)
  await flushPromises()
  expect(w.get('.question').text()).toBe('дерево')
  expect(w.findAll('.option-text')).toHaveLength(0)
  vi.mocked(wordTrainingClient.reveal).mockResolvedValue(reveal(2))
  await vi.advanceTimersByTimeAsync(1000)
  expect(w.findAll('.option-text').map(n => n.text())).toEqual(['árbol', 'flor'])
})

it('does not skip a card if prefetch was processed after the answer advanced the server', async () => {
  const pending = deferred<ReturnType<typeof card>>()
  vi.mocked(wordTrainingClient.prefetchNext).mockReturnValueOnce(pending.promise)
  const w = await mountTraining()
  await vi.advanceTimersByTimeAsync(1000)
  await flushPromises()
  await w.get('.option-btn').trigger('click')
  await flushPromises()
  // The request for the next card reached the server after it advanced to #2.
  pending.resolve(card(3))
  await flushPromises()
  vi.mocked(wordTrainingClient.current).mockResolvedValue(card(2))
  await vi.advanceTimersByTimeAsync(1000)
  expect(w.get('.question').text()).toBe('дерево')
})

it('ignores a delayed reveal after leaving the training screen', async () => {
  const pending = deferred<ReturnType<typeof reveal>>()
  vi.mocked(wordTrainingClient.reveal).mockReturnValue(pending.promise)
  const w = await mountTraining()
  await vi.advanceTimersByTimeAsync(1000)
  w.unmount()
  wrapper = undefined
  pending.resolve(reveal(2))
  await flushPromises()
  expect(wordTrainingClient.current).toHaveBeenCalledTimes(1)
})

it('ignores options from the previous card after the current card changes', async () => {
  const pending = deferred<ReturnType<typeof reveal>>()
  vi.mocked(wordTrainingClient.reveal).mockReturnValueOnce(pending.promise)
  const w = await mountTraining()
  await vi.advanceTimersByTimeAsync(1000)
  const state = (w.vm as any).$.setupState
  state.setupCard(card(2))
  pending.resolve(reveal(1))
  await flushPromises()
  expect(w.get('.question').text()).toBe('дерево')
  expect(w.findAll('.option-text')).toHaveLength(0)
  vi.mocked(wordTrainingClient.reveal).mockResolvedValue(reveal(2))
  await vi.advanceTimersByTimeAsync(1000)
  expect(w.findAll('.option-text').map(n => n.text())).toEqual(['árbol', 'flor'])
})

it('still shows a valid prefetched card and waits for server synchronization before reveal', async () => {
  const w = await mountTraining()
  await vi.advanceTimersByTimeAsync(1000)
  await w.get('.option-btn').trigger('click')
  await flushPromises()
  const sync = deferred<ReturnType<typeof card>>()
  vi.mocked(wordTrainingClient.current).mockReturnValueOnce(sync.promise)
  await vi.advanceTimersByTimeAsync(2000)
  expect(w.get('.question').text()).toBe('дерево')
  expect(wordTrainingClient.reveal).toHaveBeenCalledTimes(1)
  vi.mocked(wordTrainingClient.reveal).mockResolvedValue(reveal(2))
  sync.resolve(card(2))
  await flushPromises()
  expect(w.findAll('.option-text').map(n => n.text())).toEqual(['árbol', 'flor'])
})
