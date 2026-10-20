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
 chat = { messages: ref(state.messages), daily: { isJourney: ref(true), context: ref({ courseCode: 'es_ru' }), error: ref(''), hasPendingAction: ref(false), data: ref(state), finished: ref(false), back: vi.fn() }, draft: ref('Saved draft'), session: vi.fn(async () => state), send: vi.fn(async () => ({ ...state, messages: [...state.messages, { role: 'user', content: 'Saved draft' }, { role: 'assistant', content: 'Muy bien' }] })), retry: vi.fn(), back: vi.fn() }
 chat.send.mockImplementation(async () => { const confirmed = { ...state, messages: [...state.messages, { role: 'user', content: 'Saved draft' }, { role: 'assistant', content: 'Muy bien' }] }; chat.messages.value = confirmed.messages; chat.daily.data.value = confirmed; return confirmed })
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

it.each([['picture', PictureQuestChatView], ['conversation', PlaceChatView]] as const)('%s keeps the composer and pending user bubble mounted until the AI reply', async (_mode, component) => {
 let finish!: () => void
 const pending = new Promise<void>(resolve => { finish = resolve })
 chat.send.mockImplementation(async () => {
  chat.draft.value = ''
  chat.messages.value = [...state.messages, { role: 'user', content: 'Saved draft' }]
  chat.daily.hasPendingAction.value = true
  await pending
  const confirmed = { ...state, messages: [...state.messages, { role: 'user', content: 'Saved draft' }, { role: 'assistant', content: 'Muy bien' }] }
  chat.messages.value = confirmed.messages; chat.daily.data.value = confirmed
  chat.daily.hasPendingAction.value = false
  return confirmed
 })
 const view = mount(component, { global })
 try {
  await flushPromises()
  const composer = view.get('input').element
  await view.get('.chat-send').trigger('click'); await flushPromises()
  expect(view.get('input').element).toBe(composer)
  expect((composer as HTMLInputElement).value).toBe('')
  expect(view.findAll('.chat-user-bubble').map(bubble => bubble.text())).toEqual(['Saved draft'])
  expect(view.find('.chat-typing').exists()).toBe(true)
  expect(view.find('.chat-loading').exists()).toBe(false)
  finish(); await flushPromises()
  expect(view.get('input').element).toBe(composer)
  expect(view.findAll('.chat-user-bubble').map(bubble => bubble.text())).toEqual(['Saved draft'])
  expect(view.text()).toContain('Muy bien')
  expect(view.find('.chat-typing').exists()).toBe(false)
 } finally { finish(); view.unmount() }
})
