<template>
  <div id="app">
    <!-- Auth Error Message -->
    <div v-if="authError" class="auth-error-banner">
      <div class="auth-error-content">
        <strong><Icon name="warning" /> Ошибка авторизации:</strong> {{ authError }}
        <button @click="dismissAuthError" class="auth-error-close">×</button>
      </div>
    </div>
    
    <!-- Desktop Navbar -->
    <nav v-if="isAuthenticated" class="navbar navbar-desktop">
      <div class="container">
        <div class="nav-links">
          <div class="nav-left">
            <router-link to="/dashboard">Dashboard</router-link>
            <router-link to="/vocab">Vocabulary</router-link>
            <router-link to="/training">Training</router-link>
            <router-link to="/chat">Chat</router-link>
            <router-link v-if="isAdmin" to="/admin">Admin</router-link>
          </div>
          <div class="nav-right">
            <button @click="toggleTheme" class="theme-toggle" :title="theme === 'dark' ? 'Switch to light theme' : 'Switch to dark theme'">
              <Icon v-if="theme === 'dark'" name="sun" />
              <Icon v-else name="moon" />
            </button>
            <button v-if="!isTelegramMiniApp" @click="logout" class="btn btn-secondary">Logout</button>
          </div>
        </div>
      </div>
    </nav>
    
    <main class="container" :class="{ 'with-mobile-footer': isAuthenticated && isMobile }">
      <router-view v-if="mounted" />
      <div v-else style="padding: 20px; text-align: center;">
        Loading...
      </div>
    </main>
    
    <!-- Mobile Footer Navigation -->
    <nav v-if="isAuthenticated" class="navbar-mobile">
      <div class="mobile-nav-main">
        <router-link to="/dashboard" class="mobile-nav-item" title="Dashboard">
          <Icon name="dashboard" class="mobile-nav-icon" />
          <span class="mobile-nav-label">Dashboard</span>
        </router-link>
        <router-link to="/vocab" class="mobile-nav-item" title="Vocabulary">
          <Icon name="book" class="mobile-nav-icon" />
          <span class="mobile-nav-label">Vocab</span>
        </router-link>
        <router-link to="/training" class="mobile-nav-item" title="Training">
          <Icon name="target" class="mobile-nav-icon" />
          <span class="mobile-nav-label">Training</span>
        </router-link>
        <router-link to="/chat" class="mobile-nav-item" title="Chat">
          <Icon name="chat" class="mobile-nav-icon" />
          <span class="mobile-nav-label">Chat</span>
        </router-link>
        <button 
          v-if="hasMoreItems" 
          @click="showSidebar = !showSidebar" 
          class="mobile-nav-item mobile-nav-more"
          :class="{ active: showSidebar }"
          title="More"
        >
          <Icon name="more" class="mobile-nav-icon" />
          <span class="mobile-nav-label">More</span>
        </button>
      </div>
    </nav>
    
    <!-- Sidebar Overlay -->
    <div v-if="showSidebar" class="sidebar-overlay" @click="showSidebar = false"></div>
    
    <!-- Sidebar -->
    <aside v-if="isAuthenticated && showSidebar" class="sidebar" :class="{ open: showSidebar }">
      <div class="sidebar-header">
        <h3>Menu</h3>
        <button @click="showSidebar = false" class="sidebar-close">×</button>
      </div>
      <div class="sidebar-content">
        <router-link v-if="isAdmin" to="/admin" class="sidebar-item" @click="showSidebar = false">
          <Icon name="settings" class="sidebar-icon" />
          <span>Admin</span>
        </router-link>
        <button @click="handleThemeToggle" class="sidebar-item">
          <Icon :name="theme === 'dark' ? 'sun' : 'moon'" class="sidebar-icon" />
          <span>{{ theme === 'dark' ? 'Light' : 'Dark' }} Theme</span>
        </button>
        <button v-if="!isTelegramMiniApp" @click="handleLogout" class="sidebar-item">
          <Icon name="logout" class="sidebar-icon" />
          <span>Logout</span>
        </button>
      </div>
    </aside>
    
    <!-- Global Dialog Modals -->
    <AlertModal 
      :message="alertState.message" 
      :visible="alertState.visible"
      @close="closeAlert"
    />
    <ConfirmModal 
      :message="confirmState.message" 
      :visible="confirmState.visible"
      @confirm="() => closeConfirm(true)"
      @cancel="() => closeConfirm(false)"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuth } from './composables/useAuth'
import { useTheme } from './composables/useTheme'
import { useDialog } from './composables/useDialog'
import Icon from './components/Icon.vue'
import AlertModal from './components/AlertModal.vue'
import ConfirmModal from './components/ConfirmModal.vue'

const router = useRouter()
const route = useRoute()
const { isAuthenticated, isAdmin, logout: authLogout } = useAuth()
const { theme, toggleTheme } = useTheme()
const { alertState, confirmState, closeAlert, closeConfirm } = useDialog()

const mounted = ref(false)
const authError = ref<string | null>(null)
const isTelegramMiniApp = ref(false)
const isMobile = ref(false)
const showSidebar = ref(false)

// Check if we're in Telegram Mini App
onMounted(() => {
  const tg = (window as any).Telegram?.WebApp
  isTelegramMiniApp.value = !!tg
  
  // Check if mobile device
  const checkMobile = () => {
    isMobile.value = window.innerWidth <= 768
  }
  checkMobile()
  window.addEventListener('resize', checkMobile)
  
  mounted.value = true
  
  // Check auth status after a delay
  setTimeout(() => {
    if (!isAuthenticated.value && route.path !== '/login') {
      if (tg) {
        authError.value = 'Авторизация через Telegram не удалась. Пожалуйста, используйте OTP вход.'
      }
    }
  }, 3000)
})

// Watch route changes
watch(() => route.path, (newPath) => {
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

const handleThemeToggle = () => {
  toggleTheme()
  showSidebar.value = false
}

const handleLogout = () => {
  logout()
  showSidebar.value = false
}

// Check if we have more items to show in "More" menu
const hasMoreItems = computed(() => {
  return isAdmin.value || !isTelegramMiniApp.value
})
</script>

<style scoped>
/* Desktop Navbar */
.navbar-desktop {
  background: var(--bg-secondary);
  box-shadow: 0 2px 4px var(--navbar-shadow);
}

.nav-links {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 20px;
  flex-wrap: wrap;
}

.nav-left {
  display: flex;
  gap: 20px;
  align-items: center;
  flex-wrap: wrap;
}

.nav-right {
  display: flex;
  gap: 12px;
  align-items: center;
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

/* Mobile Footer Navigation */
.navbar-mobile {
  display: none;
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background: var(--bg-secondary);
  box-shadow: 0 -2px 8px var(--navbar-shadow);
  z-index: 1000;
  border-top: 1px solid var(--border-primary);
}

.mobile-nav-main {
  display: flex;
  align-items: stretch;
  height: 60px;
}

.mobile-nav-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  padding: 0;
  text-decoration: none;
  color: var(--color-primary);
  background: transparent;
  border: none;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.2s;
  flex: 1;
  min-width: 0;
  margin: 0;
  border-radius: 0;
}

.mobile-nav-item:hover,
.mobile-nav-item.router-link-active,
.mobile-nav-item.active {
  background: var(--bg-hover);
  color: var(--color-primary);
}

.mobile-nav-icon {
  font-size: 20px;
  line-height: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.mobile-nav-label {
  font-size: 10px;
  line-height: 1.2;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 100%;
}

/* Sidebar */
.sidebar-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 1001;
  animation: fadeIn 0.2s ease-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.sidebar {
  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  width: 280px;
  max-width: 80vw;
  background: var(--bg-secondary);
  box-shadow: 2px 0 8px var(--navbar-shadow);
  z-index: 1002;
  transform: translateX(-100%);
  transition: transform 0.3s ease-out;
  display: flex;
  flex-direction: column;
}

.sidebar.open {
  transform: translateX(0);
}

.sidebar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid var(--border-primary);
}

.sidebar-header h3 {
  margin: 0;
  font-size: 18px;
  color: var(--color-primary);
}

.sidebar-close {
  background: transparent;
  border: none;
  color: var(--color-primary);
  font-size: 28px;
  line-height: 1;
  cursor: pointer;
  padding: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.sidebar-close:hover {
  background: var(--bg-hover);
}

.sidebar-content {
  flex: 1;
  overflow-y: auto;
  padding: 10px 0;
}

.sidebar-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 20px;
  width: 100%;
  text-align: left;
  background: transparent;
  border: none;
  color: var(--color-primary);
  cursor: pointer;
  font-size: 16px;
  transition: background-color 0.2s;
  text-decoration: none;
}

.sidebar-item:hover {
  background: var(--bg-hover);
}

.sidebar-icon {
  font-size: 20px;
  width: 24px;
  text-align: center;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* Main content padding for mobile footer */
main.with-mobile-footer {
  padding-bottom: 80px;
}

/* Responsive: Hide desktop nav on mobile, show mobile nav */
@media (max-width: 768px) {
  .navbar-desktop {
    display: none;
  }
  
  .navbar-mobile {
    display: block;
  }
  
  .container {
    padding-bottom: 20px;
  }
}

/* Hide mobile nav on desktop */
@media (min-width: 769px) {
  .navbar-mobile {
    display: none !important;
  }
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

