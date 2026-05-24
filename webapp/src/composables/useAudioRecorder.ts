import { ref } from 'vue'

export type RecorderState = 'idle' | 'recording' | 'recorded' | 'error'

const MAX_DURATION_MS = 15000

export function useAudioRecorder() {
  const state = ref<RecorderState>('idle')
  const errorMessage = ref('')
  const blob = ref<Blob | null>(null)
  const durationMs = ref(0)

  let mediaRecorder: MediaRecorder | null = null
  let chunks: BlobPart[] = []
  let stream: MediaStream | null = null
  let startedAt = 0
  let stopTimer: ReturnType<typeof setTimeout> | null = null

  async function startRecording(): Promise<void> {
    errorMessage.value = ''
    blob.value = null
    durationMs.value = 0
    try {
      stream = await navigator.mediaDevices.getUserMedia({ audio: true })
      const mime = MediaRecorder.isTypeSupported('audio/webm;codecs=opus')
        ? 'audio/webm;codecs=opus'
        : 'audio/webm'
      mediaRecorder = new MediaRecorder(stream, { mimeType: mime })
      chunks = []
      mediaRecorder.ondataavailable = (e) => {
        if (e.data.size > 0) chunks.push(e.data)
      }
      mediaRecorder.onstop = () => {
        durationMs.value = Date.now() - startedAt
        blob.value = new Blob(chunks, { type: mime })
        state.value = 'recorded'
        cleanupStream()
      }
      mediaRecorder.onerror = () => {
        state.value = 'error'
        errorMessage.value = 'Recording failed'
        cleanupStream()
      }
      startedAt = Date.now()
      mediaRecorder.start()
      state.value = 'recording'
      stopTimer = setTimeout(() => {
        if (state.value === 'recording') stopRecording()
      }, MAX_DURATION_MS)
    } catch (e: unknown) {
      state.value = 'error'
      errorMessage.value = e instanceof Error ? e.message : 'Microphone access denied'
    }
  }

  function stopRecording(): void {
    if (stopTimer) {
      clearTimeout(stopTimer)
      stopTimer = null
    }
    if (mediaRecorder && mediaRecorder.state !== 'inactive') {
      mediaRecorder.stop()
    } else {
      cleanupStream()
    }
  }

  function resetRecording(): void {
    if (stopTimer) clearTimeout(stopTimer)
    stopTimer = null
    if (mediaRecorder && mediaRecorder.state === 'recording') {
      mediaRecorder.stop()
    }
    cleanupStream()
    blob.value = null
    durationMs.value = 0
    errorMessage.value = ''
    state.value = 'idle'
  }

  function cleanupStream(): void {
    if (stream) {
      stream.getTracks().forEach((t) => t.stop())
      stream = null
    }
  }

  return {
    state,
    errorMessage,
    blob,
    durationMs,
    startRecording,
    stopRecording,
    resetRecording,
  }
}
