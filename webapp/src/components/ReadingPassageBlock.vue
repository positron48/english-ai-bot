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
            v-if="canDelete"
            type="button"
            class="icon-button icon-button-danger"
            :disabled="deleting"
            :aria-label="t('reading.deleteText')"
            :title="t('reading.deleteText')"
            @click.stop="emit('delete-request')"
          >
            <Icon name="trash" />
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
            <div
              v-if="!isNarrator(segment)"
              class="speaker-icon"
              :style="{ color: speakerIconColor(segment.speaker_id) }"
              :title="speakerLabel(segment.speaker_id)"
              role="img"
              :aria-label="speakerLabel(segment.speaker_id)"
            >
              <Icon name="user" />
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
          <button
            v-if="categoryId && otherUnreadInCategoryCount > 0"
            type="button"
            class="random-unread-footer-button"
            :disabled="randomUnreadNavigating"
            @click="openRandomUnreadInCategory"
          >
            <Icon name="dice" />
            <span>{{ t('reading.anotherRandomUnread') }}</span>
          </button>
        </footer>
      </section>
    </main>

    <!-- Полный экран как в «Словарь»: карточки, SRS, формы глаголов; слово в обучение — на сервере при word-lookup -->
    <div v-if="wordModalVisible" class="word-modal-overlay" @click.self="closeWordModal">
      <div class="word-modal-panel">
        <div v-if="wordLookupLoading" class="word-modal-loading">{{ t('common.loading') }}</div>
        <div v-else-if="wordLookupError" class="word-modal-error">
          <p class="word-modal-error-text">{{ wordLookupError }}</p>
          <button type="button" class="word-modal-close-btn" @click="closeWordModal">{{ t('common.close') }}</button>
        </div>
        <VocabWordCardsDetail
          v-else
          :lemma="modalLemma"
          :preloaded="modalPreloaded"
          @close="closeWordModal"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { apiClient } from '../api/client'
import { useSettings } from '../composables/useSettings'
import Icon from './Icon.vue'
import VocabWordCardsDetail, { type VocabCardsAPIResponse } from './VocabWordCardsDetail.vue'

const props = defineProps<{
  block: any
  chapterId?: string
  textId?: string
  categoryId?: string
  isRead: boolean
  canDelete?: boolean
  deleting?: boolean
}>()

const emit = defineEmits<{
  (e: 'marked-read'): void
  (e: 'delete-request'): void
}>()

const { t } = useI18n()
const router = useRouter()
const { settings } = useSettings()

const showTranslation = ref(false)
const wordModalVisible = ref(false)
const wordLookupLoading = ref(false)
const wordLookupError = ref('')
const modalLemma = ref('')
const modalPreloaded = ref<VocabCardsAPIResponse | null>(null)
const markingRead = ref(false)
const otherUnreadInCategoryCount = ref(0)
const randomUnreadNavigating = ref(false)

const segments = computed(() => props.block?.reading_passage?.segments || [])
const activeSegmentId = ref<string | null>(null)
const selectedTokenKey = ref('')
const isAutoplaying = ref(false)
let currentAudio: HTMLAudioElement | null = null
let autoplayRun = 0

/** Distinct hues for dialogue speakers (first appearance order). */
const SPEAKER_ICON_PALETTE = [
  '#5b9cff',
  '#ffb020',
  '#3dd68c',
  '#f472b6',
  '#a78bfa',
  '#fb923c',
  '#38bdf8',
  '#f87171',
  '#4ade80',
  '#e879f9',
] as const

const speakerColorById = computed(() => {
  const map: Record<string, string> = {}
  let idx = 0
  for (const seg of segments.value) {
    const id = String(seg?.speaker_id || '').trim().toLowerCase()
    if (!id || id === 'narrator') continue
    if (map[id]) continue
    map[id] = SPEAKER_ICON_PALETTE[idx % SPEAKER_ICON_PALETTE.length]
    idx += 1
  }
  return map
})

function speakerIconColor(speakerId: string): string {
  const id = String(speakerId || '').trim().toLowerCase()
  return speakerColorById.value[id] ?? '#94a3b8'
}

function speakerLabel(id: unknown): string {
  const s = String(id ?? '').trim()
  return s || 'speaker'
}

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

const closeWordModal = () => {
  wordModalVisible.value = false
  wordLookupLoading.value = false
  wordLookupError.value = ''
  modalLemma.value = ''
  modalPreloaded.value = null
  selectedTokenKey.value = ''
}

const wordModalKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Escape' && wordModalVisible.value) {
    event.preventDefault()
    closeWordModal()
  }
}

watch(wordModalVisible, (open) => {
  if (open) {
    window.addEventListener('keydown', wordModalKeydown)
  } else {
    window.removeEventListener('keydown', wordModalKeydown)
  }
})

onUnmounted(() => {
  window.removeEventListener('keydown', wordModalKeydown)
})

const onTokenClick = async (event: MouseEvent, token: any, segment: any) => {
  if (!token?.clickable || !token?.lemma) return

  selectedTokenKey.value = tokenKey(segment.segment_id, token.token_idx)

  wordModalVisible.value = true
  wordLookupLoading.value = true
  wordLookupError.value = ''
  modalPreloaded.value = null
  modalLemma.value = token.lemma

  try {
    const data: VocabCardsAPIResponse = await apiClient.request(
      `/api/reading/word-lookup?lemma=${encodeURIComponent(token.lemma)}`,
    )
    modalLemma.value = data.lemma || token.lemma
    modalPreloaded.value = data
  } catch (error: any) {
    console.error('Word lookup failed', error)
    modalPreloaded.value = null
    const status = typeof error?.status === 'number' ? error.status : 0
    if (error?.isNetworkError) {
      wordLookupError.value = t('reading.wordLookupNetwork')
    } else if (status === 404) {
      wordLookupError.value = t('reading.wordNotFound')
    } else if (status >= 500) {
      wordLookupError.value = t('reading.wordLookupServerError')
    } else {
      wordLookupError.value = t('reading.wordLookupFailed')
    }
  } finally {
    wordLookupLoading.value = false
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
  loadAutoplayPreference()
})

onBeforeUnmount(() => {
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

const currentTextId = computed(() => String(props.textId || props.chapterId || '').trim())

async function refreshOtherUnreadInCategory() {
  const cat = String(props.categoryId || '').trim()
  const tid = currentTextId.value
  otherUnreadInCategoryCount.value = 0
  if (!cat || !tid) return
  try {
    const data: { texts?: { text_id: string; is_read: boolean }[] } = await apiClient.request(
      `/api/learning/reading/categories/${encodeURIComponent(cat)}/texts`
    )
    const texts = data.texts || []
    otherUnreadInCategoryCount.value = texts.filter((x) => !x.is_read && x.text_id !== tid).length
  } catch {
    otherUnreadInCategoryCount.value = 0
  }
}

watch(
  () => [props.categoryId, props.textId, props.chapterId, props.isRead] as const,
  () => {
    void refreshOtherUnreadInCategory()
  },
  { immediate: true }
)

const openRandomUnreadInCategory = async () => {
  const cat = String(props.categoryId || '').trim()
  const tid = currentTextId.value
  if (!cat || !tid || randomUnreadNavigating.value) return
  randomUnreadNavigating.value = true
  try {
    const data: { texts?: { text_id: string; is_read: boolean }[] } = await apiClient.request(
      `/api/learning/reading/categories/${encodeURIComponent(cat)}/texts`
    )
    const pool = (data.texts || []).filter((x) => !x.is_read && x.text_id !== tid)
    if (!pool.length) {
      otherUnreadInCategoryCount.value = 0
      return
    }
    const pick = pool[Math.floor(Math.random() * pool.length)]
    await router.push(`/learning/reading/text/${pick.text_id}`)
  } finally {
    randomUnreadNavigating.value = false
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

.icon-button-danger {
  color: #ef4444;
  border-color: color-mix(in srgb, #ef4444 40%, var(--border-primary));
}

.icon-button-danger:hover:not(:disabled) {
  background: color-mix(in srgb, #ef4444 16%, var(--bg-hover));
}

.icon-button:disabled {
  opacity: 0.55;
  cursor: default;
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
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  font-size: 22px;
  opacity: 1;
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

.word-modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  background: var(--bg-modal-overlay, rgba(0, 0, 0, 0.5));
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
}

.word-modal-panel {
  background: var(--card-bg);
  border-radius: 8px;
  max-width: 800px;
  width: 100%;
  max-height: min(90vh, 900px);
  overflow-y: auto;
  padding: 24px 28px;
  color: var(--text-primary);
  border: 1px solid var(--border-primary);
  box-shadow: 0 24px 64px var(--card-shadow);
}

.word-modal-loading {
  text-align: center;
  padding: 48px 16px;
  font-size: 16px;
  color: var(--text-secondary);
}

.word-modal-error {
  padding: 24px 8px;
  text-align: center;
}

.word-modal-error-text {
  margin: 0 0 20px;
  font-size: 15px;
  line-height: 1.45;
  color: var(--text-secondary);
}

.word-modal-close-btn {
  padding: 10px 20px;
  border-radius: 8px;
  border: 1px solid var(--border-primary);
  background: var(--bg-tertiary);
  color: var(--text-primary);
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
}

.word-modal-close-btn:hover {
  background: var(--bg-hover);
}

.footer {
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid var(--border-primary);
  display: flex;
  flex-direction: column;
  gap: 12px;
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

.random-unread-footer-button {
  width: 100%;
  min-height: 58px;
  border-radius: 12px;
  border: 1px dashed var(--border-primary);
  background: var(--bg-secondary);
  color: var(--text-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
}

.random-unread-footer-button:hover:not(:disabled) {
  background: var(--bg-hover);
}

.random-unread-footer-button:disabled {
  opacity: 0.6;
  cursor: default;
}

.random-unread-footer-button :deep(.icon) {
  font-size: 22px;
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
