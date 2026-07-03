<template>
  <div class="lg-topbar" :class="{ 'lg-topbar--full': full }">
    <span class="lg-topbar-logo" :class="{ 'lg-topbar-logo--centered': !full }">Linglow</span>
    <template v-if="full">
      <div ref="langRef" class="lg-topbar-course">
        <button
          v-if="courses.length > 1"
          class="lg-topbar-course-btn"
          type="button"
          @click="open = !open"
        >
          {{ currentCourse?.title || currentCourseCode }} <span class="lg-topbar-caret">▾</span>
        </button>
        <span v-else class="lg-topbar-course-btn">{{ currentCourse?.title || '' }}</span>
        <div v-if="open" class="lg-topbar-dropdown">
          <button
            v-for="c in courses"
            :key="c.code"
            type="button"
            class="lg-topbar-dropdown-item"
            :class="{ 'lg-topbar-dropdown-item--on': c.code === currentCourseCode }"
            @click="pick(c.code)"
          >
            {{ c.title }}
            <span v-if="c.code === currentCourseCode" class="lg-topbar-check"><LgIcon name="check" :s="14" /></span>
          </button>
        </div>
      </div>
      <div class="lg-topbar-right">
        <slot name="right" />
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { useCourse } from '../../composables/useCourse'
import LgIcon from './LgIcon.vue'

withDefaults(defineProps<{ full?: boolean }>(), { full: false })

const { courses, currentCourse, currentCourseCode, selectCourse } = useCourse()

const open = ref(false)
const langRef = ref<HTMLElement | null>(null)

const pick = async (code: string) => {
  open.value = false
  await selectCourse(code)
  // Course switch re-scopes nearly all data — full reload is simplest and safest
  window.location.reload()
}

const close = (e: MouseEvent) => {
  if (langRef.value && !langRef.value.contains(e.target as Node)) open.value = false
}
onMounted(() => document.addEventListener('mousedown', close))
onUnmounted(() => document.removeEventListener('mousedown', close))
</script>

<style scoped>
.lg-topbar {
  height: 50px;
  display: grid;
  grid-template-columns: 1fr;
  align-items: center;
  padding: 0 14px;
  background: var(--nav-bg);
  border: 1px solid var(--border);
  border-radius: 28px;
  box-shadow: 0 6px 18px rgba(86,57,22,0.08);
}
.lg-topbar--full { grid-template-columns: 1fr auto 1fr; }
.lg-topbar-logo {
  justify-self: start;
  font-family: 'Lora', serif;
  font-size: 20px;
  font-weight: 600;
  color: var(--salvia);
}
.lg-topbar-logo--centered { justify-self: center; }
.lg-topbar-course { justify-self: center; position: relative; }
.lg-topbar-course-btn {
  display: flex;
  align-items: center;
  gap: 5px;
  background: none;
  border: none;
  cursor: pointer;
  font-size: 15px;
  font-weight: 500;
  color: var(--text);
}
.lg-topbar-caret { font-size: 10px; }
.lg-topbar-dropdown {
  position: absolute;
  top: 130%;
  left: 50%;
  transform: translateX(-50%);
  background: var(--card-bg);
  border: 1px solid var(--border);
  border-radius: 12px;
  overflow: hidden;
  z-index: 300;
  min-width: 140px;
  box-shadow: var(--shadow-card);
}
.lg-topbar-dropdown-item {
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
.lg-topbar-dropdown-item:last-child { border-bottom: none; }
.lg-topbar-dropdown-item--on {
  color: var(--text);
  font-weight: 600;
}
.lg-topbar-check { color: var(--dorado); font-size: 11px; }
.lg-topbar-right {
  justify-self: end;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
</style>
