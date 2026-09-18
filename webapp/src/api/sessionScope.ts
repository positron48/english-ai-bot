// Stable across token refreshes. Cached profile data must never select an account.
export function currentUserScope(): string {
  try {
    const token = localStorage.getItem('access_token')
    if (!token) return 'anon'
    const payload = token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/')
    const { user_id } = JSON.parse(atob(payload))
    return Number.isSafeInteger(user_id) && user_id > 0 ? `user:${user_id}` : 'anon'
  } catch { return 'anon' }
}

export function captureUserScope(): () => void {
  const scope = currentUserScope()
  return () => {
    if (scope === 'anon' || currentUserScope() !== scope) throw new Error('Session changed')
  }
}
