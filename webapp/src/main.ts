import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import i18n from './i18n'
import { grammarClient } from './api/grammarClient'
import { wordTrainingClient } from './api/wordTrainingClient'
import './styles/theme.css'
import './style.css'
import './styles/markdown-content.css'

declare global {
  interface Window {
    __showQantrixRuntimeDebug?: () => void
  }
}

const OFFLINE_DEBUG_STORAGE_KEY = 'qantrix-offline-debug-state'
const RUNTIME_ERROR_STORAGE_KEY = 'qantrix-runtime-error'

const serializeError = (error: unknown) => {
  if (error instanceof Error) {
    return {
      name: error.name,
      message: error.message,
      stack: error.stack,
    }
  }
  return {
    message: typeof error === 'string' ? error : JSON.stringify(error),
  }
}

const persistRuntimeError = (kind: string, error: unknown, info?: string) => {
  const payload = {
    kind,
    info,
    at: new Date().toISOString(),
    href: window.location.href,
    online: navigator.onLine,
    serviceWorkerControlled: !!navigator.serviceWorker?.controller,
    error: serializeError(error),
    lastOfflineDebug: localStorage.getItem(OFFLINE_DEBUG_STORAGE_KEY) || null,
  }
  localStorage.setItem(RUNTIME_ERROR_STORAGE_KEY, JSON.stringify(payload, null, 2))
  return payload
}

const escapeHtml = (value: string) => value
  .replace(/&/g, '&amp;')
  .replace(/</g, '&lt;')
  .replace(/>/g, '&gt;')
  .replace(/"/g, '&quot;')
  .replace(/'/g, '&#039;')

const showRuntimeDebugOverlay = (payload: unknown) => {
  if (typeof document === 'undefined') return
  const existing = document.getElementById('runtime-debug-overlay')
  if (existing) existing.remove()

  const overlay = document.createElement('pre')
  overlay.id = 'runtime-debug-overlay'
  overlay.textContent = JSON.stringify(payload, null, 2)
  overlay.style.cssText = [
    'position:fixed',
    'inset:12px',
    'z-index:2147483647',
    'margin:0',
    'padding:12px',
    'overflow:auto',
    'white-space:pre-wrap',
    'background:#111827',
    'color:#f9fafb',
    'border:1px solid #374151',
    'border-radius:12px',
    'font:12px/1.45 ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace',
  ].join(';')
  document.body.appendChild(overlay)
}

window.__showQantrixRuntimeDebug = () => {
  const payload = localStorage.getItem(RUNTIME_ERROR_STORAGE_KEY) || localStorage.getItem(OFFLINE_DEBUG_STORAGE_KEY) || '{}'
  document.body.innerHTML = `<pre style="margin:12px;padding:12px;white-space:pre-wrap;background:#111827;color:#f9fafb;border-radius:12px;font:12px/1.45 monospace">${escapeHtml(payload)}</pre>`
}

// Clean up Telegram URL parameters before router initialization
// Telegram adds tgWebAppData to URL which needs to be extracted and removed

// Clean search params from tgWebAppData
if (window.location.search.includes('tgWebAppData')) {
  const url = new URL(window.location.href)
  const tgWebAppData = url.searchParams.get('tgWebAppData')
  if (tgWebAppData) {
    ;(window as any).__tgWebAppData = decodeURIComponent(tgWebAppData)
  }
  url.searchParams.delete('tgWebAppData')
  window.history.replaceState({}, '', url.toString())
}


// Global error handler for Vue
const errorHandler = (err: unknown, _instance: any, info: string) => {
  console.error('[Vue Error]', err, info)
  const payload = persistRuntimeError('vue-error', err, info)
  if (navigator.onLine === false) showRuntimeDebugOverlay(payload)
}

const app = createApp(App)
app.config.errorHandler = errorHandler

// Handle unhandled promise rejections
window.addEventListener('unhandledrejection', (event) => {
  console.error('[Unhandled Promise Rejection]', event.reason)
  const payload = persistRuntimeError('unhandledrejection', event.reason)
  if (navigator.onLine === false) showRuntimeDebugOverlay(payload)
})

window.addEventListener('error', (event) => {
  const error = event.error || event.message
  const payload = persistRuntimeError('window-error', error, `${event.filename}:${event.lineno}:${event.colno}`)
  if (navigator.onLine === false) showRuntimeDebugOverlay(payload)
})

router.onError((error, to) => {
  console.error('[Router Error]', error, to)
  const payload = persistRuntimeError('router-error', error, `to=${to.fullPath}`)
  if (navigator.onLine === false) showRuntimeDebugOverlay(payload)
})

if ('serviceWorker' in navigator) {
  window.addEventListener('load', () => {
    navigator.serviceWorker.register('/sw.js', { scope: '/' }).catch((error) => {
      console.warn('[PWA] Service worker registration failed:', error)
    })
  })
}

const trySyncOfflineGrammar = () => {
  if (typeof navigator !== 'undefined' && navigator.onLine === false) return
  grammarClient.syncQueuedAttempts().catch((error) => {
    console.warn('[PWA] Offline grammar sync failed:', error)
  })
  wordTrainingClient.syncQueuedAttempts().catch((error) => {
    console.warn('[PWA] Offline word training sync failed:', error)
  })
}

window.addEventListener('online', trySyncOfflineGrammar)
window.setInterval(trySyncOfflineGrammar, 30_000)

app.use(router)
app.use(i18n)

try {
  app.mount('#app')
} catch (error) {
  errorHandler(error, null, 'mount')
  showRuntimeDebugOverlay(persistRuntimeError('mount-error', error, 'mount'))
}
