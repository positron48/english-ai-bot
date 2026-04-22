import { apiClient } from '../api/client'

// Audio engine
let ctx: AudioContext | null = null
let master: GainNode | null = null
let wordAudio: HTMLAudioElement | null = null

/** Только успешные URL; отсутствие озвучки / ошибки запроса не кэшируем — повторный запрос возможен сразу. */
const pronunciationCache = new Map<string, string>()
const pronunciationInFlight = new Map<string, Promise<string | null>>()

interface AudioState {
  volume: number
  speed: number
  picked: {
    success: string | null
    fail: string | null
  }
}

const state: AudioState = {
  volume: 0.55,
  speed: 1.0,
  picked: {
    success: null,
    fail: null,
  }
}

function ensureAudio(): AudioContext {
  if (!ctx) {
    ctx = new (window.AudioContext || (window as any).webkitAudioContext)()
    master = ctx.createGain()
    master.gain.value = state.volume
    master.connect(ctx.destination)
  }
  return ctx
}

async function unlock() {
  const c = ensureAudio()
  if (c.state === 'suspended') await c.resume()
  // короткий "пинг" нулевой громкости для уверенного старта
  const o = c.createOscillator()
  const g = c.createGain()
  g.gain.value = 0.00001
  o.frequency.value = 440
  o.connect(g)
  g.connect(master!)
  const t = c.currentTime
  o.start(t)
  o.stop(t + 0.02)
}

function now(): number {
  return ensureAudio().currentTime
}

function mkGainEnvelope(
  g: GainNode,
  t0: number,
  attack: number,
  hold: number,
  release: number,
  peak: number
) {
  const a = Math.max(0.001, attack)
  const r = Math.max(0.001, release)
  const h = hold // Use hold variable
  g.gain.cancelScheduledValues(t0)
  g.gain.setValueAtTime(0.00001, t0)
  g.gain.exponentialRampToValueAtTime(Math.max(0.0002, peak), t0 + a)
  g.gain.setValueAtTime(Math.max(0.0002, peak), t0 + a + h)
  g.gain.exponentialRampToValueAtTime(0.00001, t0 + a + h + r)
}

function osc(
  type: OscillatorType,
  freq: number,
  t0: number,
  dur: number,
  peak: number = 0.25,
  attack: number = 0.01,
  hold: number = 0.02,
  release: number = 0.08
) {
  const c = ensureAudio()
  const o = c.createOscillator()
  const g = c.createGain()
  o.type = type
  o.frequency.setValueAtTime(freq, t0)
  // Use hold parameter, but ensure it doesn't exceed available duration
  const actualHold = Math.max(0, Math.min(hold, dur - attack - release))
  mkGainEnvelope(g, t0, attack, actualHold, release, peak)
  o.connect(g)
  g.connect(master!)
  o.start(t0)
  o.stop(t0 + dur + 0.02)
  return { o, g }
}

function noise(
  t0: number,
  dur: number,
  peak: number = 0.18,
  attack: number = 0.005,
  release: number = 0.08,
  toneHz: number = 0
) {
  const c = ensureAudio()
  const bufferSize = Math.max(1, Math.floor(c.sampleRate * dur))
  const buffer = c.createBuffer(1, bufferSize, c.sampleRate)
  const data = buffer.getChannelData(0)
  for (let i = 0; i < bufferSize; i++) data[i] = Math.random() * 2 - 1

  const src = c.createBufferSource()
  src.buffer = buffer

  const g = c.createGain()
  mkGainEnvelope(g, t0, attack, Math.max(0, dur - attack - release), release, peak)

  if (toneHz > 0) {
    const bp = c.createBiquadFilter()
    bp.type = 'bandpass'
    bp.frequency.value = toneHz
    bp.Q.value = 6
    src.connect(bp)
    bp.connect(g)
  } else {
    src.connect(g)
  }

  g.connect(master!)
  src.start(t0)
  src.stop(t0 + dur + 0.02)
  return { src, g }
}

function seq(steps: Array<{ fn: (t: number) => void; at: number }>) {
  // steps: [{fn:()=>void, at:secondsFromNow}, ...]
  const t0 = now()
  for (const s of steps) {
    s.fn(t0 + s.at * state.speed)
  }
}

function msToS(ms: number): number {
  return (ms / 1000) * state.speed
}

function normalizePronunciationWord(raw: string): string {
  return raw.trim().toLowerCase().replace(/\s+/g, ' ')
}

function stripInfinitiveToPrefix(normalized: string): string {
  if (!normalized.startsWith('to ')) return ''
  const stripped = normalized.slice(3).trim()
  return stripped
}

function playMelody(
  notes: Array<{ freq: number; durMs: number; type: OscillatorType; peak: number }>,
  opts: { gapMs?: number; releaseMs?: number } = {}
) {
  const gapMs = opts.gapMs || 0
  const releaseMs = opts.releaseMs || 80
  const t0 = now()
  let currentTimeMs = 0

  for (const note of notes) {
    const t = t0 + msToS(currentTimeMs)
    const dur = msToS(note.durMs)
    osc(note.type, note.freq, t, dur, note.peak, msToS(5), msToS(10), msToS(releaseMs))
    currentTimeMs += note.durMs + gapMs
  }
}

// Sound themes
interface SoundTheme {
  id: string
  name: string
  success: {
    id: string
    name: string
    hint: string
    play: () => void
  }
  fail: {
    id: string
    name: string
    hint: string
    play: () => void
  }
  victory?: {
    id: string
    name: string
    hint: string
    play: () => void
  }
  defeat?: {
    id: string
    name: string
    hint: string
    play: () => void
  }
}

const themes: Record<string, SoundTheme> = {
  tick: {
    id: 'tick',
    name: 'Tick',
    success: {
      id: 'S9',
      name: 'Быстрый "тик"',
      hint: 'очень коротко',
      play() {
        const t = now()
        osc('square', 1400, t, msToS(35), 0.07, msToS(1), 0, msToS(20))
      }
    },
    fail: {
      id: 'F4',
      name: 'Короткий "бип-бип"',
      hint: 'двойной сигнал',
      play() {
        seq([
          { at: 0, fn: (t) => osc('square', 420, t, msToS(70), 0.10, msToS(2), 0, msToS(50)) },
          { at: 0.10, fn: (t) => osc('square', 420, t, msToS(70), 0.10, msToS(2), 0, msToS(50)) },
        ])
      }
    },
    victory: {
      id: 'V5',
      name: 'Триумф',
      hint: 'подъём с паузами',
      play() {
        const t = now()
        noise(t, msToS(60), 0.08, msToS(2), msToS(40), 1400)
        setTimeout(() => {
          noise(now(), msToS(60), 0.08, msToS(2), msToS(40), 1400)
        }, 240 * state.speed)
        playMelody(
          [
            { freq: 392.0, durMs: 220, type: 'triangle', peak: 0.09 },
            { freq: 523.25, durMs: 220, type: 'triangle', peak: 0.09 },
            { freq: 659.25, durMs: 260, type: 'triangle', peak: 0.1 },
            { freq: 783.99, durMs: 420, type: 'triangle', peak: 0.1 }
          ],
          { gapMs: 70, releaseMs: 220 }
        )
      }
    },
    defeat: {
      id: 'D3',
      name: 'Тревога',
      hint: 'две ноты спорят',
      play() {
        playMelody(
          [
            { freq: 440, durMs: 180, type: 'triangle', peak: 0.08 },
            { freq: 466.16, durMs: 180, type: 'triangle', peak: 0.08 },
            { freq: 440, durMs: 180, type: 'triangle', peak: 0.08 },
            { freq: 466.16, durMs: 240, type: 'triangle', peak: 0.08 },
            { freq: 392, durMs: 520, type: 'sine', peak: 0.09 }
          ],
          { gapMs: 25, releaseMs: 220 }
        )
      }
    }
  },
  blob: {
    id: 'blob',
    name: 'Blob',
    success: {
      id: 'S6',
      name: 'Пик-ап',
      hint: 'тон с глиссандо вверх',
      play() {
        const c = ensureAudio()
        const t = c.currentTime
        const o = c.createOscillator()
        const g = c.createGain()
        o.type = 'sine'
        o.frequency.setValueAtTime(520, t)
        o.frequency.exponentialRampToValueAtTime(1040, t + msToS(140))
        mkGainEnvelope(g, t, msToS(6), msToS(40), msToS(140), 0.16)
        o.connect(g)
        g.connect(master!)
        o.start(t)
        o.stop(t + msToS(220) + 0.03)
      }
    },
    fail: {
      id: 'F5',
      name: 'Минорная "капля"',
      hint: '2 ноты вниз (минор)',
      play() {
        const base = 523.25 // C5
        const freqs = [base * 1.4983, base * 1.1892] // G Eb
        const step = 0.08
        seq(freqs.map((f, i) => ({
          at: i * step,
          fn: (t) => osc('sine', f, t, msToS(140), 0.16, msToS(5), msToS(10), msToS(90))
        })))
      }
    },
    victory: {
      id: 'V4',
      name: 'Мягкая победа',
      hint: 'мелодично и спокойно',
      play() {
        playMelody(
          [
            { freq: 587.33, durMs: 220, type: 'sine', peak: 0.09 },
            { freq: 739.99, durMs: 220, type: 'sine', peak: 0.09 },
            { freq: 880.0, durMs: 260, type: 'sine', peak: 0.09 },
            { freq: 739.99, durMs: 220, type: 'sine', peak: 0.08 },
            { freq: 987.77, durMs: 360, type: 'sine', peak: 0.1 }
          ],
          { gapMs: 55, releaseMs: 200 }
        )
      }
    },
    defeat: {
      id: 'D1',
      name: 'Печальный спад',
      hint: 'минорный спуск',
      play() {
        playMelody(
          [
            { freq: 659.25, durMs: 260, type: 'sine', peak: 0.09 },
            { freq: 587.33, durMs: 240, type: 'sine', peak: 0.085 },
            { freq: 523.25, durMs: 260, type: 'sine', peak: 0.085 },
            { freq: 392.0, durMs: 520, type: 'sine', peak: 0.09 }
          ],
          { gapMs: 65, releaseMs: 260 }
        )
      }
    }
  }
}

export function useAudio() {
  const playSuccess = (themeId: string = 'tick') => {
    const theme = themes[themeId]
    if (!theme) return
    unlock().then(() => {
      theme.success.play()
    })
  }

  const playFail = (themeId: string = 'tick') => {
    const theme = themes[themeId]
    if (!theme) return
    unlock().then(() => {
      theme.fail.play()
    })
  }

  const playVictory = (themeId: string = 'tick') => {
    const theme = themes[themeId]
    if (!theme || !theme.victory) return
    unlock().then(() => {
      theme.victory!.play()
    })
  }

  const playDefeat = (themeId: string = 'tick') => {
    const theme = themes[themeId]
    if (!theme || !theme.defeat) return
    unlock().then(() => {
      theme.defeat!.play()
    })
  }

  const setVolume = (volume: number) => {
    state.volume = volume
    if (master) {
      master.gain.value = volume
    }
  }

  const getThemes = () => {
    return Object.values(themes)
  }

  const previewTheme = async (themeId: string = 'tick') => {
    const theme = themes[themeId]
    if (!theme) return

    await unlock()

    // Play sounds sequentially with delays
    // Success
    theme.success.play()
    await new Promise(resolve => setTimeout(resolve, 500 * state.speed))

    // Fail
    theme.fail.play()
    await new Promise(resolve => setTimeout(resolve, 500 * state.speed))

    // Victory (if exists)
    if (theme.victory) {
      theme.victory.play()
      await new Promise(resolve => setTimeout(resolve, 2000 * state.speed))
    }

    // Defeat (if exists)
    if (theme.defeat) {
      theme.defeat.play()
      await new Promise(resolve => setTimeout(resolve, 2000 * state.speed))
    }
  }

  const getWordPronunciationURL = async (word: string): Promise<string | null> => {
    const normalized = normalizePronunciationWord(word)
    if (!normalized) return null

    const cached = pronunciationCache.get(normalized)
    if (cached) {
      return cached
    }

    if (pronunciationInFlight.has(normalized)) {
      return pronunciationInFlight.get(normalized) as Promise<string | null>
    }

    const requestURL = async (normalizedWord: string): Promise<string | null> => {
      try {
        const data = await apiClient.request<{ available: boolean; url: string }>(
          `/api/tts/word?word=${encodeURIComponent(normalizedWord)}`
        )
        return data?.available && data?.url ? data.url : null
      } catch {
        return null
      }
    }

    const request = (async () => {
      try {
        let url = await requestURL(normalized)
        // English infinitive display form fallback: "to X" -> "X".
        if (!url) {
          const stripped = stripInfinitiveToPrefix(normalized)
          if (stripped) {
            url = await requestURL(stripped)
            if (url) {
              pronunciationCache.set(stripped, url)
            }
          }
        }
        // Кэшируем только успешный URL: pending/ошибка не должны «замораживать» отсутствие кнопки до reload.
        if (url) {
          pronunciationCache.set(normalized, url)
        }
        return url
      } finally {
        pronunciationInFlight.delete(normalized)
      }
    })()

    pronunciationInFlight.set(normalized, request)
    return request
  }

  const playWordPronunciation = async (word: string): Promise<boolean> => {
    const url = await getWordPronunciationURL(word)
    if (!url) return false

    try {
      if (wordAudio) {
        wordAudio.pause()
        wordAudio.currentTime = 0
      }
      wordAudio = new Audio(url)
      wordAudio.preload = 'auto'
      await wordAudio.play()
      return true
    } catch {
      return false
    }
  }

  return {
    playSuccess,
    playFail,
    playVictory,
    playDefeat,
    setVolume,
    getThemes,
    previewTheme,
    getWordPronunciationURL,
    playWordPronunciation
  }
}
