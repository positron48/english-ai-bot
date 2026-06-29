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
  ;(window as any).QantrixAndroid = {
    getAppVersion: () => '0.12.10',
    checkLatestVersion: () => {
      ;(window as any).__onUpdateCheckResult?.(reply)
    },
    startUpdateDownload,
  }
  return { startUpdateDownload }
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
