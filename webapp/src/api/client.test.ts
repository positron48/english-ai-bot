import { beforeEach, afterEach, expect, it, vi } from 'vitest'
import { apiClient } from './client'

const originalFetch = globalThis.fetch

beforeEach(() => {
  localStorage.clear()
  apiClient.saveTokens('access', 'refresh')
  vi.stubGlobal('fetch', vi.fn())
})
afterEach(() => { globalThis.fetch = originalFetch })

it.each(['network', '503'])('keeps tokens after temporary refresh failure: %s', async kind => {
  vi.mocked(fetch).mockResolvedValueOnce(new Response('', { status: 401 }))
  if (kind === 'network') vi.mocked(fetch).mockRejectedValueOnce(new TypeError('Failed to fetch'))
  else vi.mocked(fetch).mockResolvedValueOnce(new Response('unavailable', { status: 503 }))
  // POST has no automatic retries, including when a refresh is unavailable.
  await expect(apiClient.request('/api/write', { method: 'POST' })).rejects.toThrow()
  expect(localStorage.getItem('refresh_token')).toBe('refresh')
})

it('clears credentials only when refresh is rejected', async () => {
  vi.mocked(fetch).mockResolvedValue(new Response('', { status: 401 }))
  await expect(apiClient.request('/api/write', { method: 'POST' })).rejects.toMatchObject({ status: 401 })
  expect(localStorage.getItem('refresh_token')).toBeNull()
})

it('does not replay a POST after losing its response', async () => {
  vi.mocked(fetch).mockRejectedValue(new TypeError('Failed to fetch'))
  const form = new FormData(); form.append('answer_text', 'apple')
  await expect(apiClient.requestFormData('/api/training/answer', form)).rejects.toMatchObject({ isNetworkError: true })
  expect(fetch).toHaveBeenCalledTimes(1)
})

it('shares one refresh between simultaneous expired requests', async () => {
  let finishRefresh!: (response: Response) => void
  const pending = new Promise<Response>(resolve => { finishRefresh = resolve })
  vi.mocked(fetch).mockImplementation(async url => {
    if (url === '/auth/refresh') return pending
    if (localStorage.getItem('access_token') === 'access') return new Response('', { status: 401 })
    return new Response('{}')
  })
  const first = apiClient.request('/api/one')
  const second = apiClient.request('/api/two')
  await vi.waitFor(() => expect(vi.mocked(fetch).mock.calls.filter(([url]) => url === '/auth/refresh')).toHaveLength(1))
  finishRefresh(new Response(JSON.stringify({ access_token: 'new-access', refresh_token: 'new-refresh' })))
  await Promise.all([first, second])
  expect(localStorage.getItem('refresh_token')).toBe('new-refresh')
})

it('does not resurrect a logged-out session from a late refresh response', async () => {
  let finish!: (response: Response) => void
  vi.mocked(fetch).mockImplementation(() => new Promise<Response>(resolve => { finish = resolve }))
  const pending = apiClient.refreshAccessToken()
  apiClient.clearTokens()
  finish(new Response(JSON.stringify({ access_token: 'new-access', refresh_token: 'new-refresh' })))
  await expect(pending).rejects.toMatchObject({ status: 409 })
  expect(localStorage.getItem('access_token')).toBeNull()
})

it('preserves explicit form headers and supports empty responses', async () => {
  vi.mocked(fetch).mockResolvedValue(new Response(null, { status: 204 }))
  const form = new FormData(); form.append('word', 'casa')
  await apiClient.requestFormData('/api/form', form)
  const options = vi.mocked(fetch).mock.calls[0][1]!
  expect(new Headers(options.headers).get('Content-Type')).toBe('application/x-www-form-urlencoded')
  expect(options.body).toBe('word=casa')
})

it('rejects a response from the previous account', async () => {
  const token = (id: number) => `header.${btoa(JSON.stringify({ user_id: id }))}.signature`
  apiClient.saveTokens(token(1), 'refresh-1')
  let finish!: (response: Response) => void
  vi.mocked(fetch).mockImplementation(() => new Promise<Response>(resolve => { finish = resolve }))
  const pending = apiClient.request('/api/me')
  apiClient.saveTokens(token(2), 'refresh-2')
  finish(new Response(JSON.stringify({ id: 1 })))
  await expect(pending).rejects.toMatchObject({ status: 409 })
  expect(localStorage.getItem('refresh_token')).toBe('refresh-2')
})
