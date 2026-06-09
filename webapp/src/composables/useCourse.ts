import { ref, computed } from 'vue'
import { courseClient, CourseSummary } from '../api/courseClient'

const courses = ref<CourseSummary[]>([])
const currentCourse = ref<CourseSummary | null>(null)
let loadPromise: Promise<void> | null = null

async function ensureCourseLoaded(): Promise<void> {
  if (loadPromise) {
    await loadPromise
    return
  }
  loadPromise = (async () => {
    try {
      const data = await courseClient.getCourses()
      courses.value = data.courses || []
      currentCourse.value = courses.value.find(c => c.is_current) || courses.value[0] || null
    } catch {
      // ignore — course selector simply won't render
    }
  })()
  await loadPromise
}

function resetCourse(): void {
  courses.value = []
  currentCourse.value = null
  loadPromise = null
}

export function useCourse() {
  const currentCourseCode = computed(() => currentCourse.value?.code ?? '')

  async function selectCourse(code: string): Promise<void> {
    const result = await courseClient.selectCourse(code)
    currentCourse.value = result.course
    courses.value = courses.value.map(c => ({ ...c, is_current: c.code === code }))
    // allow next ensureCourseLoaded to refresh the list
    loadPromise = null
  }

  return {
    courses,
    currentCourse,
    currentCourseCode,
    ensureCourseLoaded,
    selectCourse,
    resetCourse,
  }
}
