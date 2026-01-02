import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './styles/theme.css'
import './style.css'

// Clean up Telegram URL parameters before router initialization
// Telegram adds tgWebAppData to URL which breaks hash routing

// Clean hash from tgWebAppData before router initialization
let cleanedHash = window.location.hash
if (cleanedHash.includes('tgWebAppData')) {
  // Extract tgWebAppData from hash
  const tgWebAppDataMatch = cleanedHash.match(/[?&]tgWebAppData=([^&]+)/)
  if (tgWebAppDataMatch) {
    const tgWebAppData = decodeURIComponent(tgWebAppDataMatch[1])
    ;(window as any).__tgWebAppData = tgWebAppData
  }
  
  // Remove tgWebAppData from hash
  cleanedHash = cleanedHash.replace(/[?&]tgWebAppData=[^&]*/g, '')
  if (!cleanedHash || cleanedHash === '#' || cleanedHash === '#/') {
    cleanedHash = '#/'
  }
  
  // Clean search params too
  const url = new URL(window.location.href)
  url.searchParams.delete('tgWebAppData')
  
  // Replace URL
  const newUrl = url.origin + url.pathname + (url.search || '') + cleanedHash
  window.history.replaceState({}, '', newUrl)
}

// Also clean search params
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

app.use(router)

try {
  app.mount('#app')
} catch (error) {
  errorHandler(error, null, 'mount')
}

