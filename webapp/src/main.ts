import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './styles/theme.css'
import './style.css'

// Clean up Telegram URL parameters before router initialization
// Telegram sometimes adds tgWebAppData to URL which can break hash routing
if (window.location.search.includes('tgWebAppData') || window.location.hash.includes('tgWebAppData')) {
  const url = new URL(window.location.href)
  const tgWebAppData = url.searchParams.get('tgWebAppData') || url.hash.match(/tgWebAppData=([^&]+)/)?.[1]
  
  if (tgWebAppData) {
    // Store it for later use
    ;(window as any).__tgWebAppData = decodeURIComponent(tgWebAppData)
    console.log('[App] Found tgWebAppData in URL, stored for auth')
    
    // Clean URL to prevent routing issues
    url.searchParams.delete('tgWebAppData')
    const cleanHash = url.hash.replace(/[?&]tgWebAppData=[^&]*/g, '')
    url.hash = cleanHash || '#/'
    
    // Replace URL without reload
    window.history.replaceState({}, '', url.toString())
  }
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

