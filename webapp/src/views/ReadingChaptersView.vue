<template>
  <div class="reading-chapters">
    <div class="page-heading">
      <button
        type="button"
        class="back-levels-btn"
        :aria-label="t('common.back')"
        @click="router.push('/learning/reading')"
      >
        <Icon name="arrow-left" />
        <span>{{ t('common.back') }}</span>
      </button>
      <h1 class="page-title">{{ pageHeading }}</h1>
      <button
        v-if="!loading && !error && texts.length > 0 && unreadTexts.length > 0"
        type="button"
        class="random-unread-btn"
        :aria-label="t('reading.randomUnread')"
        :title="t('reading.randomUnread')"
        @click="openRandomUnread"
      >
        <Icon name="dice" />
      </button>
    </div>
    <div v-if="loading">{{ t('common.loading') }}</div>
    <div v-else-if="error">{{ error }}</div>
    <div v-else-if="texts.length === 0" class="empty">{{ t('reading.noTexts') }}</div>
    <div v-else class="split-layout">
      <div v-if="unreadTexts.length" class="list">
        <router-link
          v-for="text in unreadTexts"
          :key="text.text_id"
          :to="`/learning/reading/text/${text.text_id}`"
          class="item"
        >
          <strong>{{ getLocalizedTitle(text.title, text.title_translations) }}</strong>
          <span class="level-pill">{{ text.level }}</span>
        </router-link>
      </div>
      <p v-else class="all-read-hint">{{ t('reading.allReadInCategory') }}</p>

      <section v-if="readTexts.length" class="completed-section">
        <h2 class="section-title">{{ t('reading.completedSection') }}</h2>
        <div class="list list--read">
          <router-link
            v-for="text in readTexts"
            :key="'read-' + text.text_id"
            :to="`/learning/reading/text/${text.text_id}`"
            class="item item--read"
          >
            <span class="item-title">{{ getLocalizedTitle(text.title, text.title_translations) }}</span>
            <span class="read-badge" aria-hidden="true">{{ t('reading.alreadyRead') }}</span>
          </router-link>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { apiClient } from '../api/client'
import Icon from '../components/Icon.vue'

const { t, locale } = useI18n()
const route = useRoute()
const router = useRouter()
const categoryId = computed(() => route.params.categoryId as string)
const loading = ref(true)
const error = ref<string | null>(null)
const texts = ref<any[]>([])
const categoryLevel = ref('')
const categoryTitleFallback = ref('Reading')

const readTexts = computed(() => texts.value.filter((x) => x.is_read))
const unreadTexts = computed(() => texts.value.filter((x) => !x.is_read))

const pageHeading = computed(() => {
  const lv = categoryLevel.value.trim()
  if (lv) return lv
  return categoryTitleFallback.value
})

const getLocalizedTitle = (value: string, translations?: Record<string, string>) => {
  const currentLocale = locale.value
  if (currentLocale && currentLocale !== 'en' && translations?.[currentLocale]) {
    return translations[currentLocale]
  }
  return value
}

const openRandomUnread = () => {
  const pool = unreadTexts.value
  if (!pool.length) return
  const pick = pool[Math.floor(Math.random() * pool.length)]
  router.push(`/learning/reading/text/${pick.text_id}`)
}

onMounted(async () => {
  loading.value = true
  try {
    const data: { category?: any; texts?: any[] } = await apiClient.request(
      `/api/learning/reading/categories/${categoryId.value}/texts`
    )
    texts.value = data.texts || []
    if (data.category?.level) {
      categoryLevel.value = String(data.category.level).trim()
    }
    if (data.category?.title) {
      categoryTitleFallback.value = getLocalizedTitle(
        data.category.title,
        data.category.title_translations
      )
    }
  } catch (e: any) {
    error.value = e?.message || 'Failed to load reading texts'
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.reading-chapters { max-width: 900px; margin: 0 auto; padding: 20px; }
.page-heading {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 1rem;
}
.page-title { margin: 0; flex: 1; min-width: 0; }
.back-levels-btn {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 42px;
  padding: 0 12px;
  border: 2px solid var(--border-primary);
  border-radius: 10px;
  background: var(--card-bg);
  color: var(--text-primary);
  cursor: pointer;
  font: inherit;
}
.back-levels-btn:hover {
  border-color: var(--accent-primary, #3b82f6);
  color: var(--accent-primary, #3b82f6);
}
.random-unread-btn {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 42px;
  height: 42px;
  padding: 0;
  border: 2px solid var(--border-primary);
  border-radius: 10px;
  background: var(--card-bg);
  color: var(--text-primary);
  cursor: pointer;
  font-size: 1.35rem;
}
.random-unread-btn:hover {
  border-color: var(--accent-primary, #3b82f6);
  color: var(--accent-primary, #3b82f6);
}
.split-layout { display: flex; flex-direction: column; gap: 0; }
.list { display: flex; flex-direction: column; gap: 10px; }
.item {
  border: 2px solid var(--border-primary);
  border-radius: 8px;
  padding: 14px;
  text-decoration: none;
  color: var(--text-primary);
  background: var(--card-bg);
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
}
.item--read {
  border-style: dashed;
  opacity: 0.92;
}
.level-pill {
  flex-shrink: 0;
  font-size: 0.85rem;
  color: var(--text-secondary);
}
.item-title { font-weight: 600; }
.read-badge {
  flex-shrink: 0;
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--text-secondary);
}
.all-read-hint {
  margin: 0 0 0.5rem;
  color: var(--text-secondary);
  font-size: 0.95rem;
}
.completed-section {
  margin-top: 1.75rem;
  padding-top: 1.25rem;
  border-top: 1px solid var(--border-primary);
}
.section-title {
  margin: 0 0 0.75rem;
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--text-secondary);
  letter-spacing: 0.02em;
}
.list--read { margin-top: 0; }
.empty { color: var(--text-secondary); }
</style>

