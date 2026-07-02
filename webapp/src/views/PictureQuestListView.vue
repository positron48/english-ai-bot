<template>
  <div class="pq-page">
    <LgPageHeader
      :title="t('pictureQuest.title')"
      :show-back="true"
      @back="router.push('/learning')"
    />

    <div class="pq-sub">{{ t('pictureQuest.listSub') }}</div>

    <LgLoader v-if="loading" />
    <div v-else-if="!isPro" class="pq-empty">{{ t('chat.requiresPro') }}</div>
    <div v-else-if="!districtGroups.length" class="pq-empty">{{ loadError || t('pictureQuest.noQuests') }}</div>
    <div v-else class="pq-districts">
      <section v-for="group in districtGroups" :key="group.district.code" class="pq-district">
        <div class="pq-district-head">
          <div>
            <div class="pq-district-title">{{ group.district.title }}</div>
            <div class="pq-district-meta">{{ group.district.level_code }}</div>
          </div>
          <span class="pq-district-count">{{ group.quests.length }}</span>
        </div>

        <button
          v-for="quest in group.quests"
          :key="quest.code"
          class="pq-card"
          type="button"
          @click="openQuest(quest)"
        >
          <img v-if="quest.image_url" :src="quest.image_url" class="pq-thumb" alt="" />
          <LgActivityIcon
            v-else
            type="conversation"
            :status="quest.all_tasks_done ? 'green' : (quest.quest_passed ? 'yellow' : 'orange')"
            :size="24"
          />
          <div class="pq-card-body">
            <div class="pq-card-title">{{ quest.title }}</div>
            <div class="pq-card-meta">
              {{ quest.cefr_level }}
              <span v-if="quest.tasks.length" class="pq-tag">{{ completedCount(quest) }}/{{ quest.tasks.length }}</span>
              <span v-if="quest.all_tasks_done" class="pq-tag pq-tag--perfect">★ {{ t('chat.completed100') }}</span>
            </div>
          </div>
          <span v-if="quest.all_tasks_done" class="pq-done pq-done--perfect">★</span>
          <span v-else-if="quest.quest_passed" class="pq-done">✓</span>
          <span v-else class="pq-arrow">›</span>
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
import type { CourseMapDistrict, PictureQuestSummary } from '../api/courseClient'
import { grammarClient } from '../api/grammarClient'
import LgActivityIcon from '../components/linglow/LgActivityIcon.vue'
import LgLoader from '../components/linglow/LgLoader.vue'
import LgPageHeader from '../components/linglow/LgPageHeader.vue'
import { useCourse } from '../composables/useCourse'
import { useMe } from '../composables/useMe'
import { buildGrammarLevelAccess, isDistrictUnlocked } from '../utils/districtUnlock'

interface DistrictQuestGroup {
  district: CourseMapDistrict
  quests: PictureQuestSummary[]
}

const { t } = useI18n()
const router = useRouter()
const { currentCourseCode, ensureCourseLoaded } = useCourse()
const { ensureMe, hasFeature } = useMe()

const loading = ref(true)
const isPro = ref(false)
const loadError = ref('')
const districtGroups = ref<DistrictQuestGroup[]>([])

function completedCount(quest: PictureQuestSummary): number {
  return quest.tasks.filter(task => task.completed).length
}

function openQuest(quest: PictureQuestSummary) {
  router.push({ name: 'PictureQuestChat', params: { questCode: quest.code } })
}

onMounted(async () => {
  try {
    await ensureCourseLoaded()
    await ensureMe()
    isPro.value = hasFeature('picture_description')
    if (!isPro.value) return

    const courseCode = currentCourseCode.value || undefined
    const [map, grammarData] = await Promise.all([
      courseClient.getCourseMap(courseCode),
      grammarClient.getCategories().catch(() => ({ categories: [] })),
    ])
    const grammarAccess = buildGrammarLevelAccess(grammarData.categories || [])
    const openDistricts = (map.districts || []).filter(d => isDistrictUnlocked(d.level_code, grammarAccess))
    const results = await Promise.all(openDistricts.map(async (district) => {
      const res = await courseClient.listPictureQuests(district.code, courseCode)
      return { district, quests: res.quests || [] }
    }))
    districtGroups.value = results.filter(g => g.quests.length > 0)
  } catch (e: any) {
    const msg = String(e?.message || '')
    loadError.value = msg.includes('403') ? t('chat.requiresPro') : t('chat.notAvailable')
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.pq-page { padding-bottom: 24px; }
.pq-sub {
  padding: 0 16px 12px;
  font-size: 13px;
  line-height: 1.4;
  color: var(--subtext);
  text-align: center;
}
.pq-empty {
  padding: 40px 16px;
  text-align: center;
  color: var(--subtext);
  font-size: 14px;
}
.pq-districts {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 4px 16px 16px;
}
.pq-district {
  background: var(--card-bg);
  border: 1px solid var(--border);
  border-radius: 18px;
  box-shadow: var(--shadow-card);
  overflow: hidden;
}
.pq-district-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 14px 14px 10px;
}
.pq-district-title {
  font-family: 'Lora', serif;
  font-size: 17px;
  line-height: 1.15;
  font-weight: 700;
  color: var(--text);
}
.pq-district-meta {
  margin-top: 2px;
  font-size: 11px;
  color: var(--subtext);
}
.pq-district-count {
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
.pq-thumb {
  width: 52px;
  height: 52px;
  border-radius: 12px;
  object-fit: cover;
  flex-shrink: 0;
}
.pq-card {
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
.pq-card-body { flex: 1; min-width: 0; }
.pq-card-title {
  font-family: 'Inter', sans-serif;
  font-size: 15px;
  font-weight: 700;
  color: var(--text);
}
.pq-card-meta {
  margin-top: 2px;
  font-family: 'Inter', sans-serif;
  font-size: 12px;
  color: var(--subtext);
}
.pq-tag {
  display: inline-block;
  margin-left: 6px;
  padding: 1px 7px;
  border-radius: 10px;
  font-size: 10px;
  font-weight: 600;
  background: rgba(45,107,58,0.12);
  color: #2d6b3a;
}
.pq-tag--perfect {
  background: rgba(200,168,75,0.22);
  color: #9a7b1e;
}
.pq-done {
  color: #2d6b3a;
  font-size: 18px;
  font-weight: 800;
}
.pq-done--perfect {
  color: #c8a84b;
  text-shadow: 0 1px 4px rgba(200,168,75,0.35);
}
.pq-arrow {
  color: var(--subtext);
  font-size: 22px;
  line-height: 1;
}
</style>
