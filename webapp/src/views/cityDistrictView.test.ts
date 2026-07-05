import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import { createMemoryHistory, createRouter } from 'vue-router'
import { ref } from 'vue'
import CityDistrictView from './CityDistrictView.vue'

const hasFeature = vi.fn<(feature: string) => boolean>()

vi.mock('../api/courseClient', () => ({
  courseClient: {
    getCourseMap: vi.fn().mockResolvedValue({
      districts: [{ code: 'a1_clear_plaza', title: 'A1 Plaza', level_code: 'A1', description: 'Plaza' }],
    }),
    getProgress: vi.fn().mockResolvedValue({
      mastery: {
        levels: [{
          level_code: 'A1',
          unlocked: true,
          metrics: {
            grammar: { percent: 42, included: true },
            words: { percent: 18, included: true },
            reading: { percent: 7, included: true },
            conversation: { percent: 55, included: true },
            picture: { percent: 12, included: true },
          },
        }],
      },
    }),
  },
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
    currentCourseCode: ref('en_ru'),
    ensureCourseLoaded: vi.fn().mockResolvedValue(undefined),
  }),
}))

vi.mock('../composables/useMe', () => ({
  useMe: () => ({
    ensureMe: vi.fn().mockResolvedValue(undefined),
    hasFeature,
  }),
}))

vi.mock('../composables/useLocale', () => ({
  useLocale: () => ({ currentLocale: ref('ru') }),
}))

async function mountDistrict(code = 'a1_clear_plaza') {
  const i18n = createI18n({
    legacy: false,
    locale: 'ru',
    messages: {
      ru: {
        city: {
          district: 'District',
          areaGrammar: 'Grammar',
          areaMetaGrammar: '{pct}%',
          areaWords: 'Words',
          areaMetaWords: '{pct}%',
          areaReading: 'Reading',
          areaMetaReading: '{pct}%',
          areaChat: 'Chat',
          areaMetaChat: '{pct}%',
          areaPicture: 'Picture',
          areaMetaPicturePct: '{pct}%',
          ctaContinue: 'Continue',
          ctaRead: 'Read',
          ctaPractice: 'Practice',
          ctaOpen: 'Open',
        },
      },
    },
  })
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [{ path: '/city/district/:districtCode', name: 'CityDistrict', component: CityDistrictView }],
  })
  await router.push(`/city/district/${code}`)
  await router.isReady()
  const wrapper = mount(CityDistrictView, {
    global: {
      plugins: [i18n, router],
      stubs: {
        LgPageHeader: true,
        LgLumiFact: true,
        LgActivityIcon: true,
        LgLoader: true,
      },
    },
  })
  await flushPromises()
  return wrapper
}

describe('CityDistrictView', () => {
  beforeEach(() => {
    hasFeature.mockImplementation((feature) => feature === 'conversation' || feature === 'picture_description')
  })

  it('loads full course map and shows non-zero level metrics plus pro areas', async () => {
    const wrapper = await mountDistrict()
    const rows = wrapper.findAll('.dst-area-row')
    expect(rows.length).toBe(5)
    expect(wrapper.text()).toContain('42%')
    expect(wrapper.text()).toContain('55%')
  })
})
