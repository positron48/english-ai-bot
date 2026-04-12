import { onUnmounted, ref } from 'vue'

/** Countdown after a wrong answer (aligned with TrainingView delay UX). */
export function useTrainingAnswerDelay() {
  const waitingDelay = ref(false)
  const remainingMs = ref(0)
  const initialDelayMs = ref(0)
  const delaySeconds = ref(0)
  let intervalId: ReturnType<typeof setInterval> | null = null
  let timeoutId: ReturnType<typeof setTimeout> | null = null

  function clearAll() {
    if (intervalId != null) {
      clearInterval(intervalId)
      intervalId = null
    }
    if (timeoutId != null) {
      clearTimeout(timeoutId)
      timeoutId = null
    }
    waitingDelay.value = false
    initialDelayMs.value = 0
    remainingMs.value = 0
    delaySeconds.value = 0
  }

  onUnmounted(clearAll)

  /**
   * delaySecondsParam > 0: show ring + tick every 100ms.
   * Otherwise: short timeout (fallbackMs) then onComplete.
   */
  function runWrongAnswerDelay(delaySecondsParam: number, onComplete: () => void, fallbackMs = 150) {
    clearAll()
    if (delaySecondsParam <= 0) {
      timeoutId = setTimeout(() => {
        timeoutId = null
        onComplete()
      }, fallbackMs)
      return
    }
    const totalMs = delaySecondsParam * 1000
    initialDelayMs.value = totalMs
    remainingMs.value = totalMs
    delaySeconds.value = Math.ceil(totalMs / 1000)
    waitingDelay.value = true
    const start = Date.now()
    intervalId = setInterval(() => {
      const left = Math.max(0, totalMs - (Date.now() - start))
      remainingMs.value = left
      delaySeconds.value = Math.max(0, Math.ceil(left / 1000))
      if (left <= 0 && intervalId != null) {
        clearInterval(intervalId)
        intervalId = null
        waitingDelay.value = false
        initialDelayMs.value = 0
        remainingMs.value = 0
        delaySeconds.value = 0
        onComplete()
      }
    }, 100)
  }

  return { waitingDelay, remainingMs, initialDelayMs, delaySeconds, runWrongAnswerDelay, clearAll }
}
