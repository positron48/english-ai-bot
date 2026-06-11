<template>
  <nav class="lg-bottomnav" :class="{ 'lg-bottomnav--bordered': showBorder }">
    <router-link
      v-for="tab in NAV_TABS"
      :key="tab.id"
      :to="tab.to"
      class="lg-bottomnav-item"
    >
      <div class="lg-bottomnav-pill" :class="{ 'lg-bottomnav-pill--on': activeTab === tab.id }">
        <LgIcon :name="tab.icon" :s="20" :c="activeTab === tab.id ? 'var(--nav-active-color)' : 'var(--nav-inactive)'" />
        <span class="lg-bottomnav-label" :class="{ 'lg-bottomnav-label--on': activeTab === tab.id }">{{ t(tab.labelKey) }}</span>
      </div>
    </router-link>
  </nav>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { NAV_TABS } from './navTabs'
import LgIcon from './LgIcon.vue'

const route = useRoute()
const { t } = useI18n()
const activeTab = computed(() => (route.meta.navTab as string) || 'home')

const showBorder = ref(false)

// Capture-phase scroll listener — catches any scrollable child
const onScroll = (e: Event) => {
  const el = e.target as HTMLElement | null
  if (!el || typeof el.scrollTop === 'undefined') return
  if (el.scrollHeight <= el.clientHeight) return
  const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 8
  showBorder.value = el.scrollTop > 8 && !atBottom
}

onMounted(() => window.addEventListener('scroll', onScroll, true))
onUnmounted(() => window.removeEventListener('scroll', onScroll, true))
</script>

<style scoped>
.lg-bottomnav {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  width: 100%;
  background: var(--nav-bg);
  backdrop-filter: blur(18px);
  -webkit-backdrop-filter: blur(18px);
  border-top: 1px solid transparent;
  box-shadow: 0 -4px 20px rgba(0,0,0,0.10);
  display: flex;
  z-index: 200;
  align-items: center;
  padding-bottom: env(safe-area-inset-bottom, 0px);
  transition: border-color .2s ease;
}
.lg-bottomnav--bordered {
  border-top: 1px solid var(--border);
  box-shadow: 0 -8px 28px rgba(0,0,0,0.18);
}
.lg-bottomnav-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 8px 4px 6px;
  border: none;
  background: none;
  cursor: pointer;
  text-decoration: none;
}
.lg-bottomnav-pill {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  padding: 5px 6px 4px;
  border-radius: 18px;
  border: 1px solid transparent;
}
.lg-bottomnav-pill--on {
  padding: 5px 12px 4px;
  background: var(--nav-active-bg);
  border: 1px solid var(--nav-active-border);
}
.lg-bottomnav-label {
  font-size: 10px;
  color: var(--nav-inactive);
  font-weight: 400;
}
.lg-bottomnav-label--on {
  color: var(--nav-active-color);
  font-weight: 600;
}
</style>
