<template>
  <nav v-if="breadcrumbs.length > 0" class="breadcrumbs">
    <ol class="breadcrumbs-list">
      <li v-for="(crumb, index) in breadcrumbs" :key="index" class="breadcrumb-item">
        <router-link 
          v-if="index < breadcrumbs.length - 1" 
          :to="crumb.path" 
          class="breadcrumb-link"
        >
          <Icon v-if="index === 0" name="home" class="breadcrumb-icon" />
          <span>{{ crumb.label }}</span>
        </router-link>
        <span v-else class="breadcrumb-current">
          {{ crumb.label }}
        </span>
        <Icon 
          v-if="index < breadcrumbs.length - 1" 
          name="chevron-right" 
          class="breadcrumb-separator" 
        />
      </li>
    </ol>
  </nav>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import Icon from './Icon.vue'

interface Breadcrumb {
  label: string
  path: string
}

const route = useRoute()

// Определяем иерархию маршрутов
const routeHierarchy: Record<string, Breadcrumb[]> = {
  '/dashboard': [
    { label: 'Dashboard', path: '/dashboard' }
  ],
  '/vocab': [
    { label: 'Dashboard', path: '/dashboard' },
    { label: 'Vocabulary', path: '/vocab' }
  ],
  '/learning': [
    { label: 'Dashboard', path: '/dashboard' },
    { label: 'Learning', path: '/learning' }
  ],
  '/learning/grammar': [
    { label: 'Dashboard', path: '/dashboard' },
    { label: 'Learning', path: '/learning' },
    { label: 'Grammar', path: '/learning/grammar' }
  ],
  '/learning/words': [
    { label: 'Dashboard', path: '/dashboard' },
    { label: 'Learning', path: '/learning' },
    { label: 'Word Sets', path: '/learning/words' }
  ],
  '/chat': [
    { label: 'Dashboard', path: '/dashboard' },
    { label: 'Chat', path: '/chat' }
  ],
  '/settings': [
    { label: 'Dashboard', path: '/dashboard' },
    { label: 'Settings', path: '/settings' }
  ],
  '/admin': [
    { label: 'Dashboard', path: '/dashboard' },
    { label: 'Admin', path: '/admin' }
  ],
  '/admin/circuit-breaker': [
    { label: 'Dashboard', path: '/dashboard' },
    { label: 'Admin', path: '/admin' },
    { label: 'Circuit Breaker', path: '/admin/circuit-breaker' }
  ],
  '/admin/prompt-tester': [
    { label: 'Dashboard', path: '/dashboard' },
    { label: 'Admin', path: '/admin' },
    { label: 'Prompt Tester', path: '/admin/prompt-tester' }
  ],
  '/admin/orphaned-cards': [
    { label: 'Dashboard', path: '/dashboard' },
    { label: 'Admin', path: '/admin' },
    { label: 'Orphaned Cards', path: '/admin/orphaned-cards' }
  ],
  '/admin/word-sets': [
    { label: 'Dashboard', path: '/dashboard' },
    { label: 'Admin', path: '/admin' },
    { label: 'Word Sets', path: '/admin/word-sets' }
  ],
  '/admin/db-schema': [
    { label: 'Dashboard', path: '/dashboard' },
    { label: 'Admin', path: '/admin' },
    { label: 'DB Schema', path: '/admin/db-schema' }
  ]
}

const breadcrumbs = computed(() => {
  const currentPath = route.path
  
  // Страницы без крошек
  if (currentPath === '/training' || currentPath.match(/^\/learning\/words\/\d+\/study$/)) {
    return []
  }
  
  // Для динамических маршрутов типа /learning/words/:setId
  if (currentPath.match(/^\/learning\/words\/\d+$/) && !currentPath.endsWith('/study')) {
    const setId = currentPath.split('/').pop()
    return [
      { label: 'Dashboard', path: '/dashboard' },
      { label: 'Learning', path: '/learning' },
      { label: 'Word Sets', path: '/learning/words' },
      { label: `Word Set #${setId}`, path: currentPath }
    ]
  }
  
  // Для статических маршрутов
  return routeHierarchy[currentPath] || []
})
</script>

<style scoped>
.breadcrumbs {
  padding: 12px 0;
  margin-bottom: 16px;
  border-bottom: 1px solid var(--border-primary);
}

.breadcrumbs-list {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  list-style: none;
  margin: 0;
  padding: 0;
}

.breadcrumb-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.breadcrumb-link {
  display: flex;
  align-items: center;
  gap: 4px;
  color: var(--text-secondary);
  text-decoration: none;
  font-size: 14px;
  transition: color 0.2s;
  padding: 4px 0;
}

.breadcrumb-link:hover {
  color: var(--color-primary);
}

.breadcrumb-icon {
  font-size: 16px;
  line-height: 1;
}

.breadcrumb-current {
  color: var(--text-primary);
  font-size: 14px;
  font-weight: 500;
  padding: 4px 0;
}

.breadcrumb-separator {
  font-size: 12px;
  color: var(--text-secondary);
  opacity: 0.6;
  line-height: 1;
}

@media (max-width: 768px) {
  .breadcrumbs {
    padding: 8px 0;
    margin-bottom: 12px;
  }
  
  .breadcrumb-link,
  .breadcrumb-current {
    font-size: 12px;
  }
  
  .breadcrumb-icon {
    font-size: 14px;
  }
  
  .breadcrumb-separator {
    font-size: 10px;
  }
}
</style>
