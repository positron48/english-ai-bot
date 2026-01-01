<template>
  <div class="login-container">
    <!-- Debug Panel (visible in Telegram Mini App) -->
    <div v-if="showDebug" class="debug-panel">
      <h3>Debug Info</h3>
      <div class="debug-item">
        <strong>Telegram WebApp:</strong> 
        <span :class="debugInfo.telegramLoaded ? 'success' : 'error'">
          {{ debugInfo.telegramLoaded ? 'Loaded' : 'Not loaded' }}
        </span>
      </div>
      <div class="debug-item">
        <strong>initData:</strong> 
        <span :class="debugInfo.hasInitData ? 'success' : 'error'">
          {{ debugInfo.hasInitData ? 'Available' : 'Not available' }}
        </span>
      </div>
      <div class="debug-item" v-if="debugInfo.initDataLength > 0">
        <strong>initData length:</strong> {{ debugInfo.initDataLength }}
      </div>
      <div class="debug-item">
        <strong>Auth Status:</strong> {{ debugInfo.authStatus }}
      </div>
      <div class="debug-item" v-if="debugInfo.error">
        <strong>Error:</strong> <span class="error">{{ debugInfo.error }}</span>
      </div>
      <button @click="showDebug = false" class="btn btn-secondary" style="margin-top: 10px;">Hide Debug</button>
    </div>
    
    <div class="card" style="max-width: 400px; margin: 50px auto;">
      <h1>English Bot Login</h1>
      
      <!-- Show debug button if in Telegram Mini App -->
      <button 
        v-if="isTelegramMiniApp" 
        @click="showDebug = !showDebug" 
        class="btn btn-secondary"
        style="margin-bottom: 10px; font-size: 12px; padding: 5px 10px;"
      >
        {{ showDebug ? 'Hide' : 'Show' }} Debug
      </button>
      
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
import { ref, onMounted, reactive } from 'vue'
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
const showDebug = ref(false)

// Check if we're in Telegram Mini App
const isTelegramMiniApp = ref(false)

// Debug info
const debugInfo = reactive({
  telegramLoaded: false,
  hasInitData: false,
  initDataLength: 0,
  authStatus: 'Initializing...',
  error: ''
})

onMounted(async () => {
  // Check if Telegram WebApp is available
  const tg = (window as any).Telegram?.WebApp
  isTelegramMiniApp.value = !!tg
  
  if (tg) {
    debugInfo.telegramLoaded = true
    debugInfo.hasInitData = !!tg.initData
    debugInfo.initDataLength = tg.initData ? tg.initData.length : 0
    
    // Show debug panel automatically in Telegram Mini App if there's an issue
    if (!tg.initData) {
      showDebug.value = true
      debugInfo.authStatus = 'No initData available'
      debugInfo.error = 'Telegram WebApp is loaded but initData is missing'
    } else {
      debugInfo.authStatus = 'Attempting Telegram auth...'
      showDebug.value = true // Show debug by default in Telegram Mini App
      
      try {
        const success = await tryTelegramAuth()
        if (success) {
          debugInfo.authStatus = 'Authentication successful!'
          router.push('/dashboard')
        } else {
          debugInfo.authStatus = 'Telegram auth failed, showing login form'
          debugInfo.error = 'Telegram authentication failed. Please use OTP login.'
        }
      } catch (err: any) {
        debugInfo.authStatus = 'Error during Telegram auth'
        debugInfo.error = err.message || 'Unknown error'
        error.value = err.message || 'Telegram authentication failed'
      }
    }
  } else {
    debugInfo.telegramLoaded = false
    debugInfo.authStatus = 'Not in Telegram Mini App'
    debugInfo.error = 'Telegram WebApp script not loaded or not in Telegram Mini App'
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

.debug-panel {
  background: var(--bg-secondary, #f5f5f5);
  border: 2px solid var(--border-primary, #ddd);
  border-radius: 8px;
  padding: 15px;
  margin-bottom: 20px;
  max-width: 500px;
  width: 100%;
  font-size: 12px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.debug-panel h3 {
  margin-top: 0;
  margin-bottom: 10px;
  font-size: 14px;
}

.debug-item {
  margin-bottom: 8px;
  word-break: break-all;
}

.debug-item strong {
  display: inline-block;
  min-width: 120px;
}

.success {
  color: green;
  font-weight: bold;
}

.error {
  color: red;
  font-weight: bold;
}
</style>

