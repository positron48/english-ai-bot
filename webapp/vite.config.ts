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
            
            // For all other paths under /app/ (SPA routes), serve index.html
            // This allows direct URL access to work with history mode
            // Exclude paths that look like API endpoints or have file extensions
            if (url.startsWith('/app/') && !url.includes('.') && !isAPIEndpoint(url)) {
              // Rewrite to index.html - Vite will handle the base path automatically
              req.url = '/index.html'
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

