import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  build: {
    outDir: 'dist',
    assetsDir: 'assets',
    emptyOutDir: true,
  },
  base: '/app/',
  server: {
    port: 5173,
    proxy: {
      // Proxy only API endpoints, not static files
      '/app/dashboard': {
        target: 'http://localhost:8184',
        changeOrigin: true,
      },
      '/app/vocab': {
        target: 'http://localhost:8184',
        changeOrigin: true,
      },
      '/app/training': {
        target: 'http://localhost:8184',
        changeOrigin: true,
      },
      '/app/chat': {
        target: 'http://localhost:8184',
        changeOrigin: true,
      },
      '/app/admin': {
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

