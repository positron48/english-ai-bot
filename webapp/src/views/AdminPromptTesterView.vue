<template>
  <div class="admin-layout">
    <AdminMenu />
    
    <div class="admin-content">
      <h2>Prompt Tester</h2>
      
      <div class="card">
      <h2>Words</h2>
      <textarea
        v-model="wordsText"
        @input="saveToLocalStorage"
        class="words-input"
        placeholder="Enter words separated by commas or newlines (e.g., lack, spy, scholar)"
        rows="4"
      ></textarea>
      <div class="words-info">
        <span>{{ wordCount }} word(s)</span>
      </div>
    </div>

    <div class="card">
      <h2>Prompts</h2>
      
      <details class="prompt-section">
        <summary>
          <strong>Word Data Prompt</strong>
          <span v-if="wordPromptSource" class="prompt-source">({{ wordPromptSource }})</span>
          <span v-if="hasWordPromptChanges" class="prompt-modified">[Modified]</span>
        </summary>
        <div class="prompt-editor">
          <textarea
            v-model="wordPromptCurrent"
            @input="saveToLocalStorage"
            class="prompt-textarea"
            rows="10"
          ></textarea>
          <div class="prompt-actions">
            <button @click="resetWordPrompt" class="btn btn-secondary">Reset to Default</button>
          </div>
          <div v-if="hasWordPromptChanges" class="prompt-diff">
            <div class="prompt-diff-header">
              <h4>Changes from Default:</h4>
              <div class="prompt-diff-controls">
                <button
                  @click="showWordPromptDiff = !showWordPromptDiff; saveToLocalStorage()"
                  class="btn btn-small"
                >
                  {{ showWordPromptDiff ? 'Hide' : 'Show' }} Diff
                </button>
                <div v-if="showWordPromptDiff" class="prompt-diff-mode-selector">
                  <label>Mode:</label>
                  <div class="toggle-switch">
                    <button
                      @click="promptDiffMode = 'unified'; saveToLocalStorage()"
                      :class="['toggle-option', { active: promptDiffMode === 'unified' }]"
                    >
                      Unified
                    </button>
                    <button
                      @click="promptDiffMode = 'side-by-side'; saveToLocalStorage()"
                      :class="['toggle-option', { active: promptDiffMode === 'side-by-side' }]"
                    >
                      Side by Side
                    </button>
                  </div>
                </div>
              </div>
            </div>
            <div v-if="showWordPromptDiff" class="diff-container">
              <component :is="wordPromptDiffComponent" />
            </div>
          </div>
        </div>
      </details>

      <details class="prompt-section">
        <summary>
          <strong>Training Cards Prompt</strong>
          <span v-if="trainingPromptSource" class="prompt-source">({{ trainingPromptSource }})</span>
          <span v-if="hasTrainingPromptChanges" class="prompt-modified">[Modified]</span>
        </summary>
        <div class="prompt-editor">
          <textarea
            v-model="trainingPromptCurrent"
            @input="saveToLocalStorage"
            class="prompt-textarea"
            rows="10"
          ></textarea>
          <div class="prompt-actions">
            <button @click="resetTrainingPrompt" class="btn btn-secondary">Reset to Default</button>
          </div>
          <div v-if="hasTrainingPromptChanges" class="prompt-diff">
            <div class="prompt-diff-header">
              <h4>Changes from Default:</h4>
              <div class="prompt-diff-controls">
                <button
                  @click="showTrainingPromptDiff = !showTrainingPromptDiff; saveToLocalStorage()"
                  class="btn btn-small"
                >
                  {{ showTrainingPromptDiff ? 'Hide' : 'Show' }} Diff
                </button>
                <div v-if="showTrainingPromptDiff" class="prompt-diff-mode-selector">
                  <label>Mode:</label>
                  <div class="toggle-switch">
                    <button
                      @click="promptDiffMode = 'unified'; saveToLocalStorage()"
                      :class="['toggle-option', { active: promptDiffMode === 'unified' }]"
                    >
                      Unified
                    </button>
                    <button
                      @click="promptDiffMode = 'side-by-side'; saveToLocalStorage()"
                      :class="['toggle-option', { active: promptDiffMode === 'side-by-side' }]"
                    >
                      Side by Side
                    </button>
                  </div>
                </div>
              </div>
            </div>
            <div v-if="showTrainingPromptDiff" class="diff-container">
              <component :is="trainingPromptDiffComponent" />
            </div>
          </div>
        </div>
      </details>
    </div>

    <div class="card">
      <div class="controls">
        <button
          @click="runTests"
          :disabled="isRunning || wordCount === 0"
          class="btn btn-primary"
        >
          {{ isRunning ? 'Running...' : 'Run Tests' }}
        </button>
        <button
          v-if="results.length > 0"
          @click="clearResults"
          class="btn btn-secondary"
          :disabled="isRunning"
        >
          Clear Results
        </button>
        <div class="diff-mode-selector">
          <div class="toggle-switch">
            <button
              @click="diffMode = 'unified'; saveToLocalStorage()"
              :class="['toggle-option', { active: diffMode === 'unified' }]"
            >
              Unified
            </button>
            <button
              @click="diffMode = 'side-by-side'; saveToLocalStorage()"
              :class="['toggle-option', { active: diffMode === 'side-by-side' }]"
            >
              Side by Side
            </button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="results.length > 0" class="card">
      <h2>Results</h2>
      <div class="results-container">
        <div v-for="(wordResults, word) in groupedResults" :key="word" class="word-result">
          <h3>{{ word }}</h3>
          
          <div v-for="(result, step) in wordResults" :key="step" class="step-result">
            <h4>{{ step === 'word' ? 'Word Data' : 'Training Cards' }}</h4>
            
            <!-- Always show current result -->
            <div v-if="result.current" class="result-current">
              <div v-if="result.current.ok" class="result-success">
                <div v-if="result.current.parsed" class="json-viewer">
                  <pre><code v-html="highlightJSON(result.current.parsed)"></code></pre>
                </div>
                <div v-else class="raw-response">
                  <pre>{{ result.current.raw }}</pre>
                </div>
                <div class="result-meta">
                  Duration: {{ result.current.duration_ms }}ms
                  <span v-if="result.current.timestamp" style="margin-left: 16px;">
                    Last Updated: {{ new Date(result.current.timestamp).toLocaleString() }}
                  </span>
                </div>
              </div>
              <div v-else class="result-error">
                <strong>Error:</strong> {{ result.current.error }}
                <div class="result-meta">
                  Duration: {{ result.current.duration_ms }}ms
                  <span v-if="result.current.timestamp" style="margin-left: 16px;">
                    Last Updated: {{ new Date(result.current.timestamp).toLocaleString() }}
                  </span>
                </div>
              </div>
            </div>

            <!-- Show diff if there's both previous and current, and diff is enabled -->
            <div v-if="result.previous && result.current && showResultDiffs" class="result-diff">
              <div class="result-diff-header">
                <h5>Changes from Previous Run:</h5>
                <button
                  @click="showResultDiffs = false; saveToLocalStorage()"
                  class="btn btn-small"
                >
                  Hide Diff
                </button>
              </div>
              <component :is="getResultDiffComponent(result)" />
            </div>
            
            <!-- Show button to show diff if it's hidden -->
            <div v-if="result.previous && result.current && !showResultDiffs" class="result-diff-toggle">
              <button
                @click="showResultDiffs = true; saveToLocalStorage()"
                class="btn btn-small"
              >
                Show Diff
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, h } from 'vue'
import { apiClient } from '../api/client'
import { showAlert, showConfirm } from '../composables/useDialog'
import AdminMenu from '../components/AdminMenu.vue'
// Custom character-level diff implementation
interface DiffPart {
  value: string
  added?: boolean
  removed?: boolean
}

// Simple character-level diff using Myers algorithm
function diffChars(oldText: string, newText: string): DiffPart[] {
  const oldLen = oldText.length
  const newLen = newText.length
  
  // Simple approach: find common prefix and suffix, then diff the middle
  let prefixLen = 0
  while (prefixLen < oldLen && prefixLen < newLen && oldText[prefixLen] === newText[prefixLen]) {
    prefixLen++
  }
  
  let suffixLen = 0
  while (
    suffixLen < oldLen - prefixLen &&
    suffixLen < newLen - prefixLen &&
    oldText[oldLen - 1 - suffixLen] === newText[newLen - 1 - suffixLen]
  ) {
    suffixLen++
  }
  
  const parts: DiffPart[] = []
  
  // Add common prefix
  if (prefixLen > 0) {
    parts.push({ value: oldText.substring(0, prefixLen) })
  }
  
  // Diff the middle part
  const oldMiddle = oldText.substring(prefixLen, oldLen - suffixLen)
  const newMiddle = newText.substring(prefixLen, newLen - suffixLen)
  
  if (oldMiddle.length === 0 && newMiddle.length === 0) {
    // Nothing to diff
  } else if (oldMiddle.length === 0) {
    // All added
    parts.push({ value: newMiddle, added: true })
  } else if (newMiddle.length === 0) {
    // All removed
    parts.push({ value: oldMiddle, removed: true })
  } else {
    // Use a more sophisticated diff for the middle
    const middleDiff = diffMiddle(oldMiddle, newMiddle)
    parts.push(...middleDiff)
  }
  
  // Add common suffix
  if (suffixLen > 0) {
    parts.push({ value: oldText.substring(oldLen - suffixLen) })
  }
  
  return parts
}

// Diff the middle part using Myers algorithm (simplified)
function diffMiddle(oldText: string, newText: string): DiffPart[] {
  // Use a simple LCS-based approach for better results
  const parts: DiffPart[] = []
  
  // Build a simple edit graph
  const maxD = oldText.length + newText.length
  const v: { [key: number]: number } = { 1: 0 }
  
  // Simple approach: use dynamic programming to find common subsequence
  // For performance, we'll use a simpler greedy approach
  let oldIdx = 0
  let newIdx = 0
  
  while (oldIdx < oldText.length || newIdx < newText.length) {
    if (oldIdx >= oldText.length) {
      if (newIdx < newText.length) {
        parts.push({ value: newText.substring(newIdx), added: true })
      }
      break
    }
    if (newIdx >= newText.length) {
      if (oldIdx < oldText.length) {
        parts.push({ value: oldText.substring(oldIdx), removed: true })
      }
      break
    }
    
    // Try to find the longest common substring starting from current positions
    let bestMatch = { oldStart: oldIdx, newStart: newIdx, len: 0 }
    
    // Search for matches in a window
    const searchWindow = 100
    for (let i = 0; i < Math.min(searchWindow, oldText.length - oldIdx); i++) {
      for (let j = 0; j < Math.min(searchWindow, newText.length - newIdx); j++) {
        if (oldText[oldIdx + i] === newText[newIdx + j]) {
          // Found a match, extend it
          let len = 1
          while (
            oldIdx + i + len < oldText.length &&
            newIdx + j + len < newText.length &&
            oldText[oldIdx + i + len] === newText[newIdx + j + len]
          ) {
            len++
          }
          if (len > bestMatch.len) {
            bestMatch = { oldStart: oldIdx + i, newStart: newIdx + j, len }
          }
        }
      }
    }
    
    if (bestMatch.len >= 2) {
      // Found a good match
      // Add removed part before match
      if (bestMatch.oldStart > oldIdx) {
        parts.push({ value: oldText.substring(oldIdx, bestMatch.oldStart), removed: true })
      }
      // Add added part before match
      if (bestMatch.newStart > newIdx) {
        parts.push({ value: newText.substring(newIdx, bestMatch.newStart), added: true })
      }
      // Add common part
      parts.push({ value: oldText.substring(bestMatch.oldStart, bestMatch.oldStart + bestMatch.len) })
      oldIdx = bestMatch.oldStart + bestMatch.len
      newIdx = bestMatch.newStart + bestMatch.len
    } else {
      // No good match, compare character by character
      if (oldText[oldIdx] === newText[newIdx]) {
        parts.push({ value: oldText[oldIdx] })
        oldIdx++
        newIdx++
      } else {
        // Try to find next match
        let found = false
        for (let lookahead = 1; lookahead < 20 && oldIdx + lookahead < oldText.length; lookahead++) {
          if (oldText[oldIdx + lookahead] === newText[newIdx]) {
            parts.push({ value: oldText.substring(oldIdx, oldIdx + lookahead), removed: true })
            oldIdx += lookahead
            found = true
            break
          }
        }
        if (!found) {
          for (let lookahead = 1; lookahead < 20 && newIdx + lookahead < newText.length; lookahead++) {
            if (newText[newIdx + lookahead] === oldText[oldIdx]) {
              parts.push({ value: newText.substring(newIdx, newIdx + lookahead), added: true })
              newIdx += lookahead
              found = true
              break
            }
          }
        }
        if (!found) {
          // No match found, add both as changed
          parts.push({ value: oldText[oldIdx], removed: true })
          parts.push({ value: newText[newIdx], added: true })
          oldIdx++
          newIdx++
        }
      }
    }
  }
  
  return parts
}

// Line-level diff for side-by-side mode
function diffLines(oldText: string, newText: string): DiffPart[] {
  const oldLines = oldText.split('\n')
  const newLines = newText.split('\n')
  const parts: DiffPart[] = []
  
  let oldIdx = 0
  let newIdx = 0
  
  while (oldIdx < oldLines.length || newIdx < newLines.length) {
    if (oldIdx >= oldLines.length) {
      // Remaining in new is all added
      parts.push({ value: newLines.slice(newIdx).join('\n'), added: true })
      break
    }
    if (newIdx >= newLines.length) {
      // Remaining in old is all removed
      parts.push({ value: oldLines.slice(oldIdx).join('\n'), removed: true })
      break
    }
    
    if (oldLines[oldIdx] === newLines[newIdx]) {
      parts.push({ value: oldLines[oldIdx] + '\n' })
      oldIdx++
      newIdx++
    } else {
      // Try to find a match
      let found = false
      for (let i = oldIdx + 1; i < Math.min(oldIdx + 10, oldLines.length); i++) {
        if (oldLines[i] === newLines[newIdx]) {
          // Found a match, add removed lines
          parts.push({ value: oldLines.slice(oldIdx, i).join('\n') + '\n', removed: true })
          oldIdx = i
          found = true
          break
        }
      }
      
      if (!found) {
        for (let j = newIdx + 1; j < Math.min(newIdx + 10, newLines.length); j++) {
          if (newLines[j] === oldLines[oldIdx]) {
            // Found a match, add added lines
            parts.push({ value: newLines.slice(newIdx, j).join('\n') + '\n', added: true })
            newIdx = j
            found = true
            break
          }
        }
      }
      
      if (!found) {
        // No match found, add both as changed
        parts.push({ value: oldLines[oldIdx] + '\n', removed: true })
        parts.push({ value: newLines[newIdx] + '\n', added: true })
        oldIdx++
        newIdx++
      }
    }
  }
  
  return parts
}

interface PromptTesterEvent {
  word: string
  step: 'word' | 'cards'
  ok: boolean
  raw?: string
  parsed?: any
  error?: string
  duration_ms: number
  timestamp?: number // Add timestamp for proper sorting
  _index?: number // Internal index for sorting after localStorage load
}

interface Result {
  ok: boolean
  raw?: string
  parsed?: any
  error?: string
  duration_ms: number
  timestamp?: number
}

interface WordResults {
  word?: Result
  cards?: Result
}

const STORAGE_KEY = 'admin_prompt_tester_state_v1'

// State
const wordsText = ref('')
const wordPromptCurrent = ref('')
const wordPromptDefault = ref('')
const trainingPromptCurrent = ref('')
const trainingPromptDefault = ref('')
const wordPromptSource = ref('')
const trainingPromptSource = ref('')
const isRunning = ref(false)
const diffMode = ref<'unified' | 'side-by-side'>('unified')
const promptDiffMode = ref<'unified' | 'side-by-side'>('unified')
const showWordPromptDiff = ref(true)
const showTrainingPromptDiff = ref(true)
const showResultDiffs = ref(true)
const results = ref<PromptTesterEvent[]>([])

// Computed
const wordCount = computed(() => {
  return parseWords(wordsText.value).length
})

const hasWordPromptChanges = computed(() => {
  return wordPromptCurrent.value !== wordPromptDefault.value
})

const hasTrainingPromptChanges = computed(() => {
  return trainingPromptCurrent.value !== trainingPromptDefault.value
})

const wordPromptDiffComponent = computed(() => {
  return () => renderDiff(wordPromptDefault.value, wordPromptCurrent.value, promptDiffMode.value)
})

const trainingPromptDiffComponent = computed(() => {
  return () => renderDiff(trainingPromptDefault.value, trainingPromptCurrent.value, promptDiffMode.value)
})

const groupedResults = computed(() => {
  const grouped: Record<string, WordResults> = {}
  
  // Group events by word and step, keeping only the last two for each combination
  const eventsByKey: Record<string, PromptTesterEvent[]> = {}
  
  for (const event of results.value) {
    const key = `${event.word}:${event.step}`
    if (!eventsByKey[key]) {
      eventsByKey[key] = []
    }
    eventsByKey[key].push(event)
  }
  
  // For each word+step combination, take the last two events
  for (const [key, events] of Object.entries(eventsByKey)) {
    const [word, step] = key.split(':')
    if (!grouped[word]) {
      grouped[word] = {}
    }
    
    const stepKey = step as keyof WordResults
    
    // Sort events by their _index (if available) or by index in results array
    // _index is more reliable after localStorage load since objects may be different instances
    const sortedEvents = [...events].sort((a, b) => {
      // Prefer _index if available (from localStorage or when added)
      if (a._index !== undefined && b._index !== undefined) {
        return a._index - b._index
      }
      // Fallback to array index
      const indexA = results.value.indexOf(a)
      const indexB = results.value.indexOf(b)
      // If indices are the same (shouldn't happen), use timestamp as tiebreaker
      if (indexA === indexB && a.timestamp && b.timestamp) {
        return a.timestamp - b.timestamp
      }
      return indexA - indexB
    })
    
    // Get the latest event as current
    const latest = sortedEvents[sortedEvents.length - 1]
    // Get the second-to-last event as previous (if exists)
    const previous = sortedEvents.length > 1 ? sortedEvents[sortedEvents.length - 2] : undefined
    
    grouped[word][stepKey] = {
      current: {
        ok: latest.ok,
        raw: latest.raw || '',
        parsed: latest.parsed,
        error: latest.error,
        duration_ms: latest.duration_ms,
        timestamp: latest.timestamp
      },
      previous: previous ? {
        ok: previous.ok,
        raw: previous.raw || '',
        parsed: previous.parsed,
        error: previous.error,
        duration_ms: previous.duration_ms,
        timestamp: previous.timestamp
      } : undefined
    }
  }
  
  return grouped
})

// Methods
function parseWords(text: string): string[] {
  if (!text.trim()) return []
  
  return text
    .split(/[,\n]+/)
    .map(w => w.trim())
    .filter(w => w.length > 0)
    .filter((w, i, arr) => arr.indexOf(w) === i) // dedupe
}

function loadFromLocalStorage() {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored) {
      const data = JSON.parse(stored)
      wordsText.value = data.wordsText || ''
      wordPromptCurrent.value = data.wordPromptCurrent || ''
      trainingPromptCurrent.value = data.trainingPromptCurrent || ''
      diffMode.value = data.diffMode || 'unified'
      promptDiffMode.value = data.promptDiffMode || 'unified'
      showWordPromptDiff.value = data.showWordPromptDiff !== undefined ? data.showWordPromptDiff : true
      showTrainingPromptDiff.value = data.showTrainingPromptDiff !== undefined ? data.showTrainingPromptDiff : true
      showResultDiffs.value = data.showResultDiffs !== undefined ? data.showResultDiffs : true
      // Load results and ensure they have timestamps and indices
      const loadedResults = data.results || []
      // Add timestamps and indices to old results that don't have them (for backward compatibility)
      let baseTime = Date.now() - (loadedResults.length * 1000) // Spread over time
      results.value = loadedResults.map((r: PromptTesterEvent, index: number) => {
        if (!r.timestamp) {
          r.timestamp = baseTime + (index * 1000)
        }
        // Store index for reliable sorting
        r._index = index
        return r
      })
    }
  } catch (e) {
    console.error('Failed to load from localStorage:', e)
  }
}

function saveToLocalStorage() {
  try {
    const data = {
      wordsText: wordsText.value,
      wordPromptCurrent: wordPromptCurrent.value,
      trainingPromptCurrent: trainingPromptCurrent.value,
      wordPromptDefault: wordPromptDefault.value,
      trainingPromptDefault: trainingPromptDefault.value,
      wordPromptSource: wordPromptSource.value,
      trainingPromptSource: trainingPromptSource.value,
      diffMode: diffMode.value,
      promptDiffMode: promptDiffMode.value,
      showWordPromptDiff: showWordPromptDiff.value,
      showTrainingPromptDiff: showTrainingPromptDiff.value,
      showResultDiffs: showResultDiffs.value,
      results: results.value
    }
    localStorage.setItem(STORAGE_KEY, JSON.stringify(data))
  } catch (e) {
    console.error('Failed to save to localStorage:', e)
  }
}

async function clearResults() {
  const confirmed = await showConfirm('Are you sure you want to clear all results?')
  if (confirmed) {
    results.value = []
    saveToLocalStorage()
  }
}

async function loadDefaultPrompts() {
  try {
    const data: any = await apiClient.request('/app/admin/prompt-tester/default-prompts')
    wordPromptDefault.value = data.word_prompt || ''
    trainingPromptDefault.value = data.training_prompt || ''
    wordPromptSource.value = data.word_prompt_source || ''
    trainingPromptSource.value = data.training_prompt_source || ''
    
    // Set current prompts if not already set
    if (!wordPromptCurrent.value) {
      wordPromptCurrent.value = wordPromptDefault.value
    }
    if (!trainingPromptCurrent.value) {
      trainingPromptCurrent.value = trainingPromptDefault.value
    }
    
    saveToLocalStorage()
  } catch (error) {
    console.error('Failed to load default prompts:', error)
    await showAlert('Failed to load default prompts')
  }
}

async function resetWordPrompt() {
  try {
    const data: any = await apiClient.request('/app/admin/prompt-tester/default-prompts')
    wordPromptDefault.value = data.word_prompt || ''
    wordPromptSource.value = data.word_prompt_source || ''
    wordPromptCurrent.value = wordPromptDefault.value
    saveToLocalStorage()
  } catch (error) {
    console.error('Failed to load default word prompt:', error)
    await showAlert('Failed to load default word prompt from server')
  }
}

async function resetTrainingPrompt() {
  try {
    const data: any = await apiClient.request('/app/admin/prompt-tester/default-prompts')
    trainingPromptDefault.value = data.training_prompt || ''
    trainingPromptSource.value = data.training_prompt_source || ''
    trainingPromptCurrent.value = trainingPromptDefault.value
    saveToLocalStorage()
  } catch (error) {
    console.error('Failed to load default training prompt:', error)
    await showAlert('Failed to load default training prompt from server')
  }
}

async function runTests() {
  const words = parseWords(wordsText.value)
  if (words.length === 0) {
    await showAlert('Please enter at least one word')
    return
  }

  if (!wordPromptCurrent.value || !trainingPromptCurrent.value) {
    await showAlert('Please ensure both prompts are set')
    return
  }

  isRunning.value = true
  
  // Before clearing, mark existing results as previous by creating a copy
  // We'll handle this in groupedResults computed, but we need to preserve the structure
  // For now, just keep all results - groupedResults will handle previous/current logic

  try {
    const token = apiClient.getAccessToken()
    if (!token) {
      throw new Error('Not authenticated')
    }

    const response = await fetch('/app/admin/prompt-tester/run', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({
        words: words,
        word_prompt: wordPromptCurrent.value,
        training_prompt: trainingPromptCurrent.value
      })
    })

    if (!response.ok) {
      const errorText = await response.text()
      throw new Error(`API error: ${response.status} ${errorText}`)
    }

    // Read NDJSON stream
    const reader = response.body?.getReader()
    if (!reader) {
      throw new Error('Response body is not readable')
    }

    const decoder = new TextDecoder()
    let buffer = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || '' // Keep incomplete line in buffer

      for (const line of lines) {
        if (line.trim()) {
          try {
            const event: PromptTesterEvent = JSON.parse(line)
            // Add timestamp and index to track order
            // Use the current length of results array as the index
            const currentIndex = results.value.length
            event.timestamp = Date.now() + (currentIndex * 0.001)
            event._index = currentIndex
            results.value.push(event)
            saveToLocalStorage()
          } catch (e) {
            console.error('Failed to parse event:', e, line)
          }
        }
      }
    }

    // Process remaining buffer
    if (buffer.trim()) {
      try {
        const event: PromptTesterEvent = JSON.parse(buffer)
        // Add timestamp and index to track order
        // Use the current length of results array as the index
        const currentIndex = results.value.length
        event.timestamp = Date.now() + (currentIndex * 0.001)
        event._index = currentIndex
        results.value.push(event)
        saveToLocalStorage()
      } catch (e) {
        console.error('Failed to parse final event:', e)
      }
    }
  } catch (error: any) {
    console.error('Failed to run tests:', error)
    await showAlert('Failed to run tests: ' + (error.message || 'Unknown error'))
  } finally {
    isRunning.value = false
  }
}

function highlightJSON(obj: any): string {
  const jsonStr = JSON.stringify(obj, null, 2)
  return jsonStr
    .replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, (match) => {
      let cls = 'json-number'
      if (/^"/.test(match)) {
        if (/:$/.test(match)) {
          cls = 'json-key'
        } else {
          cls = 'json-string'
        }
      } else if (/true|false/.test(match)) {
        cls = 'json-boolean'
      } else if (/null/.test(match)) {
        cls = 'json-null'
      }
      return `<span class="${cls}">${escapeHtml(match)}</span>`
    })
}

// Highlight JSON syntax in a single line (for diff)
function highlightJSONLine(text: string): string {
  if (!text) return ''
  
  // First escape HTML to prevent XSS
  const escaped = escapeHtml(text)
  
  // Track positions to avoid double-processing
  const processed: boolean[] = new Array(escaped.length).fill(false)
  let result = escaped
  
  // Helper to check if position is already processed
  const isProcessed = (start: number, end: number): boolean => {
    for (let i = start; i < end; i++) {
      if (processed[i]) return true
    }
    return false
  }
  
  // Helper to mark as processed
  const markProcessed = (start: number, end: number) => {
    for (let i = start; i < end; i++) {
      processed[i] = true
    }
  }
  
  // Process in reverse order to maintain positions
  const replacements: Array<{ start: number; end: number; replacement: string }> = []
  
  // 1. Handle JSON keys (quoted strings followed by colon) - highest priority
  const keyRegex = /("(?:[^"\\]|\\.)*")\s*:/g
  let match
  while ((match = keyRegex.exec(escaped)) !== null) {
    if (!isProcessed(match.index, match.index + match[0].length)) {
      replacements.push({
        start: match.index,
        end: match.index + match[0].length,
        replacement: `<span class="json-key">${match[0]}</span>`
      })
      markProcessed(match.index, match.index + match[0].length)
    }
  }
  
  // 2. Handle JSON string values (quoted strings not already processed)
  const stringRegex = /("(?:[^"\\]|\\.)*")/g
  while ((match = stringRegex.exec(escaped)) !== null) {
    if (!isProcessed(match.index, match.index + match[0].length)) {
      replacements.push({
        start: match.index,
        end: match.index + match[0].length,
        replacement: `<span class="json-string">${match[0]}</span>`
      })
      markProcessed(match.index, match.index + match[0].length)
    }
  }
  
  // 3. Handle booleans
  const boolRegex = /\b(true|false)\b/g
  while ((match = boolRegex.exec(escaped)) !== null) {
    if (!isProcessed(match.index, match.index + match[0].length)) {
      replacements.push({
        start: match.index,
        end: match.index + match[0].length,
        replacement: `<span class="json-boolean">${match[0]}</span>`
      })
      markProcessed(match.index, match.index + match[0].length)
    }
  }
  
  // 4. Handle null
  const nullRegex = /\bnull\b/g
  while ((match = nullRegex.exec(escaped)) !== null) {
    if (!isProcessed(match.index, match.index + match[0].length)) {
      replacements.push({
        start: match.index,
        end: match.index + match[0].length,
        replacement: `<span class="json-null">${match[0]}</span>`
      })
      markProcessed(match.index, match.index + match[0].length)
    }
  }
  
  // 5. Handle numbers (check if not inside strings)
  const numberRegex = /-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?/g
  while ((match = numberRegex.exec(escaped)) !== null) {
    if (!isProcessed(match.index, match.index + match[0].length)) {
      // Check if inside a string
      const before = escaped.substring(0, match.index)
      const quoteCount = (before.match(/"/g) || []).length
      if (quoteCount % 2 === 0) {
        // Not inside a string
        replacements.push({
          start: match.index,
          end: match.index + match[0].length,
          replacement: `<span class="json-number">${match[0]}</span>`
        })
        markProcessed(match.index, match.index + match[0].length)
      }
    }
  }
  
  // Apply replacements in reverse order to maintain positions
  replacements.sort((a, b) => b.start - a.start)
  for (const repl of replacements) {
    result = result.substring(0, repl.start) + repl.replacement + result.substring(repl.end)
  }
  
  return result
}

function escapeHtml(text: string): string {
  const div = document.createElement('div')
  div.textContent = text
  return div.innerHTML
}

// Diff rendering functions
function escapeHtmlForDiff(text: string): string {
  const div = document.createElement('div')
  div.textContent = text
  return div.innerHTML.replace(/\n/g, '<br>')
}

function getResultDiffComponent(result: { previous?: Result; current?: Result }) {
  if (!result.previous || !result.current) {
    return () => null
  }
  
  // Use raw (already formatted JSON) or error message
  const oldText = result.previous.raw || result.previous.error || ''
  const newText = result.current.raw || result.current.error || ''
  
  // Return a function that will be called reactively, similar to prompt diff components
  return () => renderDiffWithJSONHighlight(oldText, newText, diffMode.value)
}

// Render diff with JSON syntax highlighting and character-level diff
function renderDiffWithJSONHighlight(oldText: string, newText: string, mode: string) {
  // Use character-level diff
  const charParts = diffChars(oldText || '', newText || '')
  
  // Helper to render a single character part with JSON highlighting
  const renderCharPart = (part: any, key: number): any => {
    if (part.added || part.removed) {
      // For added/removed parts, apply JSON highlighting and wrap in span with diff class
      const diffClass = part.added ? 'diff-char-add' : 'diff-char-remove'
      const highlighted = highlightJSONLine(part.value)
      // Wrap the highlighted HTML in a span with the diff class
      const wrappedHTML = `<span class="${diffClass}">${highlighted}</span>`
      // Use Vue 3 render function with innerHTML through props
      return h('span', {
        key: `char-${key}`,
        innerHTML: wrappedHTML
      })
    } else {
      // For unchanged parts, just apply JSON highlighting
      const highlighted = highlightJSONLine(part.value)
      return h('span', {
        key: `char-${key}`,
        innerHTML: highlighted
      })
    }
  }
  
  if (mode === 'unified') {
    // For unified mode, process character parts and split by newlines for display
    // Use character-level highlighting only, no line-level background colors
    const lines: any[] = []
    let currentLineParts: any[] = []
    let currentLineHasAdded = false
    let currentLineHasRemoved = false
    
    const flushLine = () => {
      if (currentLineParts.length === 0) return
      
      // Determine sign: only show + or - if line is completely added or removed
      // For mixed changes, show space (character-level highlighting will show the changes)
      let sign = ' '
      const hasUnchanged = currentLineParts.some(p => !p.added && !p.removed)
      if (currentLineHasAdded && !currentLineHasRemoved && !hasUnchanged) {
        // All parts are added, no unchanged parts
        sign = '+'
      } else if (currentLineHasRemoved && !currentLineHasAdded && !hasUnchanged) {
        // All parts are removed, no unchanged parts
        sign = '-'
      }
      
      // Always use diff-context class to avoid line-level background colors
      // Character-level spans will handle the highlighting
      const cls = 'diff-line diff-context'
      
      lines.push({ parts: [...currentLineParts], sign, cls })
      currentLineParts = []
      currentLineHasAdded = false
      currentLineHasRemoved = false
    }
    
    // Process character parts and split by newlines, preserving formatting
    for (const part of charParts) {
      const text = part.value
      let start = 0
      
      while (start < text.length) {
        const newlinePos = text.indexOf('\n', start)
        
        if (newlinePos === -1) {
          // No more newlines, add remaining text to current line
          if (start < text.length) {
            currentLineParts.push({ ...part, value: text.substring(start) })
            if (part.added) currentLineHasAdded = true
            if (part.removed) currentLineHasRemoved = true
          }
          break
        } else {
          // Add text before newline to current line
          if (newlinePos > start) {
            currentLineParts.push({ ...part, value: text.substring(start, newlinePos) })
            if (part.added) currentLineHasAdded = true
            if (part.removed) currentLineHasRemoved = true
          }
          // Flush line (newline is preserved by line structure in DOM)
          flushLine()
          start = newlinePos + 1
        }
      }
    }
    
    // Flush the last line
    flushLine()
    
    return h('div', { class: 'diff-view diff-unified' }, 
      lines.map((line, i) => {
        const highlightedContent = line.parts.map((p: any, idx: number) => renderCharPart(p, idx))
        return h('div', { key: i, class: line.cls }, [
          h('span', { class: 'diff-sign' }, line.sign),
          h('span', { class: 'diff-content' }, highlightedContent)
        ])
      })
    )
  } else {
    // side-by-side mode: use line-level diff to align, then character-level diff within each line
    // Use character-level highlighting only, no line-level background colors
    const oldLines = (oldText || '').split('\n')
    const newLines = (newText || '').split('\n')
    
    // Use simple line matching
    const leftLines: any[] = []
    const rightLines: any[] = []
    
    let oldIdx = 0
    let newIdx = 0
    
    while (oldIdx < oldLines.length || newIdx < newLines.length) {
      if (oldIdx >= oldLines.length) {
        // Only new lines
        const line = newLines[newIdx]
        const charParts = diffChars('', line)
        // Use diff-context class to avoid line-level background, character-level highlighting will show changes
        rightLines.push({ charParts, sign: '+', cls: 'diff-line diff-context' })
        leftLines.push({ charParts: null, sign: ' ', cls: 'diff-line diff-empty' })
        newIdx++
      } else if (newIdx >= newLines.length) {
        // Only old lines
        const line = oldLines[oldIdx]
        const charParts = diffChars(line, '')
        // Use diff-context class to avoid line-level background, character-level highlighting will show changes
        leftLines.push({ charParts, sign: '-', cls: 'diff-line diff-context' })
        rightLines.push({ charParts: null, sign: ' ', cls: 'diff-line diff-empty' })
        oldIdx++
      } else if (oldLines[oldIdx] === newLines[newIdx]) {
        // Identical lines - no character diff needed
        const line = oldLines[oldIdx]
        const charParts = [{ value: line }]
        leftLines.push({ charParts, sign: ' ', cls: 'diff-line diff-context' })
        rightLines.push({ charParts, sign: ' ', cls: 'diff-line diff-context' })
        oldIdx++
        newIdx++
      } else {
        // Different lines - use character-level diff
        const oldLine = oldLines[oldIdx]
        const newLine = newLines[newIdx]
        const allCharParts = diffChars(oldLine, newLine)
        const leftCharParts = allCharParts.filter(p => !p.added)
        const rightCharParts = allCharParts.filter(p => !p.removed)
        
        // Check if line is completely added/removed or partially changed
        const hasRemoved = leftCharParts.some(p => p.removed)
        const hasAdded = rightCharParts.some(p => p.added)
        const hasUnchangedLeft = leftCharParts.some(p => !p.removed && !p.added)
        const hasUnchangedRight = rightCharParts.some(p => !p.removed && !p.added)
        // Only show -/+ if line is completely removed/added (no unchanged parts and only one type of change)
        const leftSign = hasRemoved && !hasAdded && !hasUnchangedLeft ? '-' : ' '
        const rightSign = hasAdded && !hasRemoved && !hasUnchangedRight ? '+' : ' '
        
        // Use diff-context class to avoid line-level background, character-level highlighting will show changes
        leftLines.push({ charParts: leftCharParts, sign: leftSign, cls: 'diff-line diff-context' })
        rightLines.push({ charParts: rightCharParts, sign: rightSign, cls: 'diff-line diff-context' })
        oldIdx++
        newIdx++
      }
    }
    
    return h('div', { class: 'diff-view diff-side-by-side' }, [
      h('div', { class: 'diff-left' },
        leftLines.map((item, i) => {
          if (!item.charParts) {
            return h('div', { key: `left-${i}`, class: item.cls + ' diff-empty' }, [
              h('span', { class: 'diff-sign' }, item.sign),
              h('span', { class: 'diff-content' }, '\u00A0')
            ])
          }
          const highlightedContent = item.charParts.map((p: any, idx: number) => renderCharPart(p, idx))
          return h('div', { key: `left-${i}`, class: item.cls }, [
            h('span', { class: 'diff-sign' }, item.sign),
            h('span', { class: 'diff-content' }, highlightedContent)
          ])
        })
      ),
      h('div', { class: 'diff-right' },
        rightLines.map((item, i) => {
          if (!item.charParts) {
            return h('div', { key: `right-${i}`, class: item.cls + ' diff-empty' }, [
              h('span', { class: 'diff-sign' }, item.sign),
              h('span', { class: 'diff-content' }, '\u00A0')
            ])
          }
          const highlightedContent = item.charParts.map((p: any, idx: number) => renderCharPart(p, idx))
          return h('div', { key: `right-${i}`, class: item.cls }, [
            h('span', { class: 'diff-sign' }, item.sign),
            h('span', { class: 'diff-content' }, highlightedContent)
          ])
        })
      )
    ])
  }
}

function renderDiff(oldText: string, newText: string, mode: string) {
  // Use character-level diff for prompts (similar to renderDiffWithJSONHighlight but without JSON highlighting)
  const charParts = diffChars(oldText || '', newText || '')
  
  // Helper to render a single character part (without JSON highlighting)
  const renderCharPart = (part: any, key: number): any => {
    const escaped = escapeHtml(part.value)
    if (part.added || part.removed) {
      // For added/removed parts, wrap in span with diff class
      const diffClass = part.added ? 'diff-char-add' : 'diff-char-remove'
      const wrappedHTML = `<span class="${diffClass}">${escaped}</span>`
      return h('span', {
        key: `char-${key}`,
        innerHTML: wrappedHTML
      })
    } else {
      // For unchanged parts, just escape HTML
      return h('span', {
        key: `char-${key}`,
        innerHTML: escaped
      })
    }
  }
  
  if (mode === 'unified') {
    // For unified mode, process character parts and split by newlines for display
    // Use character-level highlighting only, no line-level background colors
    const lines: any[] = []
    let currentLineParts: any[] = []
    let currentLineHasAdded = false
    let currentLineHasRemoved = false
    
    const flushLine = () => {
      if (currentLineParts.length === 0) return
      
      // Determine sign: only show + or - if line is completely added or removed
      // For mixed changes, show space (character-level highlighting will show the changes)
      let sign = ' '
      const hasUnchanged = currentLineParts.some(p => !p.added && !p.removed)
      if (currentLineHasAdded && !currentLineHasRemoved && !hasUnchanged) {
        // All parts are added, no unchanged parts
        sign = '+'
      } else if (currentLineHasRemoved && !currentLineHasAdded && !hasUnchanged) {
        // All parts are removed, no unchanged parts
        sign = '-'
      }
      
      // Always use diff-context class to avoid line-level background colors
      // Character-level spans will handle the highlighting
      const cls = 'diff-line diff-context'
      
      lines.push({ parts: [...currentLineParts], sign, cls })
      currentLineParts = []
      currentLineHasAdded = false
      currentLineHasRemoved = false
    }
    
    // Process character parts and split by newlines, preserving formatting
    for (const part of charParts) {
      const text = part.value
      let start = 0
      
      while (start < text.length) {
        const newlinePos = text.indexOf('\n', start)
        
        if (newlinePos === -1) {
          // No more newlines, add remaining text to current line
          if (start < text.length) {
            currentLineParts.push({ ...part, value: text.substring(start) })
            if (part.added) currentLineHasAdded = true
            if (part.removed) currentLineHasRemoved = true
          }
          break
        } else {
          // Add text before newline to current line
          if (newlinePos > start) {
            currentLineParts.push({ ...part, value: text.substring(start, newlinePos) })
            if (part.added) currentLineHasAdded = true
            if (part.removed) currentLineHasRemoved = true
          }
          // Flush line (newline is preserved by line structure in DOM)
          flushLine()
          start = newlinePos + 1
        }
      }
    }
    
    // Flush the last line
    flushLine()
    
    return h('div', { class: 'diff-view diff-unified' }, 
      lines.map((line, i) => {
        const highlightedContent = line.parts.map((p: any, idx: number) => renderCharPart(p, idx))
        return h('div', { key: i, class: line.cls }, [
          h('span', { class: 'diff-sign' }, line.sign),
          h('span', { class: 'diff-content' }, highlightedContent)
        ])
      })
    )
  } else {
    // side-by-side mode: use line-level diff to align, then character-level diff within each line
    // Use character-level highlighting only, no line-level background colors
    const oldLines = (oldText || '').split('\n')
    const newLines = (newText || '').split('\n')
    
    // Use simple line matching
    const leftLines: any[] = []
    const rightLines: any[] = []
    
    let oldIdx = 0
    let newIdx = 0
    
    while (oldIdx < oldLines.length || newIdx < newLines.length) {
      if (oldIdx >= oldLines.length) {
        // Only new lines
        const line = newLines[newIdx]
        const charParts = diffChars('', line)
        // Use diff-context class to avoid line-level background, character-level highlighting will show changes
        rightLines.push({ charParts, sign: '+', cls: 'diff-line diff-context' })
        leftLines.push({ charParts: null, sign: ' ', cls: 'diff-line diff-empty' })
        newIdx++
      } else if (newIdx >= newLines.length) {
        // Only old lines
        const line = oldLines[oldIdx]
        const charParts = diffChars(line, '')
        // Use diff-context class to avoid line-level background, character-level highlighting will show changes
        leftLines.push({ charParts, sign: '-', cls: 'diff-line diff-context' })
        rightLines.push({ charParts: null, sign: ' ', cls: 'diff-line diff-empty' })
        oldIdx++
      } else if (oldLines[oldIdx] === newLines[newIdx]) {
        // Identical lines - no character diff needed
        const line = oldLines[oldIdx]
        const charParts = [{ value: line }]
        leftLines.push({ charParts, sign: ' ', cls: 'diff-line diff-context' })
        rightLines.push({ charParts, sign: ' ', cls: 'diff-line diff-context' })
        oldIdx++
        newIdx++
      } else {
        // Different lines - use character-level diff
        const oldLine = oldLines[oldIdx]
        const newLine = newLines[newIdx]
        const allCharParts = diffChars(oldLine, newLine)
        const leftCharParts = allCharParts.filter(p => !p.added)
        const rightCharParts = allCharParts.filter(p => !p.removed)
        
        // Check if line is completely added/removed or partially changed
        const hasRemoved = leftCharParts.some(p => p.removed)
        const hasAdded = rightCharParts.some(p => p.added)
        const hasUnchangedLeft = leftCharParts.some(p => !p.removed && !p.added)
        const hasUnchangedRight = rightCharParts.some(p => !p.removed && !p.added)
        // Only show -/+ if line is completely removed/added (no unchanged parts and only one type of change)
        const leftSign = hasRemoved && !hasAdded && !hasUnchangedLeft ? '-' : ' '
        const rightSign = hasAdded && !hasRemoved && !hasUnchangedRight ? '+' : ' '
        
        // Use diff-context class to avoid line-level background, character-level highlighting will show changes
        leftLines.push({ charParts: leftCharParts, sign: leftSign, cls: 'diff-line diff-context' })
        rightLines.push({ charParts: rightCharParts, sign: rightSign, cls: 'diff-line diff-context' })
        oldIdx++
        newIdx++
      }
    }
    
    return h('div', { class: 'diff-view diff-side-by-side' }, [
      h('div', { class: 'diff-left' },
        leftLines.map((item, i) => {
          if (!item.charParts) {
            return h('div', { key: `left-${i}`, class: item.cls + ' diff-empty' }, [
              h('span', { class: 'diff-sign' }, item.sign),
              h('span', { class: 'diff-content' }, '\u00A0')
            ])
          }
          const highlightedContent = item.charParts.map((p: any, idx: number) => renderCharPart(p, idx))
          return h('div', { key: `left-${i}`, class: item.cls }, [
            h('span', { class: 'diff-sign' }, item.sign),
            h('span', { class: 'diff-content' }, highlightedContent)
          ])
        })
      ),
      h('div', { class: 'diff-right' },
        rightLines.map((item, i) => {
          if (!item.charParts) {
            return h('div', { key: `right-${i}`, class: item.cls + ' diff-empty' }, [
              h('span', { class: 'diff-sign' }, item.sign),
              h('span', { class: 'diff-content' }, '\u00A0')
            ])
          }
          const highlightedContent = item.charParts.map((p: any, idx: number) => renderCharPart(p, idx))
          return h('div', { key: `right-${i}`, class: item.cls }, [
            h('span', { class: 'diff-sign' }, item.sign),
            h('span', { class: 'diff-content' }, highlightedContent)
          ])
        })
      )
    ])
  }
}

// Lifecycle
onMounted(async () => {
  loadFromLocalStorage()
  await loadDefaultPrompts()
})
</script>

<style scoped>
.admin-layout {
  display: flex;
  gap: 20px;
  min-height: calc(100vh - 60px);
  font-size: 16px;
}

.admin-content {
  flex: 1;
  max-width: 1200px;
  margin: 0 auto;
  padding: 10px;
  width: 100%;
}

.admin-content h1 {
  margin-bottom: 24px;
}

@media (max-width: 767px) {
  .admin-layout {
    flex-direction: column;
    gap: 0;
  }
  
  .admin-content {
    padding: 10px;
    margin-top: 60px;
  }
}

@media (min-width: 768px) {
  .admin-layout {
    padding: 20px;
  }
  
  .admin-content {
    padding: 0;
  }
}

.prompt-tester h2 {
  margin-bottom: 24px;
}

.card {
  background: var(--card-bg);
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;
}

.card h2 {
  margin-top: 0;
  margin-bottom: 16px;
}

.words-input {
  width: 100%;
  padding: 12px;
  border: 1px solid var(--input-border);
  border-radius: 4px;
  font-size: 14px;
  font-family: monospace;
  background-color: var(--input-bg);
  color: var(--text-primary);
  resize: vertical;
  box-sizing: border-box;
}

.words-info {
  margin-top: 8px;
  color: var(--text-secondary);
  font-size: 14px;
}

.prompt-section {
  margin-bottom: 20px;
  border: 1px solid var(--border-primary);
  border-radius: 4px;
  padding: 12px;
}

.prompt-section summary {
  cursor: pointer;
  font-weight: 600;
  padding: 8px;
  user-select: none;
}

.prompt-source {
  color: var(--text-secondary);
  font-weight: normal;
  font-size: 0.9em;
}

.prompt-modified {
  color: var(--color-warning, #f59e0b);
  font-size: 0.9em;
  margin-left: 8px;
}

.prompt-editor {
  margin-top: 12px;
}

.prompt-textarea {
  width: 100%;
  padding: 12px;
  border: 1px solid var(--input-border);
  border-radius: 4px;
  font-size: 13px;
  font-family: monospace;
  background-color: var(--input-bg);
  color: var(--text-primary);
  resize: vertical;
  box-sizing: border-box;
  line-height: 1.5;
}

.prompt-actions {
  margin-top: 12px;
}

.prompt-diff {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--border-primary);
}

.prompt-diff-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.prompt-diff-controls {
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
}

.prompt-diff-header h4 {
  margin: 0;
  font-size: 14px;
}

.prompt-diff-mode-selector {
  display: flex;
  gap: 8px;
  align-items: center;
}

.prompt-diff-mode-selector label {
  font-size: 12px;
  color: var(--text-secondary);
}

.controls {
  display: flex;
  gap: 20px;
  align-items: center;
  flex-wrap: wrap;
}

.diff-mode-selector {
  display: flex;
  gap: 8px;
  align-items: center;
}

.prompt-diff-mode-selector {
  display: flex;
  gap: 8px;
  align-items: center;
}

.toggle-switch {
  display: inline-flex;
  background: var(--bg-secondary);
  border: 1px solid var(--border-primary);
  border-radius: 6px;
  padding: 2px;
  gap: 2px;
}

.toggle-option {
  padding: 6px 16px;
  border: none;
  border-radius: 4px;
  background: transparent;
  color: var(--text-secondary);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.toggle-option:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.toggle-option.active {
  background: var(--color-primary);
  color: white;
  font-weight: 500;
}

.toggle-option.active:hover {
  background: var(--color-primary);
  opacity: 0.9;
}

.results-container {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.word-result {
  border: 1px solid var(--border-primary);
  border-radius: 4px;
  padding: 16px;
  background: var(--bg-secondary);
}

.word-result h3 {
  margin-top: 0;
  margin-bottom: 16px;
  color: var(--color-primary);
}

.step-result {
  margin-bottom: 20px;
  padding-bottom: 20px;
  border-bottom: 1px solid var(--border-primary);
}

.step-result:last-child {
  border-bottom: none;
  margin-bottom: 0;
  padding-bottom: 0;
}

.step-result h4 {
  margin-top: 0;
  margin-bottom: 12px;
  font-size: 16px;
}

.result-success {
  background: rgba(40, 167, 69, 0.1);
  border: 1px solid rgba(40, 167, 69, 0.3);
  border-radius: 4px;
  padding: 12px;
}

.result-error {
  background: rgba(220, 53, 69, 0.1);
  border: 1px solid rgba(220, 53, 69, 0.3);
  border-radius: 4px;
  padding: 12px;
  color: var(--color-danger);
}

.json-viewer {
  background: var(--bg-primary);
  border: 1px solid var(--border-primary);
  border-radius: 4px;
  padding: 12px;
  overflow-x: auto;
}

.json-viewer pre {
  margin: 0;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.5;
}

.json-viewer :deep(.json-key) {
  color: #268bd2;
}

.json-viewer :deep(.json-string) {
  color: #2aa198;
}

.json-viewer :deep(.json-number) {
  color: #d33682;
}

.json-viewer :deep(.json-boolean) {
  color: #b58900;
}

.json-viewer :deep(.json-null) {
  color: #859900;
}

.raw-response {
  background: var(--bg-primary);
  border: 1px solid var(--border-primary);
  border-radius: 4px;
  padding: 12px;
  overflow-x: auto;
}

.raw-response pre {
  margin: 0;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-word;
}

.result-meta {
  margin-top: 8px;
  font-size: 12px;
  color: var(--text-secondary);
}

.result-diff {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--border-primary);
}

.result-diff h5 {
  margin-top: 0;
  margin-bottom: 12px;
  font-size: 14px;
}

.result-diff-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.result-diff-header h5 {
  margin: 0;
}

.result-diff-toggle {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid var(--border-primary);
}

.btn-small {
  padding: 6px 12px;
  font-size: 12px;
}

.diff-container {
  border: 1px solid var(--border-primary);
  border-radius: 4px;
  overflow: hidden;
}

.diff-view {
  font-family: 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.5;
}

.diff-unified {
  background: var(--bg-primary);
}

.diff-side-by-side {
  display: grid;
  grid-template-columns: 1fr 1fr;
  background: var(--bg-primary);
}

.diff-left,
.diff-right {
  border-right: 1px solid var(--border-primary);
}

.diff-right {
  border-right: none;
}

.diff-line {
  display: flex;
  padding: 2px 8px;
  white-space: pre-wrap;
  word-break: break-word;
}

.diff-sign {
  display: inline-block;
  width: 20px;
  text-align: center;
  user-select: none;
  flex-shrink: 0;
}

.diff-content {
  flex: 1;
  white-space: pre-wrap;
  word-break: break-word;
}

.diff-line.diff-add {
  background: rgba(40, 167, 69, 0.25) !important;
  border-left: 3px solid #28a745 !important;
}

.diff-line.diff-add .diff-sign {
  color: #28a745 !important;
  font-weight: bold;
}

.diff-line.diff-add .diff-content {
  color: var(--text-primary);
}

.diff-line.diff-remove {
  background: rgba(220, 53, 69, 0.25) !important;
  border-left: 3px solid #dc3545 !important;
}

.diff-line.diff-remove .diff-sign {
  color: #dc3545 !important;
  font-weight: bold;
}

.diff-line.diff-remove .diff-content {
  color: var(--text-primary);
}

/* Ensure styles apply to diff inside prompt-diff */
.prompt-diff :deep(.diff-line.diff-add) {
  background: rgba(40, 167, 69, 0.25) !important;
  border-left: 3px solid #28a745 !important;
}

.prompt-diff :deep(.diff-line.diff-add .diff-sign) {
  color: #28a745 !important;
  font-weight: bold;
}

.prompt-diff :deep(.diff-line.diff-remove) {
  background: rgba(220, 53, 69, 0.25) !important;
  border-left: 3px solid #dc3545 !important;
}

.prompt-diff :deep(.diff-line.diff-remove .diff-sign) {
  color: #dc3545 !important;
  font-weight: bold;
}

/* Additional styles to ensure visibility */
.prompt-diff .diff-container :deep(.diff-line.diff-add) {
  background-color: rgba(40, 167, 69, 0.25) !important;
  border-left: 3px solid #28a745 !important;
}

.prompt-diff .diff-container :deep(.diff-line.diff-add .diff-sign) {
  color: #28a745 !important;
  font-weight: bold !important;
}

.prompt-diff .diff-container :deep(.diff-line.diff-remove) {
  background-color: rgba(220, 53, 69, 0.25) !important;
  border-left: 3px solid #dc3545 !important;
}

.prompt-diff .diff-container :deep(.diff-line.diff-remove .diff-sign) {
  color: #dc3545 !important;
  font-weight: bold !important;
}

/* Ensure styles apply to diff inside result-diff */
.result-diff :deep(.diff-line.diff-add) {
  background: rgba(40, 167, 69, 0.25) !important;
  border-left: 3px solid #28a745 !important;
}

.result-diff :deep(.diff-line.diff-add .diff-sign) {
  color: #28a745 !important;
  font-weight: bold;
}

.result-diff :deep(.diff-line.diff-remove) {
  background: rgba(220, 53, 69, 0.25) !important;
  border-left: 3px solid #dc3545 !important;
}

.result-diff :deep(.diff-line.diff-remove .diff-sign) {
  color: #dc3545 !important;
  font-weight: bold;
}

.result-diff .diff-container :deep(.diff-line.diff-add) {
  background-color: rgba(40, 167, 69, 0.25) !important;
  border-left: 3px solid #28a745 !important;
}

.result-diff .diff-container :deep(.diff-line.diff-add .diff-sign) {
  color: #28a745 !important;
  font-weight: bold !important;
}

.result-diff .diff-container :deep(.diff-line.diff-remove) {
  background-color: rgba(220, 53, 69, 0.25) !important;
  border-left: 3px solid #dc3545 !important;
}

.result-diff .diff-container :deep(.diff-line.diff-remove .diff-sign) {
  color: #dc3545 !important;
  font-weight: bold !important;
}

/* Character-level diff highlighting */
.diff-content :deep(.diff-char-add),
.diff-content .diff-char-add,
.diff-view .diff-char-add,
.diff-view :deep(.diff-char-add),
.result-diff .diff-char-add,
.result-diff :deep(.diff-char-add) {
  background: rgba(40, 167, 69, 0.7) !important;
  color: inherit !important;
  padding: 2px 1px !important;
  display: inline !important;
  font-weight: 500 !important;
}

.diff-content :deep(.diff-char-remove),
.diff-content .diff-char-remove,
.diff-view .diff-char-remove,
.diff-view :deep(.diff-char-remove),
.result-diff .diff-char-remove,
.result-diff :deep(.diff-char-remove) {
  background: rgba(220, 53, 69, 0.7) !important;
  color: inherit !important;
  padding: 2px 1px !important;
  display: inline !important;
  font-weight: 500 !important;
}

/* Ensure character-level highlighting works inside JSON syntax highlighting */
.diff-content :deep(.diff-char-add .json-key),
.diff-content :deep(.diff-char-add .json-string),
.diff-content :deep(.diff-char-add .json-number),
.diff-content :deep(.diff-char-add .json-boolean),
.diff-content :deep(.diff-char-add .json-null),
.diff-content :deep(.diff-char-remove .json-key),
.diff-content :deep(.diff-char-remove .json-string),
.diff-content :deep(.diff-char-remove .json-number),
.diff-content :deep(.diff-char-remove .json-boolean),
.diff-content :deep(.diff-char-remove .json-null) {
  background: inherit !important;
}

.diff-line.diff-context {
  background: var(--bg-primary);
}

.diff-line.diff-change {
  background: rgba(255, 193, 7, 0.15);
  border-left: 3px solid #ffc107;
}

.diff-line.diff-empty {
  background: var(--bg-primary);
  min-height: 1.5em;
}

.btn {
  padding: 10px 20px;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: var(--color-primary);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  opacity: 0.9;
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-secondary {
  background: var(--bg-secondary);
  color: var(--text-primary);
  border: 1px solid var(--border-primary);
}

.btn-secondary:hover {
  background: var(--bg-hover);
}

/* Responsive styles */
@media (max-width: 768px) {
  .prompt-tester {
    padding: 10px;
  }

  .controls {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }

  .controls .btn {
    width: 100%;
  }

  .diff-mode-selector {
    width: 100%;
    justify-content: center;
  }

  .prompt-diff-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .prompt-diff-controls {
    flex-direction: column;
    align-items: flex-start;
    width: 100%;
    gap: 8px;
  }

  .prompt-diff-mode-selector {
    width: 100%;
    flex-wrap: wrap;
  }

  .result-diff-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }

  .result-diff-header h5 {
    margin-bottom: 0;
  }

  .diff-side-by-side {
    grid-template-columns: 1fr;
  }

  .diff-left,
  .diff-right {
    border-right: none;
    border-bottom: 1px solid var(--border-primary);
  }

  .diff-right {
    border-bottom: none;
  }

  .admin-tab {
    padding: 10px 16px;
  }

  .card {
    padding: 16px;
  }

  .prompt-section {
    padding: 10px;
  }

  .word-result {
    padding: 12px;
  }
}

@media (max-width: 480px) {
  .admin-tab {
    padding: 8px 12px;
  }

  .card {
    padding: 12px;
  }

  .toggle-option {
    padding: 6px 12px;
    font-size: 12px;
  }

  .prompt-textarea {
    font-size: 12px;
  }

  .words-input {
    font-size: 13px;
  }
}
</style>
