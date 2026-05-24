<template>
  <div class="speaking-categories">
    <h1>{{ t('speaking.title') }}</h1>
    <p class="subtitle">{{ t('speaking.subtitle') }}</p>

    <div v-if="loading" class="state">{{ t('common.loading') }}</div>
    <div v-else-if="error" class="state error">{{ error }}</div>
    <div v-else-if="categories.length === 0" class="state">{{ t('speaking.noCategories') }}</div>
    <div v-else class="category-grid">
      <button
        v-for="cat in categories"
        :key="cat.category_id"
        type="button"
        class="category-card"
        @click="start(cat.category_id)"
      >
        <span class="level">{{ cat.level }}</span>
        <span class="title">{{ cat.title }}</span>
        <span class="meta">{{ t('speaking.taskCount', { n: cat.task_count }) }}</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useSpeaking, type SpeakingCategory } from '../composables/useSpeaking'

const { t } = useI18n()
const router = useRouter()
const { loadCategories, startSession } = useSpeaking()

const loading = ref(true)
const error = ref('')
const categories = ref<SpeakingCategory[]>([])

onMounted(async () => {
  try {
    categories.value = await loadCategories()
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : t('speaking.loadFailed')
  } finally {
    loading.value = false
  }
})

async function start(categoryId: string) {
  try {
    const session = await startSession(categoryId)
    await router.push(`/learning/speaking/session/${session.id}`)
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : t('speaking.startFailed')
  }
}
</script>

<style scoped>
.speaking-categories {
  max-width: 900px;
  margin: 0 auto;
  padding: 20px;
}
.subtitle {
  color: var(--text-secondary);
  margin-bottom: 24px;
}
.category-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 16px;
}
.category-card {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 6px;
  padding: 20px;
  border: 2px solid var(--border-primary);
  border-radius: 12px;
  background: var(--card-bg);
  cursor: pointer;
  text-align: left;
}
.category-card:hover {
  border-color: var(--color-primary);
}
.level {
  font-size: 12px;
  font-weight: 700;
  color: var(--color-primary);
}
.title {
  font-size: 18px;
  font-weight: 600;
}
.meta {
  font-size: 13px;
  color: var(--text-secondary);
}
.state.error {
  color: var(--color-danger, #c0392b);
}
</style>
