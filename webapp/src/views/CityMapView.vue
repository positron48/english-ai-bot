<template>
  <div class="city-map-root">
    <!-- Title overlay -->
    <div class="city-map-title-bar">
      <div class="city-map-title">{{ cityName }}</div>
      <div class="city-map-sub">{{ t('city.mapSub') }}</div>
    </div>

    <!-- Canvas -->
    <canvas ref="cvsRef" class="city-map-canvas" @click="handleCanvasClick" @mousemove="handleCanvasHover" :style="{ cursor: canvasCursor }" />

    <!-- District label overlays -->
    <div class="city-map-overlays">
      <div
        v-for="d in CITY_DISTS"
        :key="d.id"
        class="district-label"
        :class="{ 'district-label--locked': d.lv <= 1, 'district-label--wrap': d.wrap }"
        :style="{ left: d.lp[0] + '%', top: d.lp[1] + '%' }"
        @click="openDistrict(d)"
      >
        <span class="district-name">{{ d.name }}</span>
        <div v-if="d.lv <= 1" class="district-lock">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none"
            stroke="rgba(44,31,14,0.38)" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <rect x="3" y="11" width="18" height="11" rx="2" />
            <path d="M7 11V7a5 5 0 0110 0v4" />
          </svg>
        </div>
        <div v-else class="district-acts">
          <div
            v-for="(a, ai) in d.acts"
            :key="ai"
            class="district-act"
            :style="{ background: ACT_BG[a.s] || ACT_BG.started }"
          >{{ a.i }}</div>
        </div>
      </div>
    </div>

  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useCourse } from '../composables/useCourse'
import { grammarClient } from '../api/grammarClient'

const { t } = useI18n()
const router = useRouter()
const { currentCourse } = useCourse()

const cityName = computed(() => currentCourse.value?.city_name || currentCourse.value?.title || 'Ciudad Luminaria')

// ─── District definitions (static geometry, lv computed from grammar progress) ─
// lv: 1=locked, 2=unlocked/0 chapters, 3=started, 4=half+, 5=all passed
const DIST_DEFS = [
  { id: 'b2_high_district',  cefrLevel: 'B2', name: 'Distrito Alto',       lp: [23.9, 13.7], wrap: false,
    poly: [[0,0],[57,0],[57.29,13.59],[36.17,19.67],[34.58,23],[21.78,23.72],[0.37,23.96]] },
  { id: 'b1_story_bridges',  cefrLevel: 'B1', name: 'Puentes del Relato',  lp: [80.3, 21.8], wrap: true,
    poly: [[57,0],[99.72,0],[99.91,40.88],[62.34,26.94],[53.08,28.13],[34.58,23],[36.17,19.67],[57.29,13.59]] },
  { id: 'a2_living_quarter', cefrLevel: 'A2', name: 'Barrio Vivo',          lp: [18, 34.4],  wrap: false,
    poly: [[0.37,23.96],[21.78,23.72],[34.58,23],[53.08,28.13],[52.99,31.7],[50.65,34.21],[36.07,39.69],[23.55,45.17],[20.47,54.47],[21.59,57.09],[0,57]] },
  { id: 'a1_clear_plaza',    cefrLevel: 'A1', name: 'Plaza Clara',          lp: [65.3, 51.8], wrap: false,
    poly: [[99.91,40.88],[99.91,58.05],[60,73.3],[52.9,69.13],[21.59,57.09],[20.47,54.47],[23.55,45.17],[36.07,39.69],[50.65,34.21],[52.99,31.7],[53.08,28.13],[62.34,26.94]] },
  { id: 'a0_spark_gate',     cefrLevel: 'A0', name: 'Puerta de la Chispa', lp: [22.8, 73.5], wrap: true,
    poly: [[0,57],[21.59,57.09],[52.9,69.13],[60,73.3],[60.09,78.19],[52.15,83.08],[41.4,90.11],[45.79,99.88],[0.37,99.4],[0,86]] },
  { id: 'c1_mastery_campus', cefrLevel: 'C1', name: 'Campus de Maestría',  lp: [76.5, 91.1], wrap: true,
    poly: [[99.91,58.05],[100,100],[45.79,99.88],[41.4,90.11],[52.15,83.08],[60.09,78.19],[60,73.3]] },
]

// Grammar categories keyed by CEFR level (e.g. "A0" → [{passed_chapters, total_chapters, can_access}])
const grammarByLevel = ref<Record<string, { passed: number; total: number; canAccess: boolean }>>({})

async function loadGrammarProgress() {
  try {
    const { categories } = await grammarClient.getCategories()
    const map: Record<string, { passed: number; total: number; canAccess: boolean }> = {}
    for (const cat of categories) {
      const lv = (cat.level || '').toUpperCase()
      if (!lv) continue
      if (!map[lv]) map[lv] = { passed: 0, total: 0, canAccess: false }
      map[lv].passed += cat.passed_chapters || 0
      map[lv].total += cat.total_chapters || 0
      if (cat.can_access) map[lv].canAccess = true
    }
    grammarByLevel.value = map
  } catch { /* grammar not loaded – all districts stay at lv=1 */ }
}

// Compute lv for a district from grammar progress:
// 1 = no grammar access, 2 = access but 0 passed, 3 = <50% passed, 4 = ≥50% passed, 5 = 100% passed
function districtLv(cefrLevel: string): number {
  const g = grammarByLevel.value[cefrLevel]
  if (!g || !g.canAccess) return 1
  if (g.total === 0) return 2
  if (g.passed === 0) return 2
  if (g.passed >= g.total) return 5
  if (g.passed >= g.total * 0.5) return 4
  return 3
}

// Derived district list with computed lv — reactively updates when grammarByLevel changes
const CITY_DISTS = computed(() =>
  DIST_DEFS.map(d => ({
    ...d,
    lv: districtLv(d.cefrLevel),
    acts: [{ i: '📖', s: 'started' }, { i: '🌿', s: 'started' }, { i: '📚', s: 'started' }, { i: '💬', s: 'started' }],
  }))
)

const ACT_BG: Record<string, string> = {
  stable:  'rgba(36,108,56,0.9)',
  growing: 'rgba(70,155,85,0.85)',
  weak:    'rgba(196,144,28,0.9)',
  started: 'rgba(142,132,120,0.82)',
}

// Re-render canvas when grammar data arrives
watch(grammarByLevel, () => renderCity())

// ─── Canvas rendering ───────────────────────────────────────────────────────
const cvsRef = ref<HTMLCanvasElement | null>(null)
const canvasCursor = ref('default')
const imgs = new Array(5).fill(null) as (HTMLImageElement | null)[]
let readyCount = 0

function poly(ctx: CanvasRenderingContext2D, pts: number[][], W: number, H: number) {
  ctx.beginPath()
  pts.forEach(([x, y], i) => i
    ? ctx.lineTo(x * W / 100, y * H / 100)
    : ctx.moveTo(x * W / 100, y * H / 100))
  ctx.closePath()
  ctx.fill()
}

function offscreen(w: number, h: number) {
  const el = document.createElement('canvas')
  el.width = w; el.height = h
  return { el, c: el.getContext('2d')! }
}

function renderCity() {
  const cvs = cvsRef.value
  if (!cvs || !imgs[0]) return
  const tw = cvs.width, th = cvs.height
  const ctx = cvs.getContext('2d')!
  ctx.clearRect(0, 0, tw, th)
  ctx.drawImage(imgs[0]!, 0, 0, tw, th)
  for (let tier = 2; tier <= 5; tier++) {
    const eligible = CITY_DISTS.value.filter(d => d.lv >= tier)
    if (!eligible.length || !imgs[tier - 1]) continue
    const lc = offscreen(tw, th)
    lc.c.drawImage(imgs[tier - 1]!, 0, 0, tw, th)
    const mc = offscreen(tw, th)
    mc.c.filter = `blur(${Math.max(10, Math.round(tw * 0.032))}px)`
    mc.c.fillStyle = '#fff'
    eligible.forEach(d => poly(mc.c, d.poly, tw, th))
    mc.c.filter = 'none'
    lc.c.globalCompositeOperation = 'destination-in'
    lc.c.drawImage(mc.el, 0, 0)
    lc.c.globalCompositeOperation = 'source-over'
    ctx.drawImage(lc.el, 0, 0)
  }
}

function resize() {
  const cvs = cvsRef.value
  if (!cvs) return
  const r = cvs.getBoundingClientRect()
  const dpr = Math.min(window.devicePixelRatio || 1, 2)
  cvs.width = r.width * dpr
  cvs.height = r.height * dpr
  renderCity()
}

// Point-in-polygon (ray casting) for % coords
function pointInPoly(px: number, py: number, pts: number[][]): boolean {
  let inside = false
  for (let i = 0, j = pts.length - 1; i < pts.length; j = i++) {
    const [xi, yi] = pts[i], [xj, yj] = pts[j]
    if ((yi > py) !== (yj > py) && px < (xj - xi) * (py - yi) / (yj - yi) + xi) {
      inside = !inside
    }
  }
  return inside
}

function districtAtPoint(px: number, py: number) {
  // px, py in % of canvas CSS size
  return CITY_DISTS.value.find(d => pointInPoly(px, py, d.poly)) || null
}

function openDistrict(d: typeof CITY_DISTS.value[0]) {
  if (d.lv <= 1) return
  router.push({ name: 'CityDistrict', params: { districtCode: d.id } })
}

function handleCanvasHover(e: MouseEvent) {
  const cvs = cvsRef.value
  if (!cvs) return
  const r = cvs.getBoundingClientRect()
  const px = ((e.clientX - r.left) / r.width) * 100
  const py = ((e.clientY - r.top) / r.height) * 100
  const d = districtAtPoint(px, py)
  canvasCursor.value = (d && d.lv > 1) ? 'pointer' : 'default'
}

function handleCanvasClick(e: MouseEvent) {
  const cvs = cvsRef.value
  if (!cvs) return
  const r = cvs.getBoundingClientRect()
  const px = ((e.clientX - r.left) / r.width) * 100
  const py = ((e.clientY - r.top) / r.height) * 100
  const d = districtAtPoint(px, py)
  if (d) openDistrict(d)
}

onMounted(() => {
  loadGrammarProgress()
  setTimeout(resize, 50)
  const srcs = [
    '/app/linglow/city/level1.jpg',
    '/app/linglow/city/level2.jpg',
    '/app/linglow/city/level3.jpg',
    '/app/linglow/city/level4.jpg',
    '/app/linglow/city/level5.jpg',
  ]
  srcs.forEach((src, i) => {
    const img = new Image()
    img.onload = () => {
      imgs[i] = img
      if (++readyCount === 5) renderCity()
    }
    img.onerror = () => { if (++readyCount === 5) renderCity() }
    img.src = src
  })
  window.addEventListener('resize', resize)
})

onUnmounted(() => {
  window.removeEventListener('resize', resize)
})
</script>

<style scoped>
.city-map-root {
  position: relative;
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--bg);
  overflow: hidden;
}
.city-map-title-bar {
  position: absolute;
  top: 0; left: 0; right: 0;
  z-index: 20;
  padding: 18px 20px 32px;
  background: linear-gradient(to bottom, rgba(0,0,0,0.28) 0%, transparent 100%);
  pointer-events: none;
  text-align: center;
}
.city-map-title {
  font-family: 'Lora', serif;
  font-weight: 700;
  font-size: 22px;
  color: #fff;
  text-shadow: 0 1px 8px rgba(0,0,0,0.5);
  line-height: 1;
}
.city-map-sub {
  font-size: 11px;
  color: rgba(255,255,255,0.78);
  text-shadow: 0 1px 4px rgba(0,0,0,0.4);
  margin-top: 4px;
  letter-spacing: 0.1px;
}
.city-map-canvas {
  display: block;
  width: 100%;
  flex: 1;
}
.city-map-overlays {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  pointer-events: none;
}
.district-label {
  position: absolute;
  transform: translate(-50%, -50%);
  text-align: center;
  max-width: 115px;
  pointer-events: auto;
  cursor: pointer;
}
.district-label--locked {
  cursor: default;
}
.district-name {
  display: block;
  font-family: 'Lora', Georgia, serif;
  font-size: 13px;
  font-weight: 700;
  line-height: 1.25;
  white-space: nowrap;
  color: #1e1208;
  text-shadow:
    0 0 18px rgba(255,248,210,0.99),
    0 0 8px rgba(255,236,160,0.88),
    0 1px 3px rgba(255,255,255,0.7);
}
.district-label--locked .district-name {
  color: rgba(15,13,10,0.68);
  text-shadow:
    0 0 8px rgba(255,255,255,0.9),
    1px 1px 2px rgba(0,0,0,0.45),
    -1px -1px 2px rgba(0,0,0,0.35);
}
.district-label--wrap .district-name {
  white-space: normal;
}
.district-lock {
  display: flex;
  justify-content: center;
  margin-top: 4px;
}
.district-acts {
  display: flex;
  gap: 3px;
  margin-top: 5px;
  justify-content: center;
}
.district-act {
  width: 24px;
  height: 24px;
  border-radius: 7px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  border: 1px solid rgba(255,255,255,0.28);
  box-shadow: 0 1px 4px rgba(0,0,0,0.2);
}
</style>
