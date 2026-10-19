<template>
  <div class="reading-text-view">
    <div v-if="daily.isJourney.value && daily.error.value" role="alert"><p>{{ daily.error.value }}</p><button @click="daily.retry">{{ t('common.retry') }}</button><button @click="daily.back">{{ t('dailyJourney.returnToRoute') }}</button></div>
    <LgLoader v-else-if="daily.isJourney.value ? !daily.data.value && !daily.error.value : loading" />
    <div v-else-if="error">{{ error }}</div>
    <div v-else-if="daily.isJourney.value && daily.finished.value"><button @click="daily.back">{{ t('dailyJourney.returnToRoute') }}</button></div>
    <div v-else-if="!displayBlock">{{ t('reading.noTexts') }}</div>
    <ReadingPassageBlock
      v-else
      :block="displayBlock"
      :journey-quiz="daily.isJourney.value ? daily.data.value : undefined"
      :journey-busy="daily.busy.value || daily.hasPendingAction.value"
      :journey-course-code="daily.context.value?.courseCode"
      @journey-answer="answerJourney"
      @journey-open-quiz="daily.act('start', { open_quiz: true })"
      @journey-next="daily.act('ack_feedback')"
      @journey-back="backFromJourney"
      :text-id="textId"
      :category-id="daily.data.value?.document.category_id || categoryId"
      :is-read="readingIsRead"
      :cover-hero-rel-path="daily.data.value?.document.cover_hero_rel_path || coverHeroRelPath"
      @marked-read="readingIsRead = true"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { emitAppDataEvent } from '../api/cacheInvalidation'
import { apiClient } from '../api/client'
import { withCourseCode } from '../api/grammarClient'
import { useDailyJourneyExercise } from '../composables/useDailyJourneyExercise'
import type { ReadingJourneyExercise } from '../api/dailyJourneyClient'
import { useCourse } from '../composables/useCourse'
import ReadingPassageBlock from '../components/ReadingPassageBlock.vue'
import LgLoader from '../components/linglow/LgLoader.vue'

const route = useRoute()
const daily = useDailyJourneyExercise<ReadingJourneyExercise>('reading')
const { t } = useI18n()
const { ensureCourseLoaded } = useCourse()

const textId = computed(() => route.params.textId as string)
const loading = ref(true)
const error = ref<string | null>(null)
const block = ref<any>(null)
const coverHeroRelPath = ref('')
const readingIsRead = ref(false)
const categoryId = ref('')

const displayBlock = computed(() => daily.isJourney.value ? daily.data.value ? { title: daily.data.value.document.title, reading_passage: daily.data.value.document.reading_passage } : null : block.value)
async function backFromJourney() {
  if (daily.data.value?.done && daily.data.value.feedback) await daily.act('ack_feedback')
  if (!daily.error.value) await daily.back()
}
async function answerJourney(answer: string) {
  if (!daily.data.value) return
  await daily.act('answer', { index: daily.data.value.index, answer })
  if (!daily.error.value && daily.data.value?.done && daily.data.value.passed && daily.context.value) emitAppDataEvent('reading-marked-read', daily.context.value.courseCode)
}
watch(daily.finished, finished => { if (finished) void daily.back() })
const loadText = async () => {
  if (daily.isJourney.value) { loading.value = false; return }

  loading.value = true
  error.value = null
  block.value = null
  coverHeroRelPath.value = ''
  readingIsRead.value = false
  categoryId.value = ''
  try {
    await ensureCourseLoaded()
    const data: {
      block?: any
      category_id?: string
      cover_hero_rel_path?: string
      reading_progress?: { is_read?: boolean }
    } = await apiClient.request(withCourseCode(`/api/learning/reading/texts/${textId.value}`))
    block.value = data.block ?? null
    coverHeroRelPath.value = String(data.cover_hero_rel_path || '').trim()
    readingIsRead.value = !!data.reading_progress?.is_read
    categoryId.value = (data.category_id && String(data.category_id)) || ''
  } catch (e: any) {
    error.value = e?.message || 'Failed to load reading text'
  } finally {
    loading.value = false
  }
}

watch([textId, daily.isJourney], () => {
  void loadText()
}, { immediate: true })
</script>

<style scoped>
.reading-text-view { margin: 0; padding: 0; }
</style>
