<template>
  <div id="app">
    <!-- Auth Error Message -->
    <div v-if="authError" class="auth-error-banner">
      <div class="auth-error-content">
        <strong><Icon name="warning" /> {{ t('auth.error') }}</strong> {{ authError }}
        <button @click="dismissAuthError" class="auth-error-close">×</button>
      </div>
    </div>
    
    <!-- Desktop Navbar -->
    <nav v-if="isAuthenticated && !isAdminRoute" class="navbar navbar-desktop">
      <div class="container">
        <div class="nav-links">
          <div class="nav-left">
            <router-link to="/dashboard">{{ t('navigation.dashboard') }}</router-link>
            <router-link to="/learning">{{ t('navigation.learning') }}</router-link>
            <router-link to="/training">{{ t('navigation.training') }}</router-link>
            <router-link to="/chat">{{ t('navigation.chat') }}</router-link>
            <div class="nav-more-wrapper">
              <button 
                ref="moreButtonRef"
                @click.stop="showMoreDropdown = !showMoreDropdown" 
                class="nav-more-btn" 
                :title="t('navigation.more')"
                :class="{ active: showMoreDropdown }"
              >
                {{ t('navigation.more') }}
                <Icon name="chevron-down" class="nav-more-chevron" />
              </button>
              <div v-if="showMoreDropdown" ref="moreDropdownRef" class="nav-more-dropdown">
                <router-link to="/vocab" class="dropdown-item" @click="showMoreDropdown = false">
                  <Icon name="book" class="dropdown-icon" />
                  <span>{{ t('navigation.vocab') }}</span>
                </router-link>
                <router-link v-if="isAdmin" to="/admin" class="dropdown-item" @click="showMoreDropdown = false">
                  <Icon name="shield" class="dropdown-icon" />
                  <span>{{ t('navigation.admin') }}</span>
                </router-link>
                <button v-if="!isTelegramMiniApp" @click="handleMoreLogout" class="dropdown-item">
                  <Icon name="logout" class="dropdown-icon" />
                  <span>{{ t('navigation.logout') }}</span>
                </button>
              </div>
            </div>
          </div>
          <div class="nav-right">
            <div class="lang-switcher" @click.stop="handleLangSwitch">
              <div 
                class="lang-slider"
                :class="{ 'lang-slider-ru': currentLocale === 'ru' }"
              ></div>
              <button
                @click.stop="setLocale('en')"
                class="lang-btn"
                :class="{ active: currentLocale === 'en' }"
                title="English"
              >
                EN
              </button>
              <button
                @click.stop="setLocale('ru')"
                class="lang-btn"
                :class="{ active: currentLocale === 'ru' }"
                title="Русский"
              >
                RU
              </button>
            </div>
            <router-link to="/settings" class="nav-settings-btn" :title="t('navigation.settings')">
              <Icon name="gear" />
            </router-link>
          </div>
        </div>
      </div>
    </nav>
    
    <main
      class="container"
      :class="{
        'with-mobile-footer': isAuthenticated && isMobile && !isAdminRoute,
        'with-desktop-navbar': isAuthenticated && !isMobile && !isAdminRoute,
        'main-admin': isAdminRoute
      }"
    >
      <Breadcrumbs v-if="isAuthenticated && mounted && !isAdminRoute" />
      <router-view v-if="mounted" :key="route.path" />
      <div v-else style="padding: 20px; text-align: center;">
        {{ t('common.loading') }}
      </div>
    </main>
    
    <!-- Mobile Footer Navigation -->
    <nav v-if="isAuthenticated && !isAdminRoute" class="navbar-mobile">
      <div class="mobile-nav-main">
        <router-link to="/dashboard" class="mobile-nav-item" :title="t('navigation.dashboard')">
          <Icon name="dashboard" class="mobile-nav-icon" />
          <span class="mobile-nav-label">{{ t('navigation.dashboard') }}</span>
        </router-link>
        <router-link to="/learning" class="mobile-nav-item" :title="t('navigation.learning')">
          <Icon name="book" class="mobile-nav-icon" />
          <span class="mobile-nav-label">{{ t('navigation.learning') }}</span>
        </router-link>
        <router-link to="/training" class="mobile-nav-item" :title="t('navigation.training')">
          <Icon name="target" class="mobile-nav-icon" />
          <span class="mobile-nav-label">{{ t('navigation.training') }}</span>
        </router-link>
        <router-link to="/chat" class="mobile-nav-item" :title="t('navigation.chat')">
          <Icon name="chat" class="mobile-nav-icon" />
          <span class="mobile-nav-label">{{ t('navigation.chat') }}</span>
        </router-link>
        <button 
          @click="showSidebar = !showSidebar" 
          class="mobile-nav-item mobile-nav-more"
          :class="{ active: showSidebar }"
          :title="t('navigation.more')"
        >
          <Icon name="more" class="mobile-nav-icon" />
          <span class="mobile-nav-label">{{ t('navigation.more') }}</span>
        </button>
      </div>
    </nav>
    
    <!-- Sidebar Overlay -->
    <div v-if="showSidebar" class="sidebar-overlay" @click="showSidebar = false"></div>
    
    <!-- Sidebar -->
    <aside v-if="isAuthenticated && showSidebar && !isAdminRoute" class="sidebar" :class="{ open: showSidebar }">
      <div class="sidebar-header">
        <h3>{{ t('navigation.menu') }}</h3>
        <button @click="showSidebar = false" class="sidebar-close">×</button>
      </div>
      <div class="sidebar-content">
        <div class="sidebar-lang-switcher">
          <span class="sidebar-lang-label">{{ t('common.language') || 'Language' }}:</span>
          <div class="lang-switcher" @click.stop="handleLangSwitchSidebar">
            <div 
              class="lang-slider"
              :class="{ 'lang-slider-ru': currentLocale === 'ru' }"
            ></div>
            <button
              @click.stop="setLocale('en'); showSidebar = false"
              class="lang-btn"
              :class="{ active: currentLocale === 'en' }"
            >
              EN
            </button>
            <button
              @click.stop="setLocale('ru'); showSidebar = false"
              class="lang-btn"
              :class="{ active: currentLocale === 'ru' }"
            >
              RU
            </button>
          </div>
        </div>
        <router-link to="/vocab" class="sidebar-item" @click="showSidebar = false">
          <Icon name="book" class="sidebar-icon" />
          <span>{{ t('navigation.vocab') }}</span>
        </router-link>
        <router-link to="/settings" class="sidebar-item" @click="showSidebar = false">
          <Icon name="gear" class="sidebar-icon" />
          <span>{{ t('navigation.settings') }}</span>
        </router-link>
        <router-link v-if="isAdmin" to="/admin" class="sidebar-item" @click="showSidebar = false">
          <Icon name="shield" class="sidebar-icon" />
          <span>{{ t('navigation.admin') }}</span>
        </router-link>
        <button v-if="!isTelegramMiniApp" @click="handleLogout" class="sidebar-item">
          <Icon name="logout" class="sidebar-icon" />
          <span>{{ t('navigation.logout') }}</span>
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
import { useI18n } from 'vue-i18n'
import { useAuth } from './composables/useAuth'
import { useTheme } from './composables/useTheme'
import { useDialog } from './composables/useDialog'
import { useLocale } from './composables/useLocale'
import Icon from './components/Icon.vue'
import AlertModal from './components/AlertModal.vue'
import ConfirmModal from './components/ConfirmModal.vue'
import Breadcrumbs from './components/Breadcrumbs.vue'

const router = useRouter()
const route = useRoute()
const { t } = useI18n()
const { isAuthenticated, isAdmin, logout: authLogout } = useAuth()
const { currentLocale, setLocale } = useLocale()

const isAdminRoute = computed(() => {
  return route.path.startsWith('/admin')
})
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
        authError.value = t('auth.telegramAuthFailed')
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

const handleLangSwitch = (event: MouseEvent) => {
  const target = event.target as HTMLElement
  if (target.classList.contains('lang-switcher') || target.classList.contains('lang-slider')) {
    // Toggle language when clicking on switcher background or slider
    const newLocale = currentLocale.value === 'en' ? 'ru' : 'en'
    setLocale(newLocale)
  }
}

const handleLangSwitchSidebar = (event: MouseEvent) => {
  const target = event.target as HTMLElement
  if (target.classList.contains('lang-switcher') || target.classList.contains('lang-slider')) {
    // Toggle language when clicking on switcher background or slider
    const newLocale = currentLocale.value === 'en' ? 'ru' : 'en'
    setLocale(newLocale)
    showSidebar.value = false
  }
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

/* Admin: full width on desktop, no container constraint */
main.main-admin {
  max-width: none;
  margin: 0;
  padding: 0;
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

/* Language Switcher */
.lang-switcher {
  position: relative;
  display: flex;
  gap: 0;
  align-items: center;
  background: var(--bg-hover);
  border: 1px solid var(--border-primary);
  border-radius: 6px;
  padding: 2px;
  margin-right: 8px;
  overflow: hidden;
  cursor: pointer;
}

.lang-slider {
  position: absolute;
  top: 2px;
  left: 2px;
  width: calc(50% - 2px);
  height: calc(100% - 4px);
  background: var(--bg-secondary);
  border-radius: 4px;
  transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  z-index: 0;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
  pointer-events: none;
}

.lang-slider-ru {
  transform: translateX(100%);
}

.lang-btn {
  position: relative;
  background: transparent;
  border: none;
  padding: 6px 12px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
  border-radius: 4px;
  transition: color 0.2s;
  min-width: 40px;
  z-index: 1;
  flex: 1;
  user-select: none;
}

.lang-btn:hover {
  color: var(--text-primary);
}

.lang-btn.active {
  color: var(--text-primary);
}

/* Sidebar Language Switcher */
.sidebar-lang-switcher {
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-primary);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.sidebar-lang-label {
  font-size: 14px;
  color: var(--text-secondary);
  white-space: nowrap;
}

.sidebar-lang-switcher .lang-switcher {
  margin-right: 0;
  flex: 1;
  max-width: 200px;
}
</style>

