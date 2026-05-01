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
        <div
          v-if="grammarTrainingAvailable && grammarTrainingTheoryBlockCount > 0"
          class="grammar-training-stats-inline"
        >
          <span class="grammar-training-inline-seg">
            <span class="grammar-training-inline-num">{{ grammarTrainingTheoryBlockCount }}</span>
            <span class="grammar-training-inline-lbl">{{ t('grammar.trainingStatTopicsLabel') }}</span>
          </span>
          <span class="grammar-training-inline-divider" aria-hidden="true"></span>
          <span
            class="grammar-training-inline-seg grammar-training-inline-seg--due"
            :class="{ 'grammar-training-inline-seg--due-zero': grammarTrainingDueCount === 0 }"
          >
            <span class="grammar-training-inline-num">{{ grammarTrainingDueCount }}</span>
            <span class="grammar-training-inline-lbl">{{ t('grammar.trainingStatDueLabel') }}</span>
          </span>
        </div>
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
const grammarTrainingTheoryBlockCount = ref(0)
const grammarTrainingDueCount = ref(0)

onMounted(async () => {
  try {
    const data: {
      grammar_training?: {
        available?: boolean
        theory_block_count?: number
        due_theory_block_count?: number
      }
    } = await apiClient.request('/api/learning/grammar/training/availability')
    const gt = data?.grammar_training
    grammarTrainingAvailable.value = !!gt?.available
    grammarTrainingTheoryBlockCount.value =
      typeof gt?.theory_block_count === 'number' ? gt.theory_block_count : 0
    grammarTrainingDueCount.value =
      typeof gt?.due_theory_block_count === 'number' ? gt.due_theory_block_count : 0
  } catch {
    grammarTrainingAvailable.value = false
    grammarTrainingTheoryBlockCount.value = 0
    grammarTrainingDueCount.value = 0
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

.grammar-training-stats-inline {
  display: inline-flex;
  align-items: center;
  flex-wrap: wrap;
  justify-content: center;
  margin: 6px 0 8px;
  padding: 2px 2px 2px 8px;
  border-radius: 999px;
  background: var(--bg-secondary);
  background: color-mix(in srgb, var(--bg-secondary) 85%, transparent);
  border: 1px solid var(--border-primary);
  border-color: color-mix(in srgb, var(--border-primary) 65%, transparent);
  box-sizing: border-box;
}

.grammar-training-inline-seg {
  display: inline-flex;
  align-items: baseline;
  gap: 5px;
  padding: 5px 8px;
  white-space: nowrap;
}

.grammar-training-inline-num {
  font-size: 1rem;
  font-weight: 700;
  line-height: 1;
  font-variant-numeric: tabular-nums;
  color: var(--text-primary);
}

.grammar-training-inline-lbl {
  font-size: 0.72rem;
  font-weight: 600;
  color: var(--text-secondary);
  line-height: 1;
  opacity: 0.92;
}

.grammar-training-inline-divider {
  flex-shrink: 0;
  width: 1px;
  height: 14px;
  margin: 0 1px;
  background: var(--border-primary);
  opacity: 0.55;
}

.grammar-training-inline-seg--due .grammar-training-inline-num {
  color: var(--color-primary);
}

.grammar-training-inline-seg--due-zero .grammar-training-inline-num {
  color: var(--text-secondary);
  font-weight: 600;
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
