<template>
  <div class="chat">
    <div class="chat-topbar">
      <button class="chat-back" type="button" @click="router.back()">
        <LgIcon name="chevron-left" :s="20" c="var(--text)" />
      </button>
      <div class="chat-topbar-title">{{ t('navigation.chat') }}</div>
      <LgLumi :size="38" />
    </div>
    <div class="chat-container">
      <div class="chat-messages" ref="messagesContainer">
        <div v-if="messages.length === 0" class="welcome-message">
          <LgLumi :size="84" />
          <div class="welcome-icon">{{ targetLangFlag }}</div>
          <h2 class="welcome-title">{{ chatWelcome.title }}</h2>
          <p class="welcome-text">{{ chatWelcome.intro }}</p>
          <ul class="welcome-list">
            <li v-for="(line, idx) in chatWelcome.bullets" :key="idx">{{ line }}</li>
          </ul>
          <p class="welcome-text">{{ chatWelcome.closing }}</p>
        </div>
        <div
          v-for="(msg, index) in messages"
          :key="index"
          :class="['message', msg.role]"
        >
          <div 
            class="message-content"
            :class="{ 'markdown-content': msg.role === 'assistant' }"
            v-html="msg.role === 'assistant' ? renderMarkdown(msg.content) : escapeHtml(msg.content)"
          ></div>
        </div>
        <div v-if="sending" class="message assistant">
          <div class="message-content typing-indicator">
            <span class="typing-dot"></span>
            <span class="typing-dot"></span>
            <span class="typing-dot"></span>
          </div>
        </div>
      </div>
      
      <div class="chat-input-wrapper">
        <div class="chat-input-container">
          <div class="chat-input">
            <textarea
              ref="textareaRef"
              v-model="inputMessage"
              :placeholder="t('chat.placeholder')"
              @keydown.enter.exact.prevent="sendMessage"
              @input="autoResize"
            ></textarea>
            <button @click="sendMessage" class="chat-send-btn" :disabled="sending" :title="t('chat.send')">
              <LgIcon name="send" :s="18" c="white" />
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onMounted, watch, onActivated } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import LgIcon from '../components/linglow/LgIcon.vue'
import LgLumi from '../components/linglow/LgLumi.vue'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import { apiClient } from '../api/client'
import { useLearningConfig } from '../composables/useLearningConfig'

interface Message {
  role: 'user' | 'assistant'
  content: string
}

const { t, tm } = useI18n()
const { learning, ensureLearningLoaded } = useLearningConfig()

const messages = ref<Message[]>([])
const inputMessage = ref('')
const sending = ref(false)
const messagesContainer = ref<HTMLElement | null>(null)
const textareaRef = ref<HTMLTextAreaElement | null>(null)
const route = useRoute()
const router = useRouter()

const targetLangFlag = computed(() =>
  learning.value?.target_lang === 'es' ? '🇪🇸' : '🇬🇧'
)

const chatWelcome = computed(() => {
  const target = learning.value?.target_lang === 'es' ? 'es' : 'en'
  const w = tm(`chat.welcome.${target}`) as {
    title?: string
    intro?: string
    bullets?: string[]
    closing?: string
  }
  return {
    title: w?.title ?? '',
    intro: w?.intro ?? '',
    bullets: Array.isArray(w?.bullets) ? w.bullets : [],
    closing: w?.closing ?? '',
  }
})

const scrollToBottom = async () => {
  await nextTick()
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

const autoResize = () => {
  if (textareaRef.value) {
    // Reset height to calculate scrollHeight correctly
    textareaRef.value.style.height = '40px'
    const scrollHeight = textareaRef.value.scrollHeight
    // Set height to scrollHeight, but not less than 40px (button height)
    const newHeight = Math.max(40, scrollHeight)
    textareaRef.value.style.height = `${newHeight}px`
    // Hide scrollbar if content fits in max-height
    if (newHeight <= 200) {
      textareaRef.value.style.overflowY = 'hidden'
    } else {
      textareaRef.value.style.overflowY = 'auto'
    }
  }
}

const sendMessage = async () => {
  if (!inputMessage.value.trim() || sending.value) return

  const userMessage = inputMessage.value.trim()
  inputMessage.value = ''
  autoResize() // Reset textarea height after clearing
  messages.value.push({ role: 'user', content: userMessage })
  await scrollToBottom()

  sending.value = true
  await scrollToBottom() // Scroll to show typing indicator
  try {
    const formData = new FormData()
    formData.append('message', userMessage)
    
    const data: { response: string } = await apiClient.requestFormData('/api/chat', formData)
    messages.value.push({ role: 'assistant', content: data.response })
  } catch (error) {
    console.error('Failed to send message:', error)
    messages.value.push({ role: 'assistant', content: t('chat.errorSend') })
  } finally {
    sending.value = false
    await scrollToBottom()
  }
}

// Configure marked for security
marked.setOptions({
  breaks: true, // Convert line breaks to <br>
  gfm: true, // GitHub Flavored Markdown
})

const renderMarkdown = (text: string): string => {
  try {
    return marked.parse(text) as string
  } catch (error) {
    console.error('Failed to render markdown:', error)
    return escapeHtml(text)
  }
}

const escapeHtml = (text: string): string => {
  const div = document.createElement('div')
  div.textContent = text
  return div.innerHTML
}

const focusInput = async () => {
  await nextTick()
  if (textareaRef.value) {
    textareaRef.value.focus()
  }
}

// Focus input when component is mounted
onMounted(() => {
  void ensureLearningLoaded()
  scrollToBottom()
  autoResize() // Set initial height
  focusInput()
})

// Focus input when component is activated (if using keep-alive)
onActivated(() => {
  focusInput()
})

// Focus input when route changes to /chat
watch(() => route.path, (newPath) => {
  if (newPath === '/chat') {
    focusInput()
  }
})
</script>

<style scoped>
.chat {
  max-width: 880px;
  margin: 0 auto;
  padding: 0;
  height: 100vh;
  height: 100dvh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.chat-topbar {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px 8px;
  flex-shrink: 0;
}
.chat-back {
  width: 38px;
  height: 38px;
  border-radius: 50%;
  border: 1px solid var(--border);
  background: var(--card-bg);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  flex-shrink: 0;
}
.chat-topbar-title {
  flex: 1;
  text-align: center;
  font-family: 'Lora', serif;
  font-weight: 700;
  font-size: 18px;
  color: var(--text);
}
.chat-send-btn {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  border: 1px solid var(--btn-border);
  background: var(--btn-gradient);
  box-shadow: var(--btn-shadow);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  flex-shrink: 0;
}
.chat-send-btn:disabled { opacity: 0.55; cursor: default; }

.chat-container {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
  position: relative;
  background: transparent;
  box-shadow: none;
  border: none;
  padding: 0;
  overflow: hidden;
  height: 100%;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 12px 16px 20px;
  border-radius: 0;
  margin-bottom: 0;
  min-height: 0;
  box-sizing: border-box;
}

.chat-input-wrapper {
  flex-shrink: 0;
  background: var(--nav-bg);
  backdrop-filter: blur(18px);
  -webkit-backdrop-filter: blur(18px);
  border-top: 1px solid var(--border);
  padding-bottom: env(safe-area-inset-bottom, 0px);
}

.chat-input-container {
  max-width: 880px;
  margin: 0 auto;
  padding: 10px 16px;
}

.chat-input {
  display: flex;
  gap: 10px;
  align-items: flex-start;
}

.chat-input textarea {
  flex: 1;
  resize: none;
  height: 40px;
  min-height: 40px;
  max-height: 200px;
  padding: 9px 16px;
  box-sizing: border-box;
  line-height: 20px;
  overflow-y: hidden;
  border-radius: 20px;
  margin-bottom: 0;
}

.chat-input button {
  height: 40px;
  padding: 10px 20px;
  box-sizing: border-box;
  flex-shrink: 0;
  align-self: flex-start;
  margin-top: 0;
}

.welcome-message {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 40px 20px;
  color: var(--text-secondary);
}

.welcome-icon {
  font-size: 64px;
  margin-bottom: 20px;
  animation: wave 2s ease-in-out infinite;
}

@keyframes wave {
  0%, 100% {
    transform: rotate(0deg);
  }
  25% {
    transform: rotate(20deg);
  }
  75% {
    transform: rotate(-20deg);
  }
}

.welcome-title {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 16px 0;
}

.welcome-text {
  font-size: 16px;
  line-height: 1.6;
  margin: 12px 0;
  max-width: 600px;
  color: var(--text-secondary);
}

.welcome-list {
  text-align: left;
  max-width: 500px;
  margin: 20px auto;
  padding-left: 20px;
  color: var(--text-primary);
}

.welcome-list li {
  margin: 8px 0;
  line-height: 1.5;
}

.message {
  margin-bottom: 15px;
}

.message.user {
  text-align: right;
}

.message-content {
  display: inline-block;
  padding: 10px 15px;
  border-radius: 16px;
  max-width: 85%;
  word-wrap: break-word;
  font-size: 15px;
  line-height: 1.45;
}

.message.user .message-content {
  background: var(--btn-gradient);
  color: #fff;
  border-bottom-right-radius: 6px;
}

.message.assistant .message-content {
  background: var(--card-bg);
  color: var(--text);
  border: 1px solid var(--border);
  border-bottom-left-radius: 6px;
}

.typing-indicator {
  display: flex;
  align-items: flex-end;
  gap: 4px;
  padding: 12px 16px;
  min-width: 50px;
  padding-bottom: 14px;
}

.typing-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: var(--text-secondary);
  animation: typing 1.4s infinite ease-in-out;
  display: inline-block;
}

.typing-dot:nth-child(1) {
  animation-delay: -0.32s;
}

.typing-dot:nth-child(2) {
  animation-delay: -0.16s;
}

.typing-dot:nth-child(3) {
  animation-delay: 0;
}

@keyframes typing {
  0%, 60%, 100% {
    transform: translateY(0);
    opacity: 0.7;
  }
  30% {
    transform: translateY(-6px);
    opacity: 1;
  }
}

.markdown-content {
  text-align: left;
}

.markdown-content :deep(h1),
.markdown-content :deep(h2),
.markdown-content :deep(h3),
.markdown-content :deep(h4),
.markdown-content :deep(h5),
.markdown-content :deep(h6) {
  margin: 10px 0 5px 0;
  font-weight: bold;
}

.markdown-content :deep(h1) {
  font-size: 1.5em;
}

.markdown-content :deep(h2) {
  font-size: 1.3em;
}

.markdown-content :deep(h3) {
  font-size: 1.1em;
}

.markdown-content :deep(p) {
  margin: 8px 0;
  line-height: 1.5;
}

.markdown-content :deep(ul),
.markdown-content :deep(ol) {
  margin: 8px 0;
  padding-left: 25px;
}

.markdown-content :deep(li) {
  margin: 4px 0;
}

.markdown-content :deep(code) {
  background: var(--code-bg);
  color: var(--code-text);
  padding: 2px 6px;
  border-radius: 3px;
  font-family: 'Courier New', monospace;
  font-size: 0.9em;
}

.markdown-content :deep(pre) {
  background: var(--code-bg);
  color: var(--code-text);
  padding: 10px;
  border-radius: 4px;
  overflow-x: auto;
  margin: 10px 0;
}

.markdown-content :deep(pre code) {
  background: none;
  padding: 0;
}

.markdown-content :deep(strong) {
  font-weight: bold;
}

.markdown-content :deep(em) {
  font-style: italic;
}

.markdown-content :deep(a) {
  color: var(--color-primary);
  text-decoration: none;
}

.markdown-content :deep(a:hover) {
  text-decoration: underline;
}

.markdown-content :deep(blockquote) {
  border-left: 3px solid var(--border-primary);
  padding-left: 10px;
  margin: 10px 0;
  color: var(--text-secondary);
}

/* Mobile styles */
@media (max-width: 768px) {
  .chat {
    padding: 0;
    height: 100%;
    display: flex;
    flex-direction: column;
    max-width: 100%;
  }

  .chat-container {
    flex: 1;
    min-height: 0;
    height: 100%;
    display: flex;
    flex-direction: column;
    position: relative;
    background: transparent;
    box-shadow: none;
    border: none;
    padding: 0;
  }

  .chat-messages {
    flex: 1;
    overflow-y: auto;
    padding: 12px 10px 16px;
    margin-bottom: 0;
    border-radius: 0;
  }

  .chat-input-container {
    max-width: 100%;
    padding: 10px 8px;
  }

  .welcome-message {
    padding: 30px 8px;
    min-height: 200px;
  }

  .welcome-icon {
    font-size: 48px;
    margin-bottom: 16px;
  }

  .welcome-title {
    font-size: 20px;
    margin-bottom: 12px;
  }

  .welcome-text {
    font-size: 14px;
    margin: 10px 0;
  }

  .welcome-list {
    max-width: 100%;
    margin: 16px auto;
    padding-left: 16px;
    font-size: 14px;
  }

  .welcome-list li {
    margin: 6px 0;
  }
}
</style>

