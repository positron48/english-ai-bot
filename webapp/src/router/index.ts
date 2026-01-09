import { createRouter, createWebHashHistory } from 'vue-router'
import { useAuth } from '../composables/useAuth'

const router = createRouter({
  history: createWebHashHistory(),
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
      name: 'Admin',
      component: () => import('../views/AdminView.vue'),
      meta: { requiresAuth: true, requiresAdmin: true }
    },
    {
      path: '/admin/circuit-breaker',
      name: 'AdminCircuitBreaker',
      component: () => import('../views/AdminCircuitBreakerView.vue'),
      meta: { requiresAuth: true, requiresAdmin: true }
    },
    {
      path: '/admin/prompt-tester',
      name: 'AdminPromptTester',
      component: () => import('../views/AdminPromptTesterView.vue'),
      meta: { requiresAuth: true, requiresAdmin: true }
    },
    {
      path: '/admin/orphaned-cards',
      name: 'AdminOrphanedCards',
      component: () => import('../views/AdminOrphanedCardsView.vue'),
      meta: { requiresAuth: true, requiresAdmin: true }
    }
  ]
})

router.beforeEach((to, _from, next) => {
  const { isAuthenticated, isAdmin } = useAuth()
  
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
  
  if (to.meta.requiresAuth && !isAuthenticated.value) {
    next('/login')
  } else if (to.meta.requiresAdmin && !isAdmin.value) {
    next('/dashboard')
  } else {
    next()
  }
})

export default router

