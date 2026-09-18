import { currentUserScope } from './sessionScope'
const API_BASE = ''

function getCurrentLocale(): string {
  const stored = localStorage.getItem('locale')
  if (stored && ['ru', 'en', 'es'].includes(stored)) return stored
  const language = navigator.language?.toLowerCase().split('-')[0]
  return language === 'ru' || language === 'es' ? language : 'en'
}

interface AuthResponse {
  success: boolean
  message?: string
  access_token: string
  refresh_token: string
  token_type: string
  user_id?: number
}

export class ApiError extends Error {
  constructor(message: string, public status = 0, public code?: string, public isNetworkError = false) {
    super(message)
  }
}

async function responseError(response: Response): Promise<ApiError> {
  const text = await response.text()
  let message = `API error: ${response.status} ${text}`
  let code: string | undefined
  try {
    const body = JSON.parse(text)
    message = body.message || body.error || body.code || message
    code = body.code
  } catch { /* Plain-text HTTP errors are valid too. */ }
  return new ApiError(message, response.status, code)
}

type NetworkErrorCallback = (isRetrying: boolean, attempt: number, maxAttempts: number) => void

type ApiRequestOptions = Omit<RequestInit, 'body'> & { body?: BodyInit | Record<string, unknown> | null }

class ApiClient {
  private accessToken: string | null = null
  private refreshToken: string | null = null
  private refreshPromise: Promise<boolean> | null = null
  private networkErrorCallback: NetworkErrorCallback | null = null
  private networkSuccessCallback: (() => void) | null = null
  private maxRetries = 3
  private retryDelayMs = 1000

  constructor() { this.loadTokens() }
  setNetworkErrorCallback(callback: NetworkErrorCallback | null) { this.networkErrorCallback = callback }
  setNetworkSuccessCallback(callback: (() => void) | null) { this.networkSuccessCallback = callback }
  setMaxRetries(maxRetries: number) { this.maxRetries = Math.max(1, maxRetries) }

  loadTokens() {
    this.accessToken = localStorage.getItem('access_token')
    this.refreshToken = localStorage.getItem('refresh_token')
  }

  saveTokens(accessToken: string, refreshToken: string) {
    this.accessToken = accessToken
    this.refreshToken = refreshToken
    localStorage.setItem('access_token', accessToken)
    localStorage.setItem('refresh_token', refreshToken)
  }

  getAccessToken(): string | null { return this.accessToken }
  isAuthenticated(): boolean { return !!this.accessToken }

  clearTokens() {
    this.accessToken = null
    this.refreshToken = null
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
  }

  private async fetch(url: string, options: RequestInit): Promise<Response> {
    try {
      return await fetch(`${API_BASE}${url}`, options)
    } catch (error) {
      if (options.signal?.aborted || (error as Error).name === 'AbortError') throw error
      throw new ApiError((error as Error).message || 'Network error', 0, undefined, true)
    }
  }

  async refreshAccessToken(): Promise<boolean> {
    if (this.refreshPromise) return this.refreshPromise
    this.loadTokens()
    const token = this.refreshToken
    if (!token) return false
    this.refreshPromise = (async () => {
      const response = await this.fetch('/auth/refresh', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: token }),
      })
      // A logout/account switch while this request was pending wins over its result.
      if (localStorage.getItem('refresh_token') !== token) throw new ApiError('Session changed', 409)
      if (response.status === 401 || response.status === 403) {
        this.clearTokens()
        return false
      }
      if (!response.ok) throw await responseError(response)
      const data: AuthResponse = await response.json()
      if (localStorage.getItem('refresh_token') !== token) throw new ApiError('Session changed', 409)
      if (!data.access_token || !data.refresh_token) throw new ApiError('Invalid refresh response', 502)
      this.saveTokens(data.access_token, data.refresh_token)
      return true
    })()
    try { return await this.refreshPromise } finally { this.refreshPromise = null }
  }

  async request<T>(url: string, options: ApiRequestOptions = {}): Promise<T> {
    const userScope = currentUserScope()
    const checkSession = () => {
      if (currentUserScope() !== userScope) throw new ApiError('Session changed', 409)
    }
    const method = (options.method || 'GET').toUpperCase()
    // A lost response does not mean a write failed. Never automatically replay writes.
    const maxAttempts = method === 'GET' || method === 'HEAD' ? this.maxRetries : 1
    const headers = new Headers(options.headers)
    headers.set('Accept-Language', getCurrentLocale())
    let body = options.body
    if (body && typeof body === 'object' && !(body instanceof FormData) && !(body instanceof URLSearchParams) && !(body instanceof Blob) && !(body instanceof ArrayBuffer) && !ArrayBuffer.isView(body)) {
      body = JSON.stringify(body)
    }
    if (!(body instanceof FormData) && !headers.has('Content-Type')) headers.set('Content-Type', 'application/json')
    for (let attempt = 1; ; attempt++) {
      try {
        checkSession()
        this.loadTokens()
        const sentToken = this.accessToken
        if (sentToken) headers.set('Authorization', `Bearer ${sentToken}`)
        let response = await this.fetch(url, { ...options, body: body as BodyInit | null | undefined, headers })
        checkSession()
        if (response.status === 401 && this.refreshToken && !url.startsWith('/auth/')) {
          // Another request may already have refreshed the shared token.
          const current = localStorage.getItem('access_token')
          if (current === sentToken && !await this.refreshAccessToken()) throw new ApiError('Unauthorized', 401)
          this.loadTokens()
          if (!this.accessToken) throw new ApiError('Unauthorized', 401)
          headers.set('Authorization', `Bearer ${this.accessToken}`)
          response = await this.fetch(url, { ...options, body: body as BodyInit | null | undefined, headers })
        }
        checkSession()
        if (!response.ok) throw await responseError(response)
        const result = response.status === 204 ? undefined : await response.json()
        checkSession()
        this.networkSuccessCallback?.()
        return result as T
      } catch (error) {
        // Do not let a failed request from the previous account fall back to its offline data.
        if (error instanceof ApiError && error.isNetworkError) checkSession()
        const networkError = error instanceof ApiError && error.isNetworkError
        if (!networkError || attempt >= maxAttempts) {
          if (networkError) this.networkErrorCallback?.(false, attempt, maxAttempts)
          throw error
        }
        this.networkErrorCallback?.(true, attempt, maxAttempts)
        await new Promise(resolve => setTimeout(resolve, this.retryDelayMs * 2 ** (attempt - 1)))
      }
    }
  }

  requestFormData<T>(url: string, formData: FormData): Promise<T> {
    const params = new URLSearchParams()
    formData.forEach((value, key) => params.append(key, value.toString()))
    return this.request(url, {
      method: 'POST', headers: { 'Content-Type': 'application/x-www-form-urlencoded' }, body: params.toString(),
    })
  }

  authTelegram(initData: string): Promise<AuthResponse> {
    const form = new FormData()
    form.append('initData', initData)
    return this.requestFormData('/auth/telegram', form)
  }

  requestOTP(username: string): Promise<{ success: boolean; message: string; user_id: number }> {
    const form = new FormData()
    form.append('username', username)
    return this.requestFormData('/auth/request_otp', form)
  }

  async verifyOTP(userId: string, code: string): Promise<AuthResponse> {
    const form = new FormData()
    form.append('user_id', userId)
    form.append('code', code)
    const response = await this.requestFormData<AuthResponse>('/auth/otp', form)
    this.saveTokens(response.access_token, response.refresh_token)
    return response
  }
}

export const apiClient = new ApiClient()
