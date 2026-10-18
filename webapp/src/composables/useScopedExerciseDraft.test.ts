import { describe, expect, it } from 'vitest'
import { effectScope, nextTick, ref } from 'vue'
import { useScopedExerciseDraft } from './useScopedExerciseDraft'

describe('scoped exercise drafts', () => {
 it('restores after reload and removes only after server confirmation', async () => {
  sessionStorage.clear()
  const key = ref('user:1:es:j:s:q')
  const first = effectScope(); const answer = ref('')
  first.run(() => useScopedExerciseDraft(key, () => answer.value, saved => { answer.value = String(saved) }, () => false))
  answer.value = 'partial'; await nextTick(); first.stop()
  const second = effectScope(); const restored = ref(''); const confirmed = ref(false)
  second.run(() => useScopedExerciseDraft(key, () => restored.value, saved => { restored.value = String(saved) }, () => confirmed.value))
  expect(restored.value).toBe('partial')
  expect(sessionStorage.getItem(key.value)).toBe('"partial"')
  confirmed.value = true; await nextTick()
  expect(sessionStorage.getItem(key.value)).toBeNull()
  second.stop()
 })
 it('isolates new questions and ignores corrupt drafts and ordinary exercises', async () => {
  sessionStorage.clear()
  sessionStorage.setItem('other', 'broken JSON')
  const key = ref('first'); const answer = ref('first answer'); const lifetime = effectScope()
  lifetime.run(() => useScopedExerciseDraft(key, () => answer.value, saved => { answer.value = String(saved) }, () => false))
  answer.value = 'saved'; await nextTick()
  key.value = 'other'; answer.value = ''; await nextTick()
  expect(answer.value).toBe('')
  expect(sessionStorage.getItem('first')).toBe('"saved"')
  key.value = ''; answer.value = 'ordinary'; await nextTick()
  expect(sessionStorage.getItem('')).toBeNull()
  lifetime.stop()
 })
})
