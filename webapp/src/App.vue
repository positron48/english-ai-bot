<template>
  <div id="app">
    <nav v-if="isAuthenticated" class="navbar">
      <div class="container">
        <div class="nav-links">
          <router-link to="/dashboard">Dashboard</router-link>
          <router-link to="/vocab">Vocabulary</router-link>
          <router-link to="/training">Training</router-link>
          <router-link to="/chat">Chat</router-link>
          <router-link v-if="isAdmin" to="/admin">Admin</router-link>
          <button @click="toggleTheme" class="theme-toggle" :title="theme === 'dark' ? 'Switch to light theme' : 'Switch to dark theme'">
            <span v-if="theme === 'dark'">☀️</span>
            <span v-else>🌙</span>
          </button>
          <button @click="logout" class="btn btn-secondary">Logout</button>
        </div>
      </div>
    </nav>
    <main class="container">
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuth } from './composables/useAuth'
import { useTheme } from './composables/useTheme'

const router = useRouter()
const { isAuthenticated, isAdmin, logout: authLogout } = useAuth()
const { theme, toggleTheme } = useTheme()

const logout = () => {
  authLogout()
  router.push('/login')
}
</script>

<style scoped>
.navbar {
  background: var(--bg-secondary);
  box-shadow: 0 2px 4px var(--navbar-shadow);
}

.nav-links {
  display: flex;
  gap: 20px;
  align-items: center;
  flex-wrap: wrap;
}

.nav-links a {
  text-decoration: none;
  color: var(--color-primary);
  padding: 10px;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.nav-links a:hover,
.nav-links a.router-link-active {
  background-color: var(--bg-hover);
}

.theme-toggle {
  background: transparent;
  border: 1px solid var(--border-primary);
  border-radius: 6px;
  padding: 8px 12px;
  cursor: pointer;
  font-size: 18px;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 44px;
}

.theme-toggle:hover {
  background-color: var(--bg-hover);
  border-color: var(--border-secondary);
}
</style>

