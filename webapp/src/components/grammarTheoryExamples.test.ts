import { describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import GrammarTheoryExamples from './GrammarTheoryExamples.vue'
import { apiClient } from '../api/client'
import { setGrammarCourse } from '../api/grammarClient'
import ru from '../locales/ru.json'

vi.mock('../api/client', () => ({ apiClient: { request: vi.fn() } }))

describe('grammar example dictionary lookup', () => {
  it('opens the existing word card in the active Spanish course and preserves hidden translation', async () => {
    setGrammarCourse('es_ru')
    vi.mocked(apiClient.request).mockResolvedValue({ lemma: 'vivir', cards: [] })
    const wrapper = mount(GrammarTheoryExamples, {
      props: { examples: [{ id: 'ex1', text: 'Vivo en Madrid.', translation: 'Я живу в Мадриде.' }], variant: 'modal' },
      global: { plugins: [createI18n({ legacy: false, locale: 'ru', messages: { ru } })], stubs: { VocabWordCardsDetail: true, LgLumiCardLoading: true } },
    })
    expect(wrapper.text()).toContain('Vivo en Madrid.')
    expect(wrapper.find('.grammar-theory-example-translation').exists()).toBe(false)
    await wrapper.get('.ct-word').trigger('click')
    await flushPromises()
    expect(apiClient.request).toHaveBeenCalledWith('/api/reading/word-lookup?lemma=Vivo&course_code=es_ru')
    expect(document.querySelector('.ct-modal-panel')).not.toBeNull()
    expect((document.querySelector('.ct-modal-overlay') as HTMLElement).style.zIndex).toBe('2100')
    await wrapper.get('.grammar-theory-show-translation-btn').trigger('click')
    expect(wrapper.get('.grammar-theory-example-translation').text()).toBe('Я живу в Мадриде.')
    wrapper.unmount()
  })
})
