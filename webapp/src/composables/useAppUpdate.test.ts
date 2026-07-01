import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

// Pretend we are running inside the embedded Android WebView.
const setEmbeddedUserAgent = () => {
  Object.defineProperty(navigator, 'userAgent', {
    value: 'Mozilla/5.0 QantrixEmbeddedApp',
    configurable: true,
  })
}

// Installs a fake bridge whose checkLatestVersion immediately replies via the
// window.__onUpdateCheckResult callback the composable registers.
const installBridge = (reply: { latestVersion?: string; apkUrl?: string; error?: string }) => {
  const startUpdateDownload = vi.fn()
  const cancelUpdateDownload = vi.fn()
  ;(window as any).QantrixAndroid = {
    getAppVersion: () => '0.12.10',
    checkLatestVersion: () => {
      ;(window as any).__onUpdateCheckResult?.(reply)
    },
    startUpdateDownload,
    cancelUpdateDownload,
  }
  return { startUpdateDownload, cancelUpdateDownload }
}

// Each test gets a fresh module so the singleton state does not leak.
const freshUseAppUpdate = async () => {
  vi.resetModules()
  const mod = await import('./useAppUpdate')
  return mod.useAppUpdate()
}

describe('useAppUpdate', () => {
  beforeEach(() => {
    setEmbeddedUserAgent()
    localStorage.clear()
  })

  afterEach(() => {
    vi.useRealTimers()
    delete (window as any).QantrixAndroid
    delete (window as any).__onUpdateCheckResult
    delete (window as any).__onUpdateDownload
  })

  it('shows the modal on auto-check when a newer version exists', async () => {
    installBridge({ latestVersion: '0.12.49', apkUrl: 'https://example/app.apk' })
    const u = await freshUseAppUpdate()
    await u.checkForUpdate({ manual: false })
    expect(u.hasUpdate.value).toBe(true)
    expect(u.modalVisible.value).toBe(true)
    expect(u.latestVersion.value).toBe('0.12.49')
  })

  it('suppresses the auto-check modal for a skipped version', async () => {
    localStorage.setItem('appUpdate.skippedVersion', '0.12.49')
    installBridge({ latestVersion: '0.12.49', apkUrl: 'https://example/app.apk' })
    const u = await freshUseAppUpdate()
    await u.checkForUpdate({ manual: false })
    expect(u.hasUpdate.value).toBe(true)
    expect(u.modalVisible.value).toBe(false)
  })

  it('suppresses the auto-check modal while snoozed', async () => {
    localStorage.setItem('appUpdate.snoozeUntil', String(Date.now() + 60_000))
    installBridge({ latestVersion: '0.12.49', apkUrl: 'https://example/app.apk' })
    const u = await freshUseAppUpdate()
    await u.checkForUpdate({ manual: false })
    expect(u.modalVisible.value).toBe(false)
  })

  it('ignores skip/snooze for a manual check', async () => {
    localStorage.setItem('appUpdate.skippedVersion', '0.12.49')
    localStorage.setItem('appUpdate.snoozeUntil', String(Date.now() + 60_000))
    installBridge({ latestVersion: '0.12.49', apkUrl: 'https://example/app.apk' })
    const u = await freshUseAppUpdate()
    await u.checkForUpdate({ manual: true })
    expect(u.modalVisible.value).toBe(true)
  })

  it('reports up-to-date on a manual check when already current', async () => {
    installBridge({ latestVersion: '0.12.10', apkUrl: 'https://example/app.apk' })
    const u = await freshUseAppUpdate()
    await u.checkForUpdate({ manual: true })
    expect(u.hasUpdate.value).toBe(false)
    expect(u.upToDate.value).toBe(true)
    expect(u.modalVisible.value).toBe(false)
  })

  it('skipVersion persists and closes the modal', async () => {
    installBridge({ latestVersion: '0.12.49', apkUrl: 'https://example/app.apk' })
    const u = await freshUseAppUpdate()
    await u.checkForUpdate({ manual: false })
    u.skipVersion()
    expect(localStorage.getItem('appUpdate.skippedVersion')).toBe('0.12.49')
    expect(u.modalVisible.value).toBe(false)
  })

  it('snooze24h persists a future timestamp and closes the modal', async () => {
    installBridge({ latestVersion: '0.12.49', apkUrl: 'https://example/app.apk' })
    const u = await freshUseAppUpdate()
    await u.checkForUpdate({ manual: false })
    u.snooze24h()
    expect(Number(localStorage.getItem('appUpdate.snoozeUntil'))).toBeGreaterThan(Date.now())
    expect(u.modalVisible.value).toBe(false)
  })

  it('installUpdate forwards the apk url to the native bridge', async () => {
    const { startUpdateDownload } = installBridge({
      latestVersion: '0.12.49',
      apkUrl: 'https://example/app.apk',
    })
    const u = await freshUseAppUpdate()
    await u.checkForUpdate({ manual: false })
    u.installUpdate()
    expect(startUpdateDownload).toHaveBeenCalledWith('https://example/app.apk')
  })

  it('tracks native download progress', async () => {
    installBridge({
      latestVersion: '0.12.49',
      apkUrl: 'https://example/app.apk',
    })
    const u = await freshUseAppUpdate()
    await u.checkForUpdate({ manual: false })
    u.installUpdate()
    ;(window as any).__onUpdateDownload?.({
      state: 'downloading',
      bytesDownloaded: 45,
      bytesTotal: 100,
      progress: 45,
    })
    expect(u.downloadStatus.value).toBe('downloading')
    expect(u.downloadProgress.value).toBe(45)
    expect(u.downloadBytesDownloaded.value).toBe(45)
    expect(u.downloadBytesTotal.value).toBe(100)
  })

  it('marks download as failed when native bridge never finishes', async () => {
    vi.useFakeTimers()
    installBridge({
      latestVersion: '0.12.49',
      apkUrl: 'https://example/app.apk',
    })
    const u = await freshUseAppUpdate()
    await u.checkForUpdate({ manual: false })
    u.installUpdate()
    vi.advanceTimersByTime(120_000)
    expect(u.downloadStatus.value).toBe('error')
    expect(u.errorMessage.value).toBe('download_timeout')
  })

  it('cancels native download when supported', async () => {
    const { cancelUpdateDownload } = installBridge({
      latestVersion: '0.12.49',
      apkUrl: 'https://example/app.apk',
    })
    const u = await freshUseAppUpdate()
    await u.checkForUpdate({ manual: false })
    u.installUpdate()
    expect(u.canCancelDownload.value).toBe(true)
    u.cancelDownload()
    expect(cancelUpdateDownload).toHaveBeenCalled()
    expect(u.downloadStatus.value).toBe('idle')
  })

  it('is a no-op outside the embedded app', async () => {
    Object.defineProperty(navigator, 'userAgent', {
      value: 'Mozilla/5.0 (plain browser)',
      configurable: true,
    })
    installBridge({ latestVersion: '0.12.49', apkUrl: 'https://example/app.apk' })
    const u = await freshUseAppUpdate()
    await u.checkForUpdate({ manual: true })
    expect(u.modalVisible.value).toBe(false)
    expect(u.hasUpdate.value).toBe(false)
  })
})
