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

class ApiClient {
  private accessToken: string | null = null
  private refreshToken: string | null = null

  constructor() {
    this.loadTokens()
  }

  private loadTokens() {
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
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(options.headers as Record<string, string> || {}),
    }

    if (this.accessToken) {
      headers['Authorization'] = `Bearer ${this.accessToken}`
    }

    let response = await fetch(`${API_BASE}${url}`, {
      ...options,
      headers,
    })

    if (response.status === 401 && this.refreshToken) {
      const refreshed = await this.refreshAccessToken()
      if (refreshed) {
        headers['Authorization'] = `Bearer ${this.accessToken}`
        response = await fetch(`${API_BASE}${url}`, {
          ...options,
          headers,
        })
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
  }

  async requestFormData<T>(url: string, formData: FormData): Promise<T> {
    // Convert FormData to URLSearchParams for application/x-www-form-urlencoded
    const params = new URLSearchParams()
    formData.forEach((value, key) => {
      params.append(key, value.toString())
    })

    const headers: Record<string, string> = {
      'Content-Type': 'application/x-www-form-urlencoded',
    }

    if (this.accessToken) {
      headers['Authorization'] = `Bearer ${this.accessToken}`
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
      throw fetchError
    }

      if (response.status === 401 && this.refreshToken) {
        const refreshed = await this.refreshAccessToken()
        if (refreshed) {
          headers['Authorization'] = `Bearer ${this.accessToken}`
          response = await fetch(fullUrl, {
            method: 'POST',
            headers,
            body: params.toString(),
          })
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

