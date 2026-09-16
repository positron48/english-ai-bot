import workerSource from '../public/sw.js?raw'
import { describe, expect, it, vi } from 'vitest'

function worker() {
  const handlers: Record<string, (event: any) => void> = {}
  const cache = { put: vi.fn().mockResolvedValue(undefined) }
  const fetch = vi.fn().mockResolvedValue({ ok: true, clone: () => ({}) })
  const caches = {
    open: vi.fn().mockResolvedValue(cache),
    match: vi.fn().mockResolvedValue(undefined),
    keys: vi.fn().mockResolvedValue([]),
    delete: vi.fn(),
  }
  const self = {
    addEventListener: (type: string, handler: any) => { handlers[type] = handler },
    skipWaiting: vi.fn(), clients: { claim: vi.fn() },
    location: { origin: 'https://example.com' },
  }
  new Function('self', 'fetch', 'caches', workerSource)(self, fetch, caches)
  return { handlers, fetch, caches, cache }
}

describe('on-demand service worker assets', () => {
  it('installs only the shell and does not warm the entire asset manifest', async () => {
    const { handlers, fetch } = worker()
    let completion!: Promise<void>
    handlers.install({ waitUntil: (p: Promise<void>) => { completion = p } })
    await completion
    expect(fetch.mock.calls.map(([url]) => url)).toEqual([
      '/app', '/app/', '/app/manifest.webmanifest', '/telegram-web-app.js', '/favicon.svg',
    ])
    fetch.mockClear()
    handlers.activate({ waitUntil: (p: Promise<void>) => { completion = p } })
    await completion
    expect(fetch).not.toHaveBeenCalled()
  })

  it('fetches and caches only the requested asset, then serves cache hits offline', async () => {
    const { handlers, fetch, caches, cache } = worker()
    const request = { method: 'GET', url: 'https://example.com/app/assets/current.png' }
    let response!: Promise<unknown>
    handlers.fetch({ request, respondWith: (p: Promise<unknown>) => { response = p } })
    await response
    expect(fetch).toHaveBeenCalledTimes(1)
    expect(fetch).toHaveBeenCalledWith(request)
    expect(cache.put).toHaveBeenCalledWith(request, expect.anything())
    fetch.mockClear()
    caches.match.mockResolvedValueOnce({ cached: true } as any)
    handlers.fetch({ request, respondWith: (p: Promise<unknown>) => { response = p } })
    expect(await response).toEqual({ cached: true })
    expect(fetch).not.toHaveBeenCalled()
  })
})
