<template>
  <div class="lg-layout" :class="{ 'lg-layout--desktop': isDesktop }">
    <template v-if="isDesktop">
      <LgSideNav v-if="!fullscreen" />
      <div class="lg-layout-center">
        <div class="lg-layout-content" :class="{ 'lg-layout-content--map': isMapRoute, 'lg-view-pad': !fullscreen }">
          <router-view :key="currentCourseCode" />
        </div>
      </div>
    </template>
    <template v-else>
      <div
        class="lg-layout-mobile"
        :class="{
          'lg-layout-mobile--with-nav': !fullscreen,
          'lg-view-pad': !fullscreen && !isMapRoute,
          'lg-layout-mobile--map': isMapRoute,
        }"
      >
        <router-view :key="currentCourseCode" />
      </div>
      <LgBottomNav v-if="!fullscreen" />
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useAuth } from '../composables/useAuth'
import { useCourse } from '../composables/useCourse'
import { useActivityTracker } from '../composables/useActivityTracker'
import { PUBLIC_BREAKPOINT } from '../constants/layout'
import LgSideNav from '../components/linglow/LgSideNav.vue'
import LgBottomNav from '../components/linglow/LgBottomNav.vue'

const route = useRoute()
const { isAuthenticated } = useAuth()
const { ensureCourseLoaded, currentCourseCode } = useCourse()
useActivityTracker()

const isDesktop = ref(window.innerWidth >= PUBLIC_BREAKPOINT)
const check = () => { isDesktop.value = window.innerWidth >= PUBLIC_BREAKPOINT }

const fullscreen = computed(() => !!route.meta.fullscreen)
const isMapRoute = computed(() => route.name === 'City')

watch(isAuthenticated, (a) => { if (a) ensureCourseLoaded() }, { immediate: true })

onMounted(() => window.addEventListener('resize', check))
onUnmounted(() => window.removeEventListener('resize', check))
</script>

<style scoped>
.lg-layout {
  min-height: 100vh;
  min-height: 100dvh;
  background: var(--bg);
  color: var(--text);
  font-family: 'Inter', sans-serif;
}
.lg-layout--desktop {
  display: flex;
  height: 100vh;
  height: 100dvh;
}
.lg-layout--desktop .lg-sidenav { height: 100%; }
.lg-layout-center {
  flex: 1;
  height: 100%;
  display: flex;
  justify-content: center;
  position: relative;
}
.lg-layout-content {
  flex: 1;
  height: 100%;
  overflow-y: auto;
  max-width: 880px;
  width: 100%;
  position: relative;
}
.lg-layout-content--map { max-width: 760px; }
.lg-layout-mobile {
  width: 100%;
  min-height: 100vh;
  min-height: 100dvh;
  position: relative;
  /* Some Android devices (e.g. Samsung) draw under the status bar; keep content clear of it. */
  padding-top: env(safe-area-inset-top, 0px);
}
.lg-layout-mobile--with-nav {
  padding-bottom: calc(60px + env(safe-area-inset-bottom, 0px)) !important;
}
.lg-layout-mobile--map {
  height: 100vh;
  height: 100dvh;
  display: flex;
  overflow: hidden;
}
.lg-layout-mobile--map > * {
  flex: 1;
  min-height: 0;
}
.lg-view-pad {
  padding: 0 14px;
}
</style>
