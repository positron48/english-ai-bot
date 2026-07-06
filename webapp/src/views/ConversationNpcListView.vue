<template>
  <div
    class="npc-page lg-section-page"
    style="--lg-section-bg-image: url('/app/linglow/art/bg-conversation.jpg')"
  >
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
      <button
        v-for="districtGroup in districtGroups"
        :key="districtGroup.district.code"
        class="npc-district"
        :class="{ 'npc-district--locked': districtGroup.locked }"
        type="button"
        :disabled="districtGroup.locked"
        @click="openDistrict(districtGroup.district.code)"
      >
        <div class="npc-district-head">
          <div>
            <div class="npc-district-title">{{ districtGroup.district.title }}</div>
            <div class="npc-district-meta">
              {{ districtGroup.district.level_code }}
              <span class="npc-dot">·</span>
              {{ t('chat.availableQuestsCount', { count: districtGroup.availableCount }) }}
              <span class="npc-dot">·</span>
              {{ t('chat.completedQuestsCount', { count: districtGroup.completedCount }) }}
            </div>
          </div>
          <LgIcon v-if="districtGroup.locked" name="lock" :s="16" class="npc-district-lock" />
          <span v-else class="npc-district-count">{{ districtGroup.availableCount }}</span>
        </div>
      </button>
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
import LgIcon from '../components/linglow/LgIcon.vue'
import LgLoader from '../components/linglow/LgLoader.vue'
import LgPageHeader from '../components/linglow/LgPageHeader.vue'
import { useCourse } from '../composables/useCourse'
import { useMe } from '../composables/useMe'
import { buildNpcGroups, scenarioQuestPassed } from '../utils/conversations'
import type { NpcGroup } from '../utils/conversations'
import { buildGrammarLevelAccess, isDistrictUnlocked } from '../utils/districtUnlock'

interface DistrictNpcGroup {
  district: CourseMapDistrict
  npcs: NpcGroup[]
  locked: boolean
  availableCount: number
  completedCount: number
  questTotal: number
}

const { t } = useI18n()
const router = useRouter()
const { currentCourseCode, ensureCourseLoaded } = useCourse()
const { ensureMe, hasFeature } = useMe()

const loading = ref(true)
const isPro = ref(false)
const loadError = ref('')
const districtGroups = ref<DistrictNpcGroup[]>([])

function openDistrict(districtCode: string) {
  router.push({ name: 'PlaceChatList', params: { districtCode } })
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
    const results = await Promise.all((map.districts || []).map(async (district) => {
      const locked = !isDistrictUnlocked(district.level_code, grammarAccess)
      const res = await courseClient.listConversationScenarios(district.code, courseCode)
      const scenarios = res.scenarios || []
      const npcs = buildNpcGroups(scenarios, courseCode || '').filter(npc => !npc.locked)
      const quests = scenarios.filter(s => s.is_quest)
      const completedCount = quests.filter(scenarioQuestPassed).length
      const availableCount = quests.filter(s => !s.locked && !scenarioQuestPassed(s)).length
      const questTotal = quests.length
      return { district, npcs, locked, availableCount, completedCount, questTotal }
    }))
    districtGroups.value = results.filter(g => g.questTotal > 0 || g.npcs.length > 0)
  } catch (e: any) {
    const msg = String(e?.message || '')
    loadError.value = msg.includes('403') ? t('chat.requiresPro') : t('chat.notAvailable')
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.npc-page {
  padding-bottom: 24px;
}
.npc-sub {
  max-width: 560px;
  padding: 0 0 12px;
  font-size: 13px;
  line-height: 1.4;
  color: var(--subtext);
  text-align: center;
  margin: 0 auto;
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
  padding: 0 0 16px;
}
.npc-district {
  width: 100%;
  background: var(--card-bg);
  border: 1px solid var(--border);
  border-radius: 18px;
  box-shadow: var(--shadow-card);
  overflow: hidden;
  padding: 0;
  text-align: left;
  cursor: pointer;
}
.npc-district--locked { opacity: 0.55; box-shadow: none; cursor: default; }
.npc-district-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 14px 14px 12px;
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
.npc-dot { margin: 0 4px; }
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
.npc-district-lock { color: var(--subtext); flex-shrink: 0; }
</style>
