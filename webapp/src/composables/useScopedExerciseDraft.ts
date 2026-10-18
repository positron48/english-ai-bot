import { watch, type Ref } from 'vue'

// Drafts are local hints only. Confirmed answers always come from the server.
export function useScopedExerciseDraft<T>(key: Ref<string>, read: () => T, restore: (draft: unknown) => void, confirmed: () => boolean) {
  watch([key, confirmed], ([storageKey, done]) => {
    if (!storageKey) return
    try {
      if (done) { sessionStorage.removeItem(storageKey); return }
      const saved = sessionStorage.getItem(storageKey)
      if (saved) restore(JSON.parse(saved))
    } catch { /* Draft storage is optional and may contain obsolete data. */ }
  }, { immediate: true, flush: 'post' })
  watch(read, value => {
    if (!key.value || confirmed()) return
    try { sessionStorage.setItem(key.value, JSON.stringify(value)) } catch { /* storage is optional */ }
  }, { deep: true, flush: 'post' })
}
