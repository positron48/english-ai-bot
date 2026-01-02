import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './styles/theme.css'
import './style.css'

// CRITICAL: Clean up Telegram URL parameters BEFORE router initialization
// Telegram adds tgWebAppData to URL which breaks hash routing
console.log('[App] Initial URL:', window.location.href)
console.log('[App] Initial hash:', window.location.hash)
console.log('[App] Initial search:', window.location.search)

// Clean hash from tgWebAppData
let cleanedHash = window.location.hash
if (cleanedHash.includes('tgWebAppData')) {
  console.log('[App] Found tgWebAppData in hash, cleaning...')
  
  // Extract tgWebAppData from hash
  const tgWebAppDataMatch = cleanedHash.match(/[?&]tgWebAppData=([^&]+)/)
  if (tgWebAppDataMatch) {
    const tgWebAppData = decodeURIComponent(tgWebAppDataMatch[1])
    ;(window as any).__tgWebAppData = tgWebAppData
    console.log('[App] Extracted tgWebAppData from hash, length:', tgWebAppData.length)
  }
  
  // Remove tgWebAppData from hash
  cleanedHash = cleanedHash.replace(/[?&]tgWebAppData=[^&]*/g, '')
  // If hash is empty or only contains #, set to #/
  if (!cleanedHash || cleanedHash === '#' || cleanedHash === '#/') {
    cleanedHash = '#/'
  }
  
  // Clean search params too
  const url = new URL(window.location.href)
  url.searchParams.delete('tgWebAppData')
  
  // Replace URL
  const newUrl = url.origin + url.pathname + (url.search || '') + cleanedHash
  window.history.replaceState({}, '', newUrl)
  console.log('[App] Cleaned URL:', window.location.href)
  console.log('[App] Cleaned hash:', window.location.hash)
}

// Also clean search params
if (window.location.search.includes('tgWebAppData')) {
  const url = new URL(window.location.href)
  const tgWebAppData = url.searchParams.get('tgWebAppData')
  if (tgWebAppData) {
    ;(window as any).__tgWebAppData = decodeURIComponent(tgWebAppData)
    console.log('[App] Found tgWebAppData in search params, stored')
  }
  url.searchParams.delete('tgWebAppData')
  window.history.replaceState({}, '', url.toString())
  console.log('[App] Cleaned search params')
}

// Check if Telegram WebApp script is loaded
const checkTelegramScript = () => {
  // Wait a bit for the script to load
  setTimeout(() => {
    if (typeof (window as any).Telegram === 'undefined') {
      console.warn('[App] Telegram WebApp script not loaded - this is normal if not in Telegram Mini App')
    } else {
      console.log('[App] Telegram WebApp script loaded successfully')
      const tg = (window as any).Telegram?.WebApp
      if (tg) {
        console.log('[App] Telegram WebApp object available')
        if (tg.initData) {
          console.log('[App] initData available, length:', tg.initData.length)
        } else {
          console.warn('[App] initData not available')
        }
      }
    }
  }, 100)
}

checkTelegramScript()

// Global error handler for Vue
const errorHandler = (err: unknown, instance: any, info: string) => {
  console.error('[Vue Error Handler]', {
    error: err,
    instance,
    info,
    timestamp: new Date().toISOString()
  })
  
  // Display error on screen if possible
  const appElement = document.getElementById('app')
  if (appElement && !appElement.querySelector('.error-display')) {
    const errorDiv = document.createElement('div')
    errorDiv.className = 'error-display'
    errorDiv.style.cssText = `
      position: fixed;
      top: 0;
      left: 0;
      right: 0;
      background: #ff4444;
      color: white;
      padding: 10px;
      z-index: 10000;
      font-family: monospace;
      font-size: 12px;
      max-height: 200px;
      overflow: auto;
    `
    errorDiv.innerHTML = `
      <strong>Application Error:</strong><br>
      ${err instanceof Error ? err.message : String(err)}<br>
      <small>Info: ${info}</small><br>
      <button onclick="this.parentElement.remove()" style="margin-top: 5px;">Close</button>
    `
    document.body.appendChild(errorDiv)
  }
}

const app = createApp(App)
app.config.errorHandler = errorHandler

// Handle unhandled promise rejections
window.addEventListener('unhandledrejection', (event) => {
  console.error('[Unhandled Promise Rejection]', event.reason)
  errorHandler(event.reason, null, 'unhandledrejection')
})

app.use(router)

try {
  app.mount('#app')
  console.log('[App] Vue application mounted successfully')
} catch (error) {
  console.error('[App] Failed to mount Vue application:', error)
  errorHandler(error, null, 'mount')
}

