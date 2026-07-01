import { ref, computed } from 'vue'
import { apiClient } from '../api/client'
import { clearMeCache } from './useMe'

const isAuthenticated = ref(false)
const isAdmin = ref(false)
const categories = ref<number[]>([])
const permissions = ref<string[]>([])
const permissionsLoading = ref(false)
let permissionsLoadPromise: Promise<void> | null = null

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

// Get categories from JWT token (new format)
function getCategoriesFromToken(token: string | null): number[] {
  if (!token) {
    return []
  }
  
  const claims = decodeJWT(token)
  if (!claims || !claims.role) {
    return []
  }
  
  // Support new format: role = { categories: [...] }
  if (typeof claims.role === 'object' && claims.role.categories) {
    return claims.role.categories || []
  }
  
  // Legacy format: role = "admin" | "user" (return empty for backward compat)
  return []
}

// Load permissions from API
async function loadPermissions(): Promise<void> {
  if (!isAuthenticated.value) {
    return
  }

  if (permissionsLoadPromise) {
    await permissionsLoadPromise
    return
  }
  
  permissionsLoading.value = true
  permissionsLoadPromise = (async () => {
    try {
      const data: { categories: number[]; permissions: string[] } = await apiClient.request('/api/access/me')
      categories.value = data.categories || []
      permissions.value = data.permissions || []
      
      // Update isAdmin based on permissions
      // User is admin if they have any admin permission
      isAdmin.value = permissions.value.includes('full_access') || 
                     permissions.value.includes('words.read_all') ||
                     permissions.value.includes('words.edit_all') ||
                     permissions.value.includes('word_sets.read') ||
                     permissions.value.includes('word_sets.edit') ||
                     permissions.value.includes('users.read_all') ||
                     permissions.value.includes('stats.read') ||
                     categories.value.length > 0 // Also check categories as fallback
    } catch (error) {
      console.error('Failed to load permissions:', error)
      // Don't reset isAdmin on error - keep optimistic value if categories exist
      // Only reset if we're sure user has no access
      if (categories.value.length === 0) {
        isAdmin.value = false
      }
    } finally {
      permissionsLoading.value = false
      permissionsLoadPromise = null
    }
  })()

  await permissionsLoadPromise
}

export function useAuth() {
  const checkAuth = async () => {
    // Reload tokens from localStorage to ensure they're current
    // This is important when accessing the app directly via URL
    apiClient.loadTokens()
    isAuthenticated.value = apiClient.isAuthenticated()
    
    // Update categories from JWT token
    if (isAuthenticated.value) {
      const token = localStorage.getItem('access_token')
      categories.value = getCategoriesFromToken(token)
      
      // Temporarily set isAdmin to true if user has categories (optimistic)
      // This ensures admin menu is visible while permissions are loading
      // It will be updated correctly after loadPermissions() completes
      if (categories.value.length > 0) {
        isAdmin.value = true
      }
      
      // Offline navigation must not wait for permission API retries.
      if (typeof navigator === 'undefined' || navigator.onLine !== false) {
        await loadPermissions()
      }
    } else {
      categories.value = []
      permissions.value = []
      isAdmin.value = false
    }
  }

  const login = async (accessToken: string, refreshToken: string) => {
    apiClient.saveTokens(accessToken, refreshToken)
    isAuthenticated.value = true
    
    // Extract categories from JWT token
    categories.value = getCategoriesFromToken(accessToken)
    
    // Temporarily set isAdmin to true if user has categories (optimistic)
    // This ensures admin menu is visible while permissions are loading
    if (categories.value.length > 0) {
      isAdmin.value = true
    }
    
    // Load permissions from API
    await loadPermissions()
  }

  const logout = () => {
    apiClient.clearTokens()
    clearMeCache()
    isAuthenticated.value = false
    isAdmin.value = false
    categories.value = []
    permissions.value = []
  }
  
  // Check if user has a specific permission
  const can = (permission: string): boolean => {
    if (!isAuthenticated.value) {
      return false
    }
    
    // full_access grants everything
    if (permissions.value.includes('full_access')) {
      return true
    }
    
    return permissions.value.includes(permission)
  }
  
  // Check if user has any admin access (any admin permission)
  const hasAnyAdminAccess = (): boolean => {
    return isAdmin.value
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
  // Permissions will be loaded asynchronously

  return {
    isAuthenticated: computed(() => isAuthenticated.value),
    isAdmin: computed(() => isAdmin.value),
    categories: computed(() => categories.value),
    permissions: computed(() => permissions.value),
    permissionsLoading: computed(() => permissionsLoading.value),
    can,
    hasAnyAdminAccess,
    login,
    logout,
    tryTelegramAuth,
    checkAuth,
    loadPermissions
  }
}
