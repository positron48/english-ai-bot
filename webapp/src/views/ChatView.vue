<template>
  <div class="chat">
    <h1>AI Chat</h1>
    
    <div class="card chat-container">
      <div class="chat-messages" ref="messagesContainer">
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
      </div>
      
      <div class="chat-input">
        <textarea
          v-model="inputMessage"
          placeholder="Type your message..."
          @keyup.enter.ctrl="sendMessage"
          rows="3"
        ></textarea>
        <button @click="sendMessage" class="btn btn-primary" :disabled="sending">
          {{ sending ? 'Sending...' : 'Send' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick, onMounted } from 'vue'
import { marked } from 'marked'
import { apiClient } from '../api/client'

interface Message {
  role: 'user' | 'assistant'
  content: string
}

const messages = ref<Message[]>([])
const inputMessage = ref('')
const sending = ref(false)
const messagesContainer = ref<HTMLElement | null>(null)

const scrollToBottom = async () => {
  await nextTick()
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

const sendMessage = async () => {
  if (!inputMessage.value.trim() || sending.value) return

  const userMessage = inputMessage.value.trim()
  inputMessage.value = ''
  messages.value.push({ role: 'user', content: userMessage })
  await scrollToBottom()

  sending.value = true
  try {
    const formData = new FormData()
    formData.append('message', userMessage)
    
    const data: { response: string } = await apiClient.requestFormData('/app/chat', formData)
    messages.value.push({ role: 'assistant', content: data.response })
  } catch (error) {
    console.error('Failed to send message:', error)
    messages.value.push({ role: 'assistant', content: 'Sorry, an error occurred. Please try again.' })
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

onMounted(() => {
  scrollToBottom()
})
</script>

<style scoped>
.chat-container {
  display: flex;
  flex-direction: column;
  height: 600px;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  background: var(--chat-bg);
  border-radius: 4px;
  margin-bottom: 20px;
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
  border-radius: 8px;
  max-width: 70%;
  word-wrap: break-word;
}

.message.user .message-content {
  background: var(--chat-user-bg);
  color: var(--text-inverse);
}

.message.assistant .message-content {
  background: var(--chat-assistant-bg);
  color: var(--text-primary);
  border: 1px solid var(--chat-assistant-border);
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

.chat-input {
  display: flex;
  gap: 10px;
}

.chat-input textarea {
  flex: 1;
  resize: none;
}

.chat-input button {
  align-self: flex-end;
}
</style>

