import { onUnmounted, ref, shallowRef } from 'vue'

// Minimal typings for the Web Speech API, which is not in the standard TS lib.
interface SpeechRecognitionAlternativeLike {
  transcript: string
}
interface SpeechRecognitionResultLike {
  0: SpeechRecognitionAlternativeLike
  isFinal: boolean
  length: number
}
interface SpeechRecognitionResultListLike {
  length: number
  [index: number]: SpeechRecognitionResultLike
}
interface SpeechRecognitionEventLike {
  resultIndex: number
  results: SpeechRecognitionResultListLike
}
interface SpeechRecognitionLike {
  lang: string
  continuous: boolean
  interimResults: boolean
  start: () => void
  stop: () => void
  abort: () => void
  onresult: ((e: SpeechRecognitionEventLike) => void) | null
  onerror: ((e: { error?: string }) => void) | null
  onend: (() => void) | null
}
type SpeechRecognitionCtor = new () => SpeechRecognitionLike

function getRecognitionCtor(): SpeechRecognitionCtor | null {
  if (typeof window === 'undefined') return null
  const w = window as any
  return (w.SpeechRecognition || w.webkitSpeechRecognition || null) as SpeechRecognitionCtor | null
}

// Native bridge exposed by the embedded Android APK (see MainActivity.java).
// The Android WebView has no Web Speech API, so we route through the native
// SpeechRecognizer when the bridge is present.
interface NativeSpeechBridge {
  speechRecognitionAvailable?: () => boolean
  startSpeechRecognition?: (lang: string) => void
  stopSpeechRecognition?: () => void
}
function getNativeBridge(): NativeSpeechBridge | null {
  if (typeof window === 'undefined') return null
  const bridge = (window as any).QantrixAndroid as NativeSpeechBridge | undefined
  if (!bridge || typeof bridge.startSpeechRecognition !== 'function') return null
  try {
    if (bridge.speechRecognitionAvailable && !bridge.speechRecognitionAvailable()) return null
  } catch {
    return null
  }
  return bridge
}

// Map a target_lang code (e.g. "es", "en") to a BCP-47 recognition locale.
const LANG_MAP: Record<string, string> = {
  es: 'es-ES',
  en: 'en-US',
  ru: 'ru-RU',
}
function mapLang(code: string): string {
  return LANG_MAP[code] ?? code
}

function capitalizeFirst(text: string): string {
  return text.charAt(0).toLocaleUpperCase() + text.slice(1)
}

export function useSpeechRecognition(options: {
  lang: () => string
  onFinalTranscript: (text: string) => void
}) {
  const supported = getNativeBridge() !== null || getRecognitionCtor() !== null
  const listening = ref(false)
  const recognition = shallowRef<SpeechRecognitionLike | null>(null)

  // --- Native (embedded Android APK) path --------------------------------
  function startNative(bridge: NativeSpeechBridge) {
    const w = window as any
    w.__onSpeechResult = (res: { transcript?: string; error?: string }) => {
      if (res && res.transcript && res.transcript.trim()) {
        options.onFinalTranscript(capitalizeFirst(res.transcript.trim()))
      }
    }
    w.__onSpeechState = (state: string) => {
      listening.value = state === 'listening'
    }
    listening.value = true
    try {
      bridge.startSpeechRecognition!(mapLang(options.lang()))
    } catch {
      listening.value = false
    }
  }
  function stopNative(bridge: NativeSpeechBridge) {
    try { bridge.stopSpeechRecognition?.() } catch { /* ignore */ }
    listening.value = false
  }

  function stop() {
    const bridge = getNativeBridge()
    if (bridge) {
      stopNative(bridge)
      return
    }
    const rec = recognition.value
    if (rec) {
      try { rec.stop() } catch { /* ignore */ }
    }
    listening.value = false
  }

  function start() {
    // Toggle behaviour: a second tap stops an active session.
    if (listening.value) {
      stop()
      return
    }
    const bridge = getNativeBridge()
    if (bridge) {
      startNative(bridge)
      return
    }
    const Ctor = getRecognitionCtor()
    if (!Ctor) return
    const rec = new Ctor()
    rec.lang = mapLang(options.lang())
    rec.continuous = false
    rec.interimResults = false

    rec.onresult = (e) => {
      let finalText = ''
      for (let i = e.resultIndex; i < e.results.length; i++) {
        const r = e.results[i]
        if (r.isFinal) finalText += r[0].transcript
      }
      if (finalText.trim()) options.onFinalTranscript(capitalizeFirst(finalText.trim()))
    }
    rec.onerror = () => { listening.value = false }
    rec.onend = () => {
      listening.value = false
      recognition.value = null
    }

    recognition.value = rec
    try {
      rec.start()
      listening.value = true
    } catch {
      listening.value = false
      recognition.value = null
    }
  }

  onUnmounted(stop)

  return { supported, listening, start, stop }
}
