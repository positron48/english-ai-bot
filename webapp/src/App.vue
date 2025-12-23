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
import { computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuth } from './composables/useAuth'

const router = useRouter()
const { isAuthenticated, isAdmin, logout: authLogout } = useAuth()

const logout = () => {
  authLogout()
  router.push('/login')
}
</script>

<style scoped>
.navbar {
  background: white;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  margin-bottom: 20px;
}

.nav-links {
  display: flex;
  gap: 20px;
  align-items: center;
  flex-wrap: wrap;
}

.nav-links a {
  text-decoration: none;
  color: #007bff;
  padding: 10px;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.nav-links a:hover,
.nav-links a.router-link-active {
  background-color: #f0f0f0;
}
</style>

