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
      // According to Telegram Mini Apps documentation, initData should be in Telegram.WebApp.initData
      const tg = (window as any).Telegram?.WebApp
      
      if (!tg) {
        console.warn('[Telegram Auth] Telegram.WebApp object not found')
        return false
      }
      
      // Get initData directly from Telegram.WebApp.initData (as per Telegram docs)
      const initData = tg.initData
      
      if (!initData || initData.trim() === '') {
        console.warn('[Telegram Auth] Telegram.WebApp.initData is empty or not available')
        console.log('[Telegram Auth] Telegram.WebApp object:', {
          version: tg.version,
          platform: tg.platform,
          initDataUnsafe: tg.initDataUnsafe ? 'available' : 'not available'
        })
        return false
      }
      
      console.log('[Telegram Auth] Got initData from Telegram.WebApp.initData')
      console.log('[Telegram Auth] initData length:', initData.length)
      console.log('[Telegram Auth] initData preview (first 100 chars):', initData.substring(0, 100))
      console.log('[Telegram Auth] Full initData:', initData)
      
      // Try to authenticate - send initData as-is (it's already URL-encoded by Telegram)
      console.log('[Telegram Auth] Calling apiClient.authTelegram...')
      const response = await apiClient.authTelegram(initData)
      console.log('[Telegram Auth] Received response:', response)
      
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
      if (error.status) {
        console.error('[Telegram Auth] Response status:', error.status)
      }
      if (error.response) {
        console.error('[Telegram Auth] Response status:', error.response?.status)
        console.error('[Telegram Auth] Response text:', error.response?.statusText)
      }
      
      // Check for network errors
      if (error.message?.includes('Failed to fetch') || 
          error.message?.includes('NetworkError') ||
          error.name === 'TypeError' ||
          error.status === 0) {
        console.error('[Telegram Auth] Network error - request did not reach server')
        console.error('[Telegram Auth] This could be a CORS issue or server is not accessible')
        const networkError = new Error('Ошибка сети: запрос не дошел до сервера. Проверьте подключение и настройки CORS.')
        ;(networkError as any).status = 0
        ;(networkError as any).isNetworkError = true
        throw networkError
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

