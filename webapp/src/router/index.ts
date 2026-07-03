import { createRouter, createWebHistory } from 'vue-router'
import { useAuth } from '../composables/useAuth'
import { apiClient } from '../api/client'
import { grammarClient } from '../api/grammarClient'

// Public entry router. Admin routes live in the separate admin entry
// (router/admin.ts, served from admin.html at /app/admin).
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
      path: '/',
      component: () => import('../layouts/PublicLayout.vue'),
      children: [
        {
          path: 'dashboard',
          name: 'Dashboard',
          component: () => import('../views/DashboardView.vue'),
          meta: { requiresAuth: true, navTab: 'home' }
        },
        {
          path: 'city',
          name: 'City',
          component: () => import('../views/CityMapView.vue'),
          meta: { requiresAuth: true, navTab: 'city' }
        },
        {
          path: 'city/hub',
          name: 'CityHub',
          component: () => import('../views/CityView.vue'),
          meta: { requiresAuth: true, navTab: 'city' }
        },
        {
          path: 'city/daily-route',
          name: 'CityDailyRoute',
          component: () => import('../views/CityDailyRouteView.vue'),
          meta: { requiresAuth: true, navTab: 'city' }
        },
        {
          path: 'city/district/:districtCode',
          name: 'CityDistrict',
          component: () => import('../views/CityDistrictView.vue'),
          meta: { requiresAuth: true, navTab: 'city' }
        },
        {
          path: 'city/district/:districtCode/chat',
          name: 'PlaceChatList',
          component: () => import('../views/PlaceChatView.vue'),
          meta: { requiresAuth: true, navTab: 'city' }
        },
        {
          path: 'city/district/:districtCode/chat/:scenarioCode',
          name: 'PlaceChat',
          component: () => import('../views/PlaceChatView.vue'),
          meta: { requiresAuth: true, navTab: 'city' }
        },
        {
          path: 'vocab',
          name: 'Vocab',
          component: () => import('../views/VocabView.vue'),
          meta: { requiresAuth: true, navTab: 'practice' }
        },
        {
          path: 'learning',
          name: 'Learning',
          component: () => import('../views/LearningView.vue'),
          meta: { requiresAuth: true, navTab: 'practice' }
        },
        {
          path: 'learning/conversations',
          name: 'ConversationNpcList',
          component: () => import('../views/ConversationNpcListView.vue'),
          meta: { requiresAuth: true, navTab: 'practice' }
        },
        {
          path: 'learning/picture-quests',
          name: 'PictureQuestList',
          component: () => import('../views/PictureQuestListView.vue'),
          meta: { requiresAuth: true, navTab: 'practice' }
        },
        {
          path: 'learning/picture-quests/archive',
          name: 'PictureQuestArchive',
          component: () => import('../views/PictureQuestArchiveView.vue'),
          meta: { requiresAuth: true, navTab: 'practice' }
        },
        {
          path: 'learning/picture-quests/:questCode',
          name: 'PictureQuestChat',
          component: () => import('../views/PictureQuestChatView.vue'),
          meta: { requiresAuth: true, navTab: 'practice' }
        },
        {
          path: 'learning/grammar',
          name: 'LearningGrammar',
          component: () => import('../views/GrammarCategoriesView.vue'),
          meta: { requiresAuth: true, navTab: 'practice' }
        },
        {
          path: 'learning/grammar/placement-test',
          name: 'GrammarPlacementTest',
          component: () => import('../views/GrammarPlacementTestView.vue'),
          meta: { requiresAuth: true, navTab: 'practice', fullscreen: true }
        },
        {
          path: 'learning/grammar/training',
          name: 'GrammarTraining',
          component: () => import('../views/GrammarTrainingView.vue'),
          meta: { requiresAuth: true, navTab: 'practice' }
        },
        {
          path: 'learning/grammar/:sectionId',
          name: 'GrammarChapters',
          component: () => import('../views/GrammarChaptersView.vue'),
          meta: { requiresAuth: true, navTab: 'practice' }
        },
        {
          path: 'learning/grammar/:sectionId/test',
          name: 'GrammarCategoryTest',
          component: () => import('../views/GrammarTestView.vue'),
          meta: { requiresAuth: true, navTab: 'practice', fullscreen: true }
        },
        {
          path: 'learning/grammar/chapter/:chapterId',
          name: 'GrammarChapter',
          component: () => import('../views/GrammarChapterView.vue'),
          meta: { requiresAuth: true, navTab: 'practice' }
        },
        {
          path: 'learning/grammar/chapter/:chapterId/test',
          name: 'GrammarChapterTest',
          component: () => import('../views/GrammarTestView.vue'),
          meta: { requiresAuth: true, navTab: 'practice', fullscreen: true }
        },
        {
          path: 'learning/reading',
          name: 'ReadingCategories',
          component: () => import('../views/ReadingCategoriesView.vue'),
          meta: { requiresAuth: true, navTab: 'practice' }
        },
        {
          path: 'learning/reading/category/:categoryId',
          name: 'ReadingChapters',
          component: () => import('../views/ReadingChaptersView.vue'),
          meta: { requiresAuth: true, navTab: 'practice' }
        },
        {
          path: 'learning/reading/category/:categoryId/archive',
          name: 'ReadingArchive',
          component: () => import('../views/ReadingArchiveView.vue'),
          meta: { requiresAuth: true, navTab: 'practice' }
        },
        {
          path: 'learning/reading/text/:textId',
          name: 'ReadingText',
          component: () => import('../views/ReadingTextView.vue'),
          meta: { requiresAuth: true, navTab: 'practice' }
        },
        {
          path: 'learning/speaking',
          name: 'SpeakingCategories',
          component: () => import('../views/SpeakingCategoriesView.vue'),
          meta: { requiresAuth: true, requiresSpeaking: true, navTab: 'practice' }
        },
        {
          path: 'learning/speaking/session/:sessionId',
          name: 'SpeakingSession',
          component: () => import('../views/SpeakingSessionView.vue'),
          meta: { requiresAuth: true, requiresSpeaking: true, navTab: 'practice', fullscreen: true }
        },
        {
          path: 'learning/words',
          name: 'WordSets',
          component: () => import('../views/WordSetsView.vue'),
          meta: { requiresAuth: true, navTab: 'practice' }
        },
        {
          path: 'learning/words/:setId',
          name: 'WordSetDetail',
          component: () => import('../views/WordSetDetailView.vue'),
          meta: { requiresAuth: true, navTab: 'practice' }
        },
        {
          path: 'learning/words/:setId/study',
          name: 'WordSetStudy',
          component: () => import('../views/WordSetStudyView.vue'),
          meta: { requiresAuth: true, navTab: 'practice', fullscreen: true }
        },
        {
          path: 'training',
          name: 'Training',
          component: () => import('../views/TrainingView.vue'),
          meta: { requiresAuth: true, navTab: 'practice' }
        },
        {
          path: 'training/verbs',
          name: 'VerbTraining',
          component: () => import('../views/VerbTrainingView.vue'),
          meta: { requiresAuth: true, navTab: 'practice' }
        },
        {
          path: 'training/sentences',
          name: 'SentenceComposition',
          component: () => import('../views/SentenceCompositionView.vue'),
          meta: { requiresAuth: true, navTab: 'practice' }
        },
        {
          path: 'chat',
          name: 'Chat',
          component: () => import('../views/ChatView.vue'),
          meta: { requiresAuth: true, navTab: 'practice', fullscreen: true }
        },
        {
          path: 'progress',
          name: 'Progress',
          component: () => import('../views/ProgressView.vue'),
          meta: { requiresAuth: true, navTab: 'progress' }
        },
        {
          path: 'settings',
          name: 'Settings',
          component: () => import('../views/SettingsView.vue'),
          meta: { requiresAuth: true, navTab: 'profile' }
        }
      ]
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
  const { isAuthenticated, checkAuth } = useAuth()

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

  // General AI chat section is closed for everyone — redirect to dashboard.
  if (to.name === 'Chat') {
    next('/dashboard')
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
