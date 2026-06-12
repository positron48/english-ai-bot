<template>
  <div class="dst-page">

    <!-- HEADER -->
    <LgPageHeader
      :title="district?.title || t('city.district')"
      :show-back="true"
      @back="router.push('/city')"
    />

    <!-- SUBTITLE + STATUS CHIP -->
    <div class="dst-sub-row">
      <p class="dst-desc">{{ districtDesc }}</p>
      <span class="dst-chip">
        🌿 {{ confidenceLabel }}
      </span>
    </div>

    <div v-if="loading" class="dst-loading">{{ t('common.loading') }}</div>
    <template v-else>

      <!-- DISTRICT ILLUSTRATION -->
      <div class="dst-illustration">
        <img :src="districtImg" alt="" class="dst-illus-img" />
        <!-- Building labels -->
        <div v-for="(b, i) in buildings" :key="i" class="dst-bldg-label" :style="{ left: b.x, top: b.y }">
          <div class="dst-bldg-name">{{ b.name }}</div>
          <div class="dst-bldg-sub">{{ b.sub }}</div>
        </div>
      </div>

      <!-- 4 ACTIVITY AREAS -->
      <div class="dst-areas-card">
        <div v-for="(a, i) in areas" :key="i" class="dst-area-row" :class="{ 'dst-area-row--bordered': i > 0 }">
          <div class="dst-area-icon">{{ a.icon }}</div>
          <div class="dst-area-body">
            <div class="dst-area-label">{{ a.label }}</div>
            <div class="dst-area-meta">{{ a.meta }}</div>
            <div class="dst-bar-track">
              <div class="dst-bar-fill" :style="{ width: a.pct + '%', background: a.color }" />
            </div>
          </div>
          <button class="dst-area-cta" type="button" @click="a.action">
            {{ a.cta }}
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="M9 18l6-6-6-6" />
            </svg>
          </button>
        </div>
      </div>

      <!-- НОВЫЕ ОТКРЫТИЯ -->
      <div class="dst-section-title">{{ t('city.newDiscoveries') }}</div>
      <button class="dst-discovery-card" type="button" @click="router.push({ name: 'ReadingCategories', query: districtLevelQuery })">
        <div class="dst-discovery-thumb">
          <img :src="discoveryImg" alt="" class="dst-discovery-img" />
        </div>
        <div class="dst-discovery-body">
          <div class="dst-discovery-kicker">{{ t('city.discoverySub') }}</div>
          <div class="dst-discovery-title">{{ districtReadingTitle }}</div>
          <div class="dst-discovery-desc">{{ t('city.discoveryDesc') }}</div>
        </div>
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="var(--subtext)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M9 18l6-6-6-6" />
        </svg>
      </button>

      <!-- LUMI TIP -->
      <div class="dst-lumi-wrap">
        <LgLumiTip :text="lumiTip" />
      </div>

    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { courseClient, type CourseMap, type CourseProgress } from '../api/courseClient'
import { grammarClient } from '../api/grammarClient'
import LgPageHeader from '../components/linglow/LgPageHeader.vue'
import LgLumiTip from '../components/linglow/LgLumiTip.vue'
import distVida from '../assets/linglow/dist_vida.jpg'
import distParques from '../assets/linglow/dist_parques.jpg'
import distMercados from '../assets/linglow/dist_mercados.jpg'
import distCafeterias from '../assets/linglow/dist_cafeterias.jpg'
import distViajes from '../assets/linglow/dist_viajes.jpg'
import bldgLectura from '../assets/linglow/bldg_lectura.jpg'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()

const courseMap = ref<CourseMap | null>(null)
const progress = ref<CourseProgress | null>(null)
const grammarCategories = ref<any[]>([])
const loading = ref(true)
const courseCode = computed(() => (typeof route.query.course_code === 'string' ? route.query.course_code : undefined))
const districtCode = computed(() => String(route.params.districtCode || ''))

const district = computed(() => {
  const dists = Array.isArray(courseMap.value?.districts) ? courseMap.value!.districts : []
  return dists.find(d => d.code === districtCode.value) || null
})

const districtSignal = computed(() => {
  return (progress.value?.by_district || []).find(d => d.district_code === districtCode.value) || {
    foundation: 0, confidence: 0, stability: 0, weakness: 0,
    progress_percent: 0, attempted_items: 0, total_items: 0, mastered_items: 0,
  }
})

// Per-type progress for this district's locations
const typeProgress = computed(() => {
  const locs = (progress.value?.by_location || []).filter(l => l.district_code === districtCode.value)
  const out: Record<string, { attempted: number; total: number; pct: number }> = {}
  for (const loc of locs) {
    const key = loc.location_type || 'other'
    if (!out[key]) out[key] = { attempted: 0, total: 0, pct: 0 }
    out[key].attempted += loc.attempted_items
    out[key].total += loc.total_items
  }
  for (const key of Object.keys(out)) {
    const e = out[key]
    e.pct = e.total > 0 ? Math.round((e.attempted / e.total) * 100) : 0
  }
  return out
})

const wordsPct = computed(() => Math.round(districtSignal.value.confidence))
const grammarPct = computed(() => {
  const lv = districtLevelCode.value
  if (!lv || !grammarCategories.value.length) return 0
  const cats = grammarCategories.value.filter(c => String(c.level || '').toUpperCase() === lv)
  if (!cats.length) return 0
  const total = cats.reduce((s: number, c: any) => s + (c.total_chapters || 0), 0)
  const passed = cats.reduce((s: number, c: any) => s + (c.passed_chapters || 0), 0)
  return total > 0 ? Math.round((passed / total) * 100) : 0
})
const readingPct = computed(() => {
  const r = typeProgress.value['reading_text'] || typeProgress.value['reading']
  if (r) return r.pct
  return Math.round(wordsPct.value * 0.7)
})
const chatPct = computed(() => Math.round(wordsPct.value * 0.5))

// Level code for the district (e.g. "A0", "A1") — used to filter grammar categories
const districtLevelCode = computed(() => district.value?.level_code?.toUpperCase() || '')
const districtLevelQuery = computed(() => districtLevelCode.value ? { level: districtLevelCode.value } : {})

const districtDesc = computed(() => {
  const code = districtCode.value
  if (code.includes('plaza') || code.includes('a1')) return t('city.distDescA1')
  if (code.includes('barrio') || code.includes('a2')) return t('city.distDescA2')
  if (code.includes('puerta') || code.includes('a0')) return t('city.distDescA0')
  if (code.includes('alto') || code.includes('b2')) return t('city.distDescB2')
  if (code.includes('puentes') || code.includes('b1')) return t('city.distDescB1')
  if (code.includes('campus') || code.includes('c1')) return t('city.distDescC1')
  return district.value?.title || ''
})

const confidenceLabel = computed(() => {
  const c = wordsPct.value
  if (c >= 75) return t('city.confidenceHigh')
  if (c >= 40) return t('city.confidenceMid')
  return t('city.confidenceLow')
})

const lumiTip = computed(() => {
  const c = wordsPct.value
  if (c >= 75) return t('city.lumiTipHigh')
  if (c >= 40) return t('city.lumiTipMid')
  return t('city.lumiTipLow')
})

// District illustration image based on level_code
const districtImg = computed(() => {
  const lv = district.value?.level_code?.toLowerCase() || districtCode.value
  if (lv.startsWith('a0')) return distVida
  if (lv.startsWith('a1')) return distParques
  if (lv.startsWith('a2')) return distMercados
  if (lv.startsWith('b1')) return distCafeterias
  return distViajes
})

const discoveryImg = bldgLectura

// Building label positions (% coords, same layout prototype uses)
const buildings = [
  { name: 'Jardín de Frases',        sub: t('city.areaGrammar'), x: '18%', y: '22%' },
  { name: 'Mercado de Palabras',     sub: t('city.areaWords'),   x: '72%', y: '18%' },
  { name: 'Quiosco de Lectura',      sub: t('city.areaReading'), x: '12%', y: '55%' },
  { name: 'Cabinas de Conversación', sub: t('city.areaChat'),    x: '76%', y: '52%' },
]

// Reading title suggestion (first item from review queue or fallback)
const districtReadingTitle = computed(() => t('city.discoveryPhrase'))

const areas = computed(() => [
  {
    icon: '🏛', label: t('city.areaGrammar'), color: '#2d6b3a',
    meta: t('city.areaMetaGrammar', { pct: grammarPct.value }),
    pct: grammarPct.value, cta: t('city.ctaContinue'),
    action: () => router.push({ name: 'LearningGrammar', query: districtLevelQuery.value }),
  },
  {
    icon: '🌿', label: t('city.areaWords'), color: '#2d6b3a',
    meta: t('city.areaMetaWords', { pct: wordsPct.value }),
    pct: wordsPct.value, cta: t('city.ctaContinue'),
    action: () => router.push({ name: 'Training' }),
  },
  {
    icon: '📚', label: t('city.areaReading'), color: '#c8a84b',
    meta: t('city.areaMetaReading', { pct: readingPct.value }),
    pct: readingPct.value, cta: t('city.ctaRead'),
    action: () => router.push({ name: 'ReadingCategories', query: districtLevelQuery.value }),
  },
  {
    icon: '💬', label: t('city.areaChat'), color: '#2d6b3a',
    meta: t('city.areaMetaChat', { pct: chatPct.value }),
    pct: chatPct.value, cta: t('city.ctaPractice'),
    action: () => router.push({ name: 'Chat' }),
  },
])

onMounted(async () => {
  try {
    const [map, prog, grammarData] = await Promise.all([
      courseClient.getCourseMap(courseCode.value),
      courseClient.getProgress(courseCode.value),
      grammarClient.getCategories().catch(() => ({ categories: [] })),
    ])
    courseMap.value = map
    progress.value = prog
    grammarCategories.value = grammarData.categories || []
  } catch { /* ignore */ } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.dst-page { padding-bottom: 32px; }

/* SUB ROW */
.dst-sub-row {
  padding: 0 16px 14px;
  text-align: center;
}
.dst-desc {
  margin: 0 0 10px;
  font-family: 'Inter', sans-serif;
  font-size: 13px;
  color: var(--subtext);
  line-height: 1.5;
}
.dst-chip {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 5px 14px;
  border-radius: 20px;
  background: rgba(45, 107, 58, 0.1);
  border: 1px solid rgba(45, 107, 58, 0.25);
  font-family: 'Inter', sans-serif;
  font-size: 12px;
  font-weight: 600;
  color: #2d6b3a;
}
:root[data-theme="dark"] .dst-chip {
  background: rgba(45, 107, 58, 0.22);
  color: #7fd896;
  border-color: rgba(45, 107, 58, 0.4);
}

.dst-loading {
  padding: 32px 16px;
  text-align: center;
  color: var(--subtext);
  font-size: 14px;
}

/* ILLUSTRATION */
.dst-illustration {
  margin: 0 16px;
  border-radius: 18px;
  overflow: hidden;
  position: relative;
  height: 220px;
  background: var(--chip-bg);
  border: 1px solid var(--border);
}
.dst-illus-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}
.dst-bldg-label {
  position: absolute;
  transform: translate(-50%, -50%);
  background: rgba(255, 251, 240, 0.92);
  border: 1px solid rgba(200, 168, 75, 0.4);
  border-radius: 10px;
  padding: 5px 9px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  max-width: 110px;
  pointer-events: none;
}
:root[data-theme="dark"] .dst-bldg-label {
  background: rgba(20, 32, 24, 0.92);
  border-color: rgba(200, 168, 75, 0.3);
}
.dst-bldg-name {
  font-family: 'Inter', sans-serif;
  font-weight: 700;
  font-size: 9.5px;
  color: var(--text);
  line-height: 1.2;
}
.dst-bldg-sub {
  font-family: 'Inter', sans-serif;
  font-size: 9px;
  color: var(--subtext);
  margin-top: 1px;
}

/* AREAS CARD */
.dst-areas-card {
  margin: 14px 16px 0;
  background: var(--card-bg);
  border: 1px solid var(--border);
  border-radius: 18px;
  overflow: hidden;
  box-shadow: var(--shadow-card);
}
.dst-area-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
}
.dst-area-row--bordered { border-top: 1px solid var(--border); }
.dst-area-icon {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  flex-shrink: 0;
  background: var(--chip-bg);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
}
.dst-area-body { flex: 1; min-width: 0; }
.dst-area-label {
  font-family: 'Inter', sans-serif;
  font-weight: 700;
  font-size: 14px;
  color: var(--text);
  margin-bottom: 1px;
}
.dst-area-meta {
  font-family: 'Inter', sans-serif;
  font-size: 11px;
  color: var(--subtext);
}
.dst-bar-track {
  margin-top: 5px;
  height: 3px;
  border-radius: 999px;
  background: var(--progress-track, rgba(0,0,0,0.08));
  overflow: hidden;
}
.dst-bar-fill {
  height: 100%;
  border-radius: 999px;
  transition: width 0.4s ease;
}
.dst-area-cta {
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 7px 14px;
  border-radius: 20px;
  border: none;
  background: transparent;
  cursor: pointer;
  flex-shrink: 0;
  font-family: 'Inter', sans-serif;
  font-weight: 600;
  font-size: 13px;
  color: var(--salvia);
}
.dst-area-cta:hover { opacity: 0.8; }

/* SECTION TITLE */
.dst-section-title {
  margin: 18px 16px 10px;
  font-family: 'Lora', serif;
  font-weight: 700;
  font-size: 18px;
  color: var(--text);
}

/* DISCOVERY CARD */
.dst-discovery-card {
  width: 100%;
  margin: 0 0 0;
  padding: 0 16px;
  background: none;
  border: none;
  cursor: pointer;
  text-align: left;
  display: flex;
  align-items: center;
  gap: 12px;
}
.dst-discovery-thumb {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  background: var(--chip-bg);
  overflow: hidden;
  flex-shrink: 0;
  border: 1px solid var(--border);
}
.dst-discovery-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}
.dst-discovery-body { flex: 1; min-width: 0; }
.dst-discovery-kicker {
  font-family: 'Inter', sans-serif;
  font-size: 11px;
  color: var(--subtext);
  margin-bottom: 2px;
}
.dst-discovery-title {
  font-family: 'Lora', serif;
  font-weight: 600;
  font-size: 15px;
  color: var(--text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.dst-discovery-desc {
  font-family: 'Inter', sans-serif;
  font-size: 12px;
  color: var(--subtext);
  margin-top: 1px;
}

/* LUMI */
.dst-lumi-wrap { margin: 14px 16px 0; }
</style>
