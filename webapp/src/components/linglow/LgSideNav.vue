<template>
  <nav class="lg-sidenav">
    <div class="lg-sidenav-logo">
      <span class="lg-sidenav-logo-text">Linglow</span>
      <span class="lg-sidenav-logo-leaf"><LgIcon name="sprout" :s="18" c="var(--salvia, currentColor)" /></span>
    </div>
    <div class="lg-sidenav-items">
      <router-link
        v-for="tab in NAV_TABS"
        :key="tab.id"
        :to="tab.to"
        class="lg-sidenav-item"
        :class="{ 'lg-sidenav-item--on': activeTab === tab.id }"
      >
        <div v-if="activeTab === tab.id" class="lg-sidenav-marker" />
        <LgIcon :name="tab.icon" :s="20" :c="activeTab === tab.id ? 'var(--salvia)' : 'var(--subtext)'" />
        <span class="lg-sidenav-label">{{ t(tab.labelKey) }}</span>
      </router-link>
    </div>
    <div class="lg-sidenav-bottom">
      <button class="lg-sidenav-theme" type="button" @click="toggleTheme">
        <LgIcon :name="theme === 'light' ? 'moon' : 'sun'" :s="16" c="var(--text)" />
      </button>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useTheme } from '../../composables/useTheme'
import { NAV_TABS } from './navTabs'
import LgIcon from './LgIcon.vue'

const route = useRoute()
const { t } = useI18n()
const { theme, toggleTheme } = useTheme()

const activeTab = computed(() => (route.meta.navTab as string) || 'home')
</script>

<style scoped>
.lg-sidenav {
  width: 220px;
  flex-shrink: 0;
  height: 100%;
  background: var(--card-bg);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.lg-sidenav-logo {
  padding: 28px 20px 22px;
  display: flex;
  align-items: center;
  gap: 8px;
  border-bottom: 1px solid var(--border);
}
.lg-sidenav-logo-text {
  font-family: 'Lora', serif;
  font-weight: 700;
  font-size: 20px;
  color: var(--text);
}
.lg-sidenav-logo-leaf { font-size: 16px; }
.lg-sidenav-items {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 14px 10px;
  overflow-y: auto;
}
.lg-sidenav-item {
  display: flex;
  align-items: center;
  gap: 11px;
  padding: 10px 14px;
  border-radius: 12px;
  border: none;
  background: transparent;
  cursor: pointer;
  text-align: left;
  width: 100%;
  position: relative;
  text-decoration: none;
}
.lg-sidenav-item--on { background: var(--chip-bg); }
.lg-sidenav-marker {
  position: absolute;
  left: 0;
  top: 18%;
  bottom: 18%;
  width: 3px;
  border-radius: 0 3px 3px 0;
  background: var(--salvia);
}
.lg-sidenav-label {
  font-size: 14px;
  font-weight: 400;
  color: var(--subtext);
}
.lg-sidenav-item--on .lg-sidenav-label {
  font-weight: 700;
  color: var(--text);
}
.lg-sidenav-bottom {
  padding: 16px 20px;
  border-top: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}
.lg-sidenav-course {
  flex: 1;
  min-width: 0;
  height: 36px;
  padding: 0 10px;
  border: 1px solid var(--border);
  border-radius: 10px;
  background: var(--bg);
  color: var(--text);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
}
.lg-sidenav-theme {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: 1px solid var(--border);
  background: var(--bg);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  flex-shrink: 0;
  margin-left: auto;
}
</style>
