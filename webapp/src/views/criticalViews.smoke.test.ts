import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import { createMemoryHistory, createRouter } from 'vue-router'
import { ref } from 'vue'

const mountErrors: unknown[] = []

vi.mock('../api/client', () => ({
  apiClient: {
    request: vi.fn().mockResolvedValue({ course_map: { districts: [] }, progress: null }),
    loadTokens: vi.fn(),
  },
}))

vi.mock('../api/courseClient', () => ({
  courseClient: {
    getCourseMap: vi.fn().mockResolvedValue({ districts: [] }),
    getProgress: vi.fn().mockResolvedValue({ mastery: { levels: [] } }),
  },
}))

vi.mock('../api/grammarClient', () => ({
  grammarClient: { getContinueChapter: vi.fn().mockResolvedValue(null) },
}))

vi.mock('../api/wordTrainingClient', () => ({
  wordTrainingClient: { getStats: vi.fn().mockResolvedValue({}) },
}))

vi.mock('../api/appDataCache', () => ({
  getCachedScreen: vi.fn().mockResolvedValue(null),
  setCachedScreen: vi.fn().mockResolvedValue({ fetchedAt: new Date().toISOString() }),
  entryNeedsRefresh: vi.fn().mockReturnValue(true),
  resolveUserScopeFromStorage: vi.fn().mockReturnValue('test-user'),
}))

vi.mock('../composables/useAuth', () => ({
  useAuth: () => ({ isAuthenticated: ref(true) }),
}))

vi.mock('../composables/useCourse', () => ({
  useCourse: () => ({
    currentCourse: ref({ code: 'en_ru', title: 'English' }),
    currentCourseCode: ref('en_ru'),
    ensureCourseLoaded: vi.fn().mockResolvedValue(undefined),
  }),
}))

vi.mock('../composables/useMe', () => ({
  useMe: () => ({
    ensureMe: vi.fn().mockResolvedValue(undefined),
    hasFeature: vi.fn().mockReturnValue(false),
  }),
}))

vi.mock('../composables/useLocale', () => ({
  useLocale: () => ({ currentLocale: ref('ru') }),
}))

vi.mock('../composables/useStats', () => ({
  useStats: () => ({
    streakDays: ref(0),
    ensureStatsLoaded: vi.fn(),
    refreshStats: vi.fn().mockResolvedValue(undefined),
  }),
}))

vi.mock('../composables/useLearningConfig', () => ({
  useLearningConfig: () => ({
    targetLangDisplay: ref('English'),
    ensureLearningLoaded: vi.fn().mockResolvedValue(undefined),
  }),
}))

vi.mock('../composables/useGrammarContinueChapter', () => ({
  useGrammarContinueChapter: () => ({
    continueChapter: ref(null),
    applyContinueChapter: vi.fn(),
  }),
}))

vi.mock('../composables/useAppDataRefresh', () => ({
  refreshAppData: vi.fn().mockResolvedValue(undefined),
  prefetchAppData: vi.fn().mockResolvedValue(undefined),
}))

vi.mock('../composables/useOfflineAutoDownload', () => ({
  maybeRunOfflineAutoDownload: vi.fn().mockResolvedValue(undefined),
}))

const stubNames = [
  'RouterLink',
  'LgPageHeader',
  'LgLumiFact',
  'LgActivityIcon',
  'LgLoader',
  'LgIcon',
  'LgLumi',
  'LgProgressBar',
  'LgStreakBadge',
  'LgCourseSwitcher',
  'LgWordLookup',
]

async function mountCriticalView(Component: unknown, routePath = '/') {
  const i18n = createI18n({ legacy: false, locale: 'ru', messages: { ru: {} } })
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', name: 'Dashboard', component: { template: '<div />' } },
      { path: '/city/district/:districtCode', name: 'CityDistrict', component: { template: '<div />' } },
    ],
  })
  await router.push(routePath)
  await router.isReady()

  mountErrors.length = 0
  const wrapper = mount(Component as any, {
    global: {
      plugins: [i18n, router],
      stubs: Object.fromEntries(stubNames.map((name) => [name, true])),
      config: {
        errorHandler(err) {
          mountErrors.push(err)
        },
      },
    },
  })
  await flushPromises()
  return wrapper
}

describe('critical views smoke', () => {
  beforeEach(() => {
    mountErrors.length = 0
    vi.clearAllMocks()
  })

  it('mounts DashboardView without runtime errors', async () => {
    const { default: DashboardView } = await import('./DashboardView.vue')
    const wrapper = await mountCriticalView(DashboardView, '/')
    expect(wrapper.exists()).toBe(true)
    expect(mountErrors).toEqual([])
  })

  it('mounts CityDistrictView without runtime errors', async () => {
    const { default: CityDistrictView } = await import('./CityDistrictView.vue')
    const wrapper = await mountCriticalView(CityDistrictView, '/city/district/a1_clear_plaza')
    expect(wrapper.exists()).toBe(true)
    expect(mountErrors).toEqual([])
  })
})
