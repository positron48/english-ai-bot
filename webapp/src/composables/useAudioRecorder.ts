import { onUnmounted, ref } from 'vue'
import {
  estimateSpeakingMaxDurationMs,
  processRecordedSpeech,
  type ProcessedSpeechAudio,
} from '../utils/speechAudio'

export type RecorderState = 'idle' | 'recording' | 'processing' | 'ready' | 'no_speech' | 'error'

export interface StartRecordingOptions {
  referenceText?: string
  maxDurationMs?: number
}

export function useAudioRecorder() {
  const state = ref<RecorderState>('idle')
  const errorMessage = ref('')
  const blob = ref<Blob | null>(null)
  const durationMs = ref(0)
  const maxDurationMs = ref(15000)
  const remainingMs = ref(0)
  const speechAnalysis = ref<ProcessedSpeechAudio['analysis'] | null>(null)

  let mediaRecorder: MediaRecorder | null = null
  let chunks: BlobPart[] = []
  let stream: MediaStream | null = null
  let startedAt = 0
  let stopTimer: ReturnType<typeof setTimeout> | null = null
  let tickTimer: ReturnType<typeof setInterval> | null = null
  let recordedMime = 'audio/webm'
  let onCompleteCallback: ((result: ProcessedSpeechAudio | null) => void) | null = null

  function clearTimers(): void {
    if (stopTimer) {
      clearTimeout(stopTimer)
      stopTimer = null
    }
    if (tickTimer) {
      clearInterval(tickTimer)
      tickTimer = null
    }
  }

  async function handleRecordingStopped(rawBlob: Blob): Promise<void> {
    durationMs.value = Date.now() - startedAt
    state.value = 'processing'
    blob.value = null
    speechAnalysis.value = null
    try {
      const processed = await processRecordedSpeech(rawBlob)
      speechAnalysis.value = processed.analysis
      if (!processed.analysis.hasSpeech) {
        state.value = 'no_speech'
        onCompleteCallback?.(processed)
        return
      }
      blob.value = processed.blob
      state.value = 'ready'
      onCompleteCallback?.(processed)
    } catch (e: unknown) {
      state.value = 'error'
      errorMessage.value = e instanceof Error ? e.message : 'Failed to process recording'
      onCompleteCallback?.(null)
    }
  }

  async function startRecording(
    options: StartRecordingOptions = {},
    onComplete?: (result: ProcessedSpeechAudio | null) => void
  ): Promise<void> {
    errorMessage.value = ''
    blob.value = null
    durationMs.value = 0
    speechAnalysis.value = null
    onCompleteCallback = onComplete ?? null

    const limit =
      options.maxDurationMs ??
      estimateSpeakingMaxDurationMs(options.referenceText ?? '')
    maxDurationMs.value = limit
    remainingMs.value = limit

    try {
      stream = await navigator.mediaDevices.getUserMedia({
        audio: {
          echoCancellation: true,
          noiseSuppression: true,
          autoGainControl: true,
        },
      })
      recordedMime = MediaRecorder.isTypeSupported('audio/webm;codecs=opus')
        ? 'audio/webm;codecs=opus'
        : 'audio/webm'
      mediaRecorder = new MediaRecorder(stream, { mimeType: recordedMime })
      chunks = []
      mediaRecorder.ondataavailable = (e) => {
        if (e.data.size > 0) chunks.push(e.data)
      }
      mediaRecorder.onstop = () => {
        clearTimers()
        remainingMs.value = 0
        cleanupStream()
        const rawBlob = new Blob(chunks, { type: recordedMime })
        void handleRecordingStopped(rawBlob)
      }
      mediaRecorder.onerror = () => {
        clearTimers()
        state.value = 'error'
        errorMessage.value = 'Recording failed'
        cleanupStream()
        onCompleteCallback?.(null)
      }
      startedAt = Date.now()
      mediaRecorder.start()
      state.value = 'recording'
      stopTimer = setTimeout(() => {
        if (state.value === 'recording') stopRecording()
      }, limit)
      tickTimer = setInterval(() => {
        if (state.value !== 'recording') return
        remainingMs.value = Math.max(0, limit - (Date.now() - startedAt))
      }, 100)
    } catch (e: unknown) {
      state.value = 'error'
      errorMessage.value = e instanceof Error ? e.message : 'Microphone access denied'
      onCompleteCallback?.(null)
    }
  }

  function stopRecording(): void {
    if (state.value !== 'recording') return
    clearTimers()
    remainingMs.value = 0
    if (mediaRecorder && mediaRecorder.state !== 'inactive') {
      mediaRecorder.stop()
    } else {
      cleanupStream()
    }
  }

  function resetRecording(): void {
    clearTimers()
    if (mediaRecorder && mediaRecorder.state === 'recording') {
      mediaRecorder.stop()
    }
    cleanupStream()
    blob.value = null
    durationMs.value = 0
    remainingMs.value = 0
    speechAnalysis.value = null
    errorMessage.value = ''
    state.value = 'idle'
    onCompleteCallback = null
  }

  function cleanupStream(): void {
    if (stream) {
      stream.getTracks().forEach((t) => t.stop())
      stream = null
    }
  }

  onUnmounted(() => {
    resetRecording()
  })

  return {
    state,
    errorMessage,
    blob,
    durationMs,
    maxDurationMs,
    remainingMs,
    speechAnalysis,
    startRecording,
    stopRecording,
    resetRecording,
  }
}
