<template>
  <div class="admin-content admin-verb-training">
    <div class="card">
      <h1>{{ t('adminVerbTraining.title') }}</h1>
      <p class="admin-verb-training__intro">{{ t('adminVerbTraining.subtitle') }}</p>
      <p v-if="expectedCloze !== null" class="admin-verb-training__meta">
        {{ t('adminVerbTraining.expectedPack', { n: expectedCloze }) }}
      </p>

      <div class="admin-verb-training__toolbar">
        <input
          v-model="searchQuery"
          type="search"
          class="search-input"
          :placeholder="t('adminVerbTraining.searchPlaceholder')"
          @keyup.enter="reload"
        />
        <button type="button" class="btn btn-secondary" :disabled="loading" @click="reload">
          {{ t('adminVerbTraining.refresh') }}
        </button>
      </div>

      <div v-if="error" class="admin-verb-training__error">{{ error }}</div>
      <div v-else-if="loading && lemmas.length === 0" class="loading">{{ t('common.loading') }}</div>

      <div v-else class="table-wrap">
        <table class="admin-verb-training__table">
          <thead>
            <tr>
              <th>{{ t('adminVerbTraining.lemma') }}</th>
              <th>{{ t('adminVerbTraining.ruGloss') }}</th>
              <th>{{ t('adminVerbTraining.clozeCount') }}</th>
              <th>{{ t('adminVerbTraining.fullCoverage') }}</th>
              <th />
            </tr>
          </thead>
          <tbody>
            <template v-for="row in lemmas" :key="row.word_card_id">
              <tr>
                <td>
                  <strong>{{ row.lemma }}</strong>
                  <span class="admin-verb-training__wc-id">#{{ row.word_card_id }}</span>
                </td>
                <td>{{ row.ru_gloss || '—' }}</td>
                <td>{{ row.cloze_count }}</td>
                <td>
                  <span :class="row.full_coverage ? 'badge badge-ok' : 'badge badge-warn'">
                    {{ row.full_coverage ? t('adminVerbTraining.yes') : t('adminVerbTraining.no') }}
                  </span>
                </td>
                <td class="admin-verb-training__actions">
                  <button
                    type="button"
                    class="btn btn-sm btn-secondary"
                    @click="toggleCards(row.word_card_id)"
                  >
                    {{
                      expandedWordCardId === row.word_card_id
                        ? t('adminVerbTraining.hideCards')
                        : t('adminVerbTraining.showCards')
                    }}
                  </button>
                </td>
              </tr>
              <tr v-if="expandedWordCardId === row.word_card_id" class="admin-verb-training__cards-row">
                <td colspan="5">
                  <div v-if="cardsLoadingId === row.word_card_id" class="loading-inline">
                    {{ t('adminVerbTraining.loadingCards') }}
                  </div>
                  <div v-else-if="(cardsByWordId[row.word_card_id] || []).length === 0" class="muted">
                    {{ t('adminVerbTraining.noCards') }}
                  </div>
                  <div v-else class="verb-cards-list">
                    <h4 class="verb-cards-list__title">{{ t('adminVerbTraining.cardsHeading') }}</h4>
                    <div class="verb-cards-table-wrap">
                      <table class="verb-cards-detail-table">
                        <thead>
                          <tr>
                            <th class="verb-cards-detail-table__th-num">{{ t('adminVerbTraining.colId') }}</th>
                            <th>{{ t('adminVerbTraining.colTense') }}</th>
                            <th>{{ t('adminVerbTraining.colPronoun') }}</th>
                            <th>{{ t('adminVerbTraining.colForm') }}</th>
                            <th>{{ t('adminVerbTraining.colSentence') }}</th>
                            <th>{{ t('adminVerbTraining.colRu') }}</th>
                            <th>{{ t('adminVerbTraining.colOptions') }}</th>
                          </tr>
                        </thead>
                        <tbody>
                          <tr
                            v-for="c in cardsByWordId[row.word_card_id] || []"
                            :key="c.id"
                            class="verb-cards-detail-table__row"
                          >
                            <td class="verb-cards-detail-table__id">{{ c.id }}</td>
                            <td class="verb-cards-detail-table__nowrap">{{ formatTenseMood(c) }}</td>
                            <td>{{ subjectPronoun(c.person, c.number) }}</td>
                            <td class="verb-cards-detail-table__form">{{ c.surface_form || '—' }}</td>
                            <td class="verb-cards-detail-table__text">{{ promptQuestion(c) }}</td>
                            <td class="verb-cards-detail-table__text">{{ promptTranslationRu(c) }}</td>
                            <td class="verb-cards-detail-table__options">
                              <div
                                v-for="(opt, oi) in parseDistractorOptions(c)"
                                :key="`${c.id}-opt-${oi}`"
                                class="verb-cards-detail-table__option-line"
                              >
                                {{ opt }}
                              </div>
                              <span
                                v-if="parseDistractorOptions(c).length === 0"
                                class="muted"
                              >—</span>
                            </td>
                          </tr>
                        </tbody>
                      </table>
                    </div>
                  </div>
                </td>
              </tr>
            </template>
          </tbody>
        </table>
        <p v-if="!loading && lemmas.length === 0" class="muted">{{ t('adminVerbTraining.noLemmas') }}</p>
      </div>

      <div v-if="hasMore && lemmas.length > 0" class="admin-verb-training__more">
        <button type="button" class="btn btn-primary" :disabled="loading" @click="loadMore">
          {{ loading ? t('common.loading') : t('adminVerbTraining.loadMore') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { apiClient } from '../api/client'

const { t } = useI18n()

interface LemmaRow {
  word_card_id: number
  lemma: string
  cloze_count: number
  ru_gloss?: string
  full_coverage: boolean
}

interface VerbTrainingCardAPI {
  id: number
  card_type: string
  mood: string
  tense: string
  person: string
  number: string
  surface_form: string
  prompt?: unknown
  answer?: unknown
  distractors?: unknown
}

interface VerbPromptFields {
  question?: string
  example_translation?: string
}

const searchQuery = ref('')
const lemmas = ref<LemmaRow[]>([])
const nextCursor = ref(0)
const hasMore = ref(true)
const loading = ref(false)
const error = ref<string | null>(null)
const expectedCloze = ref<number | null>(null)

const expandedWordCardId = ref<number | null>(null)
const cardsByWordId = ref<Record<number, VerbTrainingCardAPI[]>>({})
const cardsLoadingId = ref<number | null>(null)

let searchDebounce: ReturnType<typeof setTimeout> | null = null

function promptFields(c: VerbTrainingCardAPI): VerbPromptFields {
  const p = c.prompt
  if (p && typeof p === 'object' && !Array.isArray(p)) {
    return p as VerbPromptFields
  }
  return {}
}

function promptQuestion(c: VerbTrainingCardAPI): string {
  const q = promptFields(c).question
  return typeof q === 'string' && q.trim() !== '' ? q.trim() : '—'
}

function promptTranslationRu(c: VerbTrainingCardAPI): string {
  const tr = promptFields(c).example_translation
  return typeof tr === 'string' && tr.trim() !== '' ? tr.trim() : '—'
}

function formatTenseMood(c: VerbTrainingCardAPI): string {
  const mood = String(c.mood || '').trim()
  const tense = String(c.tense || '').trim()
  if (!tense && !mood) return '—'
  if (!mood) return tense
  if (!tense) return mood
  return `${tense} · ${mood}`
}

function subjectPronoun(person: string, number: string): string {
  const p = String(person || '').trim()
  const n = String(number || '').trim().toLowerCase()
  if (p === '1' && n === 'singular') return 'yo'
  if (p === '2' && n === 'singular') return 'tú'
  if (p === '3' && n === 'singular') return 'él/ella/usted'
  if (p === '1' && n === 'plural') return 'nosotros'
  if (p === '2' && n === 'plural') return 'vosotros/ustedes'
  if (p === '3' && n === 'plural') return 'ellos/ellas/ustedes'
  return `${person}/${number}`
}

function parseDistractorOptions(c: VerbTrainingCardAPI): string[] {
  const d = c.distractors
  if (d == null) return []
  if (Array.isArray(d)) {
    return d.map((x) => String(x).trim()).filter(Boolean)
  }
  if (typeof d === 'string') {
    const s = d.trim()
    if (!s || s === 'null') return []
    try {
      const parsed = JSON.parse(s) as unknown
      if (Array.isArray(parsed)) {
        return parsed.map((x) => String(x).trim()).filter(Boolean)
      }
    } catch {
      /* raw string */
    }
    return s ? [s] : []
  }
  return []
}

async function fetchPage(append: boolean) {
  if (loading.value) return
  loading.value = true
  error.value = null
  try {
    const params = new URLSearchParams()
    const q = searchQuery.value.trim()
    if (q) params.set('q', q)
    params.set('limit', '50')
    if (append && nextCursor.value > 0) {
      params.set('cursor', String(nextCursor.value))
    }
    const data = (await apiClient.request(`/api/admin/verb-training/lemmas?${params.toString()}`)) as {
      lemmas: LemmaRow[]
      next_cursor: number
      expected_cloze_per_lemma: number
    }
    expectedCloze.value = data.expected_cloze_per_lemma ?? null
    const batch = data.lemmas || []
    if (!append) {
      lemmas.value = batch
      expandedWordCardId.value = null
      cardsByWordId.value = {}
    } else {
      lemmas.value = [...lemmas.value, ...batch]
    }
    nextCursor.value = data.next_cursor || 0
    hasMore.value = batch.length >= 50
  } catch (e: unknown) {
    console.error(e)
    error.value = t('adminVerbTraining.loadFailed')
    if (!append) lemmas.value = []
  } finally {
    loading.value = false
  }
}

function reload() {
  nextCursor.value = 0
  hasMore.value = true
  void fetchPage(false)
}

function loadMore() {
  void fetchPage(true)
}

async function toggleCards(wordCardId: number) {
  if (expandedWordCardId.value === wordCardId) {
    expandedWordCardId.value = null
    return
  }
  expandedWordCardId.value = wordCardId
  if (cardsByWordId.value[wordCardId]?.length) return
  cardsLoadingId.value = wordCardId
  try {
    const data = (await apiClient.request(
      `/api/admin/verb-training/cards?word_card_id=${wordCardId}`
    )) as { cards: VerbTrainingCardAPI[] }
    cardsByWordId.value = { ...cardsByWordId.value, [wordCardId]: data.cards || [] }
  } catch (e) {
    console.error(e)
    expandedWordCardId.value = null
    error.value = t('adminVerbTraining.loadFailed')
  } finally {
    cardsLoadingId.value = null
  }
}

watch(searchQuery, () => {
  if (searchDebounce) clearTimeout(searchDebounce)
  searchDebounce = setTimeout(() => {
    reload()
  }, 350)
})

onMounted(() => {
  void fetchPage(false)
})
</script>

<style scoped>
.admin-verb-training__intro,
.admin-verb-training__meta {
  color: var(--text-secondary, #666);
  margin: 0 0 12px;
  font-size: 14px;
}

.admin-verb-training__toolbar {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  margin-bottom: 16px;
  align-items: center;
}

.admin-verb-training__toolbar .search-input {
  flex: 1;
  min-width: 200px;
  padding: 8px 12px;
  border-radius: 8px;
  border: 1px solid var(--border-primary);
  background: var(--bg-primary);
  color: var(--text-primary);
}

.admin-verb-training__error {
  color: var(--color-danger, #c62828);
  margin-bottom: 12px;
}

.table-wrap {
  overflow-x: auto;
}

.admin-verb-training__table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}

.admin-verb-training__table th,
.admin-verb-training__table td {
  padding: 10px 12px;
  border-bottom: 1px solid var(--border-primary);
  text-align: left;
  vertical-align: top;
}

.admin-verb-training__wc-id {
  margin-left: 8px;
  font-size: 12px;
  color: var(--text-secondary);
  font-weight: normal;
}

.admin-verb-training__actions {
  white-space: nowrap;
}

.badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 6px;
  font-size: 12px;
}

.badge-ok {
  background: rgba(46, 125, 50, 0.15);
  color: #2e7d32;
}

.badge-warn {
  background: rgba(245, 124, 0, 0.15);
  color: #ef6c00;
}

.admin-verb-training__cards-row td {
  background: var(--bg-secondary, #f5f5f5);
}

.admin-verb-training__more {
  margin-top: 16px;
  display: flex;
  justify-content: center;
}

.loading-inline {
  padding: 8px 0;
}

.muted {
  color: var(--text-secondary);
}

.verb-cards-list__title {
  margin: 0 0 12px;
  font-size: 16px;
}

.verb-cards-table-wrap {
  overflow-x: auto;
  max-width: 100%;
}

.verb-cards-detail-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
  background: var(--bg-primary);
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  overflow: hidden;
}

.verb-cards-detail-table th,
.verb-cards-detail-table td {
  padding: 10px 12px;
  border-bottom: 1px solid var(--border-primary);
  vertical-align: top;
  text-align: left;
}

.verb-cards-detail-table thead th {
  background: var(--bg-secondary, #eee);
  font-weight: 600;
  font-size: 12px;
  white-space: nowrap;
}

.verb-cards-detail-table__th-num {
  width: 56px;
}

.verb-cards-detail-table__row:last-child td {
  border-bottom: none;
}

.verb-cards-detail-table__id {
  color: var(--text-secondary);
  font-size: 12px;
  white-space: nowrap;
}

.verb-cards-detail-table__nowrap {
  white-space: nowrap;
}

.verb-cards-detail-table__form {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
}

.verb-cards-detail-table__text {
  max-width: 280px;
  line-height: 1.45;
  word-break: break-word;
}

.verb-cards-detail-table__options {
  min-width: 140px;
  max-width: 220px;
}

.verb-cards-detail-table__option-line {
  padding: 4px 0;
  border-bottom: 1px dashed var(--border-primary);
  line-height: 1.35;
  word-break: break-word;
}

.verb-cards-detail-table__option-line:last-child {
  border-bottom: none;
  padding-bottom: 0;
}
</style>
