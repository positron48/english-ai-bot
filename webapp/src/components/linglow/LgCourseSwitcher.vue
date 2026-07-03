<template>
  <div ref="rootRef" class="lg-course-sw">
    <button class="lg-course-sw-btn" type="button" @click="toggle">
      <span class="lg-course-sw-label">{{ label }}</span>
      <span v-if="courses.length > 1" class="lg-course-sw-caret">▾</span>
    </button>
    <div v-if="open && courses.length > 1" class="lg-course-sw-dropdown">
      <button
        v-for="c in courses"
        :key="c.code"
        type="button"
        class="lg-course-sw-item"
        :class="{ 'lg-course-sw-item--on': c.code === currentCourseCode }"
        @click="pick(c.code)"
      >
        {{ c.title }}
        <span v-if="c.code === currentCourseCode" class="lg-course-sw-check"><LgIcon name="check" :s="14" /></span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useCourse } from '../../composables/useCourse'
import LgIcon from './LgIcon.vue'

const props = withDefaults(defineProps<{ label?: string }>(), { label: '' })

const { courses, currentCourse, currentCourseCode, selectCourse } = useCourse()

const open = ref(false)
const rootRef = ref<HTMLElement | null>(null)

const label = computed(() => props.label || currentCourse.value?.title || '')

const toggle = () => { if (courses.value.length > 1) open.value = !open.value }

const pick = async (code: string) => {
  open.value = false
  if (code === currentCourseCode.value) return
  await selectCourse(code)
  // Course switch re-scopes nearly all data — full reload keeps a single source of truth.
  window.location.reload()
}

const close = (e: MouseEvent) => {
  if (rootRef.value && !rootRef.value.contains(e.target as Node)) open.value = false
}
onMounted(() => document.addEventListener('mousedown', close))
onUnmounted(() => document.removeEventListener('mousedown', close))
</script>

<style scoped>
.lg-course-sw { position: relative; display: inline-block; }
.lg-course-sw-btn {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
  font: inherit;
  color: inherit;
}
.lg-course-sw-caret { font-size: 10px; opacity: 0.7; }
.lg-course-sw-dropdown {
  position: absolute;
  top: 130%;
  left: 0;
  background: var(--card-bg);
  border: 1px solid var(--border);
  border-radius: 12px;
  overflow: hidden;
  z-index: 300;
  min-width: 160px;
  box-shadow: var(--shadow-card);
}
.lg-course-sw-item {
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
.lg-course-sw-item:last-child { border-bottom: none; }
.lg-course-sw-item--on { color: var(--text); font-weight: 600; }
.lg-course-sw-check { color: var(--dorado); font-size: 11px; }
</style>
