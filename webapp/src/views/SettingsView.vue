<template>
  <div class="prf-page lg-page">

    <!-- HEADER -->
    <div class="prf-header">
      <h1 class="prf-title">{{ t('settings.profileTitle') }}</h1>
      <button class="prf-theme-btn" type="button" @click="handleThemeToggleBtn">
        <LgIcon :name="selectedTheme === 'dark' ? 'sun' : 'moon'" :s="18" c="var(--text)" />
      </button>
    </div>

    <!-- USER CARD -->
    <div class="prf-user-card">
      <div class="prf-avatar">
        <LgLumi :size="50" pose="dreaming" />
      </div>
      <div class="prf-user-info">
        <div class="prf-user-name">{{ displayName }}</div>
        <div class="prf-user-sub">{{ targetLangDisplay }}</div>
        <div class="prf-chips">
          <span v-if="confirmedLevel" class="prf-chip prf-chip--active">{{ t('settings.profileLevel', { level: confirmedLevel }) }}</span>
          <span v-if="currentCourse?.title" class="prf-chip">{{ currentCourse.title }}</span>
          <span v-if="streakDays > 0" class="prf-chip">{{ t('lg.streakDays', { n: streakDays }) }}</span>
        </div>
      </div>
    </div>

    <!-- QUICK SETTINGS ROWS (theme, language, course) -->
    <div class="prf-section-card">
      <!-- Theme -->
      <div class="prf-row">
        <span class="prf-row-label">{{ t('settings.theme') }}</span>
        <label class="prf-toggle-switch">
          <input type="checkbox" :checked="selectedTheme === 'dark'" @change="handleThemeToggle" />
          <span class="prf-toggle-slider">
            <Icon name="sun" class="prf-toggle-icon prf-toggle-icon--sun" />
            <Icon name="moon" class="prf-toggle-icon prf-toggle-icon--moon" />
          </span>
        </label>
      </div>
      <!-- Language -->
      <div class="prf-row">
        <span class="prf-row-label">{{ t('common.language') }}</span>
        <select
          :value="currentLocale"
          class="prf-select"
          @change="(e) => setLocale((e.target as HTMLSelectElement).value as any)"
        >
          <option v-for="l in availableLocales" :key="l.code" :value="l.code">{{ l.label }}</option>
        </select>
      </div>
      <!-- Course -->
      <div v-if="courses.length > 1" class="prf-row">
        <span class="prf-row-label">{{ t('city.course') }}</span>
        <select
          :value="currentCourseCode"
          class="prf-select"
          @change="(e) => selectCourse((e.target as HTMLSelectElement).value)"
        >
          <option v-for="c in courses" :key="c.code" :value="c.code">{{ c.title }}</option>
        </select>
      </div>
    </div>

    <!-- TRAINING SECTION -->
    <div class="prf-section-title">{{ t('settings.training') }}</div>
    <div class="prf-section-card">
      <div class="prf-row">
        <div class="prf-row-info">
          <span class="prf-row-label">{{ t('settings.vibration') }}</span>
          <span class="prf-row-sub">{{ t('settings.vibrationDescription') }}</span>
        </div>
        <label class="toggle-switch">
          <input type="checkbox" v-model="vibrationEnabled" @change="handleVibrationChange" />
          <span class="toggle-slider"></span>
        </label>
      </div>
      <div class="prf-row">
        <div class="prf-row-info">
          <span class="prf-row-label">{{ t('settings.sounds') }}</span>
          <span class="prf-row-sub">{{ t('settings.soundsDescription') }}</span>
        </div>
        <label class="toggle-switch">
          <input type="checkbox" v-model="soundsEnabled" @change="handleSoundsChange" />
          <span class="toggle-slider"></span>
        </label>
      </div>
      <div v-if="soundsEnabled" class="prf-row prf-row--col">
        <div class="prf-row-head">
          <span class="prf-row-label">{{ t('settings.soundTheme') }}</span>
        </div>
        <div class="prf-row-ctrl">
          <select v-model="selectedSoundTheme" @change="handleSoundThemeChange" class="prf-select">
            <option v-for="th in soundThemes" :key="th.id" :value="th.id">{{ th.name }}</option>
          </select>
          <button @click="previewSounds" class="prf-preview-btn" :disabled="previewing">
            <Icon name="play" />
          </button>
        </div>
      </div>
      <div class="prf-row prf-row--col">
        <div class="prf-row-head">
          <div class="prf-row-info">
            <span class="prf-row-label">{{ t('settings.optionsDelay') }}</span>
            <span class="prf-row-sub">{{ t('settings.optionsDelayDescription') }}</span>
          </div>
        </div>
        <div class="prf-delay-ctrl">
          <span v-if="trainingDelaysSavedAt === 'options'" class="prf-saved">{{ t('common.saved') }}</span>
          <input v-model.number="optionsDelaySeconds" type="range" min="0" max="10" class="prf-slider" @change="handleOptionsDelayChange" />
          <span class="prf-delay-val">{{ optionsDelaySeconds }} {{ t('settings.seconds') }}</span>
        </div>
      </div>
      <div class="prf-row prf-row--col">
        <div class="prf-row-head">
          <div class="prf-row-info">
            <span class="prf-row-label">{{ t('settings.wrongAnswerDelay') }}</span>
            <span class="prf-row-sub">{{ t('settings.wrongAnswerDelayDescription') }}</span>
          </div>
        </div>
        <div class="prf-delay-ctrl">
          <span v-if="trainingDelaysSavedAt === 'wrong'" class="prf-saved">{{ t('common.saved') }}</span>
          <input v-model.number="wrongAnswerDelaySeconds" type="range" min="0" max="10" class="prf-slider" @change="handleWrongAnswerDelayChange" />
          <span class="prf-delay-val">{{ wrongAnswerDelaySeconds }} {{ t('settings.seconds') }}</span>
        </div>
      </div>
      <div class="prf-row">
        <div class="prf-row-info">
          <span class="prf-row-label">{{ t('settings.spellModeEnabled') }}</span>
          <span class="prf-row-sub">{{ t('settings.spellModeEnabledDescription', { targetLang: targetLangDisplay }) }}</span>
        </div>
        <div style="display:flex;align-items:center;gap:8px">
          <span v-if="trainingDelaysSavedAt === 'spell'" class="prf-saved">{{ t('common.saved') }}</span>
          <label class="toggle-switch">
            <input v-model="spellModeEnabled" type="checkbox" @change="handleSpellSettingsChange" />
            <span class="toggle-slider"></span>
          </label>
        </div>
      </div>
      <div v-if="spellModeEnabled" class="prf-row prf-row--col">
        <div class="prf-row-head">
          <span class="prf-row-label">{{ t('settings.spellMasteringThreshold') }}</span>
        </div>
        <div class="prf-delay-ctrl">
          <input v-model.number="spellMasteringThreshold" type="range" min="0" max="100" class="prf-slider" @change="handleSpellSettingsChange" />
          <span class="prf-delay-val">{{ spellMasteringThreshold }}</span>
        </div>
      </div>
      <div class="prf-row">
        <div class="prf-row-info">
          <span class="prf-row-label">{{ t('settings.typeModeEnabled') }}</span>
          <span class="prf-row-sub">{{ t('settings.typeModeEnabledDescription', { targetLang: targetLangDisplay }) }}</span>
        </div>
        <div style="display:flex;align-items:center;gap:8px">
          <span v-if="trainingDelaysSavedAt === 'type'" class="prf-saved">{{ t('common.saved') }}</span>
          <label class="toggle-switch">
            <input v-model="typeModeEnabled" type="checkbox" @change="handleTypeSettingsChange" />
            <span class="toggle-slider"></span>
          </label>
        </div>
      </div>
      <div v-if="typeModeEnabled" class="prf-row prf-row--col">
        <div class="prf-row-head">
          <span class="prf-row-label">{{ t('settings.typeMasteringThreshold') }}</span>
        </div>
        <div class="prf-delay-ctrl">
          <input v-model.number="typeMasteringThreshold" type="range" min="0" max="100" class="prf-slider" @change="handleTypeSettingsChange" />
          <span class="prf-delay-val">{{ typeMasteringThreshold }}</span>
        </div>
      </div>
      <div class="prf-row">
        <div class="prf-row-info">
          <span class="prf-row-label">{{ t('settings.hideMorphInTraining') }}</span>
          <span class="prf-row-sub">{{ t('settings.hideMorphInTrainingDescription') }}</span>
        </div>
        <div style="display:flex;align-items:center;gap:8px">
          <span v-if="trainingDelaysSavedAt === 'morph'" class="prf-saved">{{ t('common.saved') }}</span>
          <label class="toggle-switch">
            <input v-model="hideMorphInTraining" type="checkbox" @change="handleMorphVisibilityChange" />
            <span class="toggle-slider"></span>
          </label>
        </div>
      </div>
      <div class="prf-row">
        <div class="prf-row-info">
          <span class="prf-row-label">{{ t('settings.autoplayPronunciation') }}</span>
          <span class="prf-row-sub">{{ t('settings.autoplayPronunciationDescription') }}</span>
        </div>
        <div style="display:flex;align-items:center;gap:8px">
          <span v-if="trainingDelaysSavedAt === 'autoplay'" class="prf-saved">{{ t('common.saved') }}</span>
          <label class="toggle-switch">
            <input v-model="autoplayPronunciation" type="checkbox" @change="handleAutoplayPronunciationChange" />
            <span class="toggle-slider"></span>
          </label>
        </div>
      </div>
    </div>

    <!-- NOTIFICATIONS -->
    <div class="prf-section-title">{{ t('settings.notifications') }}</div>
    <div class="prf-section-card">
      <div class="prf-row prf-row--col">
        <div class="prf-row-head">
          <div class="prf-row-info">
            <span class="prf-row-label">{{ t('settings.notificationFrequency') }}</span>
            <span class="prf-row-sub">{{ t('settings.notificationFrequencyDescription') }}</span>
          </div>
        </div>
        <div class="prf-notify-ctrl">
          <span v-if="isSaved" class="prf-saved">{{ t('common.saved') }}</span>
          <select v-model="notificationFrequency" @change="handleNotificationFrequencyChange" class="prf-select">
            <option value="daily">{{ t('settings.daily') }}</option>
            <option value="never">{{ t('settings.never') }}</option>
            <option value="custom">{{ t('settings.custom') }}</option>
          </select>
          <transition name="slide-fade">
            <div v-if="notificationFrequency === 'custom'" class="prf-custom-days">
              <input type="number" v-model.number="customDays" min="1" max="30" @input="handleCustomDaysChange" class="prf-days-input" :placeholder="t('settings.days')" />
              <span class="prf-days-label">{{ t('settings.days') }}</span>
            </div>
          </transition>
        </div>
      </div>
    </div>

    <!-- ACCOUNT -->
    <template v-if="!isTelegramMiniApp">
      <div class="prf-section-title">{{ t('settings.account') }}</div>
      <div class="prf-section-card">
        <button class="prf-logout-btn" @click="handleLogout">
          <Icon name="logout" class="prf-logout-icon" />
          <span>{{ t('settings.logout') }}</span>
        </button>
      </div>
    </template>

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
import { useLearningConfig, type SpanishVerbScopeLadderStep } from '../composables/useLearningConfig'
import { apiClient } from '../api/client'
import { useLocale } from '../composables/useLocale'
import { AVAILABLE_LOCALES } from '../i18n'
import { useCourse } from '../composables/useCourse'
import Icon from '../components/Icon.vue'
import LgIcon from '../components/linglow/LgIcon.vue'
import LgLumi from '../components/linglow/LgLumi.vue'
import { useStats } from '../composables/useStats'

const { t, locale } = useI18n()
const { currentLocale, setLocale } = useLocale()
const availableLocales = AVAILABLE_LOCALES
const { courses, currentCourse, currentCourseCode, selectCourse } = useCourse()
const { streakDays, ensureStatsLoaded } = useStats()
ensureStatsLoaded()

const meUsername = ref('')
const confirmedLevel = ref('')
const displayName = computed(() => meUsername.value || 'Linglow')

onMounted(async () => {
  try {
    const me: any = await apiClient.request('/api/me')
    meUsername.value = me?.telegram_username || ''
  } catch { /* ignore */ }
  try {
    const dash: any = await apiClient.request('/api/dashboard')
    confirmedLevel.value = dash?.grammar_stats?.confirmed_level || ''
  } catch { /* ignore */ }
})

const router = useRouter()
const { settings, setSoundsEnabled, setVibrationEnabled, setTheme, setSoundTheme, setHideMorphInTraining, setAutoplayPronunciation } = useSettings()
const { theme: currentTheme, setTheme: setThemeInTheme } = useTheme()
const { getThemes, previewTheme } = useAudio()
const { logout: authLogout } = useAuth()
const { learning, targetLangDisplay, ensureLearningLoaded } = useLearningConfig()

const soundsEnabled = ref(true)
const vibrationEnabled = ref(true)
const selectedTheme = ref<'light' | 'dark'>('light')
const selectedSoundTheme = ref('tick')
const soundThemes = ref(getThemes())
const previewing = ref(false)
const isTelegramMiniApp = ref(!!(window as any).Telegram?.WebApp)

const notificationFrequency = ref('daily')
const customDays = ref(3)
const isSaved = ref(false)
const optionsDelaySeconds = ref(5)
const wrongAnswerDelaySeconds = ref(5)
/** 'options' | 'wrong' - which delay control just saved (to show "saved" there) */
const trainingDelaysSavedAt = ref<'options' | 'wrong' | 'spell' | 'type' | 'morph' | 'autoplay' | null>(null)
const spellModeEnabled = ref(true)
const spellMasteringThreshold = ref(50)
const typeModeEnabled = ref(true)
const typeMasteringThreshold = ref(70)
const hideMorphInTraining = ref(false)
const autoplayPronunciation = ref(true)
const verbFormsProgressionIndex = ref(0)
const verbProgressionSaved = ref(false)
let trainingDelaysSavedTimeout: ReturnType<typeof setTimeout> | null = null
let verbProgressionSavedTimeout: ReturnType<typeof setTimeout> | null = null

const showVerbFormProgression = computed(
  () =>
    learning.value?.target_lang === 'es' &&
    learning.value?.spanish_verb_forms_enabled === true &&
    (learning.value?.spanish_verb_scope_ladder?.length ?? 0) > 0
)

const verbLadder = computed(() => learning.value?.spanish_verb_scope_ladder ?? [])

const verbProgressionOptionLabel = (step: { label_ru: string; label_en: string }) => {
  const label = locale.value === 'ru' ? step.label_ru : step.label_en
  return t('settings.verbFormProgressionThrough', { label })
}

onMounted(async () => {
  await ensureLearningLoaded()
  // Load current settings
  soundsEnabled.value = settings.value.soundsEnabled
  vibrationEnabled.value = settings.value.vibrationEnabled
  selectedTheme.value = currentTheme.value
  selectedSoundTheme.value = settings.value.soundTheme || 'tick'
  autoplayPronunciation.value = settings.value.autoplayPronunciation
  
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
    autoplay_pronunciation?: boolean
    verb_forms_progression_index?: number
  }
  learning?: Record<string, unknown>
}

const mergeLearningFromSettings = (data: SettingsResponse) => {
  const patch = data.learning
  if (!patch || typeof patch !== 'object') return
  const base = learning.value
  if (!base) return
  learning.value = {
    ...base,
    spanish_verb_forms_enabled:
      typeof patch.spanish_verb_forms_enabled === 'boolean'
        ? patch.spanish_verb_forms_enabled
        : base.spanish_verb_forms_enabled,
    spanish_verb_scope_ladder: Array.isArray(patch.spanish_verb_scope_ladder)
      ? (patch.spanish_verb_scope_ladder as SpanishVerbScopeLadderStep[])
      : base.spanish_verb_scope_ladder,
  }
}

const loadTrainingDelaysSettings = async () => {
  try {
    const data = await apiClient.request<SettingsResponse>('/api/settings')
    mergeLearningFromSettings(data)
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
    if (s?.autoplay_pronunciation !== undefined) {
      autoplayPronunciation.value = s.autoplay_pronunciation
      setAutoplayPronunciation(autoplayPronunciation.value)
    } else {
      autoplayPronunciation.value = true
      setAutoplayPronunciation(true)
    }
    if (s?.verb_forms_progression_index !== undefined && typeof s.verb_forms_progression_index === 'number') {
      verbFormsProgressionIndex.value = Math.max(
        0,
        Math.min((learning.value?.spanish_verb_scope_ladder?.length ?? 1) - 1, s.verb_forms_progression_index)
      )
    }
  } catch (error) {
    console.error('Failed to load training delay settings:', error)
  }
}

const handleVerbFormProgressionChange = async () => {
  const ladderLen = learning.value?.spanish_verb_scope_ladder?.length ?? 0
  if (ladderLen === 0) return
  let idx = verbFormsProgressionIndex.value
  if (idx < 0) idx = 0
  if (idx >= ladderLen) idx = ladderLen - 1
  verbFormsProgressionIndex.value = idx
  try {
    const data = await apiClient.request<SettingsResponse>('/api/settings/training', {
      method: 'POST',
      body: JSON.stringify({ verb_forms_progression_index: idx }),
    })
    mergeLearningFromSettings(data)
    if (data.settings?.verb_forms_progression_index !== undefined) {
      verbFormsProgressionIndex.value = data.settings.verb_forms_progression_index
    }
    if (verbProgressionSavedTimeout) {
      clearTimeout(verbProgressionSavedTimeout)
      verbProgressionSavedTimeout = null
    }
    verbProgressionSaved.value = true
    verbProgressionSavedTimeout = setTimeout(() => {
      verbProgressionSaved.value = false
      verbProgressionSavedTimeout = null
    }, 2500)
  } catch (error) {
    console.error('Failed to save verb form progression:', error)
  }
}

const saveTrainingDelays = async (showSavedAt: 'options' | 'wrong' | 'spell' | 'type' | 'morph' | 'autoplay') => {
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
        hide_morph_in_training: hideMorphInTraining.value,
        autoplay_pronunciation: autoplayPronunciation.value
      })
    })
    setHideMorphInTraining(hideMorphInTraining.value)
    setAutoplayPronunciation(autoplayPronunciation.value)
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
const handleAutoplayPronunciationChange = () => saveTrainingDelays('autoplay')

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

const handleThemeToggleBtn = () => {
  const newTheme = selectedTheme.value === 'dark' ? 'light' : 'dark'
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
/* ── Profile page layout ── */
.prf-page { padding-bottom: 32px; }

.prf-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 16px 12px;
}
.prf-title {
  font-family: 'Lora', serif;
  font-size: 26px;
  font-weight: 600;
  color: var(--text);
  margin: 0;
}
.prf-theme-btn {
  width: 38px; height: 38px;
  border-radius: 50%;
  border: 1px solid var(--border);
  background: var(--card-bg);
  display: flex; align-items: center; justify-content: center;
  cursor: pointer;
}

/* USER CARD */
.prf-user-card {
  display: flex; align-items: center; gap: 16px;
  padding: 0 16px 16px;
}
.prf-avatar {
  width: 72px; height: 72px; border-radius: 50%; flex-shrink: 0;
  background: var(--chip-bg);
  border: 2px solid var(--salvia);
  display: flex; align-items: center; justify-content: center;
}
.prf-user-name { font-family: 'Lora', serif; font-size: 20px; color: var(--text); }
.prf-user-sub { font-family: 'Inter', sans-serif; font-size: 13px; color: var(--subtext); margin: 2px 0; }
.prf-chips { display: flex; gap: 8px; margin-top: 4px; }
.prf-chip {
  padding: 5px 12px; border-radius: 20px;
  background: var(--chip-bg); border: 1px solid var(--border);
  font-family: 'Inter', sans-serif; font-size: 12px; font-weight: 500; color: var(--text);
}
.prf-chip--active { background: var(--salvia); color: #fff; border-color: transparent; }

/* SECTION CARDS */
.prf-section-title {
  font-family: 'Lora', serif; font-size: 17px; font-weight: 600; color: var(--text);
  margin: 14px 16px 8px;
}
.prf-section-card {
  margin: 0 16px 14px;
  background: var(--card-bg); border: 1px solid var(--border);
  border-radius: 18px; overflow: hidden;
  box-shadow: var(--shadow-soft);
}

/* ROWS */
.prf-row {
  display: flex; align-items: center; gap: 12px;
  padding: 14px 16px;
  border-bottom: 1px solid var(--border);
}
.prf-row:last-child { border-bottom: none; }
.prf-row--col { flex-direction: column; align-items: flex-start; }
.prf-row-head { display: flex; align-items: center; gap: 12px; width: 100%; }
.prf-row-icon { font-size: 18px; flex-shrink: 0; width: 22px; text-align: center; }
.prf-row-label { font-family: 'Inter', sans-serif; font-size: 14px; font-weight: 600; color: var(--text); flex: 1; }
.prf-row-info { display: flex; flex-direction: column; flex: 1; }
.prf-row-sub { font-family: 'Inter', sans-serif; font-size: 12px; color: var(--subtext); margin-top: 2px; }

/* Select */
.prf-select {
  padding: 7px 10px;
  border: 1px solid var(--border);
  border-radius: 10px;
  background: var(--input-bg);
  color: var(--text);
  font-family: inherit;
  font-size: 14px;
  cursor: pointer;
}
.prf-select:focus { outline: none; border-color: var(--salvia); }

/* Theme toggle */
.prf-toggle-switch {
  position: relative; display: inline-block;
  width: 64px; height: 32px; cursor: pointer; flex-shrink: 0;
}
.prf-toggle-switch input { opacity: 0; width: 0; height: 0; }
.prf-toggle-slider {
  position: absolute; cursor: pointer;
  top: 0; left: 0; right: 0; bottom: 0;
  background: var(--surface-2); border: 2px solid var(--border);
  border-radius: 32px; display: flex; align-items: center; padding: 2px;
  transition: 0.3s;
}
.prf-toggle-slider:before {
  content: ""; position: absolute;
  height: 24px; width: 24px; left: 4px;
  background: var(--text); border-radius: 50%; z-index: 2;
  transition: 0.3s;
}
.prf-toggle-icon {
  position: absolute; font-size: 14px;
  display: flex; align-items: center; justify-content: center; z-index: 1;
}
.prf-toggle-icon--sun { left: 8px; opacity: 1; color: var(--text); }
.prf-toggle-icon--moon { right: 8px; opacity: 0.3; color: var(--text); }
.prf-toggle-switch input:checked + .prf-toggle-slider { background: var(--salvia); border-color: var(--salvia); }
.prf-toggle-switch input:checked + .prf-toggle-slider:before { transform: translateX(32px); background: white; }

/* Preview button */
.prf-preview-btn {
  border: 1px solid var(--border); border-radius: 10px;
  padding: 7px 10px; cursor: pointer; background: transparent;
  display: flex; align-items: center; color: var(--text);
}
.prf-preview-btn:disabled { opacity: 0.5; cursor: not-allowed; }
.prf-row-ctrl { display: flex; align-items: center; gap: 8px; padding-left: 34px; }
.prf-notify-ctrl { display: flex; flex-direction: column; gap: 8px; padding-left: 34px; width: 100%; }
.prf-custom-days { display: flex; gap: 8px; align-items: center; }
.prf-days-input {
  padding: 7px 10px; border: 1px solid var(--border); border-radius: 10px;
  background: var(--input-bg); color: var(--text); font-size: 14px; width: 72px; margin-bottom: 0;
}
.prf-days-label { font-size: 13px; color: var(--subtext); }

/* Delays */
.prf-delay-ctrl { display: flex; align-items: center; gap: 10px; padding-left: 34px; }
.prf-slider { width: 110px; height: 8px; accent-color: var(--salvia); cursor: pointer; flex-shrink: 0; }
.prf-delay-val { font-size: 13px; color: var(--subtext); min-width: 3ch; }
.prf-saved { font-size: 12px; color: var(--salvia); font-weight: 500; animation: savedFadeInOut 2.5s ease-in-out; }

/* Logout */
.prf-logout-btn {
  display: flex; align-items: center; gap: 8px;
  padding: 14px 16px; width: 100%; background: none; border: none;
  cursor: pointer; color: var(--error);
  font-family: 'Inter', sans-serif; font-size: 15px; font-weight: 600;
  text-align: left;
}
.prf-logout-icon { font-size: 18px; }

/* Animations */
@keyframes savedFadeInOut {
  0% { opacity: 0; } 15% { opacity: 1; } 75% { opacity: 1; } 100% { opacity: 0; }
}

/* ── Keep existing toggle-switch/slider for prf-row items ── */
.settings {
  max-width: 800px;
  margin: 0 auto;
}

.settings .card {
  background: var(--card-bg);
  border: 1px solid var(--border);
  border-radius: 18px;
  padding: 16px;
  box-shadow: var(--shadow-soft);
  margin-bottom: 14px;
}

.settings h2 {
  margin-bottom: 12px;
  font-family: 'Lora', serif;
  font-size: 18px;
  font-weight: 600;
  color: var(--text);
}

.lg-theme-btn {
  width: 38px;
  height: 38px;
  border-radius: 50%;
  border: 1px solid var(--border);
  background: var(--card-bg);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.settings-hero {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 0 0 16px;
}

.settings-hero-avatar {
  width: 72px;
  height: 72px;
  border-radius: 50%;
  background: var(--chip-bg);
  border: 2px solid var(--salvia);
  display: flex;
  align-items: center;
  justify-content: center;
}

.settings-hero-name {
  font-family: 'Lora', serif;
  font-size: 20px;
  color: var(--text);
}

.settings-hero-sub {
  font-size: 13px;
  color: var(--subtext);
  margin-top: 2px;
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

.verb-scope-item {
  align-items: flex-start;
}

.verb-scope-select {
  min-width: 220px;
  max-width: 100%;
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
