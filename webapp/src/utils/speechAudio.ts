/** Heuristic max recording time: phrase length + lead/tail padding. */
export function estimateSpeakingMaxDurationMs(referenceText: string): number {
  const text = referenceText.trim()
  if (!text) {
    return 15000
  }

  const words = text.split(/\s+/).filter(Boolean).length
  const chars = text.length
  const phraseMs = Math.max(2500, words * 450 + chars * 35)
  const leadMs = 5000
  const tailMs = 5000
  const total = leadMs + phraseMs + tailMs
  return Math.min(45000, Math.max(12000, total))
}

export interface SpeechAnalysis {
  hasSpeech: boolean
  durationMs: number
  peakRms: number
  speechFrameRatio: number
}

export interface ProcessedSpeechAudio {
  blob: Blob
  analysis: SpeechAnalysis
}

const FRAME_MS = 20
const SPEECH_RMS = 0.018
const MIN_PEAK_RMS = 0.012
const MIN_SPEECH_RATIO = 0.06
const TRIM_PAD_MS = 180

function frameRms(samples: Float32Array, start: number, size: number): number {
  let sum = 0
  const end = Math.min(samples.length, start + size)
  for (let i = start; i < end; i++) {
    sum += samples[i] * samples[i]
  }
  if (end <= start) {
    return 0
  }
  return Math.sqrt(sum / (end - start))
}

function computeFrameRmsSeries(channel: Float32Array, sampleRate: number): number[] {
  const frameSize = Math.max(1, Math.floor((sampleRate * FRAME_MS) / 1000))
  const frames: number[] = []
  for (let i = 0; i < channel.length; i += frameSize) {
    frames.push(frameRms(channel, i, frameSize))
  }
  return frames
}

export function analyzeSpeechBuffer(buffer: AudioBuffer): SpeechAnalysis {
  const channel = mixToMono(buffer)
  const frames = computeFrameRmsSeries(channel, buffer.sampleRate)
  if (frames.length === 0) {
    return { hasSpeech: false, durationMs: 0, peakRms: 0, speechFrameRatio: 0 }
  }

  const peakRms = Math.max(...frames)
  const speechFrames = frames.filter((r) => r >= SPEECH_RMS)
  const speechFrameRatio = speechFrames.length / frames.length
  const minRms = Math.min(...frames)
  const dynamicRange = peakRms - minRms

  let hasSpeech = peakRms >= MIN_PEAK_RMS && speechFrameRatio >= MIN_SPEECH_RATIO

  if (hasSpeech && speechFrames.length >= 4) {
    const mean = speechFrames.reduce((a, b) => a + b, 0) / speechFrames.length
    const variance =
      speechFrames.reduce((acc, r) => acc + (r - mean) * (r - mean), 0) / speechFrames.length
    if (variance < 0.000008 && dynamicRange < 0.012) {
      hasSpeech = false
    }
  }

  return {
    hasSpeech,
    durationMs: Math.round((buffer.length / buffer.sampleRate) * 1000),
    peakRms,
    speechFrameRatio,
  }
}

function mixToMono(buffer: AudioBuffer): Float32Array {
  if (buffer.numberOfChannels === 1) {
    return buffer.getChannelData(0).slice()
  }
  const len = buffer.length
  const out = new Float32Array(len)
  for (let ch = 0; ch < buffer.numberOfChannels; ch++) {
    const data = buffer.getChannelData(ch)
    for (let i = 0; i < len; i++) {
      out[i] += data[i] / buffer.numberOfChannels
    }
  }
  return out
}

function findSpeechBounds(frames: number[]): { startFrame: number; endFrame: number } {
  let startFrame = -1
  let endFrame = -1
  for (let i = 0; i < frames.length; i++) {
    if (frames[i] >= SPEECH_RMS) {
      startFrame = i
      break
    }
  }
  for (let i = frames.length - 1; i >= 0; i--) {
    if (frames[i] >= SPEECH_RMS) {
      endFrame = i
      break
    }
  }
  if (startFrame < 0 || endFrame < 0 || endFrame < startFrame) {
    return { startFrame: 0, endFrame: Math.max(0, frames.length - 1) }
  }

  const padFrames = Math.ceil(TRIM_PAD_MS / FRAME_MS)
  return {
    startFrame: Math.max(0, startFrame - padFrames),
    endFrame: Math.min(frames.length - 1, endFrame + padFrames),
  }
}

function sliceBuffer(buffer: AudioBuffer, startSample: number, endSample: number): AudioBuffer {
  const length = Math.max(1, endSample - startSample)
  const out = new AudioBuffer({
    length,
    sampleRate: buffer.sampleRate,
    numberOfChannels: buffer.numberOfChannels,
  })
  for (let ch = 0; ch < buffer.numberOfChannels; ch++) {
    const src = buffer.getChannelData(ch)
    out.getChannelData(ch).set(src.subarray(startSample, endSample))
  }
  return out
}

function trimSilenceBuffer(buffer: AudioBuffer): AudioBuffer {
  const channel = mixToMono(buffer)
  const frames = computeFrameRmsSeries(channel, buffer.sampleRate)
  const { startFrame, endFrame } = findSpeechBounds(frames)
  const frameSize = Math.max(1, Math.floor((buffer.sampleRate * FRAME_MS) / 1000))
  const startSample = startFrame * frameSize
  const endSample = Math.min(buffer.length, (endFrame + 1) * frameSize)
  if (endSample <= startSample) {
    return buffer
  }
  return sliceBuffer(buffer, startSample, endSample)
}

function encodeWav(buffer: AudioBuffer): Blob {
  const numChannels = buffer.numberOfChannels
  const sampleRate = buffer.sampleRate
  const bitsPerSample = 16
  const blockAlign = (numChannels * bitsPerSample) / 8
  const byteRate = sampleRate * blockAlign
  const samples = buffer.length
  const dataSize = samples * blockAlign
  const arrayBuffer = new ArrayBuffer(44 + dataSize)
  const view = new DataView(arrayBuffer)

  const writeString = (offset: number, str: string) => {
    for (let i = 0; i < str.length; i++) {
      view.setUint8(offset + i, str.charCodeAt(i))
    }
  }

  writeString(0, 'RIFF')
  view.setUint32(4, 36 + dataSize, true)
  writeString(8, 'WAVE')
  writeString(12, 'fmt ')
  view.setUint32(16, 16, true)
  view.setUint16(20, 1, true)
  view.setUint16(22, numChannels, true)
  view.setUint32(24, sampleRate, true)
  view.setUint32(28, byteRate, true)
  view.setUint16(32, blockAlign, true)
  view.setUint16(34, bitsPerSample, true)
  writeString(36, 'data')
  view.setUint32(40, dataSize, true)

  let offset = 44
  for (let i = 0; i < samples; i++) {
    for (let ch = 0; ch < numChannels; ch++) {
      const sample = Math.max(-1, Math.min(1, buffer.getChannelData(ch)[i]))
      view.setInt16(offset, sample < 0 ? sample * 0x8000 : sample * 0x7fff, true)
      offset += 2
    }
  }

  return new Blob([arrayBuffer], { type: 'audio/wav' })
}

export async function processRecordedSpeech(blob: Blob): Promise<ProcessedSpeechAudio> {
  const arrayBuffer = await blob.arrayBuffer()
  const audioContext = new AudioContext()
  try {
    const decoded = await audioContext.decodeAudioData(arrayBuffer.slice(0))
    const trimmed = trimSilenceBuffer(decoded)
    const analysis = analyzeSpeechBuffer(trimmed)
    const wav = encodeWav(trimmed)
    return { blob: wav, analysis }
  } finally {
    await audioContext.close()
  }
}

export function recordingFileName(blob: Blob): string {
  return blob.type.includes('wav') ? 'recording.wav' : 'recording.webm'
}
