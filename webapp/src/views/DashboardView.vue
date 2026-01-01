<template>
  <div class="dashboard">
    <div class="dashboard-header">
      <h1>Dashboard</h1>
      <button @click="refreshData" class="btn-refresh" :disabled="loading">
        <span v-if="!loading">↻</span>
        <span v-else>⟳</span>
      </button>
    </div>
    
    <div v-if="loading" class="loading">Loading...</div>
    
    <div v-else class="dashboard-content">
      <!-- Main Stats Cards -->
      <div class="stats-grid">
        <div class="stat-card stat-card-primary">
          <div class="stat-icon">📚</div>
          <div class="stat-content">
            <h3>Ready for Review</h3>
            <p class="stat-number">{{ stats.dueCount }}</p>
            <p class="stat-label">Cards due</p>
          </div>
          <router-link to="/training" class="stat-action btn btn-primary">
            Start Training
          </router-link>
        </div>

        <div class="stat-card stat-card-info">
          <div class="stat-icon">✨</div>
          <div class="stat-content">
            <h3>New Cards</h3>
            <p class="stat-number">{{ stats.newCount }}</p>
            <p class="stat-label">Not started</p>
          </div>
        </div>

        <div class="stat-card stat-card-success">
          <div class="stat-icon">📖</div>
          <div class="stat-content">
            <h3>In Learning</h3>
            <p class="stat-number">{{ stats.learningCount }}</p>
            <p class="stat-label">Being learned</p>
          </div>
        </div>

        <div class="stat-card stat-card-warning">
          <div class="stat-icon">🔄</div>
          <div class="stat-content">
            <h3>In Review</h3>
            <p class="stat-number">{{ stats.reviewCount }}</p>
            <p class="stat-label">Mastered cards</p>
          </div>
        </div>
      </div>

      <!-- Progress Section -->
      <div class="progress-section">
        <div class="card">
          <h2>Your Progress</h2>
          <div class="progress-grid">
            <div class="progress-item">
              <div class="progress-header">
                <span>Total Cards</span>
                <span class="progress-value">{{ stats.totalCards }}</span>
              </div>
              <div class="progress-bar">
                <div class="progress-fill" :style="{ width: '100%' }"></div>
              </div>
            </div>
            
            <div class="progress-item">
              <div class="progress-header">
                <span>Available for Training</span>
                <span class="progress-value">{{ stats.availableForTraining }}</span>
              </div>
              <div class="progress-bar">
                <div 
                  class="progress-fill progress-fill-primary" 
                  :style="{ width: stats.totalCards > 0 ? (stats.availableForTraining / stats.totalCards * 100) + '%' : '0%' }"
                ></div>
              </div>
            </div>

            <div class="progress-item">
              <div class="progress-header">
                <span>Accuracy (30 days)</span>
                <span class="progress-value">{{ formatPercent(stats.accuracyPercent) }}%</span>
              </div>
              <div class="progress-bar">
                <div 
                  class="progress-fill progress-fill-success" 
                  :style="{ width: stats.accuracyPercent + '%' }"
                ></div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Activity Section -->
      <div class="activity-section">
        <div class="activity-grid">
          <div class="card activity-card">
            <h2>Today</h2>
            <div class="activity-stats">
              <div class="activity-stat">
                <span class="activity-label">Sessions</span>
                <span class="activity-value">{{ stats.todaySessions }}</span>
              </div>
              <div class="activity-stat">
                <span class="activity-label">Cards Completed</span>
                <span class="activity-value">{{ stats.todayCardsCompleted }}</span>
              </div>
            </div>
          </div>

          <div class="card activity-card">
            <h2>This Week</h2>
            <div class="activity-stats">
              <div class="activity-stat">
                <span class="activity-label">Sessions</span>
                <span class="activity-value">{{ stats.weekSessions }}</span>
              </div>
              <div class="activity-stat">
                <span class="activity-label">Cards Completed</span>
                <span class="activity-value">{{ stats.weekCardsCompleted }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Recent Sessions -->
      <div v-if="stats.lastSessions && stats.lastSessions.length > 0" class="recent-sessions">
        <div class="card">
          <h2>Recent Training Sessions</h2>
          <div class="sessions-list">
            <div 
              v-for="session in stats.lastSessions" 
              :key="session.id" 
              class="session-item"
            >
              <div class="session-date">{{ formatDate(session.date) }}</div>
              <div class="session-stats">
                <span class="session-stat">
                  <span class="session-stat-label">Completed:</span>
                  <span class="session-stat-value">{{ session.completed }} cards</span>
                </span>
                <span class="session-stat">
                  <span class="session-stat-label">Accuracy:</span>
                  <span class="session-stat-value">{{ formatPercent(session.accuracy) }}%</span>
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { apiClient } from '../api/client'

interface DashboardStats {
  dueCount: number
  newCount: number
  learningCount: number
  reviewCount: number
  totalCards: number
  availableForTraining: number
  todaySessions: number
  todayCardsCompleted: number
  weekSessions: number
  weekCardsCompleted: number
  accuracyPercent: number
  lastSessions: Array<{
    id: number
    date: string
    completed: number
    accuracy: number
    source: string
  }>
}

const stats = ref<DashboardStats>({
  dueCount: 0,
  newCount: 0,
  learningCount: 0,
  reviewCount: 0,
  totalCards: 0,
  availableForTraining: 0,
  todaySessions: 0,
  todayCardsCompleted: 0,
  weekSessions: 0,
  weekCardsCompleted: 0,
  accuracyPercent: 0,
  lastSessions: []
})

const loading = ref(true)

const loadData = async () => {
  try {
    loading.value = true
    const data = await apiClient.request('/app/dashboard')
    stats.value = {
      dueCount: data.due_count || 0,
      newCount: data.new_count || 0,
      learningCount: data.learning_count || 0,
      reviewCount: data.review_count || 0,
      totalCards: data.total_cards || 0,
      availableForTraining: data.available_for_training || 0,
      todaySessions: data.today_sessions || 0,
      todayCardsCompleted: data.today_cards_completed || 0,
      weekSessions: data.week_sessions || 0,
      weekCardsCompleted: data.week_cards_completed || 0,
      accuracyPercent: data.accuracy_percent || 0,
      lastSessions: data.last_sessions || []
    }
  } catch (error) {
    console.error('Failed to load dashboard:', error)
  } finally {
    loading.value = false
  }
}

const refreshData = () => {
  loadData()
}

const formatPercent = (value: number): string => {
  return value.toFixed(1)
}

const formatDate = (dateString: string): string => {
  const date = new Date(dateString)
  const now = new Date()
  const diffTime = Math.abs(now.getTime() - date.getTime())
  const diffDays = Math.floor(diffTime / (1000 * 60 * 60 * 24))
  
  if (diffDays === 0) {
    return 'Today'
  } else if (diffDays === 1) {
    return 'Yesterday'
  } else if (diffDays < 7) {
    return `${diffDays} days ago`
  } else {
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.dashboard {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

.dashboard-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
}

.dashboard-header h1 {
  margin: 0;
  font-size: 32px;
  font-weight: 600;
}

.btn-refresh {
  background: var(--card-bg);
  border: 1px solid var(--input-border);
  border-radius: 6px;
  padding: 8px 16px;
  cursor: pointer;
  font-size: 20px;
  transition: all 0.2s;
  color: var(--text-primary);
}

.btn-refresh:hover:not(:disabled) {
  background: var(--input-bg);
  transform: rotate(180deg);
}

.btn-refresh:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.dashboard-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

/* Stats Grid */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
}

.stat-card {
  background: var(--card-bg);
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 2px 8px var(--card-shadow);
  display: flex;
  flex-direction: column;
  gap: 16px;
  transition: transform 0.2s, box-shadow 0.2s;
  position: relative;
  overflow: hidden;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px var(--card-shadow);
}

.stat-card-primary {
  border-left: 4px solid var(--color-primary);
}

.stat-card-info {
  border-left: 4px solid #3498db;
}

.stat-card-success {
  border-left: 4px solid var(--color-success);
}

.stat-card-warning {
  border-left: 4px solid #f39c12;
}

.stat-icon {
  font-size: 32px;
  margin-bottom: 8px;
}

.stat-content {
  flex: 1;
}

.stat-content h3 {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-secondary);
  margin: 0 0 8px 0;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.stat-number {
  font-size: 42px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
  line-height: 1;
}

.stat-label {
  font-size: 13px;
  color: var(--text-secondary);
  margin: 8px 0 0 0;
}

.stat-action {
  margin-top: auto;
  text-align: center;
  text-decoration: none;
  display: block;
}

/* Progress Section */
.progress-section {
  margin-top: 8px;
}

.progress-grid {
  display: flex;
  flex-direction: column;
  gap: 20px;
  margin-top: 20px;
}

.progress-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.progress-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 14px;
  color: var(--text-primary);
}

.progress-value {
  font-weight: 600;
  color: var(--color-primary);
}

.progress-bar {
  width: 100%;
  height: 8px;
  background: var(--input-bg);
  border-radius: 4px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: var(--color-primary);
  border-radius: 4px;
  transition: width 0.3s ease;
}

.progress-fill-primary {
  background: var(--color-primary);
}

.progress-fill-success {
  background: var(--color-success);
}

/* Activity Section */
.activity-section {
  margin-top: 8px;
}

.activity-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
}

.activity-card h2 {
  margin-bottom: 20px;
  font-size: 20px;
}

.activity-stats {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.activity-stat {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  background: var(--input-bg);
  border-radius: 6px;
}

.activity-label {
  font-size: 14px;
  color: var(--text-secondary);
}

.activity-value {
  font-size: 24px;
  font-weight: 700;
  color: var(--color-primary);
}

/* Recent Sessions */
.recent-sessions {
  margin-top: 8px;
}

.recent-sessions h2 {
  margin-bottom: 20px;
  font-size: 20px;
}

.sessions-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.session-item {
  padding: 16px;
  background: var(--input-bg);
  border-radius: 8px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  transition: background 0.2s;
}

.session-item:hover {
  background: var(--card-bg);
}

.session-date {
  font-weight: 600;
  color: var(--text-primary);
}

.session-stats {
  display: flex;
  gap: 24px;
}

.session-stat {
  display: flex;
  gap: 8px;
  align-items: center;
}

.session-stat-label {
  font-size: 13px;
  color: var(--text-secondary);
}

.session-stat-value {
  font-weight: 600;
  color: var(--text-primary);
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .activity-grid {
    grid-template-columns: 1fr;
  }
  
  .session-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
  
  .session-stats {
    flex-direction: column;
    gap: 8px;
    width: 100%;
  }
}
</style>

