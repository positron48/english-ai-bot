<template>
  <div class="npc-page">
    <LgPageHeader
      :title="t('chat.npcTitle')"
      :show-back="true"
      @back="router.push('/learning')"
    />

    <div class="npc-sub">{{ t('chat.npcListSub') }}</div>

    <LgLoader v-if="loading" />
    <div v-else-if="!isPro" class="npc-empty">{{ t('chat.requiresPro') }}</div>
    <div v-else-if="!districtGroups.length" class="npc-empty">{{ loadError || t('chat.noPlaces') }}</div>
    <div v-else class="npc-districts">
      <section v-for="districtGroup in districtGroups" :key="districtGroup.district.code" class="npc-district">
        <div class="npc-district-head">
          <div>
            <div class="npc-district-title">{{ districtGroup.district.title }}</div>
            <div class="npc-district-meta">{{ districtGroup.district.level_code }}</div>
          </div>
          <span class="npc-district-count">{{ districtGroup.npcs.length }}</span>
        </div>

        <button
          v-for="npc in districtGroup.npcs"
          :key="npc.key"
          class="npc-card"
          type="button"
          @click="openNpc(districtGroup.district.code, npc)"
        >
          <img
            v-if="npc.npcImageUrl"
            :src="mediaUrl(npc.npcImageUrl)"
            class="npc-avatar"
            alt=""
          />
          <LgActivityIcon
            v-else
            type="conversation"
            :status="npc.allDone ? 'green' : (npc.hasCompletedQuests ? 'yellow' : 'orange')"
            :size="24"
          />
          <div class="npc-card-body">
            <div class="npc-card-title">
              {{ npc.npcName }}
              <span v-if="npc.npcRole" class="npc-role">, {{ npc.npcRole }}</span>
              <span v-if="npc.hasIncompleteQuests" class="npc-bang" :title="t('chat.newQuests')">!</span>
            </div>
            <div class="npc-card-meta">
              {{ placeLabel(npc.placeType) }} · {{ npc.level }}
              <span v-if="npc.questTotal" class="npc-tag">{{ npc.completedCount }}/{{ npc.questTotal }}</span>
              <span v-if="npc.allDone" class="npc-tag npc-tag--perfect"><LgIcon name="star-filled" :s="11" /> {{ t('chat.completed100') }}</span>
            </div>
            <div v-if="npc.questTotal > 1" class="npc-bar-track">
              <div class="npc-bar-fill" :style="{ width: npc.pct + '%' }" />
            </div>
          </div>
          <span v-if="npc.allDone" class="npc-done npc-done--perfect"><LgIcon name="star-filled" :s="16" /></span>
          <span v-else-if="npc.allPassed || npc.hasCompletedQuests" class="npc-done"><LgIcon name="check" :s="16" /></span>
          <span v-else class="npc-arrow">›</span>
        </button>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { courseClient } from '../api/courseClient'
import type { CourseMapDistrict } from '../api/courseClient'
import { grammarClient } from '../api/grammarClient'
import LgActivityIcon from '../components/linglow/LgActivityIcon.vue'
import LgIcon from '../components/linglow/LgIcon.vue'
import LgLoader from '../components/linglow/LgLoader.vue'
import LgPageHeader from '../components/linglow/LgPageHeader.vue'
import { useCourse } from '../composables/useCourse'
import { useMe } from '../composables/useMe'
import { buildNpcGroups } from '../utils/conversations'
import type { NpcGroup } from '../utils/conversations'
import { buildGrammarLevelAccess, isDistrictUnlocked } from '../utils/districtUnlock'
import { mediaUrl } from '../utils/mediaUrl'

interface DistrictNpcGroup {
  district: CourseMapDistrict
  npcs: NpcGroup[]
}

const { t } = useI18n()
const router = useRouter()
const { currentCourseCode, ensureCourseLoaded } = useCourse()
const { ensureMe, hasFeature } = useMe()

const loading = ref(true)
const isPro = ref(false)
const loadError = ref('')
const districtGroups = ref<DistrictNpcGroup[]>([])

const PLACE_LABELS: Record<string, string> = {
  cafe: 'Кафе',
  shop: 'Магазин',
  police_station: 'Полиция',
  pharmacy: 'Аптека',
  hotel: 'Отель',
  restaurant: 'Ресторан',
  market: 'Рынок',
}

function placeLabel(placeType: string): string {
  return PLACE_LABELS[placeType] || placeType
}

function openNpc(districtCode: string, npc: NpcGroup) {
  const firstAvailable = npc.questScenarios.find(s => !s.locked) || (npc.allPassed ? npc.freeScenario : null)
  if (!firstAvailable) return
  router.push({ name: 'PlaceChat', params: { districtCode, scenarioCode: firstAvailable.code } })
}

onMounted(async () => {
  try {
    await ensureCourseLoaded()
    await ensureMe()
    isPro.value = hasFeature('conversation')
    if (!isPro.value) return

    const courseCode = currentCourseCode.value || undefined
    const [map, grammarData] = await Promise.all([
      courseClient.getCourseMap(courseCode),
      grammarClient.getCategories().catch(() => ({ categories: [] })),
    ])
    const grammarAccess = buildGrammarLevelAccess(grammarData.categories || [])
    const openDistricts = (map.districts || []).filter(d => isDistrictUnlocked(d.level_code, grammarAccess))
    const results = await Promise.all(openDistricts.map(async (district) => {
      const res = await courseClient.listConversationScenarios(district.code, courseCode)
      const npcs = buildNpcGroups(res.scenarios || [], courseCode || '').filter(npc => !npc.locked)
      return { district, npcs }
    }))
    districtGroups.value = results.filter(g => g.npcs.length > 0)
  } catch (e: any) {
    const msg = String(e?.message || '')
    loadError.value = msg.includes('403') ? t('chat.requiresPro') : t('chat.notAvailable')
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.npc-page { padding-bottom: 24px; }
.npc-sub {
  padding: 0 16px 12px;
  font-size: 13px;
  line-height: 1.4;
  color: var(--subtext);
  text-align: center;
}
.npc-empty {
  padding: 40px 16px;
  text-align: center;
  color: var(--subtext);
  font-size: 14px;
}
.npc-districts {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 4px 16px 16px;
}
.npc-district {
  background: var(--card-bg);
  border: 1px solid var(--border);
  border-radius: 18px;
  box-shadow: var(--shadow-card);
  overflow: hidden;
}
.npc-district-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 14px 14px 10px;
}
.npc-district-title {
  font-family: 'Lora', serif;
  font-size: 17px;
  line-height: 1.15;
  font-weight: 700;
  color: var(--text);
}
.npc-district-meta {
  margin-top: 2px;
  font-size: 11px;
  color: var(--subtext);
}
.npc-district-count {
  min-width: 24px;
  height: 24px;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: rgba(45,107,58,0.12);
  color: #2d6b3a;
  font-size: 12px;
  font-weight: 700;
}
.npc-avatar { width: 40px; height: 40px; border-radius: 50%; object-fit: cover; flex-shrink: 0; }
.npc-card {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 13px 14px;
  border: none;
  border-top: 1px solid var(--border);
  background: transparent;
  cursor: pointer;
  text-align: left;
}
.npc-card-body { flex: 1; min-width: 0; }
.npc-card-title {
  font-family: 'Inter', sans-serif;
  font-size: 15px;
  font-weight: 700;
  color: var(--text);
}
.npc-role {
  color: var(--subtext);
  font-weight: 500;
}
.npc-card-meta {
  margin-top: 2px;
  font-family: 'Inter', sans-serif;
  font-size: 12px;
  color: var(--subtext);
}
.npc-tag {
  display: inline-block;
  margin-left: 6px;
  padding: 1px 7px;
  border-radius: 10px;
  font-size: 10px;
  font-weight: 600;
  background: rgba(45,107,58,0.12);
  color: #2d6b3a;
}
.npc-tag--perfect {
  background: rgba(200,168,75,0.22);
  color: #9a7b1e;
}
.npc-bang {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  margin-left: 6px;
  border-radius: 50%;
  background: #d97706;
  color: #fff;
  font-size: 12px;
  font-weight: 800;
  line-height: 1;
  vertical-align: middle;
}
.npc-bar-track {
  margin-top: 6px;
  height: 3px;
  border-radius: 999px;
  background: var(--progress-track, rgba(0,0,0,0.08));
  overflow: hidden;
}
.npc-bar-fill {
  height: 100%;
  border-radius: 999px;
  background: #2d6b3a;
  transition: width 0.4s ease;
}
.npc-done {
  color: #2d6b3a;
  font-size: 18px;
  font-weight: 800;
}
.npc-done--perfect {
  color: #c8a84b;
  text-shadow: 0 1px 4px rgba(200,168,75,0.35);
}
.npc-arrow {
  color: var(--subtext);
  font-size: 22px;
  line-height: 1;
}
</style>
