import { ref, computed } from 'vue'
import { apiClient } from '../api/client'

const isAuthenticated = ref(false)
const isAdmin = ref(false)

export function useAuth() {
  const checkAuth = () => {
    isAuthenticated.value = apiClient.isAuthenticated()
  }

  const checkAdmin = async () => {
    if (!isAuthenticated.value) {
      isAdmin.value = false
      return
    }

    try {
      await apiClient.request('/app/admin')
      isAdmin.value = true
    } catch (error: any) {
      if (error.message?.includes('403') || error.message?.includes('Forbidden')) {
        isAdmin.value = false
      } else {
        isAdmin.value = false
      }
    }
  }

  const login = (accessToken: string, refreshToken: string) => {
    apiClient.saveTokens(accessToken, refreshToken)
    isAuthenticated.value = true
    checkAdmin()
  }

  const logout = () => {
    apiClient.clearTokens()
    isAuthenticated.value = false
    isAdmin.value = false
  }

  const tryTelegramAuth = async (): Promise<boolean> => {
    try {
      // Check if Telegram WebApp is available
      const tg = (window as any).Telegram?.WebApp
      
      if (!tg) {
        console.warn('[Telegram Auth] Telegram WebApp object not found')
        return false
      }
      
      if (!tg.initData) {
        console.warn('[Telegram Auth] initData is not available')
        return false
      }
      
      console.log('[Telegram Auth] Attempting authentication with initData length:', tg.initData.length)
      
      // Try to authenticate
      const response = await apiClient.authTelegram(tg.initData)
      
      if (response && response.access_token && response.refresh_token) {
        console.log('[Telegram Auth] Authentication successful')
        login(response.access_token, response.refresh_token)
        return true
      } else {
        console.error('[Telegram Auth] Invalid response from server:', response)
        return false
      }
    } catch (error: any) {
      console.error('[Telegram Auth] Authentication failed:', error)
      
      // Log more details about the error
      if (error.message) {
        console.error('[Telegram Auth] Error message:', error.message)
      }
      if (error.response) {
        console.error('[Telegram Auth] Response status:', error.response?.status)
        console.error('[Telegram Auth] Response text:', error.response?.statusText)
      }
      
      // Re-throw the error so it can be caught and displayed in LoginView
      throw error
    }
  }

  checkAuth()
  if (isAuthenticated.value) {
    checkAdmin()
  }

  return {
    isAuthenticated: computed(() => isAuthenticated.value),
    isAdmin: computed(() => isAdmin.value),
    login,
    logout,
    tryTelegramAuth,
    checkAdmin
  }
}

