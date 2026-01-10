<template>
  <div class="settings">
    <h1>Settings</h1>
    
    <div class="card">
      <h2>Appearance</h2>
      <div class="settings-group">
        <div class="setting-item">
          <div class="setting-info">
            <label class="setting-label">Theme</label>
            <p class="setting-description">Choose between light and dark theme</p>
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
      <h2>Training</h2>
      <div class="settings-group">
        <div class="setting-item">
          <div class="setting-info">
            <label class="setting-label">Vibration</label>
            <p class="setting-description">Enable haptic feedback on mobile devices</p>
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
            <label class="setting-label">Sounds</label>
            <p class="setting-description">Play sounds for correct and incorrect answers</p>
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
            <label class="setting-label">Sound Theme</label>
            <p class="setting-description">Choose sound theme for training</p>
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
                title="Preview sounds"
              >
                <Icon name="play" />
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useSettings } from '../composables/useSettings'
import { useTheme } from '../composables/useTheme'
import { useAudio } from '../composables/useAudio'
import Icon from '../components/Icon.vue'

const { settings, setSoundsEnabled, setVibrationEnabled, setTheme, setSoundTheme } = useSettings()
const { theme: currentTheme, setTheme: setThemeInTheme } = useTheme()
const { getThemes, previewTheme } = useAudio()

const soundsEnabled = ref(true)
const vibrationEnabled = ref(true)
const selectedTheme = ref<'light' | 'dark'>('light')
const selectedSoundTheme = ref('tick')
const soundThemes = ref(getThemes())
const previewing = ref(false)

onMounted(() => {
  // Load current settings
  soundsEnabled.value = settings.value.soundsEnabled
  vibrationEnabled.value = settings.value.vibrationEnabled
  selectedTheme.value = currentTheme.value
  selectedSoundTheme.value = settings.value.soundTheme || 'tick'
})

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


/* Mobile styles */
@media (max-width: 768px) {
  .settings {
    padding: 8px;
  }

  .setting-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .setting-control {
    width: 100%;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
