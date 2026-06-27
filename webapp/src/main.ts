import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import i18n from './i18n'
import { initOfflineSyncRunner } from './api/offlineSyncRunner'
import { isEmbeddedAndroidApp } from './utils/runtime'
// Public entry uses ONLY the new Linglow theme; legacy styles live in admin-main.ts
import './styles/linglow-theme.css'
import './styles/linglow-markdown.css'

declare global {
  interface Window {
    __showQantrixRuntimeDebug?: () => void
    __setSafeAreaInsets?: (top: number, bottom: number) => void
  }
}

// window.__setSafeAreaInsets is defined by an inline script in index.html (so a native push that
// arrives before this bundle loads is not lost) and the embedded Android app pushes the real
// status/navigation-bar insets to it. Here we add a fallback for the brief window before that push
// arrives (and for devices that neither expose env(safe-area-inset-top) nor report a usable value):
// without it the layout briefly draws under the status bar.

// measureEnvTopInset returns the effective env(safe-area-inset-top) in CSS px (0 when unsupported).
function measureEnvTopInset(): number {
  const probe = document.createElement('div')
  probe.style.cssText =
    'position:fixed;top:0;left:0;width:0;height:env(safe-area-inset-top,0px);visibility:hidden;pointer-events:none;'
  document.body.appendChild(probe)
  const h = probe.getBoundingClientRect().height
  probe.remove()
  return h
}

function ensureAndroidTopInset(): void {
  if (!isEmbeddedAndroidApp()) return
  const root = document.documentElement
  const native = parseFloat(getComputedStyle(root).getPropertyValue('--android-inset-top')) || 0
  if (native > 0) return // native already pushed a real value
  if (measureEnvTopInset() > 0) return // env() works, the layout already uses it

  // Try the native bridge (status bar height is usually reported in device pixels).
  let top = 0
  try {
    const raw = Number((window as any).QantrixAndroid?.getStatusBarHeight?.())
    if (raw > 0) top = raw > 60 ? raw / (window.devicePixelRatio || 1) : raw
  } catch { /* bridge getter absent — fall through to the floor */ }
  if (top <= 0) top = 32 // conservative floor so content clears the status bar

  root.style.setProperty('--android-inset-top', `${Math.round(top)}px`)
}

// Run after mount, with retries, to let a late native push win before applying the fallback.
;[300, 1200].forEach((delay) => setTimeout(ensureAndroidTopInset, delay))

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

const warmOfflineRouteChunks = () => {
  if (typeof navigator !== 'undefined' && navigator.onLine === false) return
  void Promise.all([
    import('./views/LearningView.vue'),
    import('./views/TrainingView.vue'),
    import('./views/GrammarCategoriesView.vue'),
    import('./views/GrammarChaptersView.vue'),
    import('./views/GrammarChapterView.vue'),
    import('./views/GrammarTestView.vue'),
    import('./views/GrammarPlacementTestView.vue'),
    import('./views/GrammarTrainingView.vue'),
  ]).catch((error) => {
    console.debug('[PWA] Offline route warmup failed:', error)
  })
}

const warmServiceWorkerAppShell = async () => {
  if (!('serviceWorker' in navigator)) return
  if (navigator.onLine === false) return
  try {
    const registration = await navigator.serviceWorker.ready
    registration.active?.postMessage({ type: 'CACHE_APP_SHELL' })
  } catch (error) {
    console.debug('[PWA] Service worker warmup failed:', error)
  }
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
    navigator.serviceWorker.register('/sw.js', { scope: '/' })
      .then(() => {
        setTimeout(() => {
          warmOfflineRouteChunks()
          void warmServiceWorkerAppShell()
        }, 1500)
      })
      .catch((error) => {
        console.warn('[PWA] Service worker registration failed:', error)
      })
  })
  window.addEventListener('online', () => {
    warmOfflineRouteChunks()
    void warmServiceWorkerAppShell()
  })
}

initOfflineSyncRunner()

app.use(router)
app.use(i18n)

try {
  app.mount('#app')
} catch (error) {
  errorHandler(error, null, 'mount')
  showRuntimeDebugOverlay(persistRuntimeError('mount-error', error, 'mount'))
}
