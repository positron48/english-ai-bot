import {
  EVENT_DIRTY_TAGS,
  screensForTags,
  type AppDataDirtyTag,
  type AppDataScreenKey,
} from './appDataCacheLogic'
import { markScreensDirty, resolveUserScopeFromStorage, getAppDataLocale } from './appDataCache'

export type AppDataEvent = keyof typeof EVENT_DIRTY_TAGS

type AppDataListener = (event: AppDataEvent, tags: AppDataDirtyTag[], courseCode: string) => void

const listeners = new Set<AppDataListener>()
let activeCourseCode = ''
let refreshHandler: ((screens: AppDataScreenKey[], courseCode: string, reason: string) => void) | null = null
let debounceTimer: ReturnType<typeof setTimeout> | null = null
const pendingScreens = new Set<AppDataScreenKey>()

export function setActiveCourseCodeForInvalidation(courseCode: string): void {
  activeCourseCode = courseCode
}

export function getActiveCourseCodeForInvalidation(): string {
  return activeCourseCode
}

export function onAppDataEvent(listener: AppDataListener): () => void {
  listeners.add(listener)
  return () => listeners.delete(listener)
}

export function emitAppDataEvent(event: AppDataEvent, courseCode?: string, extraTags: AppDataDirtyTag[] = []): void {
  const tags = [...(EVENT_DIRTY_TAGS[event] || []), ...extraTags]
  const code = courseCode || activeCourseCode
  if (!code) return
  for (const listener of listeners) listener(event, tags, code)
  void markScreensDirty(code, tags, resolveUserScopeFromStorage(), getAppDataLocale())
  if (refreshHandler && typeof navigator !== 'undefined' && navigator.onLine !== false) {
    for (const screen of screensForTags(tags)) pendingScreens.add(screen)
    if (debounceTimer) clearTimeout(debounceTimer)
    debounceTimer = setTimeout(() => {
      const screens = [...pendingScreens]
      pendingScreens.clear()
      debounceTimer = null
      if (screens.length > 0) refreshHandler?.(screens, code, event)
    }, 400)
  }
}

export function registerAppDataRefreshHandler(
  handler: (screens: AppDataScreenKey[], courseCode: string, reason: string) => void,
): void {
  refreshHandler = handler
}

export function initAppDataInvalidation(): void {
  // Background invalidation refresh is wired in main.ts via registerAppDataRefreshHandler.
}
