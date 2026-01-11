<template>
  <div class="admin-menu">
    <h1>Admin Panel</h1>
    
    <div class="admin-tabs">
      <router-link to="/admin" class="admin-tab" :class="{ active: $route.path === '/admin' }">
        <span>Main</span>
      </router-link>
      <router-link to="/admin/circuit-breaker" class="admin-tab" :class="{ active: $route.path === '/admin/circuit-breaker' }">
        <span>Circuit Breaker</span>
      </router-link>
      <router-link to="/admin/prompt-tester" class="admin-tab" :class="{ active: $route.path === '/admin/prompt-tester' }">
        <span>Prompt Tester</span>
      </router-link>
      <router-link to="/admin/orphaned-cards" class="admin-tab" :class="{ active: $route.path === '/admin/orphaned-cards' }">
        <span>Orphaned Cards</span>
      </router-link>
      <router-link to="/admin/word-sets" class="admin-tab" :class="{ active: $route.path === '/admin/word-sets' }">
        <span>Word Sets</span>
      </router-link>
    </div>
  </div>
</template>

<script setup lang="ts">
// No props or logic needed - just navigation menu
</script>

<style scoped>
.admin-menu {
  margin-bottom: 24px;
}

.admin-menu h1 {
  margin: 0 0 16px 0;
  font-size: 28px;
  font-weight: 600;
  color: var(--text-primary);
}

.admin-tabs {
  display: flex;
  gap: 0;
  margin-bottom: 24px;
  border-bottom: 2px solid var(--border-primary);
  overflow-x: auto;
  overflow-y: hidden;
  scrollbar-width: none; /* Firefox */
  -ms-overflow-style: none; /* IE and Edge */
}

.admin-tabs::-webkit-scrollbar {
  display: none; /* Chrome, Safari, Opera */
}

.admin-tab {
  padding: 12px 24px;
  text-decoration: none;
  color: var(--text-secondary);
  border-bottom: 3px solid transparent;
  transition: all 0.2s;
  white-space: nowrap;
  position: relative;
  bottom: -2px;
  font-size: 16px;
}

.admin-tab:hover {
  color: var(--text-primary);
  background: var(--bg-hover);
}

.admin-tab.active {
  color: var(--color-primary);
  border-bottom-color: var(--color-primary);
  font-weight: 500;
}
</style>
