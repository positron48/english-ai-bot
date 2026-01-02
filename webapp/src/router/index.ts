import { createRouter, createWebHashHistory } from 'vue-router'
import { useAuth } from '../composables/useAuth'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/',
      redirect: '/dashboard'
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
    }
  ]
})

router.beforeEach((to, from, next) => {
  const { isAuthenticated, isAdmin } = useAuth()
  
  console.log('[Router] Navigation:', {
    from: from.path,
    to: to.path,
    isAuthenticated: isAuthenticated.value,
    isAdmin: isAdmin.value,
    requiresAuth: to.meta.requiresAuth,
    requiresAdmin: to.meta.requiresAdmin
  })
  
  if (to.meta.requiresAuth && !isAuthenticated.value) {
    console.log('[Router] Redirecting to /login (not authenticated)')
    next('/login')
  } else if (to.meta.requiresAdmin && !isAdmin.value) {
    console.log('[Router] Redirecting to /dashboard (not admin)')
    next('/dashboard')
  } else {
    console.log('[Router] Allowing navigation to', to.path)
    next()
  }
})

export default router

