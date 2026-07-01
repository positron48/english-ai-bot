<template>
  <Teleport to="body">
    <div v-if="modalVisible" class="modal-overlay" @click.self="dismiss">
      <div class="modal-container">
        <div class="modal-content">
          <div class="modal-body">
            <h2 class="update-title">{{ t('appUpdate.title') }}</h2>
            <p class="update-text">
              {{ t('appUpdate.newVersionAvailable', { version: latestVersion }) }}
            </p>
            <p class="update-current">
              {{ t('appUpdate.currentVersion', { version: currentVersion }) }}
            </p>
            <p v-if="downloadStatus === 'downloading'" class="update-status">
              {{ t('appUpdate.downloading') }}
            </p>
            <div v-if="downloadStatus === 'downloading'" class="update-progress" :class="{ 'update-progress--indeterminate': downloadProgress === null }">
              <div
                class="update-progress-fill"
                :style="downloadProgress !== null ? { width: `${downloadProgress}%` } : undefined"
              />
            </div>
            <p v-if="downloadStatus === 'downloading' && downloadProgress !== null" class="update-progress-text">
              {{ t('appUpdate.downloadProgress', { percent: downloadProgress }) }}
            </p>
            <p v-if="downloadStatus === 'installing'" class="update-status">
              {{ t('appUpdate.installing') }}
            </p>
            <p v-if="downloadStatus === 'error'" class="update-status update-status--error">
              {{ errorText }}
            </p>
            <p v-if="downloadStatus === 'idle'" class="update-hint">{{ t('appUpdate.sourceHint') }}</p>
          </div>
          <div class="modal-footer">
            <button
              class="btn btn-primary"
              :disabled="busy"
              @click="installUpdate"
            >
              {{ busy ? t('appUpdate.downloading') : t('appUpdate.update') }}
            </button>
            <button v-if="downloadStatus === 'downloading' && canCancelDownload" class="btn btn-secondary" @click="cancelDownload">
              {{ t('common.cancel') }}
            </button>
            <button class="btn btn-secondary" :disabled="busy" @click="snooze24h">
              {{ t('appUpdate.later') }}
            </button>
            <button class="btn btn-ghost" :disabled="busy" @click="skipVersion">
              {{ t('appUpdate.skipVersion') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppUpdate } from '../composables/useAppUpdate'

const { t } = useI18n()
const {
  modalVisible,
  latestVersion,
  currentVersion,
  downloadStatus,
  downloadProgress,
  canCancelDownload,
  errorMessage,
  installUpdate,
  cancelDownload,
  skipVersion,
  snooze24h,
  dismiss,
} = useAppUpdate()

const busy = computed(
  () => downloadStatus.value === 'downloading' || downloadStatus.value === 'installing',
)
const errorText = computed(() => {
  if (errorMessage.value === 'download_timeout') return t('appUpdate.downloadTimeout')
  return t('appUpdate.installFailed')
})
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: var(--bg-modal-overlay);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10030;
  animation: fadeIn 0.2s ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.modal-content {
  background: var(--bg-secondary);
  border-radius: 12px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.3);
  min-width: 300px;
  max-width: 460px;
  width: min(92vw, 460px);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.modal-body {
  padding: 24px;
  color: var(--text-primary);
}

.update-title {
  margin: 0 0 12px;
  font-size: 20px;
  font-weight: 800;
}

.update-text {
  margin: 0 0 6px;
  line-height: 1.5;
}

.update-current,
.update-hint {
  margin: 0;
  font-size: 13px;
  color: var(--subtext);
  line-height: 1.4;
}

.update-status {
  margin: 12px 0 0;
  font-weight: 700;
}

.update-status--error {
  color: #b91c1c;
}
.update-progress {
  position: relative;
  height: 8px;
  margin-top: 12px;
  overflow: hidden;
  border-radius: 999px;
  background: var(--progress-track, rgba(32, 53, 42, 0.12));
}

.update-progress-fill {
  height: 100%;
  border-radius: inherit;
  background: var(--btn-gradient, #3f6f3f);
  transition: width 180ms ease;
}

.update-progress--indeterminate .update-progress-fill {
  width: 42%;
  animation: update-progress-slide 1.1s ease-in-out infinite;
}

.update-progress-text {
  margin: 8px 0 0;
  font-size: 13px;
  color: var(--subtext);
}

@keyframes update-progress-slide {
  from { transform: translateX(-110%); }
  to { transform: translateX(250%); }
}

.modal-footer {
  padding: 16px 24px;
  border-top: 1px solid var(--border-primary);
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 12px;
}

.btn-ghost {
  background: transparent;
  color: var(--subtext);
}
</style>
