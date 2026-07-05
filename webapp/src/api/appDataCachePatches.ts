import { patchCachedScreen, getAppDataLocale } from './appDataCache'

type DashboardOverviewPayload = {
  dashboard?: {
    due_count?: number
    available_for_training?: number
    new_count?: number
    learning_count?: number
    review_count?: number
  }
  daily_route?: {
    today?: {
      reading_done?: boolean
      words_done?: number
    }
  }
  progress?: unknown
}

type ProgressOverviewPayload = {
  stats?: {
    month?: {
      texts_read?: number
    }
  }
}

type LearningOverviewPayload = {
  continue_chapter?: unknown
}

export async function patchDashboardAfterWordAttempt(courseCode: string): Promise<void> {
  const locale = getAppDataLocale()
  await patchCachedScreen<DashboardOverviewPayload>('dashboard', courseCode, (payload) => {
    const dashboard = { ...(payload.dashboard || {}) }
    if (typeof dashboard.due_count === 'number' && dashboard.due_count > 0) dashboard.due_count -= 1
    if (typeof dashboard.available_for_training === 'number' && dashboard.available_for_training > 0) {
      dashboard.available_for_training -= 1
    }
    return { ...payload, dashboard }
  }, { addDirtyTags: ['words', 'srs', 'daily-route', 'progress'], locale })
}

export async function patchDailyRouteAfterReadingDone(courseCode: string): Promise<void> {
  const locale = getAppDataLocale()
  await patchCachedScreen<DashboardOverviewPayload>('dashboard', courseCode, (payload) => {
    const dailyRoute = { ...(payload.daily_route || {}) }
    const today = { ...(dailyRoute.today || {}), reading_done: true }
    return { ...payload, daily_route: { ...dailyRoute, today } }
  }, { addDirtyTags: ['reading', 'daily-route'], locale })

  await patchCachedScreen<{ route?: { today?: { reading_done?: boolean } } }>('daily-route', courseCode, (payload) => ({
    ...payload,
    route: {
      ...(payload.route || {}),
      today: { ...(payload.route?.today || {}), reading_done: true },
    },
  }), { addDirtyTags: ['reading', 'daily-route'], locale })
}

export async function patchProgressStatsAfterReadingDone(courseCode: string): Promise<void> {
  const locale = getAppDataLocale()
  await patchCachedScreen<ProgressOverviewPayload>('progress', courseCode, (payload) => {
    const stats = { ...(payload.stats || {}) }
    const month = { ...(stats.month || {}) }
    month.texts_read = (month.texts_read ?? 0) + 1
    return { ...payload, stats: { ...stats, month } }
  }, { addDirtyTags: ['reading', 'stats'], locale })
}

export async function patchLearningAfterGrammarSubmit(
  courseCode: string,
  continueChapter: unknown,
): Promise<void> {
  if (!continueChapter) return
  const locale = getAppDataLocale()
  await patchCachedScreen<LearningOverviewPayload>('learning', courseCode, (payload) => ({
    ...payload,
    continue_chapter: continueChapter,
  }), { addDirtyTags: ['grammar'], locale })

  await patchCachedScreen<DashboardOverviewPayload>('dashboard', courseCode, (payload) => ({
    ...payload,
    continue_chapter: continueChapter,
  }), { addDirtyTags: ['grammar'], locale })
}
