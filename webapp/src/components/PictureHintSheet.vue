<template>
  <Teleport to="body">
    <div class="hint-overlay" @click.self="$emit('close')">
      <div class="hint-sheet" role="dialog" aria-modal="true">
        <header class="hint-head">
          <h3 class="hint-title">{{ sheetTitle }}</h3>
          <button class="hint-close" type="button" :aria-label="'close'" @click="$emit('close')">
            <LgIcon name="x" :s="18" />
          </button>
        </header>

        <div v-if="!sections.length" class="hint-empty">-</div>
        <div v-else class="hint-body">
          <section v-for="s in sections" :key="s.key" class="hint-section">
            <div class="hint-section-head">
              <span class="hint-section-icon"><LgIcon :name="s.icon" :s="16" /></span>
              <span class="hint-section-title">{{ s.title }}</span>
            </div>

            <div v-if="s.key === 'location'" class="location-card">
              <div class="location-scene" aria-hidden="true">
                <div class="location-frame">
                  <span class="location-object location-object--left"></span>
                  <span class="location-object location-object--center"></span>
                  <span class="location-object location-object--right"></span>
                  <span class="location-label location-label--top">{{ locationEntry(s.entries, 'top').term }}</span>
                  <span class="location-label location-label--bottom">{{ locationEntry(s.entries, 'bottom').term }}</span>
                  <span class="location-label location-label--left">{{ locationEntry(s.entries, 'left').term }}</span>
                  <span class="location-label location-label--right">{{ locationEntry(s.entries, 'right').term }}</span>
                  <span class="location-label location-label--center">{{ locationEntry(s.entries, 'center').term }}</span>
                  <span v-if="locationEntry(s.entries, 'foreground').term" class="location-label location-label--foreground">{{ locationEntry(s.entries, 'foreground').term }}</span>
                  <span v-if="locationEntry(s.entries, 'background').term" class="location-label location-label--background">{{ locationEntry(s.entries, 'background').term }}</span>
                </div>
              </div>
              <div class="location-pairs">
                <div v-for="item in locationSummary(s.entries)" :key="item.key" class="location-pair">
                  <span>{{ item.entry.term }}</span>
                  <small>{{ item.entry.gloss }}</small>
                </div>
              </div>
            </div>

            <div v-else-if="s.key === 'colors'" class="color-palette">
              <div
                v-for="(e, i) in s.entries"
                :key="i"
                class="color-swatch"
                :style="{ '--swatch': colorValue(e.term), '--swatch-border': colorBorder(e.term) }"
              >
                <span class="color-sample"></span>
                <span class="color-name">{{ e.term }}</span>
                <span class="color-gloss">{{ e.gloss }}</span>
              </div>
            </div>

            <div v-else class="hint-grid">
              <div v-for="(e, i) in s.entries" :key="i" class="hint-chip">
                <span class="hint-term">{{ e.term }}</span>
                <span class="hint-gloss">{{ e.gloss }}</span>
              </div>
            </div>
          </section>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import LgIcon from './linglow/LgIcon.vue'
import { getHintSections, getHintSheetTitle, type HintEntry } from '../constants/pictureHintPhrasebook'

const props = defineProps<{ targetLang: string; nativeLang: string }>()
defineEmits<{ (e: 'close'): void }>()

const sections = computed(() => getHintSections(props.targetLang, props.nativeLang))
const sheetTitle = computed(() => getHintSheetTitle(props.nativeLang))

const emptyEntry: HintEntry = { term: '', gloss: '' }

const LOCATION_TERMS: Record<string, string[]> = {
  top: ['arriba', 'above'],
  bottom: ['abajo', 'below'],
  left: ['a la izquierda', 'on the left'],
  right: ['a la derecha', 'on the right'],
  center: ['en el centro', 'in the middle'],
  foreground: ['en primer plano', 'in the foreground'],
  background: ['en el fondo', 'in the background'],
  front: ['delante de', 'in front of'],
  behind: ['detrás de', 'behind'],
  next: ['al lado de', 'next to'],
  between: ['entre', 'between'],
  inside: ['dentro de', 'inside'],
}

const COLOR_VALUES: Record<string, string> = {
  rojo: '#d94b45',
  red: '#d94b45',
  azul: '#2f72d8',
  blue: '#2f72d8',
  verde: '#3f9f5a',
  green: '#3f9f5a',
  amarillo: '#f0c84b',
  yellow: '#f0c84b',
  naranja: '#e8893d',
  orange: '#e8893d',
  morado: '#8a5bd6',
  purple: '#8a5bd6',
  rosa: '#e983a9',
  pink: '#e983a9',
  marrón: '#8a5a3b',
  brown: '#8a5a3b',
  negro: '#222222',
  black: '#222222',
  blanco: '#ffffff',
  white: '#ffffff',
  gris: '#8d949e',
  grey: '#8d949e',
  claro: 'linear-gradient(135deg, #ffffff 0%, #f3ead5 100%)',
  light: 'linear-gradient(135deg, #ffffff 0%, #f3ead5 100%)',
  oscuro: 'linear-gradient(135deg, #1f2933 0%, #5b6470 100%)',
  dark: 'linear-gradient(135deg, #1f2933 0%, #5b6470 100%)',
}

function locationEntry(entries: HintEntry[], key: string): HintEntry {
  const terms = LOCATION_TERMS[key] || []
  return entries.find(e => terms.includes(e.term.toLowerCase())) || emptyEntry
}

function locationSummary(entries: HintEntry[]) {
  return ['top', 'bottom', 'left', 'right', 'center', 'front', 'behind', 'next', 'between', 'inside']
    .map(key => ({ key, entry: locationEntry(entries, key) }))
    .filter(item => item.entry.term)
}

function colorValue(term: string): string {
  return COLOR_VALUES[term.toLowerCase()] || '#d7dce2'
}

function colorBorder(term: string): string {
  const normalized = term.toLowerCase()
  return normalized === 'white' || normalized === 'blanco' || normalized === 'light' || normalized === 'claro'
    ? 'rgba(0,0,0,0.16)'
    : 'transparent'
}
</script>

<style scoped>
.hint-overlay {
  --bottom-nav-space: calc(60px + max(env(safe-area-inset-bottom, 0px), var(--android-inset-bottom, 0px)));
  position: fixed;
  top: 0; right: 0; left: 0; bottom: var(--bottom-nav-space);
  z-index: 210;
  background: rgba(0,0,0,0.5);
  display: flex; align-items: flex-end; justify-content: center;
}
@media (min-width: 640px) {
  .hint-overlay { inset: 0; align-items: center; }
}
.hint-sheet {
  width: 100%; max-width: 560px; max-height: min(88vh, calc(100dvh - var(--bottom-nav-space) - 12px));
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

.location-card {
  display: grid;
  grid-template-columns: minmax(0, 1.1fr) minmax(150px, 0.9fr);
  gap: 12px;
  align-items: stretch;
}
.location-scene {
  min-height: 190px;
  padding: 22px;
  border-radius: 14px;
  background:
    linear-gradient(180deg, rgba(75, 136, 194, 0.10), rgba(45, 107, 58, 0.08)),
    var(--card-bg);
  border: 1px solid var(--border, rgba(0,0,0,0.08));
}
.location-frame {
  position: relative;
  height: 100%;
  min-height: 146px;
  border: 2px solid rgba(45,107,58,0.35);
  border-radius: 12px;
  background:
    linear-gradient(90deg, transparent calc(50% - 1px), rgba(45,107,58,0.15) calc(50% - 1px), rgba(45,107,58,0.15) calc(50% + 1px), transparent calc(50% + 1px)),
    linear-gradient(180deg, transparent calc(50% - 1px), rgba(45,107,58,0.15) calc(50% - 1px), rgba(45,107,58,0.15) calc(50% + 1px), transparent calc(50% + 1px));
}
.location-object {
  position: absolute;
  display: block;
  border-radius: 50%;
  background: #e8a24c;
  box-shadow: 0 8px 18px rgba(0,0,0,0.16);
}
.location-object--left { width: 28px; height: 28px; left: 16%; top: 50%; transform: translateY(-50%); }
.location-object--center { width: 38px; height: 38px; left: 50%; top: 50%; transform: translate(-50%, -50%); background: #4c88c2; }
.location-object--right { width: 24px; height: 24px; right: 16%; top: 50%; transform: translateY(-50%); background: #ca5f5b; }
.location-label {
  position: absolute;
  max-width: 42%;
  padding: 4px 7px;
  border-radius: 999px;
  background: var(--bg);
  border: 1px solid var(--border, rgba(0,0,0,0.08));
  color: #2d6b3a;
  font-family: 'Inter', sans-serif;
  font-size: 11px;
  font-weight: 800;
  line-height: 1.1;
  text-align: center;
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
}
.location-label--top { top: -14px; left: 50%; transform: translateX(-50%); }
.location-label--bottom { bottom: -14px; left: 50%; transform: translateX(-50%); }
.location-label--left { left: -16px; top: 50%; transform: translateY(-50%); }
.location-label--right { right: -16px; top: 50%; transform: translateY(-50%); }
.location-label--center { left: 50%; top: 50%; transform: translate(-50%, -50%); color: #1f527b; }
.location-label--foreground { right: 8px; bottom: 10px; color: #8b5a19; }
.location-label--background { left: 8px; top: 10px; color: #59636f; }
.location-pairs {
  display: grid;
  gap: 6px;
}
.location-pair {
  display: flex;
  flex-direction: column;
  gap: 1px;
  min-width: 0;
  padding: 7px 9px;
  border-radius: 10px;
  background: var(--card-bg);
  border: 1px solid var(--border, rgba(0,0,0,0.06));
}
.location-pair span {
  color: #2d6b3a;
  font-family: 'Inter', sans-serif;
  font-size: 13px;
  font-weight: 800;
  line-height: 1.2;
}
.location-pair small {
  color: var(--subtext);
  font-family: 'Inter', sans-serif;
  font-size: 11px;
  line-height: 1.2;
}

.color-palette {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(116px, 1fr));
  gap: 9px;
}
.color-swatch {
  min-width: 0;
  padding: 8px;
  border-radius: 12px;
  background: var(--card-bg);
  border: 1px solid var(--border, rgba(0,0,0,0.06));
}
.color-sample {
  display: block;
  height: 38px;
  border-radius: 9px;
  background: var(--swatch);
  border: 1px solid var(--swatch-border);
  box-shadow: inset 0 -10px 18px rgba(0,0,0,0.08);
}
.color-name,
.color-gloss {
  display: block;
  overflow-wrap: anywhere;
  font-family: 'Inter', sans-serif;
}
.color-name {
  margin-top: 6px;
  color: #2d6b3a;
  font-size: 13px;
  font-weight: 800;
  line-height: 1.2;
}
.color-gloss {
  margin-top: 1px;
  color: var(--subtext);
  font-size: 11px;
  line-height: 1.2;
}

@media (max-width: 520px) {
  .hint-body { padding: 12px 14px 18px; gap: 16px; }
  .location-card { grid-template-columns: 1fr; }
  .location-scene { min-height: 168px; padding: 20px; }
  .location-frame { min-height: 126px; }
  .location-pairs { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .color-palette { grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 8px; }
}
</style>
