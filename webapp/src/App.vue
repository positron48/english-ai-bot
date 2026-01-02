<template>
  <div id="app">
    <!-- Auth Error Message -->
    <div v-if="authError" class="auth-error-banner">
      <div class="auth-error-content">
        <strong>⚠️ Ошибка авторизации:</strong> {{ authError }}
        <button @click="dismissAuthError" class="auth-error-close">×</button>
      </div>
    </div>
    
    <!-- Debug info in development -->
    <div v-if="showDebugInfo" style="position: fixed; bottom: 10px; right: 10px; background: rgba(0,0,0,0.8); color: white; padding: 10px; border-radius: 5px; font-size: 11px; z-index: 10000; max-width: 300px;">
      <strong>App Debug:</strong><br>
      Auth: {{ isAuthenticated ? '✅' : '❌' }}<br>
      Admin: {{ isAdmin ? '✅' : '❌' }}<br>
      Route: {{ $route.path }}<br>
      <button @click="showDebugInfo = false" style="margin-top: 5px; padding: 2px 5px;">Hide</button>
    </div>
    
    <nav v-if="isAuthenticated" class="navbar">
      <div class="container">
        <div class="nav-links">
          <router-link to="/dashboard">Dashboard</router-link>
          <router-link to="/vocab">Vocabulary</router-link>
          <router-link to="/training">Training</router-link>
          <router-link to="/chat">Chat</router-link>
          <router-link v-if="isAdmin" to="/admin">Admin</router-link>
          <button @click="toggleTheme" class="theme-toggle" :title="theme === 'dark' ? 'Switch to light theme' : 'Switch to dark theme'">
            <span v-if="theme === 'dark'">☀️</span>
            <span v-else>🌙</span>
          </button>
          <button @click="logout" class="btn btn-secondary">Logout</button>
        </div>
      </div>
    </nav>
    <main class="container">
      <router-view v-if="mounted" />
      <div v-else style="padding: 20px; text-align: center;">
        Loading...
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuth } from './composables/useAuth'
import { useTheme } from './composables/useTheme'

const router = useRouter()
const route = useRoute()
const { isAuthenticated, isAdmin, logout: authLogout } = useAuth()
const { theme, toggleTheme } = useTheme()

const mounted = ref(false)
const showDebugInfo = ref(false)
const authError = ref<string | null>(null)

// Check if we're in Telegram Mini App and show debug
onMounted(() => {
  const tg = (window as any).Telegram?.WebApp
  if (tg) {
    showDebugInfo.value = true
  }
  
  mounted.value = true
  console.log('[App] Component mounted', {
    isAuthenticated: isAuthenticated.value,
    currentRoute: route.path
  })
  
  // Check auth status after a delay
  setTimeout(() => {
    if (!isAuthenticated.value && route.path !== '/login') {
      // If not authenticated and not on login page, show error
      const tg = (window as any).Telegram?.WebApp
      if (tg) {
        authError.value = 'Авторизация через Telegram не удалась. Пожалуйста, используйте OTP вход.'
      }
    }
  }, 3000)
})

// Watch route changes
watch(() => route.path, (newPath) => {
  console.log('[App] Route changed to:', newPath)
  // Clear error when navigating to login
  if (newPath === '/login') {
    authError.value = null
  }
})

// Watch auth status
watch(() => isAuthenticated.value, (newValue) => {
  if (newValue) {
    authError.value = null
  }
})

// Global error handler for auth errors
;(window as any).__setAuthError = (error: string) => {
  authError.value = error
}

const dismissAuthError = () => {
  authError.value = null
}

const logout = () => {
  authLogout()
  router.push('/login')
}
</script>

<style scoped>
.navbar {
  background: var(--bg-secondary);
  box-shadow: 0 2px 4px var(--navbar-shadow);
}

.nav-links {
  display: flex;
  gap: 20px;
  align-items: center;
  flex-wrap: wrap;
}

.nav-links a {
  text-decoration: none;
  color: var(--color-primary);
  padding: 10px;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.nav-links a:hover,
.nav-links a.router-link-active {
  background-color: var(--bg-hover);
}

.theme-toggle {
  background: transparent;
  border: 1px solid var(--border-primary);
  border-radius: 6px;
  padding: 8px 12px;
  cursor: pointer;
  font-size: 18px;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 44px;
}

.theme-toggle:hover {
  background-color: var(--bg-hover);
  border-color: var(--border-secondary);
}

.auth-error-banner {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  background: #ff4444;
  color: white;
  padding: 15px 20px;
  z-index: 10001;
  box-shadow: 0 2px 8px rgba(0,0,0,0.2);
}

.auth-error-content {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 15px;
}

.auth-error-close {
  background: rgba(255,255,255,0.2);
  border: none;
  color: white;
  font-size: 24px;
  line-height: 1;
  padding: 0 10px;
  cursor: pointer;
  border-radius: 3px;
  transition: background 0.2s;
}

.auth-error-close:hover {
  background: rgba(255,255,255,0.3);
}
</style>

