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
      let initData: string | null = null
      
      // Method 1: Try to get initData from Telegram.WebApp object
      const tg = (window as any).Telegram?.WebApp
      if (tg && tg.initData) {
        initData = tg.initData
        console.log('[Telegram Auth] Got initData from Telegram.WebApp.initData, length:', initData ? initData.length : 0)
      }
      
      // Method 2: Try to get initData from URL parameter tgWebAppData
      if (!initData) {
        const urlParams = new URLSearchParams(window.location.search)
        const tgWebAppData = urlParams.get('tgWebAppData')
        if (tgWebAppData) {
          initData = decodeURIComponent(tgWebAppData)
          console.log('[Telegram Auth] Got initData from URL parameter tgWebAppData, length:', initData ? initData.length : 0)
        }
      }
      
      // Method 3: Try to get from hash (if Telegram puts it there)
      if (!initData && window.location.hash) {
        const hashParams = new URLSearchParams(window.location.hash.substring(1))
        const tgWebAppData = hashParams.get('tgWebAppData')
        if (tgWebAppData) {
          initData = decodeURIComponent(tgWebAppData)
          console.log('[Telegram Auth] Got initData from hash parameter tgWebAppData, length:', initData ? initData.length : 0)
        }
      }
      
      // Method 4: Try to get from window storage (stored during URL cleanup)
      if (!initData && (window as any).__tgWebAppData) {
        initData = (window as any).__tgWebAppData
        console.log('[Telegram Auth] Got initData from window storage, length:', initData ? initData.length : 0)
      }
      
      if (!initData) {
        console.warn('[Telegram Auth] initData is not available from any source')
        return false
      }
      
      console.log('[Telegram Auth] Attempting authentication with initData length:', initData.length)
      
      // Try to authenticate
      const response = await apiClient.authTelegram(initData)
      
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

