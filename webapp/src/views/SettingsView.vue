<template>
  <div class="settings">
    <h1>{{ t('settings.title') }}</h1>
    
    <div class="card">
      <h2>{{ t('settings.appearance') }}</h2>
      <div class="settings-group">
        <div class="setting-item">
          <div class="setting-info">
            <label class="setting-label">{{ t('settings.theme') }}</label>
            <p class="setting-description">{{ t('settings.themeDescription') }}</p>
          </div>
          <div class="setting-control">
            <label class="theme-toggle-switch">
              <input 
                type="checkbox" 
                :checked="selectedTheme === 'dark'"
                @change="handleThemeToggle"
              />
              <span class="theme-toggle-slider">
                <Icon name="sun" class="theme-icon theme-icon-sun" />
                <Icon name="moon" class="theme-icon theme-icon-moon" />
              </span>
            </label>
          </div>
        </div>
      </div>
    </div>

    <div class="card">
      <h2>{{ t('settings.training') }}</h2>
      <div class="settings-group">
        <div class="setting-item">
          <div class="setting-info">
            <label class="setting-label">{{ t('settings.vibration') }}</label>
            <p class="setting-description">{{ t('settings.vibrationDescription') }}</p>
          </div>
          <div class="setting-control">
            <label class="toggle-switch">
              <input 
                type="checkbox" 
                v-model="vibrationEnabled"
                @change="handleVibrationChange"
              />
              <span class="toggle-slider"></span>
            </label>
          </div>
        </div>
        
        <div class="setting-item">
          <div class="setting-info">
            <label class="setting-label">{{ t('settings.sounds') }}</label>
            <p class="setting-description">{{ t('settings.soundsDescription') }}</p>
          </div>
          <div class="setting-control">
            <label class="toggle-switch">
              <input 
                type="checkbox" 
                v-model="soundsEnabled"
                @change="handleSoundsChange"
              />
              <span class="toggle-slider"></span>
            </label>
          </div>
        </div>
        
        <div class="setting-item" v-if="soundsEnabled">
          <div class="setting-info">
            <label class="setting-label">{{ t('settings.soundTheme') }}</label>
            <p class="setting-description">{{ t('settings.soundThemeDescription') }}</p>
          </div>
          <div class="setting-control">
            <div class="sound-theme-control">
              <select v-model="selectedSoundTheme" @change="handleSoundThemeChange" class="theme-select">
                <option v-for="theme in soundThemes" :key="theme.id" :value="theme.id">
                  {{ theme.name }}
                </option>
              </select>
              <button 
                @click="previewSounds" 
                class="preview-btn"
                :disabled="previewing"
                :title="t('settings.preview')"
              >
                <Icon name="play" />
              </button>
            </div>
          </div>
        </div>

        <div class="setting-item">
          <div class="setting-info">
            <label class="setting-label">{{ t('settings.optionsDelay') }}</label>
            <p class="setting-description">{{ t('settings.optionsDelayDescription') }}</p>
          </div>
          <div class="setting-control delay-control">
            <span class="saved-indicator-slot">
              <span v-if="trainingDelaysSavedAt === 'options'" class="saved-indicator">{{ t('common.saved') }}</span>
            </span>
            <input
              v-model.number="optionsDelaySeconds"
              type="range"
              min="0"
              max="10"
              class="delay-slider"
              @change="handleOptionsDelayChange"
            />
            <span class="delay-value">{{ optionsDelaySeconds }} {{ t('settings.seconds') }}</span>
          </div>
        </div>
        <div class="setting-item">
          <div class="setting-info">
            <label class="setting-label">{{ t('settings.wrongAnswerDelay') }}</label>
            <p class="setting-description">{{ t('settings.wrongAnswerDelayDescription') }}</p>
          </div>
          <div class="setting-control delay-control">
            <span class="saved-indicator-slot">
              <span v-if="trainingDelaysSavedAt === 'wrong'" class="saved-indicator">{{ t('common.saved') }}</span>
            </span>
            <input
              v-model.number="wrongAnswerDelaySeconds"
              type="range"
              min="0"
              max="10"
              class="delay-slider"
              @change="handleWrongAnswerDelayChange"
            />
            <span class="delay-value">{{ wrongAnswerDelaySeconds }} {{ t('settings.seconds') }}</span>
          </div>
        </div>
        <div class="setting-item">
          <div class="setting-info">
            <label class="setting-label">{{ t('settings.spellModeEnabled') }}</label>
            <p class="setting-description">{{ t('settings.spellModeEnabledDescription', { targetLang: targetLangDisplay }) }}</p>
          </div>
          <div class="setting-control">
            <span v-if="trainingDelaysSavedAt === 'spell'" class="saved-indicator">{{ t('common.saved') }}</span>
            <label class="toggle-switch">
              <input v-model="spellModeEnabled" type="checkbox" @change="handleSpellSettingsChange" />
              <span class="toggle-slider"></span>
            </label>
          </div>
        </div>
        <div class="setting-item" v-if="spellModeEnabled">
          <div class="setting-info">
            <label class="setting-label">{{ t('settings.spellMasteringThreshold') }}</label>
            <p class="setting-description">{{ t('settings.spellMasteringThresholdDescription') }}</p>
          </div>
          <div class="setting-control delay-control">
            <input
              v-model.number="spellMasteringThreshold"
              type="range"
              min="0"
              max="100"
              class="delay-slider"
              @change="handleSpellSettingsChange"
            />
            <span class="delay-value">{{ spellMasteringThreshold }}</span>
          </div>
        </div>
        <div class="setting-item">
          <div class="setting-info">
            <label class="setting-label">{{ t('settings.typeModeEnabled') }}</label>
            <p class="setting-description">{{ t('settings.typeModeEnabledDescription', { targetLang: targetLangDisplay }) }}</p>
          </div>
          <div class="setting-control">
            <span v-if="trainingDelaysSavedAt === 'type'" class="saved-indicator">{{ t('common.saved') }}</span>
            <label class="toggle-switch">
              <input v-model="typeModeEnabled" type="checkbox" @change="handleTypeSettingsChange" />
              <span class="toggle-slider"></span>
            </label>
          </div>
        </div>
        <div class="setting-item" v-if="typeModeEnabled">
          <div class="setting-info">
            <label class="setting-label">{{ t('settings.typeMasteringThreshold') }}</label>
            <p class="setting-description">{{ t('settings.typeMasteringThresholdDescription') }}</p>
          </div>
          <div class="setting-control delay-control">
            <input
              v-model.number="typeMasteringThreshold"
              type="range"
              min="0"
              max="100"
              class="delay-slider"
              @change="handleTypeSettingsChange"
            />
            <span class="delay-value">{{ typeMasteringThreshold }}</span>
          </div>
        </div>
        <div class="setting-item">
          <div class="setting-info">
            <label class="setting-label">{{ t('settings.hideMorphInTraining') }}</label>
            <p class="setting-description">{{ t('settings.hideMorphInTrainingDescription') }}</p>
          </div>
          <div class="setting-control">
            <span v-if="trainingDelaysSavedAt === 'morph'" class="saved-indicator">{{ t('common.saved') }}</span>
            <label class="toggle-switch">
              <input v-model="hideMorphInTraining" type="checkbox" @change="handleMorphVisibilityChange" />
              <span class="toggle-slider"></span>
            </label>
          </div>
        </div>
      </div>
    </div>

    <div class="card">
      <h2>{{ t('settings.notifications') }}</h2>
      <div class="settings-group">
        <div class="setting-item">
          <div class="setting-info">
            <label class="setting-label">{{ t('settings.notificationFrequency') }}</label>
            <p class="setting-description">{{ t('settings.notificationFrequencyDescription') }}</p>
          </div>
          <div class="setting-control">
            <div class="notification-control">
              <span v-if="isSaved" class="saved-indicator">{{ t('common.saved') }}</span>
              <select v-model="notificationFrequency" @change="handleNotificationFrequencyChange" class="theme-select">
                <option value="daily">{{ t('settings.daily') }}</option>
                <option value="never">{{ t('settings.never') }}</option>
                <option value="custom">{{ t('settings.custom') }}</option>
              </select>
              <transition name="slide-fade">
                <div v-if="notificationFrequency === 'custom'" class="custom-days-input">
                  <input 
                    type="number" 
                    v-model.number="customDays" 
                    min="1" 
                    max="30"
                    @input="handleCustomDaysChange"
                    class="days-input"
                    :placeholder="t('settings.days')"
                  />
                  <span class="days-label">{{ t('settings.days') }}</span>
                </div>
              </transition>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="card">
      <h2>{{ t('settings.account') }}</h2>
      <div class="settings-group">
        <div class="setting-item">
          <div class="setting-info">
            <label class="setting-label">{{ t('settings.logout') }}</label>
            <p class="setting-description">{{ t('settings.logoutDescription') }}</p>
          </div>
          <div class="setting-control">
            <button @click="handleLogout" class="logout-btn">
              <Icon name="logout" class="logout-icon" />
              <span>{{ t('settings.logout') }}</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useSettings } from '../composables/useSettings'
import { useTheme } from '../composables/useTheme'
import { useAudio } from '../composables/useAudio'
import { useAuth } from '../composables/useAuth'
import { useLearningConfig } from '../composables/useLearningConfig'
import { apiClient } from '../api/client'
import Icon from '../components/Icon.vue'

const { t } = useI18n()

const router = useRouter()
const { settings, setSoundsEnabled, setVibrationEnabled, setTheme, setSoundTheme, setHideMorphInTraining } = useSettings()
const { theme: currentTheme, setTheme: setThemeInTheme } = useTheme()
const { getThemes, previewTheme } = useAudio()
const { logout: authLogout } = useAuth()
const { targetLangDisplay, ensureLearningLoaded } = useLearningConfig()

const soundsEnabled = ref(true)
const vibrationEnabled = ref(true)
const selectedTheme = ref<'light' | 'dark'>('light')
const selectedSoundTheme = ref('tick')
const soundThemes = ref(getThemes())
const previewing = ref(false)

const notificationFrequency = ref('daily')
const customDays = ref(3)
const isSaved = ref(false)
const optionsDelaySeconds = ref(5)
const wrongAnswerDelaySeconds = ref(5)
/** 'options' | 'wrong' - which delay control just saved (to show "saved" there) */
const trainingDelaysSavedAt = ref<'options' | 'wrong' | 'spell' | 'type' | 'morph' | null>(null)
const spellModeEnabled = ref(true)
const spellMasteringThreshold = ref(50)
const typeModeEnabled = ref(true)
const typeMasteringThreshold = ref(70)
const hideMorphInTraining = ref(false)
let trainingDelaysSavedTimeout: ReturnType<typeof setTimeout> | null = null

onMounted(async () => {
  await ensureLearningLoaded()
  // Load current settings
  soundsEnabled.value = settings.value.soundsEnabled
  vibrationEnabled.value = settings.value.vibrationEnabled
  selectedTheme.value = currentTheme.value
  selectedSoundTheme.value = settings.value.soundTheme || 'tick'
  
  // Load notification settings from API
  await loadNotificationSettings()
  // Load training delay settings from API
  await loadTrainingDelaysSettings()
})

const loadNotificationSettings = async () => {
  try {
    const data = await apiClient.request<{ settings: { notification_frequency?: string } }>('/api/settings')
    const freq = data.settings?.notification_frequency || 'daily'
    
    // If it's a custom frequency (number), set to 'custom' and extract days
    if (freq !== 'daily' && freq !== 'never' && !isNaN(Number(freq))) {
      notificationFrequency.value = 'custom'
      customDays.value = Number(freq)
    } else {
      notificationFrequency.value = freq
    }
  } catch (error) {
    console.error('Failed to load notification settings:', error)
  }
}

interface SettingsResponse {
  settings?: {
    options_delay_seconds?: number
    wrong_answer_delay_seconds?: number
    spell_mode_enabled?: boolean
    spell_mastering_threshold?: number
    type_mode_enabled?: boolean
    type_mastering_threshold?: number
    hide_morph_in_training?: boolean
  }
}

const loadTrainingDelaysSettings = async () => {
  try {
    const data = await apiClient.request<SettingsResponse>('/api/settings')
    const s = data.settings
    if (s?.options_delay_seconds !== undefined) {
      optionsDelaySeconds.value = Math.max(0, Math.min(10, s.options_delay_seconds))
    }
    if (s?.wrong_answer_delay_seconds !== undefined) {
      wrongAnswerDelaySeconds.value = Math.max(0, Math.min(10, s.wrong_answer_delay_seconds))
    }
    if (s?.spell_mode_enabled !== undefined) {
      spellModeEnabled.value = s.spell_mode_enabled
    }
    if (s?.spell_mastering_threshold !== undefined) {
      spellMasteringThreshold.value = Math.max(0, Math.min(100, s.spell_mastering_threshold))
    }
    if (s?.type_mode_enabled !== undefined) {
      typeModeEnabled.value = s.type_mode_enabled
    }
    if (s?.type_mastering_threshold !== undefined) {
      typeMasteringThreshold.value = Math.max(0, Math.min(100, s.type_mastering_threshold))
    }
    if (s?.hide_morph_in_training !== undefined) {
      hideMorphInTraining.value = s.hide_morph_in_training
      setHideMorphInTraining(hideMorphInTraining.value)
    }
  } catch (error) {
    console.error('Failed to load training delay settings:', error)
  }
}

const saveTrainingDelays = async (showSavedAt: 'options' | 'wrong' | 'spell' | 'type' | 'morph') => {
  const opts = Math.max(0, Math.min(10, optionsDelaySeconds.value))
  const wrong = Math.max(0, Math.min(10, wrongAnswerDelaySeconds.value))
  const spellThreshold = Math.max(0, Math.min(100, spellMasteringThreshold.value))
  const typeThreshold = Math.max(0, Math.min(100, typeMasteringThreshold.value))
  optionsDelaySeconds.value = opts
  wrongAnswerDelaySeconds.value = wrong
  spellMasteringThreshold.value = spellThreshold
  typeMasteringThreshold.value = typeThreshold
  try {
    await apiClient.request<{ success: boolean }>('/api/settings/training', {
      method: 'POST',
      body: JSON.stringify({
        options_delay_seconds: opts,
        wrong_answer_delay_seconds: wrong,
        spell_mode_enabled: spellModeEnabled.value,
        spell_mastering_threshold: spellThreshold,
        type_mode_enabled: typeModeEnabled.value,
        type_mastering_threshold: typeThreshold,
        hide_morph_in_training: hideMorphInTraining.value
      })
    })
    setHideMorphInTraining(hideMorphInTraining.value)
    if (trainingDelaysSavedTimeout) {
      clearTimeout(trainingDelaysSavedTimeout)
      trainingDelaysSavedTimeout = null
    }
    trainingDelaysSavedAt.value = showSavedAt
    trainingDelaysSavedTimeout = setTimeout(() => {
      trainingDelaysSavedAt.value = null
      trainingDelaysSavedTimeout = null
    }, 2500)
  } catch (error) {
    console.error('Failed to save training delay settings:', error)
  }
}

const handleOptionsDelayChange = () => saveTrainingDelays('options')
const handleWrongAnswerDelayChange = () => saveTrainingDelays('wrong')
const handleSpellSettingsChange = () => saveTrainingDelays('spell')
const handleTypeSettingsChange = () => saveTrainingDelays('type')
const handleMorphVisibilityChange = () => saveTrainingDelays('morph')

const handleNotificationFrequencyChange = async () => {
  const freq = notificationFrequency.value
  
  // If it's a predefined option (daily or never), save immediately
  if (freq === 'daily' || freq === 'never') {
    await saveNotificationFrequency(freq)
  }
  // If it's 'custom', save with current value (or default to 3)
  if (freq === 'custom') {
    if (!customDays.value || customDays.value < 1) {
      customDays.value = 3
    }
    await saveNotificationFrequency(String(customDays.value))
  }
}

const handleCustomDaysChange = async () => {
  // Auto-save when user changes the number
  if (customDays.value && customDays.value >= 1 && customDays.value <= 30) {
    // Keep 'custom' selected in the dropdown
    notificationFrequency.value = 'custom'
    await saveNotificationFrequency(String(customDays.value))
  }
}

const saveNotificationFrequency = async (frequency: string) => {
  try {
    const data = await apiClient.request<{ frequency: string }>('/api/settings/notifications', {
      method: 'POST',
      body: JSON.stringify({ frequency }),
    })
    
    // Update frequency value
    if (frequency === 'daily' || frequency === 'never') {
      notificationFrequency.value = data.frequency
    } else {
      // For custom frequency, keep 'custom' selected
      notificationFrequency.value = 'custom'
      customDays.value = Number(data.frequency)
    }
    
    // Show saved indicator
    isSaved.value = true
    setTimeout(() => {
      isSaved.value = false
    }, 2000) // Hide after 2 seconds
  } catch (error) {
    console.error('Failed to save notification settings:', error)
  }
}

// Watch for theme changes from outside
watch(() => currentTheme.value, (newTheme) => {
  selectedTheme.value = newTheme
  setTheme(newTheme)
})

const handleSoundsChange = () => {
  setSoundsEnabled(soundsEnabled.value)
}

const handleVibrationChange = () => {
  setVibrationEnabled(vibrationEnabled.value)
}

const handleThemeToggle = (event: Event) => {
  const target = event.target as HTMLInputElement
  const newTheme = target.checked ? 'dark' : 'light'
  selectedTheme.value = newTheme
  setTheme(newTheme)
  setThemeInTheme(newTheme)
}

const handleSoundThemeChange = () => {
  setSoundTheme(selectedSoundTheme.value)
}

const previewSounds = async () => {
  if (previewing.value) return
  previewing.value = true
  try {
    await previewTheme(selectedSoundTheme.value)
  } catch (error) {
    console.error('Failed to preview sounds:', error)
  } finally {
    previewing.value = false
  }
}

const handleLogout = () => {
  authLogout()
  router.push('/login')
}
</script>

<style scoped>
.settings {
  max-width: 800px;
  margin: 0 auto;
  padding: 10px;
}

.settings h1 {
  margin-bottom: 24px;
}

.settings h2 {
  margin-bottom: 20px;
  font-size: 20px;
  color: var(--text-primary);
}

.settings-group {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.setting-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 20px;
  padding: 16px 0;
  border-bottom: 1px solid var(--border-primary);
}

.setting-item:last-child {
  border-bottom: none;
}

.setting-info {
  flex: 1;
}

.setting-label {
  display: block;
  font-weight: 600;
  font-size: 16px;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.setting-description {
  font-size: 14px;
  color: var(--text-secondary);
  margin: 0;
}

.setting-control {
  flex-shrink: 0;
}

/* Toggle Switch */
.toggle-switch {
  position: relative;
  display: inline-block;
  width: 52px;
  height: 28px;
  cursor: pointer;
}

.toggle-switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.toggle-slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: var(--bg-secondary);
  border: 2px solid var(--border-primary);
  transition: 0.3s;
  border-radius: 28px;
}

.toggle-slider:before {
  position: absolute;
  content: "";
  height: 20px;
  width: 20px;
  left: 2px;
  bottom: 2px;
  background-color: var(--text-primary);
  transition: 0.3s;
  border-radius: 50%;
}

.toggle-switch input:checked + .toggle-slider {
  background-color: var(--color-primary);
  border-color: var(--color-primary);
}

.toggle-switch input:checked + .toggle-slider:before {
  transform: translateX(24px);
}

.toggle-switch input:focus + .toggle-slider {
  box-shadow: 0 0 1px var(--color-primary);
}

/* Theme Toggle Switch with Icons */
.theme-toggle-switch {
  position: relative;
  display: inline-block;
  width: 64px;
  height: 32px;
  cursor: pointer;
}

.theme-toggle-switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.theme-toggle-slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: var(--bg-secondary);
  border: 2px solid var(--border-primary);
  transition: 0.3s;
  border-radius: 32px;
  display: flex;
  align-items: center;
  padding: 2px;
}

.theme-toggle-slider:before {
  content: "";
  position: absolute;
  height: 24px;
  width: 24px;
  left: 4px;
  background-color: var(--text-primary);
  transition: 0.3s;
  border-radius: 50%;
  z-index: 2;
}

.theme-icon {
  position: absolute;
  font-size: 14px;
  transition: 0.3s;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1;
}

.theme-icon-sun {
  left: 8px;
  opacity: 1;
  color: var(--text-primary);
}

.theme-icon-moon {
  right: 8px;
  opacity: 0.3;
  color: var(--text-primary);
}

.theme-toggle-switch input:checked + .theme-toggle-slider {
  background-color: var(--color-primary);
  border-color: var(--color-primary);
}

.theme-toggle-switch input:checked + .theme-toggle-slider:before {
  transform: translateX(32px);
  background-color: white;
}

.theme-toggle-switch input:checked + .theme-toggle-slider .theme-icon-sun {
  opacity: 0.6;
  color: rgba(255, 255, 255, 0.9);
}

.theme-toggle-switch input:checked + .theme-toggle-slider .theme-icon-moon {
  opacity: 1;
  color: white;
}

.theme-toggle-switch input:not(:checked) + .theme-toggle-slider {
  background-color: var(--bg-secondary);
  border-color: var(--border-primary);
}

.theme-toggle-switch input:not(:checked) + .theme-toggle-slider:before {
  background-color: var(--text-primary);
}

.theme-toggle-switch input:not(:checked) + .theme-toggle-slider .theme-icon-sun {
  color: var(--text-primary);
  opacity: 1;
}

.theme-toggle-switch input:not(:checked) + .theme-toggle-slider .theme-icon-moon {
  color: var(--text-primary);
  opacity: 0.3;
}

.theme-toggle-switch input:focus + .theme-toggle-slider {
  box-shadow: 0 0 1px var(--color-primary);
}

/* Sound Theme Select */
.theme-select {
  padding: 8px 12px;
  border: 1px solid var(--border-primary);
  border-radius: 6px;
  background-color: var(--input-bg);
  color: var(--text-primary);
  font-size: 16px;
  cursor: pointer;
  min-width: 120px;
  transition: border-color 0.2s ease, background-color 0.3s ease;
}

.theme-select:focus {
  outline: none;
  border-color: var(--input-focus-border);
}

/* Sound Theme Control */
.sound-theme-control {
  display: flex;
  align-items: center;
  gap: 8px;
}

.preview-btn {
  background: transparent;
  border: 1px solid var(--border-primary);
  border-radius: 6px;
  padding: 8px 10px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  color: var(--text-primary);
  min-width: 36px;
  height: 36px;
}

.preview-btn:hover:not(:disabled) {
  background-color: var(--bg-hover);
  border-color: var(--border-secondary);
}

.preview-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.preview-btn .icon {
  font-size: 16px;
}

/* Logout Button */
.logout-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-primary);
  border-radius: 6px;
  color: var(--text-primary);
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.logout-btn:hover {
  background-color: var(--bg-hover);
  border-color: var(--border-secondary);
}

.logout-icon {
  font-size: 18px;
}

/* Mobile styles */
@media (max-width: 768px) {
  .settings {
    padding: 8px;
  }

  .setting-item {
    flex-direction: row;
    align-items: center;
    gap: 12px;
  }

  .setting-control {
    flex-shrink: 0;
  }
  
  .sound-theme-control {
    flex-direction: row;
    gap: 8px;
  }
}

/* Notification Control */

.notification-control {
  display: flex;
  flex-direction: row;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}

.custom-days-input {
  display: flex;
  gap: 8px;
  align-items: center;
}

/* Slide-fade animation */
.slide-fade-enter-active {
  transition: all 0.3s ease-out;
}

.slide-fade-leave-active {
  transition: all 0.2s ease-in;
}

.slide-fade-enter-from {
  transform: translateX(-10px);
  opacity: 0;
}

.slide-fade-leave-to {
  transform: translateX(-10px);
  opacity: 0;
}

.days-input {
  padding: 8px 12px;
  border: 1px solid var(--border-primary);
  border-radius: 6px;
  background-color: var(--input-bg);
  color: var(--text-primary);
  font-size: 16px;
  width: 80px;
  margin-bottom: 0;
  transition: border-color 0.2s ease, background-color 0.3s ease;
}

.days-input:focus {
  outline: none;
  border-color: var(--input-focus-border);
}

.days-label {
  font-size: 14px;
  color: var(--text-secondary);
  margin-left: 4px;
}

.saved-indicator {
  font-size: 14px;
  color: var(--color-primary);
  margin-right: 8px;
  font-weight: 500;
  opacity: 0;
  animation: fadeInOut 2s ease-in-out;
}

.delay-control {
  display: flex;
  flex-wrap: nowrap;
  align-items: center;
  gap: 10px;
  min-width: 140px;
}

.delay-slider {
  width: 100px;
  height: 8px;
  flex-shrink: 0;
  accent-color: var(--color-primary);
  cursor: pointer;
}

.delay-value {
  font-size: 14px;
  color: var(--text-secondary);
  min-width: 3ch;
}

/* Фиксированная ширина под "saved", чтобы блок не сдвигался при появлении надписи */
.saved-indicator-slot {
  display: inline-block;
  width: 72px;
  min-width: 72px;
  text-align: left;
}

.saved-indicator-slot .saved-indicator {
  font-size: 14px;
  color: var(--color-primary);
  font-weight: 500;
  animation: savedFadeInOut 2.5s ease-in-out;
}

@keyframes savedFadeInOut {
  0% { opacity: 0; }
  15% { opacity: 1; }
  75% { opacity: 1; }
  100% { opacity: 0; }
}

@keyframes fadeInOut {
  0% {
    opacity: 0;
  }
  20% {
    opacity: 1;
  }
  80% {
    opacity: 1;
  }
  100% {
    opacity: 0;
  }
}
</style>
