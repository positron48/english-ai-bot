import { defineComponent, ref } from 'vue'
import { mount, flushPromises } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useCachedOverviewScreen } from './useCachedOverviewScreen'
import { emitAppDataEvent, registerAppDataRefreshHandler } from '../api/cacheInvalidation'
import { getCachedScreen, setCachedScreen, markScreensDirty } from '../api/appDataCache'

const { activeUser } = vi.hoisted(() => ({ activeUser: { value: 'user:1' } }))
vi.mock('../api/appDataCache', () => ({
  getCachedScreen: vi.fn().mockResolvedValue(null),
  entryNeedsRefresh: vi.fn().mockReturnValue(true),
  resolveUserScopeFromStorage: () => activeUser.value,
  getAppDataLocale: () => 'ru',
  setCachedScreen: vi.fn().mockResolvedValue({ fetchedAt: 'now' }),
  markScreensDirty: vi.fn().mockResolvedValue(undefined),
}))

function screen(fetcher = vi.fn().mockResolvedValue({})) {
  let cache!: ReturnType<typeof useCachedOverviewScreen>
  const wrapper = mount(defineComponent({
    setup() {
      cache = useCachedOverviewScreen({
        screenKey: 'dashboard', courseCode: ref('es_ru'), fetcher, applyPayload: vi.fn(),
      })
      return () => null
    },
  }))
  return { wrapper, cache, fetcher }
}

describe('page-scoped data loading', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    activeUser.value = 'user:1'
    registerAppDataRefreshHandler((screens, courseCode) => {
      window.dispatchEvent(new CustomEvent('linglow-app-data-refresh', { detail: { screens, courseCode } }))
    })
  })

  it('deduplicates simultaneous page loads from mount and watchers', async () => {
    const { wrapper, cache, fetcher } = screen()
    await Promise.all([cache.load(), cache.load(), cache.load()])
    expect(fetcher).toHaveBeenCalledTimes(1)
    wrapper.unmount()
  })

  it('invalidates after an answer without fetching an unmounted dashboard', async () => {
    vi.useFakeTimers()
    const { wrapper, cache, fetcher } = screen()
    await cache.load()
    wrapper.unmount()
    fetcher.mockClear()
    emitAppDataEvent('word-review-recorded', 'es_ru')
    await vi.advanceTimersByTimeAsync(401)
    expect(markScreensDirty).toHaveBeenCalled()
    expect(fetcher).not.toHaveBeenCalled()
    vi.useRealTimers()
  })

  it('refreshes a mounted affected page once and ignores other courses', async () => {
    vi.useFakeTimers()
    const { wrapper, fetcher } = screen()
    emitAppDataEvent('word-review-recorded', 'en_ru')
    await vi.advanceTimersByTimeAsync(401)
    expect(fetcher).not.toHaveBeenCalled()
    emitAppDataEvent('word-review-recorded', 'es_ru')
    emitAppDataEvent('word-review-recorded', 'es_ru')
    await vi.advanceTimersByTimeAsync(401)
    await flushPromises()
    expect(fetcher).toHaveBeenCalledTimes(1)
    wrapper.unmount()
    vi.useRealTimers()
  })
})

it('drops a late overview after changing account or course without caching it under the new scope', async () => {
 for (const change of ['account','course']) {
  activeUser.value='user:1';vi.mocked(getCachedScreen).mockResolvedValue(null);vi.mocked(setCachedScreen).mockClear()
  const code=ref('es_ru');const apply=vi.fn();let resolve!: (value:any)=>void
  const pending=new Promise(done=>{resolve=done})
  let cache!: ReturnType<typeof useCachedOverviewScreen>
  const wrapper=mount(defineComponent({setup(){cache=useCachedOverviewScreen({screenKey:'dashboard',courseCode:code,fetcher:()=>pending,applyPayload:apply});return()=>null}}))
  const loading=cache.load();await flushPromises()
  if(change==='account')activeUser.value='user:2';else code.value='en_ru'
  resolve({daily_journey:{enabled:true,journey:{id:'private-old-journey'}}});await loading
  expect(apply).not.toHaveBeenCalled();expect(setCachedScreen).not.toHaveBeenCalled()
  wrapper.unmount()
 }
})

it('does not hydrate a cached journey after its course changes while IndexedDB is pending', async () => {
 const code=ref('es_ru');const apply=vi.fn();let resolve!: (value:any)=>void
 vi.mocked(getCachedScreen).mockReturnValueOnce(new Promise(done=>{resolve=done}))
 let cache!: ReturnType<typeof useCachedOverviewScreen>
 const fetcher=vi.fn()
 const wrapper=mount(defineComponent({setup(){cache=useCachedOverviewScreen({screenKey:'dashboard',courseCode:code,fetcher,applyPayload:apply});return()=>null}}))
 const loading=cache.load();code.value='en_ru';resolve({payload:{daily_journey:{journey:{id:'old'}}}});await loading
 expect(apply).not.toHaveBeenCalled();expect(fetcher).not.toHaveBeenCalled()
 wrapper.unmount()
})
