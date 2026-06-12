<template>
  <div class="lg-feedback" :class="correct ? 'lg-feedback--ok' : 'lg-feedback--bad'">
    <div class="lg-feedback-main">
      <span class="lg-feedback-leaf">🌿</span>
      <span class="lg-feedback-msg"><slot>{{ message }}</slot></span>
    </div>
    <div v-if="correct && points" class="lg-feedback-points">+{{ points }} ✦</div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  correct: boolean
  message?: string
  points?: number | null
}>()
</script>

<style scoped>
.lg-feedback {
  margin-top: 12px;
  padding: 12px 14px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}
.lg-feedback--ok {
  background: rgba(106,143,99,0.12);
  border: 1.5px solid var(--salvia);
}
.lg-feedback--bad {
  background: rgba(220,60,60,0.1);
  border: 1.5px solid #E05A5A;
}
.lg-feedback-main {
  display: flex;
  align-items: center;
  gap: 8px;
}
.lg-feedback-leaf { font-size: 18px; }
.lg-feedback-msg {
  font-size: 13px;
  color: var(--text);
  line-height: 1.4;
}
.lg-feedback-points {
  padding: 4px 10px;
  border-radius: 20px;
  background: var(--salvia);
  color: white;
  font-weight: 700;
  font-size: 13px;
  flex-shrink: 0;
}
</style>
