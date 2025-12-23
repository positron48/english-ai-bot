<template>
  <div class="login-container">
    <div class="card" style="max-width: 400px; margin: 50px auto;">
      <h1>English Bot Login</h1>
      
      <div v-if="step === 'username'" class="login-step">
        <p>Enter your Telegram username or ID:</p>
        <input
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
import { ref, onMounted } from 'vue'
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

onMounted(async () => {
  const success = await tryTelegramAuth()
  if (success) {
    router.push('/dashboard')
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
  align-items: center;
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

