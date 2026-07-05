<template>
  <div class="pq-page">
    <LgPageHeader
      :title="t('pictureQuest.title')"
      :show-back="true"
      @back="router.push('/learning')"
    />

    <div class="pq-sub">{{ t('pictureQuest.categoriesSub') }}</div>

    <LgLoader v-if="loading" />
    <div v-else-if="!isPro" class="pq-empty">{{ t('chat.requiresPro') }}</div>
    <div v-else-if="loadError" class="pq-empty">{{ loadError }}</div>
    <div v-else-if="!districts.length" class="pq-empty">{{ t('pictureQuest.noQuests') }}</div>
    <div v-else class="pq-cats">
      <button
        v-for="d in districts"
        :key="d.code"
        class="pq-cat"
        :class="{ 'pq-cat--locked': d.locked }"
        type="button"
        :disabled="d.locked"
        @click="openDistrict(d)"
      >
        <div class="pq-cat-level">{{ d.level_code }}</div>
        <div class="pq-cat-body">
          <div class="pq-cat-title">{{ d.title }}</div>
          <div class="pq-cat-meta">
            {{ (t as any)('pictureQuest.picturesCount', d.total, { n: d.total }) }}
            <span v-if="!d.locked && d.total > 0" class="pq-cat-pct">· {{ Math.round(d.passed / d.total * 100) }}%</span>
          </div>
        </div>
        <LgIcon v-if="d.locked" name="lock" :s="16" class="pq-cat-lock" />
        <span v-else class="pq-cat-arrow">›</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { courseClient, type PictureQuestDistrict } from '../api/courseClient'
import { grammarClient } from '../api/grammarClient'
import LgIcon from '../components/linglow/LgIcon.vue'
import LgLoader from '../components/linglow/LgLoader.vue'
import LgPageHeader from '../components/linglow/LgPageHeader.vue'
import { useCourse } from '../composables/useCourse'
import { useMe } from '../composables/useMe'
import { buildGrammarLevelAccess, isDistrictUnlocked } from '../utils/districtUnlock'

interface CategoryDistrict extends PictureQuestDistrict {
  locked: boolean
}

const { t } = useI18n()
const router = useRouter()
const { currentCourseCode, ensureCourseLoaded } = useCourse()
const { ensureMe, hasFeature } = useMe()

const loading = ref(true)
const isPro = ref(false)
const loadError = ref('')
const districts = ref<CategoryDistrict[]>([])

function openDistrict(d: CategoryDistrict) {
  if (d.locked) return
  router.push({ name: 'PictureQuestDistrict', params: { districtCode: d.code } })
}

onMounted(async () => {
  try {
    await ensureCourseLoaded()
    await ensureMe()
    isPro.value = hasFeature('picture_description')
    if (!isPro.value) return

    const courseCode = currentCourseCode.value || undefined
    const [res, grammarData] = await Promise.all([
      courseClient.listPictureQuestDistricts(courseCode),
      grammarClient.getCategories().catch(() => ({ categories: [] })),
    ])
    const access = buildGrammarLevelAccess(grammarData.categories || [])
    districts.value = (res.districts || []).map(d => ({
      ...d,
      locked: !isDistrictUnlocked(d.level_code, access),
    }))
  } catch (e: any) {
    const msg = String(e?.message || '')
    loadError.value = msg.includes('403') ? t('chat.requiresPro') : t('chat.notAvailable')
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.pq-page {
  padding-bottom: 24px;
  background:
    linear-gradient(180deg, color-mix(in srgb, var(--bg, var(--bg-primary)) 20%, transparent) 0%, var(--bg, var(--bg-primary)) 380px),
    url('/app/linglow/art/bg-picture-quest.jpg') top center / 100% auto no-repeat;
}
.pq-sub {
  max-width: 560px;
  padding: 0 16px 12px;
  font-size: 13px;
  line-height: 1.4;
  color: var(--subtext);
  text-align: center;
  margin: 0 auto;
}
.pq-empty {
  padding: 40px 16px;
  text-align: center;
  color: var(--subtext);
  font-size: 14px;
}
.pq-cats {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 4px 16px 16px;
}
.pq-cat {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px 16px;
  border: 1px solid var(--border);
  border-radius: 16px;
  background: var(--card-bg);
  box-shadow: var(--shadow-card);
  cursor: pointer;
  text-align: left;
}
.pq-cat--locked { opacity: 0.55; cursor: default; box-shadow: none; }
.pq-cat-level {
  flex-shrink: 0;
  min-width: 46px;
  height: 46px;
  border-radius: 12px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: rgba(45,107,58,0.12);
  color: #2d6b3a;
  font-family: 'Lora', serif;
  font-weight: 700;
  font-size: 17px;
}
.pq-cat--locked .pq-cat-level { background: var(--chip-bg); color: var(--subtext); }
.pq-cat-body { flex: 1; min-width: 0; }
.pq-cat-title {
  font-family: 'Inter', sans-serif;
  font-size: 15px;
  font-weight: 700;
  color: var(--text);
}
.pq-cat-meta {
  margin-top: 2px;
  font-family: 'Inter', sans-serif;
  font-size: 12px;
  color: var(--subtext);
}
.pq-cat-lock { color: var(--subtext); flex-shrink: 0; }
.pq-cat-arrow { color: var(--subtext); font-size: 22px; line-height: 1; flex-shrink: 0; }
</style>
