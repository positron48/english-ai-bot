<template>
  <div id="app">
    <div v-if="networkToast.visible" class="network-toast" :class="`network-toast--${networkToast.kind}`">
      <span>{{ networkToast.message }}</span>
      <button type="button" class="network-toast-close" @click="hideNetworkToast">×</button>
    </div>
    <!-- Auth Error Message -->
    <div v-if="authError" class="auth-error-banner">
      <div class="auth-error-content">
        <strong>{{ t('auth.error') }}</strong> {{ authError }}
        <button @click="dismissAuthError" class="auth-error-close">×</button>
      </div>
    </div>

    <router-view v-if="mounted" />
    <div v-else class="app-loading">{{ t('common.loading') }}</div>

    <!-- Global Dialog Modals -->
    <AlertModal
      :message="alertState.message"
      :visible="alertState.visible"
      @close="closeAlert"
    />
    <ConfirmModal
      :message="confirmState.message"
      :visible="confirmState.visible"
      @confirm="() => closeConfirm(true)"
      @cancel="() => closeConfirm(false)"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuth } from './composables/useAuth'
import { useTheme } from './composables/useTheme'
import { useDialog } from './composables/useDialog'
import { useLocale } from './composables/useLocale'
import { useLearningConfig } from './composables/useLearningConfig'
import { isEmbeddedAndroidApp } from './utils/runtime'
import AlertModal from './components/AlertModal.vue'
import ConfirmModal from './components/ConfirmModal.vue'

const route = useRoute()
const { t } = useI18n()
const { isAuthenticated } = useAuth()
const { currentLocale } = useLocale()
const { learning, ensureLearningLoaded } = useLearningConfig()
const { theme } = useTheme()
const { alertState, confirmState, closeAlert, closeConfirm } = useDialog()

const mounted = ref(false)
const authError = ref<string | null>(null)
const networkToast = ref<{ visible: boolean; kind: 'offline' | 'online'; message: string }>({
  visible: false,
  kind: 'online',
  message: '',
})
let networkToastTimer: ReturnType<typeof setTimeout> | null = null

const hideNetworkToast = () => {
  networkToast.value.visible = false
  if (networkToastTimer) {
    clearTimeout(networkToastTimer)
    networkToastTimer = null
  }
}

const showNetworkToast = (kind: 'offline' | 'online') => {
  networkToast.value = {
    visible: true,
    kind,
    message: kind === 'offline' ? t('offline.connectionLost') : t('offline.connectionRestored'),
  }
  if (networkToastTimer) {
    clearTimeout(networkToastTimer)
    networkToastTimer = null
  }
  if (kind === 'online') {
    networkToastTimer = setTimeout(() => {
      networkToast.value.visible = false
      networkToastTimer = null
    }, 3000)
  }
  // offline toast stays until dismissed or until online event fires
}

const handleOffline = () => showNetworkToast('offline')
const handleOnline = () => showNetworkToast('online')

const updateThemeMetaColor = () => {
  const bg = getComputedStyle(document.documentElement).getPropertyValue('--bg').trim() || '#F8F1E4'
  let meta = document.querySelector('meta[name="theme-color"]') as HTMLMetaElement | null
  if (!meta) {
    meta = document.createElement('meta')
    meta.name = 'theme-color'
    document.head.appendChild(meta)
  }
  meta.content = bg
  if (isEmbeddedAndroidApp()) {
    ;(window as any).QantrixAndroid?.setSystemBarsColor?.(bg)
  }
}

const updateDocumentTitle = () => {
  const tl = learning.value?.target_lang ?? 'en'
  document.title = t(tl === 'es' ? 'app.titleEsInstance' : 'app.titleEnInstance')
}

watch([learning, currentLocale], updateDocumentTitle, { deep: true, immediate: true })

onMounted(() => {
  ensureLearningLoaded()
  const tg = (window as any).Telegram?.WebApp

  window.addEventListener('offline', handleOffline)
  window.addEventListener('online', handleOnline)

  mounted.value = true
  updateThemeMetaColor()

  // Check auth status after a delay
  setTimeout(() => {
    if (!isAuthenticated.value && route.path !== '/login') {
      if (tg) {
        authError.value = t('auth.telegramAuthFailed')
      }
    }
  }, 3000)
})

onUnmounted(() => {
  window.removeEventListener('offline', handleOffline)
  window.removeEventListener('online', handleOnline)
  hideNetworkToast()
})

watch(theme, () => setTimeout(updateThemeMetaColor, 0), { immediate: true })

// Watch route changes
watch(() => route.path, (newPath) => {
  if (newPath === '/login') {
    authError.value = null
  }
})

// Watch auth status
watch(() => isAuthenticated.value, (newValue) => {
  if (newValue) {
    authError.value = null
  }
})

// Global error handler for auth errors
;(window as any).__setAuthError = (error: string) => {
  authError.value = error
}

const dismissAuthError = () => {
  authError.value = null
}
</script>

<style scoped>
.app-loading {
  padding: 40px 20px;
  text-align: center;
  color: var(--subtext);
}

.auth-error-banner {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  background: #b91c1c;
  color: white;
  padding: 15px 20px;
  z-index: 10001;
  box-shadow: 0 2px 8px rgba(0,0,0,0.2);
}

.auth-error-content {
  max-width: 880px;
  margin: 0 auto;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 15px;
}

.auth-error-close {
  background: rgba(255,255,255,0.2);
  border: none;
  color: white;
  font-size: 24px;
  line-height: 1;
  padding: 0 10px;
  cursor: pointer;
  border-radius: 3px;
  transition: background 0.2s;
}

.auth-error-close:hover {
  background: rgba(255,255,255,0.3);
}

.network-toast {
  position: fixed;
  top: 14px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 10020;
  display: flex;
  align-items: center;
  gap: 12px;
  max-width: min(92vw, 560px);
  padding: 12px 14px;
  border-radius: 12px;
  color: #fff;
  box-shadow: 0 12px 28px rgba(0, 0, 0, 0.22);
}

.network-toast--offline {
  background: #b91c1c;
}

.network-toast--online {
  background: #047857;
}

.network-toast-close {
  border: 0;
  background: rgba(255, 255, 255, 0.2);
  color: #fff;
  width: 26px;
  height: 26px;
  border-radius: 50%;
  font-size: 16px;
  line-height: 1;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  padding: 0;
}
</style>
