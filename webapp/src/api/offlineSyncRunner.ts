import { grammarClient } from './grammarClient'
import { wordTrainingClient } from './wordTrainingClient'
import { contentReportClient } from './contentReportClient'

let syncInFlight = false
let syncScheduled = false
let lastSyncAt = 0

const MIN_SYNC_INTERVAL_MS = 5_000

async function runOfflineSync(): Promise<void> {
  if (typeof navigator !== 'undefined' && navigator.onLine === false) return
  const now = Date.now()
  if (syncInFlight) {
    syncScheduled = true
    return
  }
  if (now - lastSyncAt < MIN_SYNC_INTERVAL_MS) return

  syncInFlight = true
  try {
    await Promise.all([
      grammarClient.syncQueuedAttempts().catch((error) => {
        console.warn('[PWA] Offline grammar sync failed:', error)
      }),
      wordTrainingClient.syncQueuedAttempts().catch((error) => {
        console.warn('[PWA] Offline word training sync failed:', error)
      }),
      contentReportClient.syncQueuedReports().catch((error) => {
        console.warn('[PWA] Offline content report sync failed:', error)
      }),
    ])
    lastSyncAt = Date.now()
  } finally {
    syncInFlight = false
    if (syncScheduled) {
      syncScheduled = false
      void runOfflineSync()
    }
  }
}

export function scheduleOfflineSync(): void {
  void runOfflineSync()
}

export function initOfflineSyncRunner(): void {
  window.addEventListener('online', scheduleOfflineSync)
  document.addEventListener('visibilitychange', () => {
    if (document.visibilityState === 'visible') scheduleOfflineSync()
  })
  window.setInterval(scheduleOfflineSync, 30_000)
  scheduleOfflineSync()
}
