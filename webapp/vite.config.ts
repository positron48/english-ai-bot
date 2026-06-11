import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// API endpoints that should be proxied to backend
// All API endpoints are now under /api/
const apiEndpoints = [
  '/api/',
  '/auth',
  '/swagger',
  '/health',
]

// Check if path is an API endpoint
function isAPIEndpoint(path: string): boolean {
  return apiEndpoints.some(endpoint => path.startsWith(endpoint))
}

// Check if path is a static asset (has file extension)
function isStaticAsset(path: string): boolean {
  return /\.\w+$/.test(path) && !path.endsWith('.html')
}

export default defineConfig({
  plugins: [
    vue(),
    {
      name: 'spa-fallback',
      configureServer(server) {
        return () => {
          server.middlewares.use((req, res, next) => {
            const url = req.url || ''
            
            // Don't interfere with API endpoints (they are handled by proxy)
            if (isAPIEndpoint(url)) {
              return next()
            }
            
            // Don't interfere with static assets
            if (isStaticAsset(url)) {
              return next()
            }
            
            // For all other paths under /app/ (SPA routes), serve the right entry:
            // /app/admin* -> admin.html (admin entry), everything else -> index.html.
            // Must match the Go fallback in internal/web/webapp_routes.go.
            if (url.startsWith('/app/') && !url.includes('.') && !isAPIEndpoint(url)) {
              req.url = url === '/app/admin' || url.startsWith('/app/admin/') ? '/admin.html' : '/index.html'
            } else if (url === '/app' || url === '/app/') {
              // Handle root /app path
              req.url = '/index.html'
            }
            
            next()
          })
        }
      },
    },
  ],
  build: {
    outDir: 'dist',
    assetsDir: 'assets',
    emptyOutDir: true,
    rollupOptions: {
      input: {
        index: new URL('./index.html', import.meta.url).pathname,
        admin: new URL('./admin.html', import.meta.url).pathname,
      },
    },
  },
  base: '/app/',
  server: {
    port: 5173,
    proxy: {
      // Proxy all API endpoints to backend
      '/api': {
        target: 'http://localhost:8184',
        changeOrigin: true,
      },
      '/auth': {
        target: 'http://localhost:8184',
        changeOrigin: true,
      },
      '/swagger': {
        target: 'http://localhost:8184',
        changeOrigin: true,
      },
      '/health': {
        target: 'http://localhost:8184',
        changeOrigin: true,
      },
    },
  },
})

