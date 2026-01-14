<template>
  <!-- Mobile menu button -->
  <button 
    v-if="isMobile" 
    @click="showSidebar = true" 
    class="admin-menu-toggle"
    aria-label="Open admin menu"
  >
    <Icon name="menu" />
  </button>

  <!-- Sidebar overlay for mobile -->
  <div 
    v-if="isMobile && showSidebar" 
    class="admin-sidebar-overlay"
    @click="showSidebar = false"
  ></div>

  <!-- Sidebar -->
  <aside 
    class="admin-sidebar" 
    :class="{ 'open': showSidebar || !isMobile }"
  >
    <div class="admin-sidebar-header">
      <h1>Admin Panel</h1>
      <button 
        v-if="isMobile" 
        @click="showSidebar = false" 
        class="admin-sidebar-close"
        aria-label="Close admin menu"
      >
        <Icon name="close" />
      </button>
    </div>
    
    <nav class="admin-sidebar-nav">
      <router-link 
        to="/admin" 
        class="admin-sidebar-item" 
        :class="{ active: $route.path === '/admin' }"
        @click="isMobile && (showSidebar = false)"
      >
        <Icon name="home" class="admin-sidebar-icon" />
        <span>Main</span>
      </router-link>
      <router-link 
        to="/admin/circuit-breaker" 
        class="admin-sidebar-item" 
        :class="{ active: $route.path === '/admin/circuit-breaker' }"
        @click="isMobile && (showSidebar = false)"
      >
        <Icon name="settings" class="admin-sidebar-icon" />
        <span>Circuit Breaker</span>
      </router-link>
      <router-link 
        to="/admin/prompt-tester" 
        class="admin-sidebar-item" 
        :class="{ active: $route.path === '/admin/prompt-tester' }"
        @click="isMobile && (showSidebar = false)"
      >
        <Icon name="code" class="admin-sidebar-icon" />
        <span>Prompt Tester</span>
      </router-link>
      <router-link 
        to="/admin/orphaned-cards" 
        class="admin-sidebar-item" 
        :class="{ active: $route.path === '/admin/orphaned-cards' }"
        @click="isMobile && (showSidebar = false)"
      >
        <Icon name="alert" class="admin-sidebar-icon" />
        <span>Orphaned Cards</span>
      </router-link>
      <router-link 
        to="/admin/word-sets" 
        class="admin-sidebar-item" 
        :class="{ active: $route.path === '/admin/word-sets' }"
        @click="isMobile && (showSidebar = false)"
      >
        <Icon name="book" class="admin-sidebar-icon" />
        <span>Word Sets</span>
      </router-link>
      <router-link 
        to="/admin/db-schema" 
        class="admin-sidebar-item" 
        :class="{ active: $route.path === '/admin/db-schema' }"
        @click="isMobile && (showSidebar = false)"
      >
        <Icon name="database" class="admin-sidebar-icon" />
        <span>DB Schema</span>
      </router-link>
    </nav>
  </aside>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import Icon from './Icon.vue'

const showSidebar = ref(false)
const windowWidth = ref(window.innerWidth)

const isMobile = computed(() => windowWidth.value < 768)

const updateWindowWidth = () => {
  windowWidth.value = window.innerWidth
  // Auto-close sidebar on mobile when resizing to desktop
  if (!isMobile.value) {
    showSidebar.value = false
  }
}

onMounted(() => {
  window.addEventListener('resize', updateWindowWidth)
})

onUnmounted(() => {
  window.removeEventListener('resize', updateWindowWidth)
})
</script>

<style scoped>
/* Mobile menu toggle button */
.admin-menu-toggle {
  position: fixed;
  top: 70px;
  left: 10px;
  z-index: 1001;
  background: var(--bg-primary);
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  padding: 10px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transition: all 0.2s;
}

.admin-menu-toggle:hover {
  background: var(--bg-hover);
}

.admin-menu-toggle svg {
  width: 24px;
  height: 24px;
  color: var(--text-primary);
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
}

.admin-sidebar.open {
  transform: translateX(0);
}

/* Desktop: always visible */
@media (min-width: 768px) {
  .admin-sidebar {
    position: relative;
    transform: translateX(0);
    height: auto;
    min-height: 100vh;
    box-shadow: none;
  }
  
  .admin-menu-toggle {
    display: none;
  }
  
  .admin-sidebar-overlay {
    display: none;
  }
}

.admin-sidebar-header {
  padding: 20px;
  border-bottom: 1px solid var(--border-primary);
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
}

.admin-sidebar-header h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: var(--text-primary);
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
  padding: 12px 0;
  overflow-y: auto;
}

.admin-sidebar-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 20px;
  text-decoration: none;
  color: var(--text-secondary);
  transition: all 0.2s;
  border-left: 3px solid transparent;
}

.admin-sidebar-item:hover {
  color: var(--text-primary);
  background: var(--bg-hover);
}

.admin-sidebar-item.active {
  color: var(--color-primary);
  background: var(--bg-hover);
  border-left-color: var(--color-primary);
  font-weight: 500;
}

.admin-sidebar-icon {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.admin-sidebar-item span {
  flex: 1;
}
</style>
