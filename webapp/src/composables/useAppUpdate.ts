import { ref, computed } from 'vue'
import { isEmbeddedAndroidApp } from '../utils/runtime'
import { compareVersions } from '../utils/version'

// Bridge exposed by the embedded Android WebView (see MainActivity.AndroidBridge).
interface QantrixAndroid {
  getAppVersion?: () => string
  checkLatestVersion?: () => void
  startUpdateDownload?: (apkUrl: string) => void
  setSystemBarsColor?: (color: string) => void
}

interface UpdateCheckResult {
  latestVersion?: string
  apkUrl?: string
  error?: string
}

interface DownloadState {
  state?: 'downloading' | 'installing' | 'error'
  error?: string
}

type DownloadStatus = 'idle' | 'downloading' | 'installing' | 'error'

const SKIP_KEY = 'appUpdate.skippedVersion'
const SNOOZE_KEY = 'appUpdate.snoozeUntil'
const SNOOZE_MS = 24 * 60 * 60 * 1000

// Module-level (singleton) state so the startup check in App.vue and the manual
// check in SettingsView share one source of truth.
const currentVersion = ref('')
const latestVersion = ref('')
const apkUrl = ref('')
const hasUpdate = ref(false)
const modalVisible = ref(false)
const checking = ref(false)
const upToDate = ref(false)
const downloadStatus = ref<DownloadStatus>('idle')
const errorMessage = ref('')

let pendingResolve: ((value: UpdateCheckResult) => void) | null = null
let bridgeReady = false

const bridge = (): QantrixAndroid | undefined => (window as any).QantrixAndroid

const readSnoozeUntil = (): number => {
  const raw = localStorage.getItem(SNOOZE_KEY)
  const n = raw ? Number(raw) : 0
  return Number.isFinite(n) ? n : 0
}

const ensureBridgeCallbacks = () => {
  if (bridgeReady) return
  bridgeReady = true
  ;(window as any).__onUpdateCheckResult = (result: UpdateCheckResult) => {
    if (pendingResolve) {
      pendingResolve(result || {})
      pendingResolve = null
    }
  }
  ;(window as any).__onUpdateDownload = (payload: DownloadState) => {
    const state = payload?.state
    if (state === 'downloading') {
      downloadStatus.value = 'downloading'
    } else if (state === 'installing') {
      downloadStatus.value = 'installing'
    } else if (state === 'error') {
      downloadStatus.value = 'error'
      errorMessage.value = payload?.error || 'download_failed'
    }
  }
}

const requestLatestVersion = (): Promise<UpdateCheckResult> => {
  ensureBridgeCallbacks()
  const api = bridge()
  if (!api?.checkLatestVersion) {
    return Promise.resolve({ error: 'no_bridge' })
  }
  return new Promise<UpdateCheckResult>((resolve) => {
    const timeout = setTimeout(() => {
      if (pendingResolve !== settle) return // already settled by the bridge callback
      pendingResolve = null
      resolve({ error: 'timeout' })
    }, 20000)
    const settle = (value: UpdateCheckResult) => {
      clearTimeout(timeout)
      resolve(value)
    }
    pendingResolve = settle
    api.checkLatestVersion!()
  })
}

export const useAppUpdate = () => {
  const skipVersion = () => {
    if (latestVersion.value) {
      localStorage.setItem(SKIP_KEY, latestVersion.value)
    }
    modalVisible.value = false
  }

  const snooze24h = () => {
    localStorage.setItem(SNOOZE_KEY, String(Date.now() + SNOOZE_MS))
    modalVisible.value = false
  }

  const dismiss = () => {
    modalVisible.value = false
  }

  const installUpdate = () => {
    const api = bridge()
    if (!api?.startUpdateDownload || !apkUrl.value) return
    downloadStatus.value = 'downloading'
    errorMessage.value = ''
    api.startUpdateDownload(apkUrl.value)
  }

  // manual: triggered from Settings — ignores skip/snooze and surfaces the
  // "up to date" result. auto: startup check — suppressed by skip/snooze.
  const checkForUpdate = async ({ manual }: { manual: boolean }) => {
    if (!isEmbeddedAndroidApp()) return
    currentVersion.value = bridge()?.getAppVersion?.() || currentVersion.value
    upToDate.value = false
    errorMessage.value = ''
    checking.value = true
    try {
      const result = await requestLatestVersion()
      if (result.error || !result.latestVersion || !result.apkUrl) {
        if (manual) errorMessage.value = result.error || 'check_failed'
        return
      }
      latestVersion.value = result.latestVersion
      apkUrl.value = result.apkUrl
      const newer = compareVersions(result.latestVersion, currentVersion.value) > 0
      hasUpdate.value = newer

      if (!newer) {
        if (manual) upToDate.value = true
        return
      }

      if (!manual) {
        const skipped = localStorage.getItem(SKIP_KEY)
        if (skipped === result.latestVersion) return
        if (Date.now() < readSnoozeUntil()) return
      }

      downloadStatus.value = 'idle'
      modalVisible.value = true
    } finally {
      checking.value = false
    }
  }

  return {
    currentVersion,
    latestVersion,
    hasUpdate,
    modalVisible: computed(() => modalVisible.value),
    checking: computed(() => checking.value),
    upToDate: computed(() => upToDate.value),
    downloadStatus: computed(() => downloadStatus.value),
    errorMessage: computed(() => errorMessage.value),
    checkForUpdate,
    installUpdate,
    skipVersion,
    snooze24h,
    dismiss,
  }
}
