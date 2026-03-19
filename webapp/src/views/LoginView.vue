<template>
  <div class="login-container">
    <div class="card" style="max-width: 400px; margin: 50px auto;">
      <h1>{{ t('auth.loginTitle') }}</h1>

      <!-- Registration hint when login form is shown (no Telegram WebApp or auth failed) -->
      <div v-if="!isCheckingTelegramAuth" class="register-hint">
        <p>{{ t('auth.registerInBotMessage') }}</p>
        <a href="https://t.me/positroid_english_bot" target="_blank" rel="noopener noreferrer" class="bot-link">{{ t('auth.registerInBotLinkText') }}</a>
      </div>
      
      <!-- Show loading indicator while checking Telegram auth -->
      <div v-if="isCheckingTelegramAuth" class="login-loading">
        <div class="spinner"></div>
        <p>{{ t('auth.telegramAuthInProgress') }}</p>
      </div>
      
      <!-- Show login form only after Telegram auth check is complete -->
      <template v-else>
        <div v-if="step === 'username'" class="login-step">
          <p>{{ t('auth.enterUsername') }}</p>
          <div class="info-box">
            <svg class="info-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/>
              <line x1="12" y1="16" x2="12" y2="12"/>
              <line x1="12" y1="8" x2="12.01" y2="8"/>
            </svg>
            <div class="info-content">
              <p class="info-text" v-html="t('auth.usernameHint')"></p>
            </div>
          </div>
          <input
            ref="usernameInput"
            v-model="username"
            type="text"
            :placeholder="t('auth.usernamePlaceholder')"
            @blur="username = username.trim()"
            @keyup.enter="requestOTP"
          />
          <button @click="requestOTP" class="btn btn-primary" :disabled="loading">
            {{ loading ? t('auth.sending') : t('auth.sendOTP') }}
          </button>
          <p v-if="error" class="error">{{ error }}</p>
        </div>

        <div v-if="step === 'otp'" class="login-step">
          <p>{{ t('auth.enterOTP') }}</p>
          <div class="otp-input-container">
            <input
              v-for="(digit, index) in otpDigits"
              :key="index"
              :ref="(el) => { if (el) otpInputs[index] = el as HTMLInputElement }"
              v-model="otpDigits[index]"
              type="text"
              inputmode="numeric"
              pattern="[0-9]"
              maxlength="1"
              class="otp-digit"
              @input="handleOTPInput(index, $event)"
              @keydown="handleOTPKeydown(index, $event)"
              @paste="handleOTPPaste($event)"
              @focus="handleOTPFocus(index)"
            />
          </div>
          <button @click="verifyOTP" class="btn btn-primary" :disabled="loading || !isOTPComplete">
            {{ loading ? t('auth.verifying') : t('auth.verify') }}
          </button>
          <button @click="step = 'username'" class="btn btn-secondary">{{ t('common.back') }}</button>
          <p v-if="error" class="error">{{ error }}</p>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuth } from '../composables/useAuth'
import { apiClient } from '../api/client'

const { t } = useI18n()

const router = useRouter()
const { login, tryTelegramAuth } = useAuth()

const step = ref<'username' | 'otp'>('username')
const username = ref('')
const otpDigits = ref<string[]>(['', '', '', '', '', ''])
const otpInputs = ref<(HTMLInputElement | null)[]>([])
const userId = ref('')
const loading = ref(false)
const error = ref('')
const isCheckingTelegramAuth = ref(false)

const usernameInput = ref<HTMLInputElement | null>(null)

const isOTPComplete = computed(() => {
  return otpDigits.value.every(digit => digit !== '')
})

onMounted(async () => {
  // Check if Telegram WebApp is available and try to authenticate
  const tg = (window as any).Telegram?.WebApp
  if (tg && tg.initData) {
    isCheckingTelegramAuth.value = true
    try {
      const success = await tryTelegramAuth()
      if (success) {
        router.push('/dashboard')
        return
      }
    } catch (err: any) {
      // Silent fail - show OTP login form
      error.value = t('auth.telegramAuthFailed')
    } finally {
      isCheckingTelegramAuth.value = false
    }
  } else {
    // No Telegram WebApp available, show login form immediately
    isCheckingTelegramAuth.value = false
  }
  
  // Focus on username input when component is mounted
  if (step.value === 'username' && !isCheckingTelegramAuth.value) {
    await nextTick()
    usernameInput.value?.focus()
  }
})

// Watch for step changes to focus on the appropriate input
watch(step, async (newStep) => {
  if (isCheckingTelegramAuth.value) {
    return // Don't focus if we're still checking Telegram auth
  }
  await nextTick()
  if (newStep === 'username') {
    usernameInput.value?.focus()
  } else if (newStep === 'otp') {
    // Reset OTP digits when switching to OTP step
    otpDigits.value = ['', '', '', '', '', '']
    otpInputs.value = []
    await nextTick()
    otpInputs.value[0]?.focus()
  }
})

// Watch for Telegram auth check completion to focus on input
watch(isCheckingTelegramAuth, async (isChecking) => {
  if (!isChecking && step.value === 'username') {
    await nextTick()
    usernameInput.value?.focus()
  }
})

const requestOTP = async () => {
  const cleaned = username.value.trim()
  if (!cleaned) {
    error.value = t('auth.usernameRequired')
    return
  }
  username.value = cleaned

  loading.value = true
  error.value = ''

  try {
    const response = await apiClient.requestOTP(cleaned)
    userId.value = response.user_id.toString()
    step.value = 'otp'
    // Focus will be set automatically by watch on step change
  } catch (err: any) {
    error.value = err.message || t('auth.otpSendFailed')
  } finally {
    loading.value = false
  }
}

const handleOTPInput = (index: number, event: Event) => {
  const target = event.target as HTMLInputElement
  const value = target.value
  
  // Only allow digits
  if (value && !/^\d$/.test(value)) {
    otpDigits.value[index] = ''
    return
  }
  
  otpDigits.value[index] = value
  
  // Move to next field if digit entered
  if (value && index < 5) {
    nextTick(() => {
      otpInputs.value[index + 1]?.focus()
    })
  }
  
  // Auto-verify when last digit is entered
  if (value && index === 5 && isOTPComplete.value) {
    nextTick(() => {
      verifyOTP()
    })
  }
}

const handleOTPKeydown = (index: number, event: KeyboardEvent) => {
  if (event.key === 'Backspace') {
    if (otpDigits.value[index]) {
      // If current field has value, clear it
      otpDigits.value[index] = ''
    } else if (index > 0) {
      // If current field is empty, go to previous and clear it
      otpDigits.value[index - 1] = ''
      nextTick(() => {
        otpInputs.value[index - 1]?.focus()
      })
    }
    event.preventDefault()
  } else if (event.key === 'ArrowLeft' && index > 0) {
    nextTick(() => {
      otpInputs.value[index - 1]?.focus()
    })
  } else if (event.key === 'ArrowRight' && index < 5) {
    nextTick(() => {
      otpInputs.value[index + 1]?.focus()
    })
  }
}

const handleOTPPaste = (event: ClipboardEvent) => {
  event.preventDefault()
  const pastedData = event.clipboardData?.getData('text') || ''
  const digits = pastedData.replace(/\D/g, '').slice(0, 6).split('')
  
  // Fill digits from current position
  const startIndex = otpInputs.value.findIndex(input => input === event.target)
  if (startIndex === -1) return
  
  for (let i = 0; i < digits.length && startIndex + i < 6; i++) {
    otpDigits.value[startIndex + i] = digits[i]
  }
  
  // Focus on the next empty field or last field
  const nextIndex = Math.min(startIndex + digits.length, 5)
  nextTick(() => {
    otpInputs.value[nextIndex]?.focus()
    if (isOTPComplete.value) {
      verifyOTP()
    }
  })
}

const handleOTPFocus = (index: number) => {
  // When focusing on a field, if it's in the middle and has value,
  // we'll allow overwriting (the input handler will handle it)
  // Just ensure we're at the right position
  if (otpDigits.value[index]) {
    // If field has value, select it for easy overwrite
    nextTick(() => {
      otpInputs.value[index]?.select()
    })
  }
}

const verifyOTP = async () => {
  const code = otpDigits.value.join('')
  if (!code || code.length !== 6) {
    error.value = t('auth.otpIncomplete')
    return
  }

  loading.value = true
  error.value = ''

  try {
    const response = await apiClient.verifyOTP(userId.value, code)
    login(response.access_token, response.refresh_token)
    router.push('/dashboard')
  } catch (err: any) {
    error.value = err.message || t('auth.otpInvalid')
    // Clear OTP on error
    otpDigits.value = ['', '', '', '', '', '']
    nextTick(() => {
      otpInputs.value[0]?.focus()
    })
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 20px;
}

.login-step {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

h1 {
  margin-bottom: 20px;
  text-align: center;
}

.info-box {
  display: flex;
  gap: 12px;
  padding: 12px 16px;
  margin-bottom: 16px;
  background-color: var(--info-bg, #e3f2fd);
  border: 1px solid var(--info-border, #90caf9);
  border-radius: 8px;
  align-items: flex-start;
}

.info-icon {
  width: 20px;
  height: 20px;
  min-width: 20px;
  color: var(--info-icon, #1976d2);
  margin-top: 2px;
  flex-shrink: 0;
}

.info-content {
  flex: 1;
}

.info-text {
  margin: 0;
  font-size: 0.875em;
  line-height: 1.5;
  color: var(--info-text, #1565c0);
}

.info-text code {
  background-color: var(--info-code-bg, rgba(25, 118, 210, 0.1));
  color: var(--info-code-text, #0d47a1);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'Courier New', monospace;
  font-size: 0.95em;
  font-weight: 600;
}

.login-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 20px;
  padding: 40px 20px;
  min-height: 200px;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 4px solid var(--border-primary, #e0e0e0);
  border-top-color: var(--color-primary, #1976d2);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.login-loading p {
  margin: 0;
  color: var(--color-primary);
  font-size: 1em;
}

.register-hint {
  margin-bottom: 20px;
  padding: 12px 16px;
  background-color: var(--info-bg, #e3f2fd);
  border: 1px solid var(--info-border, #90caf9);
  border-radius: 8px;
  text-align: center;
}

.register-hint p {
  margin: 0 0 8px 0;
  font-size: 0.9em;
  color: var(--info-text, #1565c0);
}

.register-hint .bot-link {
  display: inline-block;
  font-weight: 600;
  color: var(--color-primary, #1976d2);
  text-decoration: none;
}

.register-hint .bot-link:hover {
  text-decoration: underline;
}

/* Dark theme support */
[data-theme="dark"] .info-box {
  background-color: var(--info-bg, rgba(25, 118, 210, 0.15));
  border-color: var(--info-border, rgba(144, 202, 249, 0.3));
}

[data-theme="dark"] .info-icon {
  color: var(--info-icon, #64b5f6);
}

[data-theme="dark"] .info-text {
  color: var(--info-text, #90caf9);
}

[data-theme="dark"] .info-text code {
  background-color: var(--info-code-bg, rgba(100, 181, 246, 0.2));
  color: var(--info-code-text, #bbdefb);
}

[data-theme="dark"] .register-hint {
  background-color: var(--info-bg, rgba(25, 118, 210, 0.15));
  border-color: var(--info-border, rgba(144, 202, 249, 0.3));
}

[data-theme="dark"] .register-hint p {
  color: var(--info-text, #90caf9);
}

[data-theme="dark"] .register-hint .bot-link {
  color: var(--info-icon, #64b5f6);
}

.otp-input-container {
  display: flex;
  gap: 8px;
  justify-content: center;
  margin: 20px 0;
}

.otp-digit {
  width: 48px;
  height: 56px;
  text-align: center;
  font-size: 24px;
  font-weight: 600;
  border: 2px solid var(--border-primary, #e0e0e0);
  border-radius: 8px;
  background-color: var(--input-bg, #fff);
  color: var(--text-primary);
  transition: all 0.2s;
}

.otp-digit:focus {
  outline: none;
  border-color: var(--color-primary, #1976d2);
  box-shadow: 0 0 0 3px rgba(25, 118, 210, 0.1);
}

.otp-digit:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

@media (max-width: 480px) {
  .otp-digit {
    width: 40px;
    height: 48px;
    font-size: 20px;
  }
  
  .otp-input-container {
    gap: 6px;
  }
}
</style>

