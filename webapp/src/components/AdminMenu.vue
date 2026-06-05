<template>
  <!-- Swipe indicator at left edge - тонкая полоска -->
  <div 
    v-if="isMobile && !showSidebar" 
    class="admin-swipe-indicator"
    @click="showSidebar = true"
  >
    <div class="admin-swipe-indicator-handle"></div>
  </div>

  <!-- Sidebar overlay for mobile -->
  <div 
    v-if="isMobile && showSidebar" 
    class="admin-sidebar-overlay"
    data-admin-sidebar-overlay="true"
    @click="handleCloseSidebar"
    @touchstart="handleTouchStart"
    @touchmove="handleTouchMove"
    @touchend="handleTouchEnd"
  ></div>

  <!-- Sidebar -->
  <aside 
    class="admin-sidebar" 
    :class="{ 'open': showSidebar || !isMobile }"
    data-admin-sidebar="true"
  >
    <div class="admin-sidebar-header">
      <div class="admin-sidebar-header-left">
        <router-link to="/dashboard" class="admin-home-button" title="Go to main site">
          <Icon name="home" />
        </router-link>
        <h1>Admin Panel</h1>
      </div>
      <button 
        v-if="isMobile" 
        @click="handleCloseSidebar" 
        class="admin-sidebar-close"
        aria-label="Close admin menu"
      >
        <Icon name="close" />
      </button>
    </div>
    
    <nav class="admin-sidebar-nav">
      <router-link 
        v-if="can('words.read_all')"
        to="/admin" 
        class="admin-sidebar-item" 
        :class="{ active: $route.path === '/admin' }"
        @click="isMobile && handleCloseSidebar()"
      >
        <Icon name="home" class="admin-sidebar-icon" />
        <span>Words Management</span>
      </router-link>
      <router-link
        v-if="can('words.read_all') && showSpanishVerbFormsAdmin"
        to="/admin/verb-forms"
        class="admin-sidebar-item"
        :class="{ active: $route.path === '/admin/verb-forms' }"
        @click="isMobile && handleCloseSidebar()"
      >
        <Icon name="list" class="admin-sidebar-icon" />
        <span>{{ t('navigation.adminVerbFormsDb') }}</span>
      </router-link>
      <router-link 
        v-if="can('full_access')"
        to="/admin/circuit-breaker" 
        class="admin-sidebar-item" 
        :class="{ active: $route.path === '/admin/circuit-breaker' }"
        @click="isMobile && handleCloseSidebar()"
      >
        <Icon name="toggle" class="admin-sidebar-icon" />
        <span>Circuit Breaker</span>
      </router-link>
      <router-link 
        v-if="can('full_access')"
        to="/admin/prompt-tester" 
        class="admin-sidebar-item" 
        :class="{ active: $route.path === '/admin/prompt-tester' }"
        @click="isMobile && handleCloseSidebar()"
      >
        <Icon name="code" class="admin-sidebar-icon" />
        <span>Prompt Tester</span>
      </router-link>
      <router-link
        v-if="can('full_access')"
        to="/admin/content-reports"
        class="admin-sidebar-item"
        :class="{ active: $route.path === '/admin/content-reports' }"
        @click="isMobile && handleCloseSidebar()"
      >
        <Icon name="warning" class="admin-sidebar-icon" />
        <span>Content Reports</span>
      </router-link>
      <router-link 
        v-if="can('word_sets.read')"
        to="/admin/word-sets" 
        class="admin-sidebar-item" 
        :class="{ active: $route.path.startsWith('/admin/word-sets') }"
        @click="isMobile && handleCloseSidebar()"
      >
        <Icon name="book" class="admin-sidebar-icon" />
        <span>Word Sets</span>
      </router-link>
      <router-link 
        v-if="can('full_access')"
        to="/admin/grammar" 
        class="admin-sidebar-item" 
        :class="{ active: $route.path === '/admin/grammar' }"
        @click="isMobile && handleCloseSidebar()"
      >
        <Icon name="book-open" class="admin-sidebar-icon" />
        <span>Grammar</span>
      </router-link>
      <router-link
        v-if="can('full_access')"
        to="/admin/reading-texts"
        class="admin-sidebar-item"
        :class="{ active: $route.path === '/admin/reading-texts' }"
        @click="isMobile && handleCloseSidebar()"
      >
        <Icon name="book" class="admin-sidebar-icon" />
        <span>Reading Texts</span>
      </router-link>
      <router-link 
        v-if="can('full_access')"
        to="/admin/app-settings" 
        class="admin-sidebar-item" 
        :class="{ active: $route.path === '/admin/app-settings' }"
        @click="isMobile && handleCloseSidebar()"
      >
        <Icon name="gear" class="admin-sidebar-icon" />
        <span>App Settings</span>
      </router-link>
      <router-link
        v-if="can('full_access')"
        to="/admin/linglow-srs"
        class="admin-sidebar-item"
        :class="{ active: $route.path === '/admin/linglow-srs' }"
        @click="isMobile && handleCloseSidebar()"
      >
        <Icon name="chart" class="admin-sidebar-icon" />
        <span>Linglow SRS</span>
      </router-link>
      <router-link 
        v-if="can('full_access')"
        to="/admin/access" 
        class="admin-sidebar-item" 
        :class="{ active: $route.path.startsWith('/admin/access') }"
        @click="isMobile && handleCloseSidebar()"
      >
        <Icon name="lock" class="admin-sidebar-icon" />
        <span>Access Control</span>
      </router-link>
      <router-link 
        v-if="can('users.read_all')"
        to="/admin/users" 
        class="admin-sidebar-item" 
        :class="{ active: $route.path === '/admin/users' }"
        @click="isMobile && handleCloseSidebar()"
      >
        <Icon name="users" class="admin-sidebar-icon" />
        <span>Users</span>
      </router-link>
      <router-link 
        v-if="can('stats.read')"
        to="/admin/stats" 
        class="admin-sidebar-item" 
        :class="{ active: $route.path === '/admin/stats' }"
        @click="isMobile && handleCloseSidebar()"
      >
        <Icon name="chart" class="admin-sidebar-icon" />
        <span>Statistics</span>
      </router-link>
    </nav>
    
    <!-- Theme Settings -->
    <div class="admin-sidebar-footer">
      <div class="admin-theme-settings">
        <div class="admin-theme-label">
          <Icon name="gear" class="admin-sidebar-icon" />
          <span>Theme</span>
        </div>
        <label class="admin-theme-toggle">
          <input 
            type="checkbox" 
            :checked="selectedTheme === 'dark'"
            @change="handleThemeToggle"
          />
          <span class="admin-theme-slider">
            <Icon name="sun" class="admin-theme-icon admin-theme-icon-sun" />
            <Icon name="moon" class="admin-theme-icon admin-theme-icon-moon" />
          </span>
        </label>
      </div>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { useTheme } from '../composables/useTheme'
import { useAuth } from '../composables/useAuth'
import { useLearningConfig } from '../composables/useLearningConfig'
import Icon from './Icon.vue'

const { t } = useI18n()
const { can, loadPermissions } = useAuth()
const { learning, ensureLearningLoaded } = useLearningConfig()

const showSidebar = ref(false)
const windowWidth = ref(window.innerWidth)
const { theme: currentTheme, toggleTheme, setTheme } = useTheme()
const selectedTheme = ref<'light' | 'dark'>(currentTheme.value)

const isMobile = computed(() => windowWidth.value < 768)
const showSpanishVerbFormsAdmin = computed(() => Boolean(learning.value?.spanish_verb_forms_enabled))

let isClosingSidebar = false // Флаг для предотвращения двойного закрытия

const handleCloseSidebar = () => {
  if (isClosingSidebar || !showSidebar.value) return
  
  isClosingSidebar = true
  showSidebar.value = false
  
  // Сбрасываем флаг через небольшую задержку
  setTimeout(() => {
    isClosingSidebar = false
  }, 300)
}

// Swipe gesture handling
const touchStartX = ref(0)
const touchStartY = ref(0)
const touchStartTime = ref(0)
const SWIPE_THRESHOLD = 50 // Минимальное расстояние для свайпа
const SWIPE_EDGE_ZONE = 20 // Зона у края экрана для активации свайпа
const SWIPE_MAX_VERTICAL = 100 // Максимальное вертикальное отклонение

const handleTouchStart = (e: TouchEvent) => {
  touchStartX.value = e.touches[0].clientX
  touchStartY.value = e.touches[0].clientY
  touchStartTime.value = Date.now()
}

const handleTouchMove = (e: TouchEvent) => {
  // Предотвращаем скролл страницы при свайпе от края
  if (touchStartX.value < SWIPE_EDGE_ZONE && !showSidebar.value) {
    e.preventDefault()
  }
}

const handleTouchEnd = (e: TouchEvent) => {
  if (!touchStartX.value || !touchStartY.value) return
  
  const touchEndX = e.changedTouches[0].clientX
  const touchEndY = e.changedTouches[0].clientY
  const deltaX = touchEndX - touchStartX.value
  const deltaY = Math.abs(touchEndY - touchStartY.value)
  const deltaTime = Date.now() - touchStartTime.value
  
  // Если начали свайп от левого края
  if (touchStartX.value < SWIPE_EDGE_ZONE) {
    // Свайп вправо для открытия меню
    if (deltaX > SWIPE_THRESHOLD && deltaY < SWIPE_MAX_VERTICAL && deltaTime < 500) {
      showSidebar.value = true
    }
  } else if (showSidebar.value && touchStartX.value > 100) {
    // Свайп влево для закрытия меню (если меню открыто)
    if (deltaX < -SWIPE_THRESHOLD && deltaY < SWIPE_MAX_VERTICAL && deltaTime < 500) {
      handleCloseSidebar()
    }
  }
  
  // Сброс
  touchStartX.value = 0
  touchStartY.value = 0
  touchStartTime.value = 0
}

// Watch for theme changes
watch(() => currentTheme.value, (newTheme) => {
  selectedTheme.value = newTheme
})

onMounted(() => {
  selectedTheme.value = currentTheme.value
})

const handleThemeToggle = (event: Event) => {
  const target = event.target as HTMLInputElement
  const newTheme = target.checked ? 'dark' : 'light'
  selectedTheme.value = newTheme
  setTheme(newTheme)
}

const updateWindowWidth = () => {
  windowWidth.value = window.innerWidth
  // Auto-close sidebar on mobile when resizing to desktop
  if (!isMobile.value) {
    handleCloseSidebar()
  }
}

let isOpeningMenu = false // Флаг для предотвращения двойного открытия
let menuEventListenerAdded = false // Флаг для предотвращения двойной регистрации слушателя

const handleOpenMenu = (e?: Event) => {
  // Предотвращаем обработку, если меню уже открывается или уже открыто
  if (isOpeningMenu || !isMobile.value || showSidebar.value) return
  
  isOpeningMenu = true
  showSidebar.value = true
  
  // Останавливаем распространение события, если оно было передано
  if (e) {
    e.stopPropagation()
    e.preventDefault()
  }
  
  // Сбрасываем флаг через небольшую задержку
  setTimeout(() => {
    isOpeningMenu = false
  }, 300)
}

onMounted(async () => {
  await loadPermissions()
  await ensureLearningLoaded()
  
  // Проверяем, нет ли уже другого сайдбара в DOM
  const existingSidebars = document.querySelectorAll('[data-admin-sidebar="true"]')
  if (existingSidebars.length > 1) {
    console.warn('Multiple admin sidebars detected, removing duplicates')
    // Оставляем только первый, удаляем остальные
    for (let i = 1; i < existingSidebars.length; i++) {
      existingSidebars[i].remove()
    }
  }
  
  // Проверяем, нет ли уже другого overlay в DOM
  const existingOverlays = document.querySelectorAll('[data-admin-sidebar-overlay="true"]')
  if (existingOverlays.length > 1) {
    console.warn('Multiple admin sidebar overlays detected, removing duplicates')
    // Оставляем только первый, удаляем остальные
    for (let i = 1; i < existingOverlays.length; i++) {
      existingOverlays[i].remove()
    }
  }
  
  window.addEventListener('resize', updateWindowWidth)
  // Слушаем событие открытия меню из layout, но только один раз
  if (!menuEventListenerAdded) {
    window.addEventListener('openMenu', handleOpenMenu, { once: false, passive: true })
    menuEventListenerAdded = true
  }
})

// Watch для удаления дубликатов overlay при изменении showSidebar
watch(() => showSidebar.value, (newValue) => {
  if (newValue) {
    // Когда сайдбар открывается, проверяем на дубликаты overlay
    nextTick(() => {
      const existingOverlays = document.querySelectorAll('[data-admin-sidebar-overlay="true"]')
      if (existingOverlays.length > 1) {
        // Оставляем только первый, удаляем остальные
        for (let i = 1; i < existingOverlays.length; i++) {
          existingOverlays[i].remove()
        }
      }
    })
  }
})

onUnmounted(() => {
  window.removeEventListener('resize', updateWindowWidth)
  if (menuEventListenerAdded) {
    window.removeEventListener('openMenu', handleOpenMenu)
    menuEventListenerAdded = false
  }
})
</script>

<style scoped>
/* Swipe indicator at left edge - тонкая полоска-подсказка */
.admin-swipe-indicator {
  position: fixed;
  left: 0;
  top: 0;
  bottom: 0;
  width: 8px;
  z-index: 1001;
  cursor: pointer;
  touch-action: none;
}

.admin-swipe-indicator-handle {
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 4px;
  height: 60px;
  background: var(--text-secondary, #666);
  border-radius: 0 2px 2px 0;
  opacity: 0.5;
  transition: opacity 0.2s, width 0.2s;
}

.admin-swipe-indicator:hover .admin-swipe-indicator-handle {
  opacity: 0.8;
  width: 6px;
  background: var(--color-primary, #1976d2);
}

/* Sidebar overlay for mobile */
.admin-sidebar-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 999;
  animation: fadeIn 0.2s ease-out;
  touch-action: pan-y;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

/* Sidebar */
.admin-sidebar {
  position: fixed;
  top: 0;
  left: 0;
  height: 100vh;
  width: 280px;
  background: var(--bg-primary);
  border-right: 1px solid var(--border-primary);
  z-index: 1000;
  display: flex;
  flex-direction: column;
  transform: translateX(-100%);
  transition: transform 0.3s ease-out;
  overflow-y: auto;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.1);
  touch-action: pan-y;
}

.admin-sidebar.open {
  transform: translateX(0);
}

/* Desktop: always visible, fixed at left edge */
@media (min-width: 768px) {
  .admin-sidebar {
    position: fixed;
    top: 0;
    left: 0;
    transform: translateX(0);
    height: 100vh;
    box-shadow: 2px 0 8px rgba(0, 0, 0, 0.1);
    width: 240px;
    flex-shrink: 0;
  }
  
  .admin-swipe-indicator {
    display: none;
  }
  
  .admin-sidebar-overlay {
    display: none;
  }
}

.admin-sidebar-header {
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-primary);
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
}

.admin-sidebar-header-left {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
}

.admin-home-button {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: transparent;
  border: 1px solid var(--border-primary);
  color: var(--text-secondary);
  text-decoration: none;
  transition: all 0.2s;
  flex-shrink: 0;
}

.admin-home-button:hover {
  background: var(--bg-hover);
  color: var(--color-primary);
  border-color: var(--color-primary);
}

.admin-home-button svg {
  width: 20px;
  height: 20px;
}

.admin-sidebar-header h1 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary);
  flex: 1;
}

.admin-sidebar-close {
  background: none;
  border: none;
  cursor: pointer;
  padding: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-secondary);
  transition: color 0.2s;
}

.admin-sidebar-close:hover {
  color: var(--text-primary);
}

.admin-sidebar-close svg {
  width: 24px;
  height: 24px;
}

@media (min-width: 768px) {
  .admin-sidebar-close {
    display: none;
  }
}

.admin-sidebar-nav {
  flex: 1;
  padding: 8px 0;
  overflow-y: auto;
}

.admin-sidebar-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  text-decoration: none;
  color: var(--text-secondary);
  transition: all 0.2s;
  border-left: 3px solid transparent;
  min-height: 44px;
}

.admin-sidebar-item:hover {
  color: var(--text-primary);
  background: var(--bg-hover);
}

.admin-sidebar-item.active {
  color: var(--text-primary);
  background: var(--bg-hover);
  border-left-color: var(--color-primary);
  font-weight: 500;
}

.admin-sidebar-icon {
  flex-shrink: 0;
  opacity: 0.9;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  min-width: 20px;
  min-height: 20px;
  max-width: 20px;
  max-height: 20px;
}

.admin-sidebar-icon :deep(.icon) {
  width: 20px !important;
  height: 20px !important;
  min-width: 20px !important;
  min-height: 20px !important;
  max-width: 20px !important;
  max-height: 20px !important;
  font-size: 20px !important;
  display: inline-flex !important;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.admin-sidebar-icon :deep(svg) {
  width: 20px !important;
  height: 20px !important;
  min-width: 20px !important;
  min-height: 20px !important;
  max-width: 20px !important;
  max-height: 20px !important;
  display: block;
  flex-shrink: 0;
}

.admin-sidebar-item span {
  flex: 1;
}

.admin-sidebar-footer {
  padding: 16px 20px;
  border-top: 1px solid var(--border-primary);
  flex-shrink: 0;
  margin-top: auto;
}

.admin-theme-settings {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.admin-theme-label {
  display: flex;
  align-items: center;
  gap: 10px;
  color: var(--text-secondary);
  font-size: 14px;
}

.admin-theme-toggle {
  position: relative;
  display: inline-block;
  width: 52px;
  height: 28px;
  cursor: pointer;
  flex-shrink: 0;
}

.admin-theme-toggle input {
  opacity: 0;
  width: 0;
  height: 0;
}

.admin-theme-slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: var(--bg-secondary, #ccc);
  transition: 0.3s;
  border-radius: 28px;
  display: flex;
  align-items: center;
  padding: 2px;
}

.admin-theme-slider:before {
  position: absolute;
  content: "";
  height: 24px;
  width: 24px;
  left: 2px;
  bottom: 2px;
  background-color: white;
  transition: 0.3s;
  border-radius: 50%;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.admin-theme-toggle input:checked + .admin-theme-slider {
  background-color: var(--color-primary, #1976d2);
}

.admin-theme-toggle input:checked + .admin-theme-slider:before {
  transform: translateX(24px);
}

.admin-theme-icon {
  position: absolute;
  width: 16px;
  height: 16px;
  transition: opacity 0.3s;
  z-index: 1;
}

.admin-theme-icon-sun {
  left: 6px;
  opacity: 1;
}

.admin-theme-icon-moon {
  right: 6px;
  opacity: 0;
}

.admin-theme-toggle input:checked + .admin-theme-slider .admin-theme-icon-sun {
  opacity: 0;
}

.admin-theme-toggle input:checked + .admin-theme-slider .admin-theme-icon-moon {
  opacity: 1;
}
</style>
