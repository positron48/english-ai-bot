import { expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import WordSetDetailView from './WordSetDetailView.vue'
import { apiClient } from '../api/client'

vi.mock('../api/client', () => ({ apiClient: { request: vi.fn() } }))
vi.mock('vue-router', () => ({ useRoute: () => ({ params: { setId: '1' } }), useRouter: () => ({ push: vi.fn() }) }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
vi.mock('../composables/useAudio', () => ({ useAudio: () => ({ getWordPronunciationURL: vi.fn().mockResolvedValue(null), playWordPronunciation: vi.fn() }) }))
vi.mock('../composables/useDialog', () => ({ showAlert: vi.fn() }))

it('closes the added word before the background refresh completes, without hiding the list', async () => {
  const request = vi.mocked(apiClient.request)
  request.mockResolvedValueOnce({ word_set: { id: 1, title: 'Words', total_words: 1, known_words: 0, words_in_vocab: 0, unknown_words: 1, progress_percent: 0 }, words: [{ word_card_id: 2, word: 'casa', status: 'unknown' }] })
  const wrapper = mount(WordSetDetailView)
  await flushPromises()
  request.mockResolvedValueOnce({ training_card: { word_en: 'casa', word_ru: 'дом' } })
  await wrapper.get('.word-item').trigger('click')
  await flushPromises()
  expect(wrapper.find('.modal').exists()).toBe(true)
  request.mockResolvedValueOnce({ success: true })
  let resolveRefresh!: (value: unknown) => void
  request.mockImplementationOnce(() => new Promise(resolve => { resolveRefresh = resolve }))
  await wrapper.get('.btn-learn').trigger('click')
  await flushPromises()
  expect(wrapper.find('.modal').exists()).toBe(false)
  expect(wrapper.find('.loading').exists()).toBe(false)
  expect(wrapper.get('.word-item').classes()).toContain('status-in_vocab')
  resolveRefresh({ word_set: { id: 1, title: 'Words', unknown_words: 0 }, words: [{ word_card_id: 2, word: 'casa', status: 'in_vocab' }] })
  await flushPromises()
  wrapper.unmount()
})
