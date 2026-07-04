const MEDIA_CACHE_VERSION = 'media-pvc-20260704'

export function mediaUrl(rawUrl?: string | null): string {
  if (!rawUrl) return ''

  try {
    const url = new URL(rawUrl, window.location.origin)
    if (url.origin !== window.location.origin || !url.pathname.startsWith('/api/media/')) {
      return rawUrl
    }
    url.searchParams.set('v', MEDIA_CACHE_VERSION)
    return url.pathname + url.search + url.hash
  } catch {
    return rawUrl
  }
}
