import { ref, computed } from 'vue'
import { apiClient } from '../api/client'

const isAuthenticated = ref(false)
const isAdmin = ref(false)

// Decode JWT token to extract claims
function decodeJWT(token: string): any | null {
  try {
    const parts = token.split('.')
    if (parts.length !== 3) {
      return null
    }
    
    // Decode base64url payload (second part)
    const payload = parts[1]
    // Replace base64url characters with base64
    const base64 = payload.replace(/-/g, '+').replace(/_/g, '/')
    // Add padding if needed
    const padded = base64 + '='.repeat((4 - base64.length % 4) % 4)
    const decoded = atob(padded)
    return JSON.parse(decoded)
  } catch (error) {
    console.error('Failed to decode JWT token:', error)
    return null
  }
}

// Get role from JWT token
function getRoleFromToken(token: string | null): string {
  if (!token) {
    return 'user'
  }
  
  const claims = decodeJWT(token)
  if (!claims || !claims.role) {
    return 'user' // Default to user if role is not set
  }
  
  return claims.role
}

export function useAuth() {
  const checkAuth = () => {
    // Reload tokens from localStorage to ensure they're current
    // This is important when accessing the app directly via URL
    apiClient.loadTokens()
    isAuthenticated.value = apiClient.isAuthenticated()
    
    // Update admin status from JWT token
    if (isAuthenticated.value) {
      const token = localStorage.getItem('access_token')
      const role = getRoleFromToken(token)
      isAdmin.value = role === 'admin'
    } else {
      isAdmin.value = false
    }
  }

  const login = (accessToken: string, refreshToken: string) => {
    apiClient.saveTokens(accessToken, refreshToken)
    isAuthenticated.value = true
    
    // Extract role from JWT token
    const role = getRoleFromToken(accessToken)
    isAdmin.value = role === 'admin'
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
  // Don't automatically check admin status - it's read from JWT token

  return {
    isAuthenticated: computed(() => isAuthenticated.value),
    isAdmin: computed(() => isAdmin.value),
    login,
    logout,
    tryTelegramAuth,
    checkAuth
  }
}

