<template>
  <div class="grammar-chapter">
    <div v-if="loading" class="loading">
      <p>{{ t('grammar.loadingChapter') }}</p>
    </div>
    
    <div v-else-if="error" class="error">
      <p>{{ error }}</p>
      <button @click="loadChapter" class="btn btn-primary">{{ t('common.retry') }}</button>
    </div>
    
    <div v-else-if="chapter">
      <div class="chapter-header">
        <h1>{{ chapterTitle }}</h1>
        <div class="chapter-meta">
          <span v-if="chapter.level" class="meta-badge">{{ chapter.level }}</span>
          <span v-if="chapter.estimated_minutes" class="meta-badge">
            ~{{ chapter.estimated_minutes }} {{ t('grammar.min') }}
          </span>
        </div>
      </div>
      
      <div v-if="chapter.description" class="chapter-description">
        {{ chapter.description }}
      </div>
      
      <div class="chapter-content">
        <div
          v-for="block in chapter.blocks"
          :key="block.id"
          class="content-block"
        >
          <!-- Theory Block -->
          <div v-if="block.type === 'theory'" class="theory-block">
            <div 
              v-if="block.theory?.content_md" 
              class="theory-content markdown-content"
              v-html="renderMarkdown(block.theory.content_md)"
            ></div>
            
            <!-- Key Points (callout) -->
            <div v-if="block.theory?.key_points && block.theory.key_points.length > 0" class="key-points key-points-callout">
              <h3 class="key-points-title">{{ t('grammar.keyPoints') }}</h3>
              <ul class="key-points-list">
                <li v-for="(point, idx) in block.theory.key_points" :key="idx" class="key-point-item">
                  {{ point }}
                </li>
              </ul>
            </div>
            
            <!-- Common Mistakes -->
            <div v-if="block.theory?.common_mistakes && block.theory.common_mistakes.length > 0" class="common-mistakes">
              <h3>{{ t('grammar.commonMistakes') }}</h3>
              <div
                v-for="(mistake, idx) in block.theory.common_mistakes"
                :key="idx"
                class="mistake-item"
              >
                <div class="mistake-wrong">
                  <strong>{{ t('grammar.wrong') }}:</strong> {{ mistake.wrong }}
                </div>
                <div class="mistake-right">
                  <strong>{{ t('grammar.correct') }}:</strong> {{ mistake.right }}
                </div>
                <div class="mistake-why">
                  <strong>{{ t('grammar.why') }}:</strong> {{ mistake.why }}
                </div>
              </div>
            </div>
            
            <!-- Examples -->
            <div v-if="block.theory?.examples && block.theory.examples.length > 0" class="examples">
              <h3>{{ t('grammar.examples') }}</h3>
              <div
                v-for="example in block.theory.examples"
                :key="example.id"
                class="example-item"
              >
                <div class="example-text">{{ example.text }}</div>
                <div class="example-translation">{{ example.translation }}</div>
                <div v-if="example.notes" class="example-notes">{{ example.notes }}</div>
              </div>
            </div>
          </div>
          
          <!-- Inline Quiz Block -->
          <div v-if="block.type === 'quiz_inline' && hasQuizQuestions(block)" class="quiz-block">
            <h2 v-if="block.title" class="block-title">{{ block.title }}</h2>
            <GrammarQuestion
              v-for="questionId in block.quiz_inline?.question_ids || []"
              :key="questionId"
              :question="getQuestionById(questionId)"
              :theory-chapter-context="theoryChapterContextForCourse"
              :show-answers="block.quiz_inline?.show_answers_immediately"
              :show-theory-help-button="false"
              @answer="handleQuizAnswer(questionId, $event)"
            />
          </div>

        </div>
      </div>
      
      <div class="chapter-footer">
        <button @click="startTest" class="btn btn-primary btn-large">
          {{ t('grammar.startChapterTest') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { marked } from 'marked'
import { grammarClient } from '../api/grammarClient'
import GrammarQuestion from '../components/GrammarQuestion.vue'
import { useSettings } from '../composables/useSettings'
import { useAudio } from '../composables/useAudio'

const { t, locale } = useI18n()
const route = useRoute()
const router = useRouter()
const chapterId = computed(() => route.params.chapterId as string)

const { settings } = useSettings()
const { playSuccess, playFail } = useAudio()

const chapter = ref<any>(null)
const chapterTitleOverride = ref<string | null>(null)
const sectionMeta = ref<{
  title?: string
  title_translations?: Record<string, string>
  level?: string
} | null>(null)
const loading = ref(true)
const error = ref<string | null>(null)
const questionMap = ref<Map<string, any>>(new Map())

const getLocalizedTitle = (title: string, titleTranslations?: Record<string, string>) => {
  const currentLocale = locale.value
  if (currentLocale && currentLocale !== 'en' && titleTranslations?.[currentLocale]) {
    return titleTranslations[currentLocale]
  }
  return title
}

const chapterTitle = computed(() => {
  const baseTitle = chapterTitleOverride.value || chapter.value?.title || chapterId.value
  const translations = chapter.value?.title_translations
  return getLocalizedTitle(baseTitle, translations)
})

const theoryChapterContextForCourse = computed(() => {
  if (!chapter.value) return null
  const sec = sectionMeta.value
  const baseTitle = chapterTitleOverride.value || chapter.value?.title || chapterId.value
  const level =
    (typeof chapter.value.level === 'string' && chapter.value.level.length > 0 ? chapter.value.level : '') ||
    (typeof sec?.level === 'string' && sec.level.length > 0 ? sec.level : '')
  return {
    categoryTitle: typeof sec?.title === 'string' ? sec.title : '',
    categoryTitleTranslations: sec?.title_translations,
    chapterTitle: baseTitle,
    chapterTitleTranslations: chapter.value.title_translations,
    level
  }
})

const loadChapter = async () => {
  loading.value = true
  error.value = null
  try {
    const data: {
      chapter: any
      title: string
      title_translations?: Record<string, string>
      section?: { title?: string; title_translations?: Record<string, string>; level?: string }
    } = await grammarClient.getChapter(chapterId.value)
    sectionMeta.value = data.section ?? null
    chapter.value = data.chapter
    chapterTitleOverride.value = data.title || null
    if (data.title_translations && chapter.value) {
      chapter.value.title_translations = data.title_translations
    }
    // Build question map for quick lookup
    if (chapter.value?.question_bank?.questions) {
      questionMap.value = new Map()
      chapter.value.question_bank.questions.forEach((q: any) => {
        questionMap.value.set(q.id, q)
      })
    }
  } catch (err: any) {
    error.value = err.message || 'Failed to load chapter'
    console.error('Failed to load grammar chapter:', err)
  } finally {
    loading.value = false
  }
}

const getQuestionById = (questionId: string) => {
  return questionMap.value.get(questionId) || null
}

const hasQuizQuestions = (block: any): boolean => {
  return Array.isArray(block?.quiz_inline?.question_ids) && block.quiz_inline.question_ids.length > 0
}

// Helper function to compare answers
const compareAnswers = (userAnswer: any, correctAnswer: any): boolean => {
  if (typeof correctAnswer === 'string') {
    return userAnswer?.trim().toLowerCase() === correctAnswer.trim().toLowerCase()
  }
  if (Array.isArray(correctAnswer)) {
    if (!Array.isArray(userAnswer)) return false
    if (userAnswer.length !== correctAnswer.length) return false
    // Sort both arrays for comparison
    const sortedUser = [...userAnswer].sort()
    const sortedCorrect = [...correctAnswer].sort()
    return sortedUser.every((val, idx) => val === sortedCorrect[idx])
  }
  return userAnswer === correctAnswer
}

// Haptic feedback helper function
const triggerHapticFeedback = (isCorrect: boolean) => {
  if (!settings.value.vibrationEnabled) return
  
  const tg = (window as any).Telegram?.WebApp
  
  // Try Telegram Web App API first
  if (tg?.HapticFeedback) {
    try {
      const haptic = tg.HapticFeedback
      if (isCorrect) {
        if (typeof haptic.notificationOccurred === 'function') {
          haptic.notificationOccurred('success')
        } else if (typeof haptic.impactOccurred === 'function') {
          haptic.impactOccurred('medium')
        }
      } else {
        if (typeof haptic.notificationOccurred === 'function') {
          haptic.notificationOccurred('error')
        } else if (typeof haptic.impactOccurred === 'function') {
          haptic.impactOccurred('heavy')
        }
      }
      return
    } catch (error) {
      console.warn('Telegram haptic feedback failed:', error)
    }
  }
  
  // Fallback to native Vibration API
  if ('vibrate' in navigator && typeof navigator.vibrate === 'function') {
    try {
      if (isCorrect) {
        navigator.vibrate(50)
      } else {
        navigator.vibrate([100, 50, 100])
      }
    } catch (error) {
      console.warn('Native vibration failed:', error)
    }
  }
}

const handleQuizAnswer = (questionId: string, answer: any) => {
  // For inline quizzes, we can show immediate feedback
  // Answers are already included in the chapter data
  
  // Only play sounds/vibration if answers are shown immediately
  const question = getQuestionById(questionId)
  if (!question || !question.correct_answer) {
    return
  }
  
  // Check if this quiz block shows answers immediately
  // We need to find the block that contains this question
  let showAnswersImmediately = false
  if (chapter.value?.blocks) {
    for (const block of chapter.value.blocks) {
      if (block.type === 'quiz_inline' && block.quiz_inline?.question_ids) {
        if (block.quiz_inline.question_ids.includes(questionId)) {
          showAnswersImmediately = block.quiz_inline.show_answers_immediately || false
          break
        }
      }
    }
  }
  
  // Only play feedback if answers are shown immediately
  if (showAnswersImmediately) {
    const isCorrect = compareAnswers(answer, question.correct_answer)
    
    // Play sound
    if (settings.value.soundsEnabled) {
      if (isCorrect) {
        playSuccess(settings.value.soundTheme)
      } else {
        playFail(settings.value.soundTheme)
      }
    }
    
    // Trigger haptic feedback
    triggerHapticFeedback(isCorrect)
  }
}

const startTest = () => {
  router.push(`/learning/grammar/chapter/${chapterId.value}/test`)
}

const renderMarkdown = (text: string): string => {
  if (!text) return ''
  try {
    marked.setOptions({
      breaks: true,
      gfm: true,
    })
    return marked.parse(text) as string
  } catch (error) {
    console.error('Failed to render markdown:', error)
    return text
  }
}

onMounted(() => {
  loadChapter()
})
</script>

<style scoped>
.grammar-chapter {
  max-width: 900px;
  margin: 0 auto;
  padding: 20px;
}

.loading, .error {
  text-align: center;
  padding: 40px 20px;
}

.error {
  color: var(--color-danger);
}

.chapter-header {
  margin-bottom: 24px;
}

.chapter-header h1 {
  margin: 0 0 12px 0;
  overflow-wrap: anywhere;
  word-break: break-word;
}

.chapter-meta {
  display: flex;
  gap: 8px;
}

.meta-badge {
  padding: 4px 8px;
  background: var(--color-primary-light);
  color: var(--color-primary);
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.chapter-description {
  margin-bottom: 32px;
  padding: 16px;
  background: var(--bg-tertiary);
  border-radius: 8px;
  color: var(--text-secondary);
  font-size: 14px;
}

.chapter-content {
  margin-bottom: 32px;
}

.content-block {
  margin-bottom: 32px;
}

.block-title {
  margin: 0 0 16px 0;
  font-size: 20px;
  color: var(--text-primary);
}

.theory-block {
  background: var(--card-bg);
  border: 2px solid var(--border-primary);
  border-radius: 8px;
  padding: 24px;
}

.theory-content {
  margin-bottom: 24px;
  line-height: 1.7;
}

.common-mistakes, .examples {
  margin-top: 24px;
  padding-top: 24px;
  border-top: 1px solid var(--border-primary);
}

.key-points-callout {
  margin-top: 24px;
  padding: 20px 24px;
  background: var(--color-primary-light);
  border-left: 4px solid var(--color-primary);
  border-radius: 8px;
  border-top: none;
}

.key-points-title {
  margin: 0 0 12px 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--color-primary);
}

.key-points-list {
  margin: 0;
  padding-left: 0;
  list-style: none;
}

.key-point-item {
  position: relative;
  margin-bottom: 10px;
  padding-left: 24px;
  line-height: 1.6;
}

.key-point-item::before {
  content: '✓';
  position: absolute;
  left: 0;
  color: var(--color-primary);
  font-weight: 700;
  font-size: 14px;
}

.key-point-item:last-child {
  margin-bottom: 0;
}

.common-mistakes h3, .examples h3 {
  margin: 0 0 12px 0;
  font-size: 16px;
  color: var(--text-primary);
}

.mistake-item {
  margin-bottom: 16px;
  padding: 12px;
  background: var(--bg-tertiary);
  border-radius: 6px;
}

.mistake-wrong {
  color: var(--color-danger);
  margin-bottom: 4px;
}

.mistake-right {
  color: var(--color-success);
  margin-bottom: 4px;
}

.mistake-why {
  color: var(--text-secondary);
  font-size: 14px;
}

.example-item {
  margin-bottom: 16px;
  padding: 12px;
  background: var(--example-bg);
  border-radius: 6px;
  border-left: 4px solid var(--color-primary);
}

.example-text {
  font-weight: 600;
  margin-bottom: 4px;
  color: var(--text-primary);
}

.example-translation {
  color: var(--text-secondary);
  font-size: 14px;
  margin-bottom: 4px;
}

.example-notes {
  color: var(--text-tertiary);
  font-size: 12px;
  font-style: italic;
}

.quiz-block {
  margin-top: 32px;
}

.chapter-footer {
  margin-top: 48px;
  padding-top: 24px;
  border-top: 2px solid var(--border-primary);
  text-align: center;
}

.btn {
  padding: 12px 24px;
  border-radius: 8px;
  border: none;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-primary {
  background: var(--color-primary);
  color: white;
}

.btn-primary:hover {
  background: var(--color-primary-hover);
}

.btn-large {
  padding: 16px 32px;
  font-size: 16px;
}

@media (max-width: 768px) {
  .chapter-header h1 {
    line-height: 1.25;
    hyphens: auto;
  }
}
</style>
