<template>
  <div class="learning">
    <h1>{{ t('learning.title') }}</h1>
    
    <div class="learning-sections">
      <router-link to="/learning/grammar" class="learning-card grammar-card">
        <div class="card-icon">
          <Icon name="book-open" />
        </div>
        <h2>{{ t('learning.grammar') }}</h2>
        <p>{{ t('learning.grammarDescription') }}</p>
        <button
          v-if="grammarTrainingAvailable"
          class="btn btn-primary grammar-training-btn"
          @click.stop="goGrammarTraining"
        >
          {{ t('grammar.trainingTitle') || 'Grammar Training' }}
        </button>
      </router-link>
      
      <router-link to="/learning/words" class="learning-card words-card">
        <div class="card-icon">
          <Icon name="book" />
        </div>
        <h2>{{ t('learning.words') }}</h2>
        <p>{{ t('learning.wordsDescription') }}</p>
      </router-link>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { apiClient } from '../api/client'
import Icon from '../components/Icon.vue'

const { t } = useI18n()
const router = useRouter()
const grammarTrainingAvailable = ref(false)

onMounted(async () => {
  try {
    const data: any = await apiClient.request('/api/learning/grammar/training/availability')
    grammarTrainingAvailable.value = !!data?.grammar_training?.available
  } catch {
    grammarTrainingAvailable.value = false
  }
})

const goGrammarTraining = () => {
  router.push('/learning/grammar/training')
}
</script>

<style scoped>
.learning {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

.learning h1 {
  margin-bottom: 32px;
}

.learning-sections {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 24px;
}

.learning-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 32px 24px;
  background: var(--card-bg);
  border: 2px solid var(--border-primary);
  border-radius: 12px;
  text-decoration: none;
  color: var(--text-primary);
  transition: all 0.3s ease;
  position: relative;
}

.learning-card:hover {
  border-color: var(--color-primary);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}


.card-icon {
  font-size: 48px;
  margin-bottom: 16px;
  color: var(--color-primary);
}

.learning-card h2 {
  margin: 0 0 8px 0;
  font-size: 24px;
}

.learning-card p {
  margin: 0;
  color: var(--text-secondary);
  text-align: center;
}

.grammar-training-btn {
  margin-top: 14px;
}

.coming-soon {
  position: absolute;
  top: 12px;
  right: 12px;
  font-size: 12px;
  padding: 4px 8px;
  background: var(--color-warning, #f59e0b);
  color: white;
  border-radius: 4px;
  font-weight: 600;
}

@media (max-width: 768px) {
  .learning-sections {
    grid-template-columns: 1fr;
  }
}
</style>
