const APP_SHELL_CACHE = 'qantrix-app-shell-v10'
const ASSET_MANIFEST_URL = '/app/asset-manifest.json'
const APP_SHELL_URLS = ['/app', '/app/', '/app/manifest.webmanifest', '/telegram-web-app.js', '/favicon.svg']

const cacheURL = async (cache, url) => {
  const response = await fetch(url, { cache: 'no-store' }).catch(() => null)
  if (response?.ok) await cache.put(url, response.clone())
}

const cacheAppShell = async () => {
  const cache = await caches.open(APP_SHELL_CACHE)
  await Promise.all(APP_SHELL_URLS.map((url) => cacheURL(cache, url).catch(() => undefined)))

  const manifestResponse = await fetch(ASSET_MANIFEST_URL, { cache: 'no-store' }).catch(() => null)
  if (!manifestResponse || !manifestResponse.ok) return

  await cache.put(ASSET_MANIFEST_URL, manifestResponse.clone())
  const manifest = await manifestResponse.json().catch(() => null)
  const assets = Array.isArray(manifest?.assets) ? manifest.assets : []
  await Promise.all(assets.map((url) => cacheURL(cache, url).catch(() => undefined)))
}

self.addEventListener('install', (event) => {
  event.waitUntil(cacheAppShell().catch(() => undefined))
  self.skipWaiting()
})

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((keys) =>
      Promise.all(keys.filter((key) => key !== APP_SHELL_CACHE).map((key) => caches.delete(key)))
    ).then(() => cacheAppShell()).catch(() => undefined)
  )
  self.clients.claim()
})

self.addEventListener('message', (event) => {
  if (event.data?.type !== 'CACHE_APP_SHELL') return
  event.waitUntil(cacheAppShell().catch(() => undefined))
})

self.addEventListener('fetch', (event) => {
  const request = event.request
  if (request.method !== 'GET') return

  const url = new URL(request.url)
  if (url.origin !== self.location.origin) return
  if (url.pathname.startsWith('/api/') || url.pathname.startsWith('/auth/')) return

  // Admin entry (/app/admin*) is online-only; never cache it as the app shell
  if (url.pathname === '/app/admin' || url.pathname.startsWith('/app/admin/')) return

  if (request.mode === 'navigate' || (url.pathname.startsWith('/app') && !url.pathname.includes('.'))) {
    event.respondWith(
      fetch(request)
        .then((response) => {
          const copy = response.clone()
          caches.open(APP_SHELL_CACHE).then((cache) => cache.put('/app/', copy))
          return response
        })
        .catch(() => caches.match('/app/').then((cached) => cached || caches.match('/app')))
    )
    return
  }

  if (
    url.pathname.startsWith('/app/assets/') ||
    url.pathname.startsWith('/app/fonts/') ||
    url.pathname.startsWith('/app/linglow/') ||
    url.pathname === '/app/manifest.webmanifest' ||
    url.pathname === ASSET_MANIFEST_URL ||
    url.pathname === '/telegram-web-app.js' ||
    url.pathname === '/favicon.svg'
  ) {
    event.respondWith(
      caches.match(request).then((cached) => {
        if (cached) return cached
        return fetch(request).then((response) => {
          if (response.ok) {
            const copy = response.clone()
            caches.open(APP_SHELL_CACHE).then((cache) => cache.put(request, copy))
          }
          return response
        }).catch(() => caches.match(request))
      })
    )
  }
})
