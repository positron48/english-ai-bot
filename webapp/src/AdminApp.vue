<template>
  <div id="app">
    <main class="container main-admin">
      <router-view v-if="mounted" />
      <div v-else style="padding: 20px; text-align: center;">
        {{ t('common.loading') }}
      </div>
    </main>

    <AlertModal
      :message="alertState.message"
      :visible="alertState.visible"
      @close="closeAlert"
    />
    <ConfirmModal
      :message="confirmState.message"
      :visible="confirmState.visible"
      @confirm="() => closeConfirm(true)"
      @cancel="() => closeConfirm(false)"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useDialog } from './composables/useDialog'
import { useLearningConfig } from './composables/useLearningConfig'
import AlertModal from './components/AlertModal.vue'
import ConfirmModal from './components/ConfirmModal.vue'

const { t } = useI18n()
const { alertState, confirmState, closeAlert, closeConfirm } = useDialog()
const { ensureLearningLoaded } = useLearningConfig()

const mounted = ref(false)

onMounted(() => {
  ensureLearningLoaded()
  mounted.value = true
})
</script>

<style scoped>
main.main-admin {
  max-width: none;
  margin: 0;
  padding: 0;
}
</style>
