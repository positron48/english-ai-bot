<template>
  <div class="pq-page">
    <LgPageHeader
      :title="districtTitle || t('pictureQuest.title')"
      :show-back="true"
      @back="router.push({ name: 'PictureQuestCategories' })"
    />

    <div class="pq-sub">{{ t('pictureQuest.listSub') }}</div>

    <LgLoader v-if="loading" />
    <div v-else-if="!isPro" class="pq-empty">{{ t('chat.requiresPro') }}</div>
    <div v-else-if="loadError" class="pq-empty">{{ loadError }}</div>
    <template v-else>
      <div v-if="quests.length" class="pq-toolbar">
        <button type="button" class="pq-random" :title="t('pictureQuest.randomQuest')" @click="openRandom">
          <LgIcon name="dice" :s="18" /> {{ t('pictureQuest.randomQuest') }}
        </button>
      </div>

      <div v-if="!quests.length" class="pq-empty">{{ t('pictureQuest.allDoneMain') }}</div>
      <div v-else class="pq-list">
        <button
          v-for="quest in pagedQuests"
          :key="quest.code"
          class="pq-card"
          type="button"
          @click="openQuest(quest)"
        >
          <img v-if="quest.image_url" :src="quest.image_url" class="pq-thumb" alt="" />
          <LgActivityIcon
            v-else
            type="conversation"
            :status="quest.quest_passed ? 'yellow' : 'orange'"
            :size="24"
          />
          <div class="pq-card-body">
            <div class="pq-card-title">{{ quest.title }}</div>
            <div class="pq-card-meta">
              {{ quest.cefr_level }}
              <span v-if="quest.tasks.length" class="pq-tag">{{ completedCount(quest) }}/{{ quest.tasks.length }}</span>
            </div>
          </div>
          <span class="pq-arrow">›</span>
        </button>
      </div>

      <div class="pq-footer">
        <button v-if="hasMore" type="button" class="pq-more" @click="page++">
          {{ t('pictureQuest.loadMore') }}
        </button>
        <router-link :to="{ name: 'PictureQuestDistrictArchive', params: { districtCode } }" class="pq-archive-link">
          {{ t('pictureQuest.viewArchive') }} →
        </router-link>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { courseClient, type PictureQuestSummary } from '../api/courseClient'
import LgActivityIcon from '../components/linglow/LgActivityIcon.vue'
import LgIcon from '../components/linglow/LgIcon.vue'
import LgLoader from '../components/linglow/LgLoader.vue'
import LgPageHeader from '../components/linglow/LgPageHeader.vue'
import { useCourse } from '../composables/useCourse'
import { useMe } from '../composables/useMe'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const { currentCourseCode, ensureCourseLoaded } = useCourse()
const { ensureMe, hasFeature } = useMe()

const districtCode = computed(() => String(route.params.districtCode || ''))
const districtTitle = ref('')

const loading = ref(true)
const isPro = ref(false)
const loadError = ref('')
const quests = ref<PictureQuestSummary[]>([])

const page = ref(1)
const perPage = 20
const pagedQuests = computed(() => quests.value.slice(0, page.value * perPage))
const hasMore = computed(() => pagedQuests.value.length < quests.value.length)

function completedCount(quest: PictureQuestSummary): number {
  return quest.tasks.filter(task => task.completed).length
}

function openQuest(quest: PictureQuestSummary) {
  router.push({ name: 'PictureQuestChat', params: { questCode: quest.code } })
}

function openRandom() {
  const pool = pagedQuests.value
  if (!pool.length) return
  const pick = pool[Math.floor(Math.random() * pool.length)]
  openQuest(pick)
}

onMounted(async () => {
  try {
    await ensureCourseLoaded()
    await ensureMe()
    isPro.value = hasFeature('picture_description')
    if (!isPro.value) return
    const courseCode = currentCourseCode.value || undefined
    const [res, map] = await Promise.all([
      courseClient.listPictureQuests(districtCode.value, courseCode, false),
      courseClient.getCourseMap(courseCode).catch(() => null),
    ])
    quests.value = res.quests || []
    districtTitle.value = map?.districts?.find(d => d.code === districtCode.value)?.title || ''
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
.pq-toolbar {
  display: flex;
  justify-content: flex-end;
  padding: 0 16px 8px;
}
.pq-random {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border: 1px solid var(--border);
  border-radius: 12px;
  background: var(--card-bg);
  color: var(--text);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
}
.pq-random:hover { border-color: #2d6b3a; color: #2d6b3a; }
.pq-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 0 16px;
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
  padding: 12px 14px;
  border: 1px solid var(--border);
  border-radius: 14px;
  background: var(--card-bg);
  box-shadow: var(--shadow-card);
  cursor: pointer;
  text-align: left;
}
.pq-card-body { flex: 1; min-width: 0; }
.pq-card-title {
  font-family: 'Inter', sans-serif;
  font-size: 15px;
  font-weight: 700;
  color: var(--text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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
.pq-arrow {
  color: var(--subtext);
  font-size: 22px;
  line-height: 1;
}
.pq-footer {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  flex-wrap: wrap;
  padding: 18px 16px 0;
}
.pq-more {
  padding: 9px 22px;
  border: 1px solid var(--border);
  border-radius: 12px;
  background: var(--card-bg);
  color: var(--text);
  cursor: pointer;
  font-size: 14px;
}
.pq-more:hover { border-color: #2d6b3a; color: #2d6b3a; }
.pq-archive-link {
  font-size: 13px;
  color: var(--subtext);
  text-decoration: underline;
}
</style>
