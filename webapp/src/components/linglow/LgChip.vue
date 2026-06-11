<template>
  <span class="lg-chip" :class="{ 'lg-chip--active': active, 'lg-chip--new': variant === 'new' }">
    <slot>{{ label }}</slot>
  </span>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  label?: string
  active?: boolean
  variant?: 'level' | 'new'
}>(), { active: false, variant: 'level' })
</script>

<style scoped>
.lg-chip {
  padding: 3px 10px;
  border-radius: 8px;
  font-size: 12px;
  font-weight: 700;
  display: inline-block;
  background: var(--chip-bg);
  color: var(--chip-text);
}
.lg-chip--active {
  background: var(--salvia);
  color: white;
}
.lg-chip--new {
  padding: 2px 8px;
  font-size: 11px;
  background: #E05A30;
  color: white;
}
</style>
