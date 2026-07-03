<template>
  <div ref="rootRef" class="lg-lang">
    <button class="lg-lang-btn" type="button" @click="open = !open">
      {{ currentLabel }} <span class="lg-lang-caret">▾</span>
    </button>
    <div v-if="open" class="lg-lang-dropdown">
      <button
        v-for="l in locales"
        :key="l.code"
        type="button"
        class="lg-lang-item"
        :class="{ 'lg-lang-item--on': l.code === currentLocale }"
        @click="pick(l.code)"
      >
        {{ l.label }}
        <span v-if="l.code === currentLocale" class="lg-lang-check"><LgIcon name="check" :s="14" /></span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { AVAILABLE_LOCALES, type SupportedLocale } from '../../i18n'
import { useLocale } from '../../composables/useLocale'
import LgIcon from './LgIcon.vue'

const { currentLocale, setLocale } = useLocale()

const locales = AVAILABLE_LOCALES
const open = ref(false)
const rootRef = ref<HTMLElement | null>(null)

const currentLabel = computed(
  () => locales.find(l => l.code === currentLocale.value)?.code.toUpperCase() || 'RU',
)

const pick = async (code: SupportedLocale) => {
  open.value = false
  if (code !== currentLocale.value) await setLocale(code)
}

const close = (e: MouseEvent) => {
  if (rootRef.value && !rootRef.value.contains(e.target as Node)) open.value = false
}
onMounted(() => document.addEventListener('mousedown', close))
onUnmounted(() => document.removeEventListener('mousedown', close))
</script>

<style scoped>
.lg-lang { position: relative; }
.lg-lang-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: none;
  border: none;
  cursor: pointer;
  font-size: 14px;
  font-weight: 600;
  color: var(--text);
  padding: 4px 6px;
}
.lg-lang-caret { font-size: 9px; }
.lg-lang-dropdown {
  position: absolute;
  top: 130%;
  right: 0;
  background: var(--card-bg);
  border: 1px solid var(--border);
  border-radius: 12px;
  overflow: hidden;
  z-index: 300;
  min-width: 130px;
  box-shadow: var(--shadow-card);
}
.lg-lang-item {
  width: 100%;
  padding: 10px 14px;
  border: none;
  background: none;
  cursor: pointer;
  text-align: left;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  font-size: 13px;
  color: var(--subtext);
  font-weight: 400;
  border-bottom: 1px solid var(--border);
  white-space: nowrap;
}
.lg-lang-item:last-child { border-bottom: none; }
.lg-lang-item--on { color: var(--text); font-weight: 600; }
.lg-lang-check { color: var(--dorado); font-size: 11px; }
</style>
