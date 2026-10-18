import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ref } from 'vue'
import DashboardView from '../views/DashboardView.vue'
const { push } = vi.hoisted(() => ({ push: vi.fn() }))
const course = ref('es_ru')
let apply: (data: any, meta: any) => void
vi.mock('vue-router', () => ({ useRouter: () => ({ push }) }))
vi.mock('vue-i18n', async importOriginal => ({ ...await importOriginal<typeof import('vue-i18n')>(), useI18n: () => ({ t: (key: string) => key }) }))
vi.mock('../composables/useAuth', () => ({ useAuth: () => ({ isAuthenticated: ref(true) }) }))
vi.mock('../composables/useCourse', () => ({ useCourse: () => ({ currentCourseCode: course, currentCourse: ref(null) }) }))
vi.mock('../composables/useStats', () => ({ useStats: () => ({ streakDays: ref(0), ensureStatsLoaded: vi.fn(), refreshStats: vi.fn() }) }))
vi.mock('../composables/useLearningConfig', () => ({ useLearningConfig: () => ({ targetLangDisplay: ref('Spanish'), ensureLearningLoaded: async () => {} }) }))
vi.mock('../composables/useLocale', () => ({ useLocale: () => ({ currentLocale: ref('ru') }) }))
vi.mock('../composables/useGrammarContinueChapter', () => ({ useGrammarContinueChapter: () => ({ continueChapter: ref(null), applyContinueChapter: vi.fn() }) }))
vi.mock('../composables/useOfflineAutoDownload', () => ({ maybeRunOfflineAutoDownload: vi.fn() }))
vi.mock('../api/client', () => ({ apiClient: { loadTokens: vi.fn(), request: vi.fn() } }))
vi.mock('../composables/useCachedOverviewScreen', () => ({ useCachedOverviewScreen: (options: any) => { apply = options.applyPayload; return { loadingInitial: ref(false), refreshing: ref(false), load: vi.fn() } } }))
const stubs = Object.fromEntries(['LgLoader','LgIcon','LgLumi','LgProgressBar','LgLumiFact','LgStreakBadge','LgCourseSwitcher','LgWordLookup','LgActivityIcon','RouterLink'].map(name => [name, true]))
beforeEach(() => { course.value = 'es_ru'; push.mockClear() })
describe('dashboard journey entry', () => {
 it('switches by server flag and navigates to the saved journey in its course', async () => {
  const view = mount(DashboardView, { global: { stubs: { ...stubs, DailyJourneyCard: { props: ['journey'], template: '<button data-test="journey" @click="$emit(\'open\')">{{ journey?.id || "new" }}</button>' } } } })
  await flushPromises()
  apply({ daily_journey: { enabled: false, journey: null } }, { fromCache: false })
  await flushPromises()
  expect(view.find('.lg-path-wrap').exists()).toBe(true)
  const journey = { id: 'saved', courseCode: 'es_ru', status: 'active', steps: [] }
  apply({ daily_journey: { enabled: true, courseCode: 'es_ru', journey } }, { fromCache: false })
  await flushPromises()
  expect(view.find('.lg-path-wrap').exists()).toBe(false)
  await view.get('[data-test="journey"]').trigger('click')
  expect(push).toHaveBeenCalledWith({ name: 'CityDailyRoute', query: { course_code: 'es_ru', journey_id: 'saved' } })
  course.value = 'en_ru'
  await flushPromises()
  expect(view.find('[data-test="journey"]').exists()).toBe(false)
  view.unmount()
 })
 it('clears old availability when an overview section is omitted', async () => {
  const view = mount(DashboardView, { global: { stubs: { ...stubs, DailyJourneyCard: true } } })
  await flushPromises()
  apply({ daily_journey: { enabled: true, courseCode: 'es_ru', journey: null } }, { fromCache: false })
  await flushPromises()
  expect(view.find('daily-journey-card-stub').exists()).toBe(true)
  apply({ dashboard: {} }, { fromCache: false })
  await flushPromises()
  expect(view.find('daily-journey-card-stub').exists()).toBe(false)
  view.unmount()
 })
})
