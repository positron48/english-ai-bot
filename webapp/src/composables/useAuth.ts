import { ref, computed } from 'vue'
import { apiClient } from '../api/client'

const isAuthenticated = ref(false)
const isAdmin = ref(false)

export function useAuth() {
  const checkAuth = () => {
    // Reload tokens from localStorage to ensure they're current
    // This is important when accessing the app directly via URL
    apiClient.loadTokens()
    isAuthenticated.value = apiClient.isAuthenticated()
  }

  const checkAdmin = async () => {
    if (!isAuthenticated.value) {
      isAdmin.value = false
      return
    }

    try {
      await apiClient.request('/api/admin')
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
      const tg = (window as any).Telegram?.WebApp
      
      if (!tg) {
        return false
      }
      
      const initData = tg.initData
      
      if (!initData || initData.trim() === '') {
        return false
      }
      
      const response = await apiClient.authTelegram(initData)
      
      if (response && response.access_token && response.refresh_token) {
        login(response.access_token, response.refresh_token)
        return true
      }
      
      return false
    } catch (error: any) {
      // Check for network errors
      if (error.message?.includes('Failed to fetch') || 
          error.message?.includes('NetworkError') ||
          error.name === 'TypeError' ||
          error.status === 0) {
        const networkError = new Error('Ошибка сети: запрос не дошел до сервера. Проверьте подключение и настройки CORS.')
        ;(networkError as any).status = 0
        ;(networkError as any).isNetworkError = true
        throw networkError
      }
      
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
    checkAdmin,
    checkAuth
  }
}

