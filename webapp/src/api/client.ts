const API_BASE = ''

interface AuthResponse {
  success: boolean
  message?: string
  access_token: string
  refresh_token: string
  token_type: string
  user_id?: number
}

interface RefreshResponse {
  success: boolean
  message?: string
  access_token: string
  refresh_token: string
  token_type: string
}

// Network error callback type
type NetworkErrorCallback = (isRetrying: boolean, attempt: number, maxAttempts: number) => void
type NetworkSuccessCallback = () => void

class ApiClient {
  private accessToken: string | null = null
  private refreshToken: string | null = null
  private networkErrorCallback: NetworkErrorCallback | null = null
  private networkSuccessCallback: NetworkSuccessCallback | null = null
  private maxRetries: number = 3
  private retryDelayMs: number = 1000 // Initial delay

  constructor() {
    this.loadTokens()
  }

  setNetworkErrorCallback(callback: NetworkErrorCallback | null) {
    this.networkErrorCallback = callback
  }

  setNetworkSuccessCallback(callback: NetworkSuccessCallback | null) {
    this.networkSuccessCallback = callback
  }

  setMaxRetries(maxRetries: number) {
    this.maxRetries = maxRetries
  }

  private isNetworkError(error: any): boolean {
    // Check for network errors
    if (error.name === 'TypeError' && 
        (error.message?.includes('Failed to fetch') || 
         error.message?.includes('NetworkError') ||
         error.message?.includes('network'))) {
      return true
    }
    
    // Check for fetch errors (no response)
    if (error.message?.includes('Failed to fetch') || 
        error.message?.includes('NetworkError') ||
        error.isNetworkError) {
      return true
    }
    
    return false
  }

  private async sleep(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms))
  }

  private async retryWithBackoff<T>(
    fn: () => Promise<T>,
    url: string,
    attempt: number = 1
  ): Promise<T> {
    try {
      const result = await fn()
      // Notify about successful request (hide error notification)
      // Call on any successful attempt to hide notification after retry
      if (this.networkSuccessCallback) {
        this.networkSuccessCallback()
      }
      return result
    } catch (error: any) {
      const isNetworkErr = this.isNetworkError(error)
      
      // Only retry network errors, not HTTP errors (4xx, 5xx)
      if (!isNetworkErr || attempt >= this.maxRetries) {
        if (isNetworkErr && this.networkErrorCallback) {
          this.networkErrorCallback(false, attempt, this.maxRetries)
        }
        throw error
      }

      // Notify about retry
      if (this.networkErrorCallback) {
        this.networkErrorCallback(true, attempt, this.maxRetries)
      }

      // Exponential backoff: delay = initialDelay * 2^(attempt-1)
      const delay = this.retryDelayMs * Math.pow(2, attempt - 1)
      await this.sleep(delay)

      return this.retryWithBackoff(fn, url, attempt + 1)
    }
  }

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

  getAccessToken(): string | null {
    return this.accessToken
  }

  isAuthenticated(): boolean {
    return !!this.accessToken
  }

  clearTokens() {
    this.accessToken = null
    this.refreshToken = null
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
  }

  async refreshAccessToken(): Promise<boolean> {
    if (!this.refreshToken) {
      return false
    }

    try {
      const response = await fetch(`${API_BASE}/auth/refresh`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          refresh_token: this.refreshToken
        })
      })

      if (!response.ok) {
        return false
      }

      const data: RefreshResponse = await response.json()
      this.saveTokens(data.access_token, data.refresh_token)
      return true
    } catch (error) {
      console.error('Failed to refresh token:', error)
      return false
    }
  }

  async request<T>(url: string, options: RequestInit = {}): Promise<T> {
    return this.retryWithBackoff(async () => {
      // CRITICAL: Load tokens from localStorage RIGHT BEFORE creating headers
      // This ensures tokens are always fresh, especially on direct URL access
      this.loadTokens()
      
      const headers: Record<string, string> = {
        'Content-Type': 'application/json',
        ...(options.headers as Record<string, string> || {}),
      }

      // Get token directly from localStorage to ensure it's current
      // This is critical for direct URL access when components mount before tokens are loaded
      const token = localStorage.getItem('access_token')
      if (token) {
        headers['Authorization'] = `Bearer ${token}`
        // Also update instance variable for consistency
        this.accessToken = token
      } else if (this.accessToken) {
        // Fallback to instance variable if localStorage is empty
        headers['Authorization'] = `Bearer ${this.accessToken}`
      } else {
        // Debug: log when no token is available
        console.warn('[ApiClient] No access token available for request:', url)
      }

      let response: Response
      try {
        response = await fetch(`${API_BASE}${url}`, {
          ...options,
          headers,
        })
      } catch (fetchError: any) {
        // Wrap fetch errors as network errors
        const networkError = new Error(fetchError.message || 'Network error')
        ;(networkError as any).name = fetchError.name || 'TypeError'
        ;(networkError as any).isNetworkError = true
        throw networkError
      }

      if (response.status === 401 && this.refreshToken) {
        const refreshed = await this.refreshAccessToken()
        if (refreshed) {
          // Reload token after refresh to ensure we have the latest one
          this.loadTokens()
          const refreshedToken = localStorage.getItem('access_token') || this.accessToken
          if (refreshedToken) {
            headers['Authorization'] = `Bearer ${refreshedToken}`
          }
          try {
            response = await fetch(`${API_BASE}${url}`, {
              ...options,
              headers,
            })
          } catch (fetchError: any) {
            const networkError = new Error(fetchError.message || 'Network error')
            ;(networkError as any).name = fetchError.name || 'TypeError'
            ;(networkError as any).isNetworkError = true
            throw networkError
          }
        } else {
          this.clearTokens()
          throw new Error('Unauthorized')
        }
      }

      if (!response.ok) {
        const errorText = await response.text()
        let errorMessage = `API error: ${response.status} ${errorText}`
        
        // Try to parse JSON error if possible
        try {
          const errorJson = JSON.parse(errorText)
          if (errorJson.message) {
            errorMessage = errorJson.message
          } else if (errorJson.error) {
            errorMessage = errorJson.error
          }
        } catch {
          // Not JSON, use text as-is
        }
        
        const error = new Error(errorMessage)
        ;(error as any).status = response.status
        ;(error as any).response = response
        throw error
      }

      return response.json()
    }, url)
  }

  async requestFormData<T>(url: string, formData: FormData): Promise<T> {
    return this.retryWithBackoff(async () => {
      // CRITICAL: Load tokens from localStorage RIGHT BEFORE creating headers
      // This ensures tokens are always fresh, especially on direct URL access
      this.loadTokens()
      
      // Convert FormData to URLSearchParams for application/x-www-form-urlencoded
      const params = new URLSearchParams()
      formData.forEach((value, key) => {
        params.append(key, value.toString())
      })

      const headers: Record<string, string> = {
        'Content-Type': 'application/x-www-form-urlencoded',
      }

      // Get token directly from localStorage to ensure it's current
      // This is critical for direct URL access when components mount before tokens are loaded
      const token = localStorage.getItem('access_token')
      if (token) {
        headers['Authorization'] = `Bearer ${token}`
        // Also update instance variable for consistency
        this.accessToken = token
      } else if (this.accessToken) {
        // Fallback to instance variable if localStorage is empty
        headers['Authorization'] = `Bearer ${this.accessToken}`
      } else {
        // Debug: log when no token is available
        console.warn('[ApiClient] No access token available for requestFormData:', url)
      }

      const fullUrl = `${API_BASE}${url}`

      let response: Response
      try {
        response = await fetch(fullUrl, {
          method: 'POST',
          headers,
          body: params.toString(),
        })
      } catch (fetchError: any) {
        // Wrap fetch errors as network errors
        const networkError = new Error(fetchError.message || 'Network error')
        ;(networkError as any).name = fetchError.name || 'TypeError'
        ;(networkError as any).isNetworkError = true
        throw networkError
      }

      if (response.status === 401 && this.refreshToken) {
        const refreshed = await this.refreshAccessToken()
        if (refreshed) {
          // Reload token after refresh to ensure we have the latest one
          this.loadTokens()
          const refreshedToken = localStorage.getItem('access_token') || this.accessToken
          if (refreshedToken) {
            headers['Authorization'] = `Bearer ${refreshedToken}`
          }
          try {
            response = await fetch(fullUrl, {
              method: 'POST',
              headers,
              body: params.toString(),
            })
          } catch (fetchError: any) {
            const networkError = new Error(fetchError.message || 'Network error')
            ;(networkError as any).name = fetchError.name || 'TypeError'
            ;(networkError as any).isNetworkError = true
            throw networkError
          }
        } else {
          this.clearTokens()
          throw new Error('Unauthorized')
        }
      }

      if (!response.ok) {
        const errorText = await response.text()
        let errorMessage = `API error: ${response.status} ${errorText}`
        
        // Try to parse JSON error if possible
        try {
          const errorJson = JSON.parse(errorText)
          if (errorJson.message) {
            errorMessage = errorJson.message
          } else if (errorJson.error) {
            errorMessage = errorJson.error
          }
        } catch {
          // Not JSON, use text as-is
        }
        
        const error = new Error(errorMessage)
        ;(error as any).status = response.status
        ;(error as any).response = response
        throw error
      }

      return response.json()
    }, url)
  }

  async authTelegram(initData: string): Promise<AuthResponse> {
    const formData = new FormData()
    formData.append('initData', initData)
    
    try {
      const response = await this.requestFormData<AuthResponse>('/auth/telegram', formData)
      return response
    } catch (error: any) {
      // Check for network/CORS errors
      if (error.message?.includes('Failed to fetch') || 
          error.message?.includes('NetworkError') ||
          error.name === 'TypeError') {
        const networkError = new Error('Ошибка сети: не удалось отправить запрос. Возможно, проблема с CORS или сервер недоступен.')
        ;(networkError as any).status = 0
        ;(networkError as any).isNetworkError = true
        throw networkError
      }
      
      throw error
    }
  }

  async requestOTP(username: string): Promise<{ success: boolean; message: string; user_id: number }> {
    const formData = new FormData()
    formData.append('username', username)
    return this.requestFormData('/auth/request_otp', formData)
  }

  async verifyOTP(userId: string, code: string): Promise<AuthResponse> {
    const formData = new FormData()
    formData.append('user_id', userId)
    formData.append('code', code)
    const response = await this.requestFormData<AuthResponse>('/auth/otp', formData)
    this.saveTokens(response.access_token, response.refresh_token)
    return response
  }
}

export const apiClient = new ApiClient()

