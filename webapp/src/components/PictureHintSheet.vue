<template>
  <div class="hint-overlay" @click.self="$emit('close')">
    <div class="hint-sheet" role="dialog" aria-modal="true">
      <header class="hint-head">
        <h3 class="hint-title">{{ sheetTitle }}</h3>
        <button class="hint-close" type="button" :aria-label="'close'" @click="$emit('close')">
          <LgIcon name="x" :s="18" />
        </button>
      </header>

      <div v-if="!sections.length" class="hint-empty">—</div>
      <div v-else class="hint-body">
        <section v-for="s in sections" :key="s.key" class="hint-section">
          <div class="hint-section-head">
            <span class="hint-section-icon"><LgIcon :name="s.icon" :s="16" /></span>
            <span class="hint-section-title">{{ s.title }}</span>
          </div>
          <div class="hint-grid">
            <div v-for="(e, i) in s.entries" :key="i" class="hint-chip">
              <span class="hint-term">{{ e.term }}</span>
              <span class="hint-gloss">{{ e.gloss }}</span>
            </div>
          </div>
        </section>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import LgIcon from './linglow/LgIcon.vue'
import { getHintSections, getHintSheetTitle } from '../constants/pictureHintPhrasebook'

const props = defineProps<{ targetLang: string; nativeLang: string }>()
defineEmits<{ (e: 'close'): void }>()

const sections = computed(() => getHintSections(props.targetLang, props.nativeLang))
const sheetTitle = computed(() => getHintSheetTitle(props.nativeLang))
</script>

<style scoped>
.hint-overlay {
  position: fixed; inset: 0; z-index: 210;
  background: rgba(0,0,0,0.5);
  display: flex; align-items: flex-end; justify-content: center;
}
@media (min-width: 640px) {
  .hint-overlay { align-items: center; }
}
.hint-sheet {
  width: 100%; max-width: 560px; max-height: 88vh;
  display: flex; flex-direction: column;
  background: var(--bg, #fff);
  border-radius: 20px 20px 0 0;
  overflow: hidden;
}
@media (min-width: 640px) {
  .hint-sheet { border-radius: 20px; max-height: 82vh; }
}
.hint-head {
  display: flex; align-items: center; justify-content: space-between;
  padding: 16px 18px 12px; flex-shrink: 0;
  border-bottom: 1px solid var(--border, rgba(0,0,0,0.08));
}
.hint-title {
  margin: 0; font-family: 'Lora', serif; font-size: 18px; font-weight: 700; color: var(--text);
}
.hint-close {
  display: inline-flex; align-items: center; justify-content: center;
  width: 32px; height: 32px; border-radius: 50%; border: none;
  background: var(--card-bg); color: var(--text); cursor: pointer;
}
.hint-body { overflow-y: auto; padding: 14px 18px 24px; display: flex; flex-direction: column; gap: 18px; }
.hint-empty { padding: 40px; text-align: center; color: var(--subtext); }
.hint-section-head { display: flex; align-items: center; gap: 8px; margin-bottom: 10px; }
.hint-section-icon { display: inline-flex; color: #2d6b3a; }
.hint-section-title {
  font-family: 'Inter', sans-serif; font-size: 13px; font-weight: 700;
  text-transform: uppercase; letter-spacing: 0.04em; color: var(--subtext);
}
.hint-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(150px, 1fr)); gap: 8px; }
.hint-chip {
  display: flex; flex-direction: column; gap: 1px;
  padding: 8px 10px; border-radius: 12px;
  background: var(--card-bg); border: 1px solid var(--border, rgba(0,0,0,0.06));
}
.hint-term { font-family: 'Inter', sans-serif; font-size: 14px; font-weight: 700; color: #2d6b3a; }
.hint-gloss { font-family: 'Inter', sans-serif; font-size: 12px; color: var(--subtext); }
</style>
