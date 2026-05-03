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
                    <div
                      v-for="c in cardsByWordId[row.word_card_id] || []"
                      :key="c.id"
                      class="verb-card-block"
                    >
                      <div class="verb-card-block__head">
                        <span class="verb-card-block__id">{{ t('adminVerbTraining.cardId') }}: {{ c.id }}</span>
                        <span class="verb-card-block__slice">{{ formatSlice(c) }}</span>
                        <span v-if="c.surface_form" class="verb-card-block__surface">{{
                          c.surface_form
                        }}</span>
                      </div>
                      <div class="verb-card-block__grid">
                        <div v-if="c.prompt != null" class="verb-card-block__col">
                          <div class="verb-card-block__label">{{ t('adminVerbTraining.promptJson') }}</div>
                          <pre class="json-block">{{ pretty(c.prompt) }}</pre>
                        </div>
                        <div v-if="c.answer != null" class="verb-card-block__col">
                          <div class="verb-card-block__label">{{ t('adminVerbTraining.answerJson') }}</div>
                          <pre class="json-block">{{ pretty(c.answer) }}</pre>
                        </div>
                        <div
                          v-if="c.distractors != null && hasContent(c.distractors)"
                          class="verb-card-block__col"
                        >
                          <div class="verb-card-block__label">{{ t('adminVerbTraining.distractorsJson') }}</div>
                          <pre class="json-block">{{ pretty(c.distractors) }}</pre>
                        </div>
                      </div>
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

function pretty(v: unknown): string {
  try {
    return JSON.stringify(v, null, 2)
  } catch {
    return String(v)
  }
}

function hasContent(v: unknown): boolean {
  if (v === null || v === undefined) return false
  if (typeof v === 'string') return v.trim() !== '' && v !== '{}'
  if (Array.isArray(v)) return v.length > 0
  if (typeof v === 'object') return Object.keys(v as object).length > 0
  return true
}

function formatSlice(c: VerbTrainingCardAPI): string {
  return `${c.mood} · ${c.tense} · ${c.person}/${c.number}`
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

.verb-card-block {
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 12px;
  background: var(--bg-primary);
}

.verb-card-block__head {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 16px;
  align-items: baseline;
  margin-bottom: 8px;
  font-size: 13px;
}

.verb-card-block__id {
  color: var(--text-secondary);
}

.verb-card-block__slice {
  font-weight: 600;
}

.verb-card-block__surface {
  font-family: ui-monospace, monospace;
}

.verb-card-block__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 12px;
}

.verb-card-block__label {
  font-size: 12px;
  color: var(--text-secondary);
  margin-bottom: 4px;
}

.json-block {
  margin: 0;
  padding: 8px;
  border-radius: 6px;
  background: var(--bg-secondary, #f0f0f0);
  font-size: 11px;
  line-height: 1.4;
  max-height: 220px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-word;
}
</style>
