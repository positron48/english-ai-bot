<template>
  <div v-if="ready && allowed" class="verb-training-page">
    <VerbFormsTrainingPanel :embedded="false" :auto-start="autoStart" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import VerbFormsTrainingPanel from '../components/VerbFormsTrainingPanel.vue'
import { useLearningConfig, ensureLearningLoaded } from '../composables/useLearningConfig'

const router = useRouter()
const route = useRoute()
const { learning } = useLearningConfig()
const ready = ref(false)

const autoStart = computed(() => {
  const v = route.query.start
  if (v === true) return true
  if (Array.isArray(v)) return v.some((x) => String(x).toLowerCase() === '1' || String(x).toLowerCase() === 'true')
  return String(v || '').toLowerCase() === '1' || String(v || '').toLowerCase() === 'true'
})

const allowed = computed(
  () =>
    (learning.value?.target_lang || '').toLowerCase() === 'es' &&
    learning.value?.spanish_verb_forms_enabled === true
)

onMounted(async () => {
  await ensureLearningLoaded()
  ready.value = true
  if (!allowed.value) {
    router.replace('/training')
  }
})
</script>

<style scoped>
.verb-training-page {
  max-width: 1200px;
  margin: 0 auto;
  padding: 10px;
}
</style>
