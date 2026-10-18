import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { reactive, ref } from 'vue'
import CityDailyRouteView from '../views/CityDailyRouteView.vue'
import DailyJourneyRouteView from '../views/DailyJourneyRouteView.vue'
const { current, start, get } = vi.hoisted(() => ({ current: vi.fn(), start: vi.fn(), get: vi.fn() }))
const routeState = reactive({ query: {} as Record<string, string> })
const course = ref('en_ru')
const authenticated = ref(true)
vi.mock('vue-router', () => ({ useRoute: () => routeState }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))
vi.mock('../composables/useCourse', () => ({ useCourse: () => ({ currentCourseCode: course }) }))
vi.mock('../composables/useAuth', () => ({ useAuth: () => ({ isAuthenticated: authenticated }) }))
vi.mock('../api/dailyJourneyClient', () => ({ dailyJourneyClient: { current, start, get } }))
vi.mock('../views/LegacyDailyRouteView.vue', () => ({ default: { template: '<div data-test="legacy">Legacy</div>' } }))
vi.mock('../views/DailyJourneyRouteView.vue', () => ({ default: { props: ['courseCode', 'initialJourney'], template: '<div data-test="journey">{{ courseCode }}</div>' } }))
beforeEach(() => {
  current.mockReset()
  start.mockReset()
  get.mockReset()
  routeState.query = {}
  course.value = 'en_ru'
  authenticated.value = true
})
describe('dailyJourney production readiness', () => {
  it('keeps the previous screen when the server reports disabled', async () => {
    current.mockResolvedValue({ enabled: false, journey: null })
    const view = mount(CityDailyRouteView)
    await flushPromises()
    expect(view.find('[data-test="legacy"]').exists()).toBe(true)
    expect(view.find('[data-test="journey"]').exists()).toBe(false)
    view.unmount()
  })
  it('shows an explicit journey error without silently opening standalone recommendations', async () => {
    routeState.query = { journey_id: 'saved-id' }
    current.mockResolvedValue({ enabled: false, journey: null })
    const view = mount(CityDailyRouteView)
    await flushPromises()
    expect(view.get('[role="alert"]').text()).toContain('dailyJourney.messages.disabled')
    expect(view.find('[data-test="legacy"]').exists()).toBe(false)
    view.unmount()
  })
  it('ignores an old readiness response after switching courses', async () => {
    let resolve!: (value: unknown) => void
    current.mockImplementationOnce(() => new Promise(done => { resolve = done }))
    current.mockResolvedValue({ enabled: false, journey: null })
    const view = mount(CityDailyRouteView)
    course.value = 'es_ru'
    await flushPromises()
    resolve({ enabled: true, journey: null })
    await flushPromises()
    expect(view.find('[data-test="journey"]').exists()).toBe(false)
    expect(current).toHaveBeenCalledWith('es_ru')
    view.unmount()
  })
})

it('creates today’s route immediately without an intermediate start screen', async () => {
  const journey = { id: 'today', courseCode: 'en_ru', steps: [] }
  current.mockResolvedValue({ enabled: true, journey: null })
  let resolve!: (value: unknown) => void
  start.mockImplementation(() => new Promise(done => { resolve = done }))
  const view = mount(CityDailyRouteView)
  expect(view.find('[data-test="legacy"]').exists()).toBe(false)
  await flushPromises()
  expect(start).toHaveBeenCalledExactlyOnceWith('en_ru')
  expect(view.find('[data-test="legacy"]').exists()).toBe(false)
  resolve(journey)
  await flushPromises()
  expect(view.getComponent(DailyJourneyRouteView).props('initialJourney')).toEqual(journey)
  expect(current).toHaveBeenCalledOnce()
  view.unmount()
})
it('reuses the already loaded route without a second read or start', async () => {
  current.mockResolvedValue({ enabled: true, journey: { id: 'today', courseCode: 'en_ru', steps: [] } })
  const view = mount(CityDailyRouteView)
  await flushPromises()
  expect(view.find('[data-test="journey"]').exists()).toBe(true)
  expect(current).toHaveBeenCalledOnce()
  expect(start).not.toHaveBeenCalled()
  expect(get).not.toHaveBeenCalled()
  view.unmount()
})
