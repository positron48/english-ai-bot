import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import i18n from './i18n'
import { grammarClient } from './api/grammarClient'
import './styles/theme.css'
import './style.css'
import './styles/markdown-content.css'

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
}

const app = createApp(App)
app.config.errorHandler = errorHandler

// Handle unhandled promise rejections
window.addEventListener('unhandledrejection', (event) => {
  console.error('[Unhandled Promise Rejection]', event.reason)
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
}

window.addEventListener('online', trySyncOfflineGrammar)
window.setInterval(trySyncOfflineGrammar, 30_000)

app.use(router)
app.use(i18n)

try {
  app.mount('#app')
} catch (error) {
  errorHandler(error, null, 'mount')
}
