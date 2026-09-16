import { defineComponent, ref } from 'vue'
import { mount, flushPromises } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useCachedOverviewScreen } from './useCachedOverviewScreen'
import { emitAppDataEvent, registerAppDataRefreshHandler } from '../api/cacheInvalidation'
import { markScreensDirty } from '../api/appDataCache'

vi.mock('../api/appDataCache', () => ({
  getCachedScreen: vi.fn().mockResolvedValue(null),
  entryNeedsRefresh: vi.fn().mockReturnValue(true),
  resolveUserScopeFromStorage: () => 'user:1',
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
