<template>
  <div v-if="ready && allowed" class="verb-training lg-page">
    <LgPageHeader
      :title="t('verbTraining.title')"
      :show-back="true"
      @back="goBack"
    />
    <VerbFormsTrainingPanel embedded hide-page-title :auto-start="autoStart" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import LgPageHeader from '../components/linglow/LgPageHeader.vue'
import VerbFormsTrainingPanel from '../components/VerbFormsTrainingPanel.vue'
import { useLearningConfig, ensureLearningLoaded } from '../composables/useLearningConfig'

const router = useRouter()
const route = useRoute()
const { t } = useI18n()
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

function goBack() {
  if (typeof window !== 'undefined' && window.history.length > 1) {
    router.back()
    return
  }
  router.push('/learning')
}

onMounted(async () => {
  await ensureLearningLoaded()
  ready.value = true
  if (!allowed.value) {
    router.replace('/learning')
  }
})
</script>

<style scoped>
.verb-training {
  max-width: 980px;
  margin: 0 auto;
}
</style>
