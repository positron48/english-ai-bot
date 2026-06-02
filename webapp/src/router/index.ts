import { createRouter, createWebHistory } from 'vue-router'
import { useAuth } from '../composables/useAuth'
import { apiClient } from '../api/client'
import { grammarClient } from '../api/grammarClient'

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
      component: () => import('../views/GrammarCategoriesView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/learning/grammar/placement-test',
      name: 'GrammarPlacementTest',
      component: () => import('../views/GrammarPlacementTestView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/learning/grammar/training',
      name: 'GrammarTraining',
      component: () => import('../views/GrammarTrainingView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/learning/grammar/:sectionId',
      name: 'GrammarChapters',
      component: () => import('../views/GrammarChaptersView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/learning/grammar/:sectionId/test',
      name: 'GrammarCategoryTest',
      component: () => import('../views/GrammarTestView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/learning/grammar/chapter/:chapterId',
      name: 'GrammarChapter',
      component: () => import('../views/GrammarChapterView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/learning/reading',
      name: 'ReadingCategories',
      component: () => import('../views/ReadingCategoriesView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/learning/reading/category/:categoryId',
      name: 'ReadingChapters',
      component: () => import('../views/ReadingChaptersView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/learning/reading/text/:textId',
      name: 'ReadingText',
      component: () => import('../views/ReadingTextView.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/learning/speaking',
      name: 'SpeakingCategories',
      component: () => import('../views/SpeakingCategoriesView.vue'),
      meta: { requiresAuth: true, requiresSpeaking: true }
    },
    {
      path: '/learning/speaking/session/:sessionId',
      name: 'SpeakingSession',
      component: () => import('../views/SpeakingSessionView.vue'),
      meta: { requiresAuth: true, requiresSpeaking: true }
    },
    {
      path: '/learning/grammar/chapter/:chapterId/test',
      name: 'GrammarChapterTest',
      component: () => import('../views/GrammarTestView.vue'),
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
      path: '/training/verbs',
      name: 'VerbTraining',
      component: () => import('../views/VerbTrainingView.vue'),
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
          path: 'prompt-tester',
          name: 'AdminPromptTester',
          component: () => import('../views/AdminPromptTesterView.vue')
        },
        {
          path: 'content-reports',
          name: 'AdminContentReports',
          component: () => import('../views/AdminContentReportsView.vue')
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
          path: 'app-settings',
          name: 'AdminAppSettings',
          component: () => import('../views/AdminAppSettingsView.vue')
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
    if (typeof navigator !== 'undefined' && navigator.onLine === false) {
      next('/learning/grammar')
      return
    }
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

  const isOffline = typeof navigator !== 'undefined' && navigator.onLine === false
  const isOfflineAllowedBaseRoute = typeof to.name === 'string' && [
    'Dashboard',
    'Training',
    'Learning',
    'LearningGrammar',
    'GrammarChapters',
    'GrammarChapter',
    'GrammarChapterTest',
    'GrammarCategoryTest',
    'GrammarPlacementTest',
    'GrammarTraining',
  ].includes(to.name)
  const isOfflineAllowedRoute = isOfflineAllowedBaseRoute
  if (isOffline && to.meta.requiresAuth && !isOfflineAllowedRoute) {
    next(to.path.startsWith('/learning') ? '/learning' : '/dashboard')
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

    if (to.name === 'AdminVerbTraining') {
      try {
        const settings = await apiClient.request<{ learning?: { spanish_verb_forms_enabled?: boolean } }>('/api/settings')
        if (!settings?.learning?.spanish_verb_forms_enabled) {
          next('/admin')
          return
        }
      } catch (_error) {
        next('/admin')
        return
      }
    }
  }
  
  // Check grammar chapter access (guard client-side, backend will enforce too)
  if (to.name === 'GrammarChapter' && to.params.chapterId) {
    try {
      const chapterId = to.params.chapterId as string
      const response: { can_access: boolean } = await grammarClient.canAccessChapter(chapterId)
      if (!response.can_access) {
        // Extract sectionId from chapterId (format: section.chapter)
        const sectionMatch = chapterId.match(/^(.+)\.[^.]+$/)
        if (sectionMatch) {
          const sectionId = sectionMatch[1]
          next({
            path: `/learning/grammar/${sectionId}`,
            query: { error: 'previous_chapter_not_passed' }
          })
        } else {
          next('/learning/grammar')
        }
        return
      }
    } catch (error: any) {
      console.error('Failed to check chapter access:', error)
    }
  }
  
  // Check grammar section access
  if (to.name === 'GrammarChapters' && to.params.sectionId) {
    try {
      const sectionId = to.params.sectionId as string
      const response: { can_access: boolean } = await grammarClient.canAccessSection(sectionId)
      if (!response.can_access) {
        next({
          path: '/learning/grammar',
          query: { error: 'previous_section_not_complete' }
        })
        return
      }
    } catch (error: any) {
      console.error('Failed to check section access:', error)
    }
  }

  if (to.name === 'GrammarTraining') {
    try {
      const response: any = await grammarClient.getTrainingAvailability()
      if (!response?.grammar_training?.available) {
        next('/learning/grammar')
        return
      }
    } catch (error: any) {
      console.error('Failed to check grammar training availability:', error)
      next('/learning/grammar')
      return
    }
  }

  if (to.meta.requiresSpeaking) {
    try {
      const response: any = await apiClient.request('/api/learning/speaking/availability')
      if (!response?.can_access || !response?.available) {
        next('/learning')
        return
      }
    } catch (error: any) {
      console.error('Failed to check speaking availability:', error)
      next('/learning')
      return
    }
  }
  
  next()
})

export default router
