import { useSettings } from './useSettings'
import { grammarClient } from '../api/grammarClient'
import { wordTrainingClient } from '../api/wordTrainingClient'

const NEXT_AT_KEY = 'offline.autoDownload.nextAt'
const INTERVAL_MS = 60 * 60 * 1000 // 1 hour

function readNextAt(): number {
  if (typeof window === 'undefined') return 0
  const raw = localStorage.getItem(NEXT_AT_KEY)
  const parsed = raw ? Number(raw) : 0
  return Number.isFinite(parsed) ? parsed : 0
}

function writeNextAt(ts: number): void {
  if (typeof window === 'undefined') return
  try {
    localStorage.setItem(NEXT_AT_KEY, String(ts))
  } catch {
    // ignore storage failures
  }
}

let running = false

/**
 * Background, throttled offline preload. When the auto-download setting is on,
 * preloads grammar (only if not already downloaded) and the current word-training
 * pack, at most once per hour. Fire-and-forget — callers should `void` it.
 */
export async function maybeRunOfflineAutoDownload(): Promise<void> {
  if (running) return
  const { settings } = useSettings()
  if (!settings.value.offlineAutoDownload) return
  if (typeof navigator !== 'undefined' && navigator.onLine === false) return
  if (Date.now() < readNextAt()) return

  // Reserve the window up-front so a slow or failed run still respects throttle.
  writeNextAt(Date.now() + INTERVAL_MS)
  running = true
  try {
    try {
      const grammar = await grammarClient.getOfflineStatus()
      if (!grammar.ready) {
        await grammarClient.preload()
      }
    } catch (e) {
      console.warn('Offline auto-download (grammar) failed:', e)
    }
    try {
      await wordTrainingClient.preload()
    } catch (e) {
      console.warn('Offline auto-download (words) failed:', e)
    }
  } finally {
    running = false
  }
}
