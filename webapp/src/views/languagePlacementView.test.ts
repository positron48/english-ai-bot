import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import { createMemoryHistory, createRouter } from 'vue-router'
import { ref } from 'vue'
import LanguagePlacementView from './LanguagePlacementView.vue'
import { placementClient, type PlacementSession } from '../api/placementClient'
import ru from '../locales/ru.json'

const course = ref('es_ru')
const me = ref({ id: 7 })
vi.mock('../composables/useCourse', () => ({ useCourse: () => ({ currentCourseCode: course, ensureCourseLoaded: vi.fn().mockResolvedValue(undefined) }) }))
vi.mock('../composables/useMe', () => ({ useMe: () => ({ me, ensureMe: vi.fn().mockResolvedValue({ id: 7 }) }) }))
vi.mock('../api/placementClient', () => ({ placementClient: { start: vi.fn(), get: vi.fn(), answer: vi.fn(), finish: vi.fn() } }))
vi.mock('../api/grammarClient', () => ({ clearCategoriesCache: vi.fn() }))
vi.mock('../api/cacheInvalidation', () => ({ emitAppDataEvent: vi.fn() }))

function fixture(answered = 0, count = 30): PlacementSession {
  const questions = Array.from({ length: count }, (_, i) => ({
    id: `q${i + 1}`, context: 'Вы ищете аптеку.', instruction: 'Дополните ответ.',
    prompt: 'La farmacia ___ al lado del banco.',
    choices: [{ id: 'a', text: 'está' }, { id: 'b', text: 'es' }, { id: 'c', text: 'hay' }],
  }))
  return { id: 'a'.repeat(32), course_code: 'es_ru', status: 'active', questions,
    answers: Object.fromEntries(questions.slice(0, answered).map(q => [q.id, 'a'])),
    base_closed: answered >= 30, clarifying: count > 30 }
}

async function mountView() {
  const router = createRouter({ history: createMemoryHistory(), routes: [
    { path: '/learning/placement-test', component: LanguagePlacementView },
    { path: '/learning/grammar', component: { template: '<div />' } },
    { path: '/learning/grammar/chapter/:chapterId', name: 'GrammarChapter', component: { template: '<div />' } },
  ] })
  await router.push('/learning/placement-test')
  await router.isReady()
  const wrapper = mount(LanguagePlacementView, { global: {
    plugins: [router, createI18n({ legacy: false, locale: 'ru', messages: { ru } })],
  } })
  await flushPromises()
  return wrapper
}

describe('LanguagePlacementView', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
    course.value = 'es_ru'
    me.value = { id: 7 }
    vi.mocked(placementClient.start).mockResolvedValue(fixture())
  })

  it('shows an independent welcome and starts only on an explicit click', async () => {
    const wrapper = await mountView()
    expect(wrapper.text()).toContain('Читать курс заранее не нужно')
    expect(wrapper.text()).toContain('30 вопросов')
    expect(placementClient.start).not.toHaveBeenCalled()
    await wrapper.get('[data-test="start"]').trigger('click')
    await flushPromises()
    expect(placementClient.start).toHaveBeenCalledWith('es_ru', expect.any(String), false)
    expect(wrapper.text()).toContain('Вопрос 1 из 30')
    expect(wrapper.get('[data-test="save"]').attributes('disabled')).toBeDefined()
    wrapper.unmount()
  })

  it('renders plain text, preserves an unsent unknown answer, and restores it for the same user and course', async () => {
    const initial = fixture()
    initial.questions[0].context = '<img src=x onerror=alert(1)>'
    vi.mocked(placementClient.start).mockResolvedValue(initial)
    vi.mocked(placementClient.answer).mockRejectedValue(new TypeError('Failed to fetch'))
    let wrapper = await mountView()
    await wrapper.get('[data-test="start"]').trigger('click')
    await flushPromises()
    expect(wrapper.find('img').exists()).toBe(false)
    expect(wrapper.text()).toContain('<img src=x onerror=alert(1)>')
    await wrapper.get('[data-test="unknown"]').setValue()
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(placementClient.answer).toHaveBeenCalledWith(initial.id, 'es_ru', 'q1', '')
    expect(wrapper.text()).toContain('Выбранный ответ сохранён на этом устройстве')
    const saved = JSON.parse(localStorage.getItem('linglow.placement.v1:7:es_ru')!)
    expect(saved.pending).toEqual({ questionID: 'q1', answer: '' })
    wrapper.unmount()

    vi.mocked(placementClient.get).mockResolvedValue(initial)
    wrapper = await mountView()
    expect(placementClient.get).not.toHaveBeenCalled()
    await wrapper.get('[data-test="start"]').trigger('click')
    await flushPromises()
    expect(placementClient.get).toHaveBeenCalledWith(initial.id, 'es_ru')
    expect((wrapper.get('[data-test="unknown"]').element as HTMLInputElement).checked).toBe(true)
    expect(placementClient.answer).toHaveBeenCalledTimes(1)
    wrapper.unmount()

    me.value = { id: 8 }
    wrapper = await mountView()
    expect(wrapper.get('[data-test="start"]').text()).toContain('Начать тест')
    wrapper.unmount()
  })

  it('closes the 30-question block and shows all six server-selected follow-ups', async () => {
    vi.mocked(placementClient.start).mockResolvedValue(fixture(29))
    const extended = fixture(30, 36)
    vi.mocked(placementClient.answer).mockResolvedValue(extended)
    const wrapper = await mountView()
    await wrapper.get('[data-test="start"]').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('Вопрос 30 из 30')
    await wrapper.get('input[value="a"]').setValue()
    await wrapper.get('form').trigger('submit')
    await flushPromises()
    expect(wrapper.text()).toContain('Вопрос 31 из 36')
    expect(wrapper.text()).toContain('Добавили 6 вопросов')
    expect(wrapper.find('.placement-question-actions .placement-secondary').exists()).toBe(false)
    wrapper.unmount()
  })

  it('finishes explicitly, shows skill recommendations and starts a fresh retake', async () => {
    const complete = fixture(30)
    vi.mocked(placementClient.start).mockResolvedValue(complete)
    vi.mocked(placementClient.finish).mockResolvedValue({ ...complete, status: 'completed', available_chapter_ids: ['es.grammar.location.estar'], result: {
      level: 'A2', upper_level: 'B1', estimated: true, correct: 20, total: 30,
      profile: [{ level: 'A2', correct: 4, total: 6, status: 'borderline' }],
      review: [{ ...complete.questions[0], level: 'A1', skill_title: 'Место', answer: '', correct_answer: 'a', correct: false, explanation: 'Говорим о местоположении.', chapter_ids: [] }],
      recommended_skills: [{ id: 'location', level: 'A1', title: 'Местоположение', description: 'Где находится знакомое место.', section_id: 'es.grammar.location', chapter_ids: ['es.grammar.location.estar'] }],
      opened_sections: ['es.grammar.location'],
    } })
    const wrapper = await mountView()
    await wrapper.get('[data-test="start"]').trigger('click')
    await flushPromises()
    expect(placementClient.finish).not.toHaveBeenCalled()
    await wrapper.get('[data-test="finish"]').trigger('click')
    await flushPromises()
    expect(wrapper.text()).toContain('A2–B1')
    expect(wrapper.text()).toContain('Местоположение')
    expect(wrapper.get('.placement-recommendations a').attributes('href')).toContain('course_code=es_ru')
    expect(wrapper.findAll('details')).toHaveLength(1)
    const oldKey = vi.mocked(placementClient.start).mock.calls[0][1]
    vi.mocked(placementClient.start).mockResolvedValue(fixture())
    await wrapper.get('[data-test="retake"]').trigger('click')
    await flushPromises()
    expect(placementClient.start).toHaveBeenLastCalledWith('es_ru', expect.any(String), true)
    expect(vi.mocked(placementClient.start).mock.calls[1][1]).not.toBe(oldKey)
    wrapper.unmount()
  })

  it('links only to available result chapters and keeps locked recommended topics visible', async () => {
    const complete = fixture(30)
    const availableChapter = 'es.grammar.location.estar'
    const lockedChapter = 'es.grammar.past.imperfecto'
    vi.mocked(placementClient.start).mockResolvedValue({ ...complete, status: 'completed',
      available_chapter_ids: [availableChapter], result: {
        level: 'A1', upper_level: 'A1', estimated: true, correct: 12, total: 30,
        profile: [], review: [], opened_sections: ['es.grammar.location'],
        recommended_skills: [
          { id: 'location', level: 'A1', title: 'Местоположение', description: 'Где находится знакомое место.', section_id: 'es.grammar.location', chapter_ids: ['es.grammar.location.later', availableChapter] },
          { id: 'past', level: 'A2', title: 'События в прошлом', description: 'Как рассказать о прошлом.', section_id: 'es.grammar.past', chapter_ids: [lockedChapter] },
        ],
      },
    })
    const wrapper = await mountView()
    await wrapper.get('[data-test="start"]').trigger('click')
    await flushPromises()
    const topics = wrapper.findAll('.placement-recommendations li')
    expect(topics).toHaveLength(2)
    expect(topics[0].get('a').attributes('href')).toContain(availableChapter)
    expect(topics[0].get('a').attributes('href')).toContain('course_code=es_ru')
    expect(topics[1].text()).toContain('События в прошлом')
    expect(topics[1].text()).toContain(ru.placement.topicLater)
    expect(topics[1].find('a').exists()).toBe(false)
    expect(wrapper.findAll('.placement-recommendations a')).toHaveLength(1)
    expect(wrapper.findAll('a').some(link => link.text() === ru.placement.goToCourse && link.attributes('href') === '/learning/grammar')).toBe(true)
    wrapper.unmount()
  })

  it('ignores an in-flight response after a course switch', async () => {
    let resolve!: (value: PlacementSession) => void
    vi.mocked(placementClient.start).mockReturnValue(new Promise(done => { resolve = done }))
    const wrapper = await mountView()
    await wrapper.get('[data-test="start"]').trigger('click')
    course.value = 'en_ru'
    await flushPromises()
    resolve(fixture())
    await flushPromises()
    expect(wrapper.text()).toContain('Английский')
    expect(wrapper.find('form').exists()).toBe(false)
    expect(localStorage.getItem('linglow.placement.v1:7:en_ru')).toBeNull()
    wrapper.unmount()
  })
})
