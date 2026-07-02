import { createRouter, createWebHistory } from 'vue-router'
import { useAuth } from '../composables/useAuth'
import { apiClient } from '../api/client'

// Admin entry router: base /app/admin, served from admin.html.
// Public routes live in the public entry (router/index.ts); cross-entry
// navigation is a full page load via plain <a href>.
const router = createRouter({
  history: createWebHistory('/app/admin'),
  routes: [
    {
      path: '/',
      component: () => import('../layouts/AdminLayout.vue'),
      meta: { requiresAuth: true, requiresAdmin: true },
      children: [
        {
          path: '',
          name: 'Admin',
          component: () => import('../views/AdminView.vue')
        },
        {
          path: 'verb-forms',
          name: 'AdminVerbTraining',
          component: () => import('../views/AdminVerbTrainingView.vue')
        },
        {
          path: 'circuit-breaker',
          name: 'AdminCircuitBreaker',
          component: () => import('../views/AdminCircuitBreakerView.vue')
        },
        {
          path: 'conversations',
          name: 'AdminConversations',
          component: () => import('../views/AdminConversationsView.vue')
        },
        {
          path: 'picture-quests',
          name: 'AdminPictureQuests',
          component: () => import('../views/AdminPictureQuestsView.vue')
        },
        {
          path: 'content-reports',
          name: 'AdminContentReports',
          component: () => import('../views/AdminContentReportsView.vue')
        },
        {
          path: 'sentence-composition',
          name: 'AdminSentenceComposition',
          component: () => import('../views/AdminSentenceCompositionView.vue')
        },
        {
          path: 'grammar',
          name: 'AdminGrammar',
          component: () => import('../views/AdminGrammarView.vue')
        },
        {
          path: 'reading-texts',
          name: 'AdminReadingTexts',
          component: () => import('../views/AdminReadingTextsView.vue')
        },
        {
          path: 'word-sets',
          redirect: '/word-sets/categories'
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
          path: 'app-settings',
          name: 'AdminAppSettings',
          component: () => import('../views/AdminAppSettingsView.vue')
        },
        {
          path: 'linglow-srs',
          name: 'AdminLinglowSRS',
          component: () => import('../views/AdminLinglowSRSView.vue')
        },
        {
          path: 'lumi-facts',
          name: 'AdminLumiFacts',
          component: () => import('../views/AdminLumiFactsView.vue')
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
        },
        {
          path: 'stats',
          name: 'AdminStats',
          component: () => import('../views/AdminStatsView.vue')
        },
        {
          path: 'help',
          name: 'AdminHelp',
          component: () => import('../views/AdminHelpView.vue')
        }
      ]
    },
    {
      // Anything unknown inside the admin subtree goes back to the admin home
      path: '/:pathMatch(.*)*',
      redirect: '/'
    }
  ]
})

router.beforeEach(async (to, _from, next) => {
  const { isAuthenticated, hasAnyAdminAccess, checkAuth, loadPermissions } = useAuth()

  await checkAuth()

  if (!isAuthenticated.value) {
    // Login lives in the public entry — full page navigation
    window.location.href = '/app/login'
    return
  }

  await loadPermissions()
  if (!hasAnyAdminAccess()) {
    window.location.href = '/app/dashboard'
    return
  }

  if (to.name === 'AdminVerbTraining') {
    try {
      const settings = await apiClient.request<{ learning?: { spanish_verb_forms_enabled?: boolean } }>('/api/settings')
      if (!settings?.learning?.spanish_verb_forms_enabled) {
        next('/')
        return
      }
    } catch (_error) {
      next('/')
      return
    }
  }

  next()
})

export default router
