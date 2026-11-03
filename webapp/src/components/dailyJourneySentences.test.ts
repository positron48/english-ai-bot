import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { ref } from 'vue'
import SentenceCompositionView from '../views/SentenceCompositionView.vue'

const { useDaily, today, start, current, answer, act, back, push, emit } = vi.hoisted(() => ({
  useDaily: vi.fn(), today: vi.fn(), start: vi.fn(), current: vi.fn(), answer: vi.fn(),
  act: vi.fn(), back: vi.fn(), push: vi.fn(), emit: vi.fn(),
}))
vi.mock('../composables/useDailyJourneyExercise', () => ({ useDailyJourneyExercise: useDaily }))
vi.mock('../api/sentenceClient', () => ({ sentenceClient: { today, start, current, answer } }))
vi.mock('../api/cacheInvalidation', () => ({ emitAppDataEvent: emit }))
vi.mock('../api/sessionScope', () => ({ currentUserScope: () => 'user-a' }))
vi.mock('vue-router', () => ({ useRouter: () => ({ push, back, replace: push }) }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
vi.mock('../composables/useCourse', () => ({ useCourse: () => ({ currentCourseCode: ref('es_ru') }) }))
vi.mock('../composables/useMe', () => ({ useMe: () => ({ ensureMe: vi.fn(), hasFeature: () => true }) }))
vi.mock('../composables/useLearningConfig', () => ({ useLearningConfig: () => ({ learning: ref({ target_lang: 'es' }), targetLangDisplay: ref('Spanish') }) }))
let daily: any
const item = { id: 42, position: 1, attempted: false, prompt_ru: 'Я дома' }
const set = { status: 'started', total: 10, attempted: 0, stars: 0, passed: 0, failed: 0 }
const grade = { outcome: 'star', error_count: 0, corrected_es: 'Estoy en casa' }
function savedFeedback() {
  return { current: { ...item, attempted: true, user_input: 'Estoy en casa', grading: grade }, set: { ...set, attempted: 1, stars: 1 }, done: false }
}
const global = { stubs: { LgPageHeader: true, LgIcon: true, VoiceMicButton: true, SentenceGrading: true } }
beforeEach(() => {
  vi.clearAllMocks(); sessionStorage.clear()
  daily = {
    isJourney: ref(true), context: ref({ courseCode: 'es_ru', journeyId: 'j1', stepId: 's1' }),
    data: ref(null), busy: ref(false), hasPendingAction: ref(false), error: ref(''), finished: ref(false),
    ready: vi.fn(async () => { daily.data.value = { current: item, set, done: false } }),
    retry: vi.fn(), act, back,
  }
  useDaily.mockReturnValue(daily)
})
describe('sentence composition in the daily journey', () => {
  it('uses only the bound ten-item session, preserves feedback and returns after completion', async () => {
    const view = mount(SentenceCompositionView, { global }); await flushPromises()
    expect(view.text()).toContain('0 / 10')
    await view.get('textarea').setValue('Estoy en casa')
    act.mockImplementationOnce(async () => { daily.data.value = savedFeedback() })
    await view.findAll('button').find(b => b.text() === 'sentence.checkBtn')!.trigger('click'); await flushPromises()
    expect(act).toHaveBeenCalledExactlyOnceWith('answer', { item_id: 42, user_input: 'Estoy en casa' })
    expect(view.text()).toContain('Estoy en casa')
    expect(view.text()).toContain('1 / 10')
    expect(back).not.toHaveBeenCalled()
    act.mockImplementationOnce(async () => { daily.data.value = { current: null, set: { ...set, status: 'completed', attempted: 10, stars: 10 }, done: true } })
    await view.findAll('button').find(b => b.text() === 'sentence.nextBtn')!.trigger('click'); await flushPromises()
    expect(act).toHaveBeenLastCalledWith('ack_feedback')
    expect(view.text()).toContain('sentence.doneTitle')
    expect(back).not.toHaveBeenCalled()
    await view.findAll('button').find(b => b.text() === 'dailyJourney.returnToRoute')!.trigger('click')
    expect(back).toHaveBeenCalledOnce()
    for (const call of [today, start, current, answer]) expect(call).not.toHaveBeenCalled()
    view.unmount()
  })
  it('restores the confirmed answer after a reload without grading it again', async () => {
    daily.ready.mockImplementation(async () => { daily.data.value = savedFeedback() })
    const view = mount(SentenceCompositionView, { global }); await flushPromises()
    expect(view.text()).toContain('Estoy en casa')
    expect(view.get('textarea').attributes('disabled')).toBeDefined()
    expect(act).not.toHaveBeenCalled()
    view.unmount()
  })
  it('restores a scoped draft and clears it only after confirmation', async () => {
    const key = 'daily-journey:sentence-draft:user-a:es_ru:j1:s1:42'
    sessionStorage.setItem(key, JSON.stringify('Estoy'))
    const view = mount(SentenceCompositionView, { global }); await flushPromises()
    expect((view.get('textarea').element as HTMLTextAreaElement).value).toBe('Estoy')
    await view.get('textarea').setValue('Estoy en casa')
    expect(JSON.parse(sessionStorage.getItem(key)!)).toBe('Estoy en casa')
    daily.data.value = { current: { ...item }, set: { ...set }, done: false }; await flushPromises()
    expect((view.get('textarea').element as HTMLTextAreaElement).value).toBe('Estoy en casa')
    daily.data.value = savedFeedback(); await flushPromises()
    expect(sessionStorage.getItem(key)).toBeNull()
    view.unmount()
  })
  it('does not fall back to a standalone set when the journey cannot load', async () => {
    daily.error.value = 'Retry the journey'
    daily.ready.mockRejectedValue(new Error('offline'))
    const view = mount(SentenceCompositionView, { global }); await flushPromises()
    expect(view.text()).toContain('Retry the journey')
    for (const call of [today, start, current, answer]) expect(call).not.toHaveBeenCalled()
    view.unmount()
  })
  it('keeps the ordinary training entrance working', async () => {
    daily.isJourney.value = false
    today.mockResolvedValue({ available: true, remaining: 20 })
    current.mockResolvedValue({ ...item, total: 20, attempted_count: 0 })
    const view = mount(SentenceCompositionView, { global }); await flushPromises()
    expect(start).toHaveBeenCalledWith('es_ru')
    expect(view.text()).toContain('0 / 20')
    expect(daily.ready).not.toHaveBeenCalled()
    view.unmount()
  })
})
