<template>
  <div class="journey-error" role="alert">
    <p>{{ message }}</p>
    <div class="journey-error-actions">
      <button type="button" :disabled="busy" @click="$emit('retry')">{{ t('common.retry') }}</button>
      <button v-if="showBack" type="button" class="secondary" :disabled="busy" @click="$emit('back')">{{ t('dailyJourney.returnToRoute') }}</button>
    </div>
  </div>
</template>
<script setup lang="ts">
import { useI18n } from 'vue-i18n'
withDefaults(defineProps<{ message: string; busy?: boolean; showBack?: boolean }>(), { busy: false, showBack: true })
defineEmits<{ retry: []; back: [] }>()
const { t } = useI18n()
</script>
<style scoped>
.journey-error { padding: 20px; margin-bottom: 20px; border: 1px solid var(--error, #b75145); border-radius: 18px; background: var(--card-bg); color: var(--text); }
p { margin: 0; line-height: 1.6; overflow-wrap: anywhere; }
.journey-error-actions { display: flex; flex-wrap: wrap; gap: 10px; margin-top: 16px; }
button { min-height: 44px; padding: 10px 16px; border: 1px solid transparent; border-radius: 12px; font: inherit; font-size: .85rem; font-weight: 600; color: var(--crema); background: var(--btn-gradient); cursor: pointer; }
.secondary { border-color: var(--border); background: transparent; color: var(--text); }
button:disabled { opacity: .55; cursor: default; }
button:focus-visible { outline: 2px solid var(--salvia); outline-offset: 3px; }
</style>
