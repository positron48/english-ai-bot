<template>
  <div class="admin-content">
    <h2>Reading Texts</h2>

    <div class="course-selector">
      <label for="reading-course">Course:</label>
      <select id="reading-course" v-model="selectedCourseCode" class="level-select">
        <option disabled value="">Select a course</option>
        <option v-for="course in availableCourses" :key="course.code" :value="course.code">
          {{ course.title || course.code }}
        </option>
      </select>
    </div>

    <div v-if="coursesLoading" class="loading">Loading courses...</div>
    <div v-else-if="coursesError" class="error">{{ coursesError }}</div>
    <div v-else-if="!selectedCourseCode" class="empty-message">Select a course to manage its reading texts.</div>

    <div v-if="selectedCourseCode" class="toolbar">
      <input
        v-model="search"
        type="text"
        class="search-input"
        placeholder="Search by title..."
        @input="loadTexts"
      />
      <select v-model="level" class="level-select" @change="loadTexts">
        <option value="">All levels</option>
        <option v-for="value in levels" :key="value" :value="value">{{ value }}</option>
      </select>
      <button class="btn btn-primary" @click="loadTexts">Refresh</button>
    </div>

    <div v-if="selectedCourseCode && loading" class="loading">Loading reading texts...</div>
    <div v-else-if="selectedCourseCode && error" class="error">{{ error }}</div>
    <div v-else-if="selectedCourseCode && texts.length === 0" class="empty-message">No texts found.</div>

    <div v-else-if="selectedCourseCode" class="table-wrap">
      <table class="texts-table">
        <thead>
          <tr>
            <th>Title</th>
            <th>Level</th>
            <th>Lang</th>
            <th>Category</th>
            <th>Segments</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="text in texts" :key="text.text_id">
            <td>{{ text.title }}</td>
            <td>{{ text.level || '-' }}</td>
            <td>{{ text.target_language || '-' }}</td>
            <td class="mono">{{ text.category_id || '-' }}</td>
            <td>{{ text.segments_count }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { apiClient } from '../api/client'
import { courseClient, type CourseSummary } from '../api/courseClient'

interface ReadingTextAdminItem {
  text_id: string
  course_code: string
  category_id: string
  title: string
  level: string
  target_language: string
  segments_count: number
}

const levels = ['A0', 'A1', 'A2', 'B1', 'B2', 'C1']
const texts = ref<ReadingTextAdminItem[]>([])
const search = ref('')
const level = ref('')
const loading = ref(false)
const error = ref<string | null>(null)
const availableCourses = ref<CourseSummary[]>([])
const selectedCourseCode = ref('')
const coursesLoading = ref(false)
const coursesError = ref<string | null>(null)

const loadTexts = async () => {
  loading.value = true
  error.value = null
  try {
    const params = new URLSearchParams()
    params.set('course_code', selectedCourseCode.value)
    if (search.value.trim()) {
      params.set('search', search.value.trim())
    }
    if (level.value.trim()) {
      params.set('level', level.value.trim())
    }
    const query = params.toString()
    const url = query ? `/api/admin/reading/texts?${query}` : '/api/admin/reading/texts'
    const data: { texts?: ReadingTextAdminItem[] } = await apiClient.request(url)
    texts.value = data.texts || []
  } catch (e: any) {
    error.value = e?.message || 'Failed to load reading texts'
  } finally {
    loading.value = false
  }
}

watch(selectedCourseCode, courseCode => {
  texts.value = []
  if (courseCode) {
    loadTexts()
  }
})

onMounted(async () => {
  coursesLoading.value = true
  try {
    const data = await courseClient.getCourses()
    availableCourses.value = data.courses || []
    selectedCourseCode.value = availableCourses.value.find(course => course.is_current)?.code
      || availableCourses.value[0]?.code
      || ''
  } catch (e: any) {
    coursesError.value = e?.message || 'Failed to load courses'
  } finally {
    coursesLoading.value = false
  }
})
</script>

<style scoped>
.toolbar {
  display: flex;
  gap: 10px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}

.course-selector {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
}

.search-input,
.level-select {
  border: 1px solid var(--border-primary);
  border-radius: 6px;
  padding: 8px 10px;
  background: var(--bg-secondary);
  color: var(--text-primary);
}

.search-input {
  min-width: 260px;
}

.table-wrap {
  overflow-x: auto;
}

.texts-table {
  width: 100%;
  border-collapse: collapse;
}

.texts-table th,
.texts-table td {
  text-align: left;
  border-bottom: 1px solid var(--border-primary);
  padding: 10px 8px;
}

.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
  font-size: 12px;
}
</style>
