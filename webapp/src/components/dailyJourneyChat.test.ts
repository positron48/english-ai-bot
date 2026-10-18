import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import PictureQuestChatView from '../views/PictureQuestChatView.vue'
import PlaceChatView from '../views/PlaceChatView.vue'
const { source, startPicture, startConversation, list, ensureCourse, ensureLearning } = vi.hoisted(() => ({ source: vi.fn(), startPicture: vi.fn(), startConversation: vi.fn(), list: vi.fn(), ensureCourse: vi.fn(), ensureLearning: vi.fn() }))
vi.mock('../composables/useDailyJourneyChat', () => ({ useDailyJourneyChat: source }))
vi.mock('../api/courseClient', () => ({ courseClient: { startPictureQuestSession: startPicture, startConversationSession: startConversation, listConversationScenarios: list } }))
vi.mock('../composables/useCourse', () => ({ useCourse: () => ({ currentCourse: ref({ target_language: 'en', native_language: 'ru' }), currentCourseCode: ref('en_ru'), ensureCourseLoaded: ensureCourse }) }))
vi.mock('../composables/useLearningConfig', () => ({ useLearningConfig: () => ({ learning: ref({ target_lang: 'en' }), ensureLearningLoaded: ensureLearning }) }))
vi.mock('vue-i18n', async original => ({ ...await original<typeof import('vue-i18n')>(), useI18n: () => ({ t: (key: string) => key }) }))
vi.mock('vue-router', () => ({ useRoute: () => ({ params: { questCode: 'picture-a', scenarioCode: 'cafe', districtCode: 'a0' }, query: {} }), useRouter: () => ({ push: vi.fn() }) }))
let chat: any, state: any
beforeEach(() => {
 vi.clearAllMocks()
 state = { session_id: 12, title: 'Bound chat', npc_name: 'Marta', image_url: '/test.jpg', messages: [{ role: 'assistant', content: 'Hola' }], tasks: [{ code: 'q1', title: 'Required', is_required: true, completed: false }], status: 'open', quest_passed: false, budget_exhausted: false, attempt_finished: false }
 chat = { daily: { isJourney: ref(true), context: ref({ courseCode: 'es_ru' }), error: ref(''), hasPendingAction: ref(false), data: ref(state), finished: ref(false), back: vi.fn() }, draft: ref('Saved draft'), session: vi.fn(async () => state), send: vi.fn(async () => ({ ...state, messages: [...state.messages, { role: 'user', content: 'Saved draft' }, { role: 'assistant', content: 'Muy bien' }] })), retry: vi.fn(), back: vi.fn() }
 source.mockReturnValue(chat)
})
const global = { stubs: { LgPageHeader: { template: '<header><slot name="right" /></header>' }, LgIcon: true, LgLumi: true, LgSpeechBubble: { template: '<div><slot /></div>' }, ClickableText: { props: ['text', 'courseCode'], template: '<span>{{ text }}</span>' }, LgButton: { template: '<button><slot /></button>' }, PictureHintSheet: true, VoiceMicButton: true } }
describe('daily chat on ordinary pages', () => {
 for (const [mode,component] of [['picture',PictureQuestChatView],['conversation',PlaceChatView]] as const) {
  it(`restores ${mode} messages and sends without a normal chat start or course switch`, async () => {
   const view = mount(component, { global });await flushPromises()
   expect(source).toHaveBeenCalledWith(mode)
   expect(view.text()).toContain('Hola')
   expect(view.find('.chat-reset').exists()).toBe(false)
   expect((view.get('input').element as HTMLInputElement).value).toBe('Saved draft')
   await view.get('.chat-send').trigger('click');await flushPromises()
   expect(chat.send).toHaveBeenCalledWith(12, 'Saved draft')
   expect(view.text()).toContain('Muy bien')
   expect(startPicture).not.toHaveBeenCalled();expect(startConversation).not.toHaveBeenCalled();expect(list).not.toHaveBeenCalled()
   expect(ensureCourse).not.toHaveBeenCalled();expect(ensureLearning).not.toHaveBeenCalled()
   view.unmount()
  })
 }
})
