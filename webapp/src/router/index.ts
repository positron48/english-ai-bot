import { createRouter, createWebHistory } from 'vue-router'
import { useAuth } from '../composables/useAuth'
import { apiClient } from '../api/client'

const router = createRouter({
  history: createWebHistory('/app'),
  routes: [
    {
      path: '/',
      redirect: '/login'  // Start with login, let auth handle redirect
    },
    {
      path: '/login',
      name: 'Login',
      component: () => import('../views/LoginView.vue'),
      meta: { requiresAuth: false }
    },
    {
      path: '/dashboard',
      name: 'Dashboard',
      component: () => import('../views/DashboardView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/vocab',
      name: 'Vocab',
      component: () => import('../views/VocabView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/learning',
      name: 'Learning',
      component: () => import('../views/LearningView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/learning/grammar',
      name: 'LearningGrammar',
      component: () => import('../views/LearningGrammarView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/learning/words',
      name: 'WordSets',
      component: () => import('../views/WordSetsView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/learning/words/:setId',
      name: 'WordSetDetail',
      component: () => import('../views/WordSetDetailView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/learning/words/:setId/study',
      name: 'WordSetStudy',
      component: () => import('../views/WordSetStudyView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/training',
      name: 'Training',
      component: () => import('../views/TrainingView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/chat',
      name: 'Chat',
      component: () => import('../views/ChatView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/admin',
      component: () => import('../layouts/AdminLayout.vue'),
      meta: { requiresAuth: true, requiresAdmin: true },
      children: [
        {
          path: '',
          name: 'Admin',
          component: () => import('../views/AdminView.vue')
        },
        {
          path: 'circuit-breaker',
          name: 'AdminCircuitBreaker',
          component: () => import('../views/AdminCircuitBreakerView.vue')
        },
        {
          path: 'prompt-tester',
          name: 'AdminPromptTester',
          component: () => import('../views/AdminPromptTesterView.vue')
        },
        {
          path: 'orphaned-cards',
          name: 'AdminOrphanedCards',
          component: () => import('../views/AdminOrphanedCardsView.vue')
        },
        {
          path: 'word-sets',
          redirect: '/admin/word-sets/categories'
        },
        {
          path: 'word-sets/categories',
          name: 'AdminWordSetsCategories',
          component: () => import('../views/AdminWordSetsView.vue')
        },
        {
          path: 'word-sets/sets',
          name: 'AdminWordSetsSets',
          component: () => import('../views/AdminWordSetsView.vue')
        },
        {
          path: 'db-schema',
          name: 'AdminDBSchema',
          component: () => import('../views/AdminDBSchemaView.vue')
        },
        {
          path: 'access',
          name: 'AdminAccess',
          component: () => import('../views/AdminAccessView.vue')
        },
        {
          path: 'users',
          name: 'AdminUsers',
          component: () => import('../views/AdminUsersView.vue')
        }
      ]
    },
    {
      path: '/settings',
      name: 'Settings',
      component: () => import('../views/SettingsView.vue'),
      meta: { requiresAuth: true }
    }
  ]
})

router.beforeEach(async (to, _from, next) => {
  // Clean path if it contains tgWebAppData (shouldn't happen, but just in case)
  let cleanPath = to.path
  if (cleanPath.includes('tgWebAppData')) {
    cleanPath = '/'
  }
  
  // If path was cleaned, redirect
  if (cleanPath !== to.path) {
    next(cleanPath)
    return
  }
  
  // Get auth state - this will check tokens from localStorage
  const { isAuthenticated, hasAnyAdminAccess, checkAuth, loadPermissions } = useAuth()
  
  // Ensure auth state is up to date (especially important on direct URL access)
  // Reload tokens from localStorage and update auth state
  await checkAuth()
  
  // If user is trying to access login page, check if they're already authenticated
  // by making a request to the backend
  if (to.path === '/login' && isAuthenticated.value) {
    try {
      // Check authentication via backend request
      await apiClient.request('/api/dashboard')
      // If request succeeds, user is authenticated, redirect to dashboard
      next('/dashboard')
      return
    } catch (error: any) {
      // If request fails (401 or network error), user is not authenticated
      // Continue to login page
      // Clear invalid tokens if it's an auth error
      if (error.message?.includes('401') || error.message?.includes('Unauthorized')) {
        apiClient.clearTokens()
        await checkAuth()
      }
    }
  }
  
  // Check authentication
  if (to.meta.requiresAuth && !isAuthenticated.value) {
    next('/login')
    return
  }
  
  // Check admin access (for /admin routes)
  if (to.meta.requiresAdmin) {
    // Load permissions if not already loaded
    if (isAuthenticated.value) {
      await loadPermissions()
    }
    
    if (!hasAnyAdminAccess()) {
      next('/dashboard')
      return
    }
  }
  
  next()
})

export default router

