<template>
  <div class="learning-grammar">
    <div class="card">
      <h1>Grammar</h1>
      <p class="message">This section is under development.</p>
      <button @click="goBack" class="btn btn-primary">Go Back</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { onMounted } from 'vue'
import { showAlert } from '../composables/useDialog'

const router = useRouter()

onMounted(async () => {
  await showAlert('This section is under development.')
})

const goBack = () => {
  router.push('/learning')
}
</script>

<style scoped>
.learning-grammar {
  max-width: 800px;
  margin: 0 auto;
  padding: 20px;
}

.card {
  text-align: center;
  padding: 60px 40px;
}

.card h1 {
  margin-bottom: 24px;
}

.message {
  font-size: 18px;
  color: var(--text-secondary);
  margin-bottom: 32px;
}

.btn {
  padding: 12px 24px;
  font-size: 16px;
  border-radius: 8px;
  cursor: pointer;
  border: none;
  font-weight: 600;
  transition: all 0.2s;
}

.btn-primary {
  background-color: var(--color-primary, #3b82f6);
  color: white;
}

.btn-primary:hover {
  background-color: var(--color-primary-hover, #2563eb);
}
</style>
