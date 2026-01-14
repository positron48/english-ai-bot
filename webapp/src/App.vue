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
            <router-link to="/learning">Learning</router-link>
            <router-link to="/training">Training</router-link>
            <router-link to="/chat">Chat</router-link>
            <div class="nav-more-wrapper">
              <button 
                ref="moreButtonRef"
                @click.stop="showMoreDropdown = !showMoreDropdown" 
                class="nav-more-btn" 
                title="More"
                :class="{ active: showMoreDropdown }"
              >
                More
                <Icon name="chevron-down" class="nav-more-chevron" />
              </button>
              <div v-if="showMoreDropdown" ref="moreDropdownRef" class="nav-more-dropdown">
                <router-link to="/vocab" class="dropdown-item" @click="showMoreDropdown = false">
                  <Icon name="book" class="dropdown-icon" />
                  <span>Vocabulary</span>
                </router-link>
                <router-link v-if="isAdmin" to="/admin" class="dropdown-item" @click="showMoreDropdown = false">
                  <Icon name="shield" class="dropdown-icon" />
                  <span>Admin</span>
                </router-link>
                <button v-if="!isTelegramMiniApp" @click="handleMoreLogout" class="dropdown-item">
                  <Icon name="logout" class="dropdown-icon" />
                  <span>Logout</span>
                </button>
              </div>
            </div>
          </div>
          <div class="nav-right">
            <router-link to="/settings" class="nav-settings-btn" title="Settings">
              <Icon name="gear" />
            </router-link>
          </div>
        </div>
      </div>
    </nav>
    
    <main class="container" :class="{ 'with-mobile-footer': isAuthenticated && isMobile, 'with-desktop-navbar': isAuthenticated && !isMobile }">
      <Breadcrumbs v-if="isAuthenticated && mounted" />
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
        <router-link to="/learning" class="mobile-nav-item" title="Learning">
          <Icon name="book" class="mobile-nav-icon" />
          <span class="mobile-nav-label">Learning</span>
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
        <router-link to="/vocab" class="sidebar-item" @click="showSidebar = false">
          <Icon name="book" class="sidebar-icon" />
          <span>Vocabulary</span>
        </router-link>
        <router-link to="/settings" class="sidebar-item" @click="showSidebar = false">
          <Icon name="gear" class="sidebar-icon" />
          <span>Settings</span>
        </router-link>
        <router-link v-if="isAdmin" to="/admin" class="sidebar-item" @click="showSidebar = false">
          <Icon name="shield" class="sidebar-icon" />
          <span>Admin</span>
        </router-link>
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
import { ref, onMounted, onUnmounted, watch, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuth } from './composables/useAuth'
import { useTheme } from './composables/useTheme'
import { useDialog } from './composables/useDialog'
import Icon from './components/Icon.vue'
import AlertModal from './components/AlertModal.vue'
import ConfirmModal from './components/ConfirmModal.vue'
import Breadcrumbs from './components/Breadcrumbs.vue'

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
const showMoreDropdown = ref(false)
const moreDropdownRef = ref<HTMLElement | null>(null)
const moreButtonRef = ref<HTMLElement | null>(null)

// Handle click outside dropdown
const handleClickOutside = (event: MouseEvent) => {
  if (showMoreDropdown.value) {
    const target = event.target as HTMLElement
    if (moreDropdownRef.value && !moreDropdownRef.value.contains(target) &&
        moreButtonRef.value && !moreButtonRef.value.contains(target)) {
      showMoreDropdown.value = false
    }
  }
}

// Check if mobile device
const checkMobile = () => {
  isMobile.value = window.innerWidth <= 768
}

// Check if we're in Telegram Mini App
onMounted(() => {
  const tg = (window as any).Telegram?.WebApp
  isTelegramMiniApp.value = !!tg
  
  checkMobile()
  window.addEventListener('resize', checkMobile)
  
  // Add click outside handler
  document.addEventListener('click', handleClickOutside)
  
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

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  window.removeEventListener('resize', checkMobile)
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

const handleMoreLogout = () => {
  logout()
  showMoreDropdown.value = false
}

// Settings is always in More menu on mobile
</script>

<style scoped>
/* Desktop Navbar */
.navbar-desktop {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  background: var(--bg-secondary);
  box-shadow: 0 2px 4px var(--navbar-shadow);
  z-index: 1000;
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

.nav-more-wrapper {
  position: relative;
}

.nav-more-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  background: transparent;
  border: none;
  padding: 10px;
  cursor: pointer;
  font-size: inherit;
  color: var(--text-primary);
  transition: background-color 0.2s;
  text-decoration: none;
  border-radius: 4px;
}

.nav-more-btn:hover,
.nav-more-btn.active {
  background-color: var(--bg-hover);
}

.nav-more-chevron {
  font-size: 0.8em;
  transition: transform 0.2s;
}

.nav-more-btn.active .nav-more-chevron {
  transform: rotate(180deg);
}

.nav-more-dropdown {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  background: var(--bg-secondary);
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  box-shadow: 0 4px 12px var(--navbar-shadow);
  min-width: 180px;
  z-index: 1002;
  overflow: hidden;
  animation: fadeIn 0.2s ease-out;
  pointer-events: auto;
}

.dropdown-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  width: 100%;
  text-align: left;
  background: transparent;
  border: none;
  color: var(--text-primary);
  cursor: pointer;
  font-size: 14px;
  transition: background-color 0.2s;
  text-decoration: none;
  pointer-events: auto;
  position: relative;
  z-index: 1;
}

.dropdown-item:hover {
  background: var(--bg-hover);
}

.dropdown-icon {
  font-size: 18px;
  width: 20px;
  text-align: center;
  display: flex;
  align-items: center;
  justify-content: center;
}

.nav-settings-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: 1px solid var(--border-primary);
  border-radius: 6px;
  padding: 8px 12px;
  cursor: pointer;
  font-size: 18px;
  color: var(--text-primary);
  transition: all 0.2s;
  text-decoration: none;
  min-width: 44px;
}

.nav-settings-btn:hover,
.nav-settings-btn.router-link-active {
  background-color: var(--bg-hover);
  border-color: var(--border-secondary);
}


.nav-links a {
  text-decoration: none;
  color: var(--text-primary);
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

.theme-toggle:hover,
.theme-toggle.active {
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
  color: var(--text-primary);
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
  color: var(--text-primary);
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
  color: var(--text-primary);
}

.sidebar-close {
  background: transparent;
  border: none;
  color: var(--text-primary);
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
  color: var(--text-primary);
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

/* Main content padding for desktop navbar */
main.with-desktop-navbar {
  padding-top: 80px;
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
    padding-left: 8px;
    padding-right: 8px;
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

