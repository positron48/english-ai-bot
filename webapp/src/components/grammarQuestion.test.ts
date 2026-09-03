import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import GrammarQuestion from './GrammarQuestion.vue'
import ru from '../locales/ru.json'

describe('GrammarQuestion reorder', () => {
  it('shows the Russian sentence before answering and keeps it while selecting words', async () => {
    const wrapper = mount(GrammarQuestion, {
      props: { question: {
        id: 'q', type: 'reorder', prompt: 'Расставьте слова в правильном порядке:',
        correct_answer: 'I work at home.', translation_ru: 'Я работаю дома.',
      } },
      global: { plugins: [createI18n({ legacy: false, locale: 'ru', messages: { ru } })] },
    })
    expect(wrapper.get('.reorder-translation').text()).toBe('Я работаю дома.')
    expect(wrapper.get('.reorder-translation').attributes('lang')).toBe('ru')
    expect(wrapper.get('.reorder-translation').element.nextElementSibling?.className).toBe('reorder-container')
    expect(wrapper.findAll('.available-word')).toHaveLength(4)
    await wrapper.get('.available-word').trigger('click')
    expect(wrapper.findAll('.sentence-word')).toHaveLength(1)
    expect(wrapper.get('.reorder-translation').text()).toBe('Я работаю дома.')
    expect(wrapper.emitted('answer')).toBeUndefined()
    for (let i = 0; i < 3; i++) await wrapper.get('.available-word').trigger('click')
    expect(wrapper.emitted('answer')).toHaveLength(1)
    wrapper.unmount()
  })

  it('renders the translation as text, not HTML', () => {
    const translation = '<img src=x onerror=alert(1)> Я работаю дома.'
    const wrapper = mount(GrammarQuestion, {
      props: { question: { id: 'q', type: 'reorder', prompt: '', correct_answer: 'I work.', translation_ru: translation } },
      global: { plugins: [createI18n({ legacy: false, locale: 'ru', messages: { ru } })] },
    })
    expect(wrapper.get('.reorder-translation').text()).toBe(translation)
    expect(wrapper.find('.reorder-translation img').exists()).toBe(false)
    wrapper.unmount()
  })

  it('also shows the full translation when an ambiguous reorder was converted to a gap', () => {
    const wrapper = mount(GrammarQuestion, {
      props: { question: { id: 'q', type: 'fill_blank', prompt: 'She ___ at home. (work)', correct_answer: 'works', translation_ru: 'Она работает дома.' } },
      global: { plugins: [createI18n({ legacy: false, locale: 'ru', messages: { ru } })] },
    })
    expect(wrapper.get('.reorder-translation').text()).toBe('Она работает дома.')
    expect(wrapper.find('.question-input').exists()).toBe(true)
    wrapper.unmount()
  })
})
