<template>
  <div class="login-container">
    <div class="card" style="max-width: 400px; margin: 50px auto;">
      <h1>English Bot Login</h1>
      
      <div v-if="step === 'username'" class="login-step">
        <p>Enter your Telegram username or ID:</p>
        <input
          ref="usernameInput"
          v-model="username"
          type="text"
          placeholder="Username or Telegram ID"
          @keyup.enter="requestOTP"
        />
        <button @click="requestOTP" class="btn btn-primary" :disabled="loading">
          {{ loading ? 'Sending...' : 'Send OTP' }}
        </button>
        <p v-if="error" class="error">{{ error }}</p>
      </div>

      <div v-if="step === 'otp'" class="login-step">
        <p>Enter the OTP code sent to your Telegram:</p>
        <input
          ref="otpInput"
          v-model="otpCode"
          type="text"
          placeholder="OTP Code"
          maxlength="6"
          @keyup.enter="verifyOTP"
        />
        <button @click="verifyOTP" class="btn btn-primary" :disabled="loading">
          {{ loading ? 'Verifying...' : 'Verify' }}
        </button>
        <button @click="step = 'username'" class="btn btn-secondary">Back</button>
        <p v-if="error" class="error">{{ error }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useAuth } from '../composables/useAuth'
import { apiClient } from '../api/client'

const router = useRouter()
const { login, tryTelegramAuth } = useAuth()

const step = ref<'username' | 'otp'>('username')
const username = ref('')
const otpCode = ref('')
const userId = ref('')
const loading = ref(false)
const error = ref('')

const usernameInput = ref<HTMLInputElement | null>(null)
const otpInput = ref<HTMLInputElement | null>(null)

onMounted(async () => {
  // Check if Telegram WebApp is available and try to authenticate
  const tg = (window as any).Telegram?.WebApp
  if (tg && tg.initData) {
    try {
      const success = await tryTelegramAuth()
      if (success) {
        router.push('/dashboard')
        return
      }
    } catch (err: any) {
      // Silent fail - show OTP login form
      error.value = 'Авторизация через Telegram не удалась. Пожалуйста, используйте OTP вход.'
    }
  }
  
  // Focus on username input when component is mounted
  if (step.value === 'username') {
    await nextTick()
    usernameInput.value?.focus()
  }
})

// Watch for step changes to focus on the appropriate input
watch(step, async (newStep) => {
  await nextTick()
  if (newStep === 'username') {
    usernameInput.value?.focus()
  } else if (newStep === 'otp') {
    otpInput.value?.focus()
  }
})

const requestOTP = async () => {
  if (!username.value.trim()) {
    error.value = 'Please enter username or Telegram ID'
    return
  }

  loading.value = true
  error.value = ''

  try {
    const response = await apiClient.requestOTP(username.value.trim())
    userId.value = response.user_id.toString()
    step.value = 'otp'
    // Focus will be set automatically by watch on step change
  } catch (err: any) {
    error.value = err.message || 'Failed to send OTP'
  } finally {
    loading.value = false
  }
}

const verifyOTP = async () => {
  if (!otpCode.value.trim()) {
    error.value = 'Please enter OTP code'
    return
  }

  loading.value = true
  error.value = ''

  try {
    const response = await apiClient.verifyOTP(userId.value, otpCode.value.trim())
    login(response.access_token, response.refresh_token)
    router.push('/dashboard')
  } catch (err: any) {
    error.value = err.message || 'Invalid OTP code'
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
</style>

