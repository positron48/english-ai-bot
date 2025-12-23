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
    const tg = (window as any).Telegram?.WebApp
    if (!tg || !tg.initData) {
      return false
    }

    try {
      const response = await apiClient.authTelegram(tg.initData)
      login(response.access_token, response.refresh_token)
      return true
    } catch (error) {
      console.error('Telegram auth failed:', error)
      return false
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

