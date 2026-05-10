<template>
  <div class="screen">
    <main class="reader-shell">
      <header class="header">
        <button type="button" class="back-button" aria-label="Назад" @click="goBack">
          <Icon name="arrow-left" />
        </button>
        <h1 class="title">{{ block.reading_passage?.title || block.title }}</h1>
        <div class="header-actions">
          <button type="button" class="icon-button" aria-label="Показать перевод" @click="showTranslation = !showTranslation">
            文A
          </button>
          <button
            type="button"
            class="icon-button"
            :class="{ active: isAutoplaying }"
            aria-label="Автовоспроизведение"
            @click="toggleAutoplay"
          >
            <Icon name="play" />
          </button>
        </div>
      </header>

      <section class="content">
        <div class="text-flow">
          <div
            v-for="segment in segments"
            :key="segment.segment_id"
            class="sentence-row"
            :class="[{ narrator: isNarrator(segment), active: activeSegmentId === segment.segment_id }]"
            @click="playSingleSegment(segment)"
          >
            <div v-if="!isNarrator(segment)" class="speaker-icon" :class="speakerClass(segment.speaker_id)">
              <Icon name="users" />
            </div>
            <div v-else></div>

            <div>
              <div class="sentence-text">
                <span
                  v-for="(token, tokenIndex) in segment.tokens || []"
                  :key="`${segment.segment_id}-${token.token_idx}`"
                  class="token"
                  :class="{
                    clickable: token.clickable,
                    'word-selected': selectedTokenKey === tokenKey(segment.segment_id, token.token_idx),
                  }"
                  @click.stop="onTokenClick($event, token, segment)"
                >
                  {{ tokenText(segment.tokens || [], tokenIndex) }}
                </span>
              </div>
              <div v-if="segment.text_translation_ru" class="translation" :class="{ hidden: !showTranslation }">
                {{ segment.text_translation_ru }}
              </div>
            </div>

            <button
              v-if="segment.audio_rel_path"
              type="button"
              class="sentence-audio-button"
              aria-label="Озвучить предложение"
              @click.stop="playSingleSegment(segment)"
            >
              <Icon name="play" />
            </button>
          </div>
        </div>

        <aside
          v-if="wordPopoverVisible"
          ref="popoverRef"
          class="dictionary-popover"
          :style="{
            top: `${popoverTop}px`,
            left: `${popoverLeft}px`,
            '--arrow-top': `${popoverArrowTop}px`,
          }"
        >
          <div v-if="wordLoading">{{ t('common.loading') }}</div>
          <template v-else-if="wordPopoverError">
            <div class="dictionary-header">
              <div class="dictionary-word dictionary-word--compact">{{ popoverLemmaDisplay }}</div>
            </div>
            <p class="dictionary-popover-message">{{ wordPopoverError }}</p>
          </template>
          <template v-else-if="wordModalData && hasDictionaryBody">
            <div class="dictionary-header">
              <div class="dictionary-word">{{ wordModalData.lemma }}</div>
              <button type="button" class="dictionary-audio" aria-label="Озвучить слово" @click="playWordAudio">
                <Icon name="play" />
              </button>
            </div>
            <div v-if="dictionaryMeta" class="dictionary-meta">{{ dictionaryMeta }}</div>
            <div class="dictionary-meanings">
              <div v-for="(meaning, idx) in dictionaryMeanings" :key="`${meaning}-${idx}`" class="meaning-row">
                <span class="meaning-index">{{ idx + 1 }}.</span>
                <span class="meaning-text">{{ meaning }}</span>
              </div>
            </div>
            <div class="dictionary-footer">
              <span>Добавить в слова</span>
              <button type="button" class="add-word-button" disabled>
                <Icon name="plus" />
              </button>
            </div>
          </template>
          <template v-else-if="wordModalData">
            <div class="dictionary-header">
              <div class="dictionary-word">{{ wordModalData.lemma || popoverLemmaDisplay }}</div>
              <button type="button" class="dictionary-audio" aria-label="Озвучить слово" @click="playWordAudio">
                <Icon name="play" />
              </button>
            </div>
            <p class="dictionary-popover-message">{{ t('reading.wordNoDefinition') }}</p>
          </template>
        </aside>

        <footer class="footer">
          <button
            type="button"
            class="mark-read-button"
            :disabled="isRead || markingRead"
            @click="markRead"
          >
            <span>✓</span>
            <span>{{ isRead ? t('reading.alreadyRead') : t('reading.markRead') }}</span>
          </button>
        </footer>
      </section>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { apiClient } from '../api/client'
import { useAudio } from '../composables/useAudio'
import { useSettings } from '../composables/useSettings'
import Icon from './Icon.vue'

const props = defineProps<{
  block: any
  chapterId?: string
  textId?: string
  isRead: boolean
}>()

const emit = defineEmits<{
  (e: 'marked-read'): void
}>()

const { t } = useI18n()
const router = useRouter()
const { playWordPronunciation } = useAudio()
const { settings } = useSettings()

const showTranslation = ref(false)
const wordPopoverVisible = ref(false)
const wordLoading = ref(false)
const wordModalData = ref<any>(null)
const wordPopoverError = ref('')
const markingRead = ref(false)
const selectedTokenKey = ref('')
const selectedLemma = ref('')
const popoverTop = ref(0)
const popoverLeft = ref(0)
const popoverArrowTop = ref(24)
const popoverRef = ref<HTMLElement | null>(null)
const activeSegmentId = ref<string | null>(null)
const isAutoplaying = ref(false)
let currentAudio: HTMLAudioElement | null = null
let autoplayRun = 0

const segments = computed(() => props.block?.reading_passage?.segments || [])

const noSpaceBefore = new Set(['.', ',', '!', '?', ';', ':', ')', ']', '}', '%', '…'])
const noSpaceAfter = new Set(['(', '[', '{', '¿', '¡', '«'])

const tokenText = (tokens: any[], index: number) => {
  const current = String(tokens?.[index]?.surface ?? '')
  const next = String(tokens?.[index + 1]?.surface ?? '')
  if (!current) return ''
  if (!next) return current
  if (noSpaceAfter.has(current) || noSpaceBefore.has(next)) return current
  return `${current} `
}

const waitMs = (ms: number) => new Promise<void>((resolve) => setTimeout(resolve, ms))

const goBack = () => {
  router.back()
}

const tokenKey = (segmentId: string, tokenIdx: number) => `${segmentId}-${tokenIdx}`

const isNarrator = (segment: any) => String(segment?.speaker_id || '').toLowerCase() === 'narrator'

const speakerClass = (speakerId: string) => {
  const id = String(speakerId || '').toLowerCase()
  if (id === 'speaker_b') return 'yellow'
  return 'blue'
}

const posMap: Record<string, string> = {
  noun: 'существительное',
  verb: 'глагол',
  adjective: 'прилагательное',
  adverb: 'наречие',
  pronoun: 'местоимение',
  preposition: 'предлог',
}

const genderMap: Record<string, string> = {
  masculine: 'мужской род',
  feminine: 'женский род',
  neuter: 'средний род',
}

const dictionaryMeanings = computed(() => {
  const cards = Array.isArray(wordModalData.value?.cards) ? wordModalData.value.cards : []
  const out: string[] = []
  const keys = ['meaning_target', 'meaning_en', 'word_native', 'word_ru'] as const
  for (const card of cards) {
    for (const k of keys) {
      const value = String(card?.[k] || '').trim()
      if (value && !out.includes(value)) out.push(value)
      if (out.length >= 4) break
    }
    if (out.length >= 4) break
  }
  return out
})

const dictionaryMeta = computed(() => {
  const first = (wordModalData.value?.cards || [])[0] || {}
  const pos = posMap[String(first?.pos || '').toLowerCase()] || String(first?.pos || '').trim()
  const gender = genderMap[String(first?.noun_gender || '').toLowerCase()] || ''
  return [pos, gender].filter(Boolean).join(', ')
})

const hasDictionaryBody = computed(() => dictionaryMeanings.value.length > 0 || !!dictionaryMeta.value)

const popoverLemmaDisplay = computed(() => String(selectedLemma.value || '').trim() || '—')

const stopCurrentAudio = () => {
  if (!currentAudio) return
  currentAudio.pause()
  currentAudio.currentTime = 0
  currentAudio = null
}

const playSegmentAudio = async (audioRelPath: string) => {
  const url = `/api/learning/reading/audio?path=${encodeURIComponent(audioRelPath)}`
  stopCurrentAudio()
  const audio = new Audio(url)
  currentAudio = audio
  await audio.play().catch((error) => {
    console.error('Failed to play segment audio', error)
  })
  await new Promise<void>((resolve) => {
    const finish = () => resolve()
    audio.addEventListener('ended', finish, { once: true })
    audio.addEventListener('error', finish, { once: true })
  })
}

const playSingleSegment = async (segment: any) => {
  if (!segment?.audio_rel_path) return
  isAutoplaying.value = false
  autoplayRun += 1
  activeSegmentId.value = segment.segment_id
  await playSegmentAudio(segment.audio_rel_path)
  if (!isAutoplaying.value) {
    activeSegmentId.value = null
  }
}

const toggleAutoplay = async () => {
  if (isAutoplaying.value) {
    isAutoplaying.value = false
    autoplayRun += 1
    activeSegmentId.value = null
    stopCurrentAudio()
    return
  }

  const runId = ++autoplayRun
  isAutoplaying.value = true
  for (const segment of segments.value) {
    if (!isAutoplaying.value || runId !== autoplayRun) break
    if (!segment?.audio_rel_path) continue
    activeSegmentId.value = segment.segment_id
    await playSegmentAudio(segment.audio_rel_path)
    if (!isAutoplaying.value || runId !== autoplayRun) break
    await waitMs(500)
  }
  if (runId === autoplayRun) {
    isAutoplaying.value = false
    activeSegmentId.value = null
  }
}

const playWordAudio = async () => {
  if (!selectedLemma.value) return
  await playWordPronunciation(selectedLemma.value)
}

const updatePopoverPosition = (target: HTMLElement) => {
  const rect = target.getBoundingClientRect()
  const width = 370
  const maxLeft = window.innerWidth - width - 12
  const top = Math.max(12, rect.top - 18)
  const left = Math.max(12, Math.min(maxLeft, rect.right + 16))
  popoverTop.value = top
  popoverLeft.value = left
  popoverArrowTop.value = Math.max(18, Math.min(56, rect.top - top + rect.height / 2))
}

const closePopover = () => {
  wordPopoverVisible.value = false
  selectedTokenKey.value = ''
  wordPopoverError.value = ''
  wordModalData.value = null
}

const onDocumentMouseDown = (e: MouseEvent) => {
  const target = e.target as Node
  if (popoverRef.value?.contains(target)) return
  const el = target as HTMLElement | null
  if (el?.closest('.token.clickable')) return
  closePopover()
}

const onTokenClick = async (event: MouseEvent, token: any, segment: any) => {
  if (!token?.clickable || !token?.lemma) return
  const target = event.currentTarget as HTMLElement | null
  if (!target) return

  selectedLemma.value = token.lemma
  selectedTokenKey.value = tokenKey(segment.segment_id, token.token_idx)
  updatePopoverPosition(target)
  wordPopoverVisible.value = true
  wordLoading.value = true
  wordPopoverError.value = ''
  wordModalData.value = null

  try {
    const data = await apiClient.request(`/api/reading/word-lookup?lemma=${encodeURIComponent(token.lemma)}`)
    wordModalData.value = data
  } catch (error: any) {
    console.error('Word lookup failed', error)
    wordModalData.value = null
    const status = typeof error?.status === 'number' ? error.status : 0
    if (error?.isNetworkError) {
      wordPopoverError.value = t('reading.wordLookupNetwork')
    } else if (status === 404) {
      wordPopoverError.value = t('reading.wordNotFound')
    } else if (status >= 500) {
      wordPopoverError.value = t('reading.wordLookupServerError')
    } else {
      wordPopoverError.value = t('reading.wordLookupFailed')
    }
  } finally {
    wordLoading.value = false
  }
}

const loadAutoplayPreference = async () => {
  let enabled = !!settings.value.autoplayPronunciation
  try {
    const data = await apiClient.request<{ settings?: { autoplay_pronunciation?: boolean } }>('/api/settings')
    if (typeof data?.settings?.autoplay_pronunciation === 'boolean') {
      enabled = data.settings.autoplay_pronunciation
    }
  } catch (error) {
    console.error('Failed to load autoplay setting for reading mode:', error)
  }
  if (enabled && segments.value.length > 0) {
    toggleAutoplay()
  }
}

onMounted(() => {
  document.addEventListener('mousedown', onDocumentMouseDown)
  loadAutoplayPreference()
})

onBeforeUnmount(() => {
  document.removeEventListener('mousedown', onDocumentMouseDown)
  stopCurrentAudio()
})

const markRead = async () => {
  markingRead.value = true
  try {
    const resourceId = props.textId || props.chapterId
    if (!resourceId) throw new Error('text id is required')
    await apiClient.request(`/api/learning/reading/texts/${resourceId}/mark-read`, { method: 'POST' })
    emit('marked-read')
  } catch (error) {
    console.error('Failed to mark reading text as read', error)
  } finally {
    markingRead.value = false
  }
}
</script>

<style scoped>
.screen {
  min-height: 100vh;
  background: var(--bg-primary);
  color: var(--text-primary);
  font-family: Inter, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  display: flex;
  justify-content: center;
  padding: 16px;
}

.reader-shell {
  width: 100%;
  max-width: 960px;
  min-height: calc(100vh - 32px);
  background: var(--bg-secondary);
  border: 1px solid var(--border-primary);
  border-radius: 20px;
  overflow: hidden;
  box-shadow: 0 24px 80px var(--card-shadow);
}

.header {
  min-height: 84px;
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 0 20px;
  border-bottom: 1px solid var(--border-primary);
  background: var(--bg-secondary);
}

.back-button {
  width: 44px;
  height: 44px;
  display: grid;
  place-items: center;
  border: 0;
  background: transparent;
  color: var(--text-primary);
  cursor: pointer;
  font-size: 22px;
}

.title {
  flex: 1;
  font-size: 22px;
  line-height: 1.2;
  font-weight: 700;
  letter-spacing: -0.02em;
  margin: 0;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.icon-button {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  border: 1px solid var(--border-primary);
  background: var(--bg-tertiary);
  color: var(--text-primary);
  display: grid;
  place-items: center;
  cursor: pointer;
  font-size: 16px;
}

.icon-button.active {
  color: var(--color-primary);
  border-color: var(--color-primary);
}

.icon-button:hover {
  background: var(--bg-hover);
}

.content {
  padding: 22px 16px 24px;
  position: relative;
}

.text-flow {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.sentence-row {
  position: relative;
  display: grid;
  grid-template-columns: 30px 1fr 40px;
  column-gap: 8px;
  align-items: start;
  min-height: 48px;
  padding: 8px 8px;
  border-radius: 14px;
  cursor: pointer;
}

.sentence-row.narrator {
  grid-template-columns: 0 1fr 40px;
}

.sentence-row:hover {
  background: var(--bg-hover);
}

.sentence-row.active {
  background: linear-gradient(90deg, rgba(44, 116, 255, 0.2), rgba(44, 116, 255, 0.08));
  box-shadow: inset 0 0 0 1px rgba(84, 145, 255, 0.12);
}

.speaker-icon {
  width: 24px;
  height: 24px;
  margin-top: 7px;
  opacity: 0.95;
}

.speaker-icon.blue {
  color: #4f91ff;
}

.speaker-icon.yellow {
  color: #ffca3a;
}

.sentence-text {
  font-size: 21px;
  line-height: 1.28;
  font-weight: 500;
  color: var(--text-primary);
  letter-spacing: -0.01em;
}

.translation {
  margin-top: 6px;
  font-size: 14px;
  line-height: 1.35;
  font-weight: 400;
  color: var(--text-secondary);
}

.translation.hidden {
  display: none;
}

.token {
  white-space: pre-wrap;
}

.token.clickable {
  cursor: pointer;
}

.word-selected {
  background: #3b82f6;
  color: #ffffff;
  border-radius: 6px;
  padding: 1px 4px 3px;
}

.sentence-audio-button {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: 1px solid var(--border-primary);
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  display: grid;
  place-items: center;
  margin-top: 2px;
  cursor: pointer;
}

.sentence-row.active .sentence-audio-button {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.dictionary-popover {
  position: fixed;
  z-index: 30;
  width: min(370px, calc(100vw - 24px));
  padding: 20px 20px 16px;
  border-radius: 18px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-primary);
  box-shadow: 0 24px 70px var(--card-shadow);
}

.dictionary-popover::before {
  content: "";
  position: absolute;
  top: var(--arrow-top);
  left: -12px;
  width: 18px;
  height: 18px;
  background: var(--bg-secondary);
  border-left: 1px solid var(--border-primary);
  border-bottom: 1px solid var(--border-primary);
  transform: rotate(45deg);
}

.dictionary-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.dictionary-word {
  font-size: 24px;
  line-height: 1.15;
  font-weight: 700;
  color: var(--text-primary);
}

.dictionary-word--compact {
  font-size: 20px;
  font-weight: 600;
}

.dictionary-popover-message {
  margin: 0;
  font-size: 14px;
  line-height: 1.45;
  color: var(--text-secondary);
}

.dictionary-audio {
  width: 34px;
  height: 34px;
  display: grid;
  place-items: center;
  color: var(--text-secondary);
  background: transparent;
  border: 0;
  cursor: pointer;
}

.dictionary-meta {
  font-size: 14px;
  color: var(--text-secondary);
  margin-bottom: 12px;
}

.dictionary-meanings {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding-top: 12px;
  border-top: 1px solid var(--border-primary);
}

.meaning-row {
  display: grid;
  grid-template-columns: 20px 1fr;
  gap: 8px;
  font-size: 16px;
  line-height: 1.32;
}

.meaning-index {
  color: var(--text-secondary);
}

.meaning-text {
  color: var(--text-primary);
}

.dictionary-footer {
  margin-top: 16px;
  padding-top: 12px;
  border-top: 1px solid var(--border-primary);
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 10px;
  color: var(--text-secondary);
  font-size: 14px;
}

.add-word-button {
  width: 34px;
  height: 34px;
  border-radius: 50%;
  border: 1px solid var(--border-primary);
  background: var(--bg-tertiary);
  color: var(--text-primary);
  display: grid;
  place-items: center;
}

.footer {
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid var(--border-primary);
}

.mark-read-button {
  width: 100%;
  min-height: 58px;
  border-radius: 12px;
  border: 1px solid var(--border-primary);
  background: var(--bg-tertiary);
  color: var(--text-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
}

.mark-read-button:hover {
  background: var(--bg-hover);
}

.mark-read-button:disabled {
  opacity: 0.6;
  cursor: default;
}

@media (max-width: 768px) {
  .screen {
    padding: 8px;
  }

  .reader-shell {
    min-height: calc(100vh - 16px);
    border-radius: 16px;
  }

  .header {
    min-height: 74px;
    padding: 0 12px;
    gap: 10px;
  }

  .title {
    font-size: 18px;
  }

  .icon-button {
    width: 40px;
    height: 40px;
  }

  .content {
    padding: 14px 10px 20px;
  }

  .sentence-text {
    font-size: 19px;
  }

  .translation {
    font-size: 13px;
  }
}
</style>
