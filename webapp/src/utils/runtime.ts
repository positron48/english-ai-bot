export const isEmbeddedAndroidApp = (): boolean => {
  if (typeof navigator === 'undefined') return false
  return navigator.userAgent.includes('QantrixEmbeddedApp')
}
