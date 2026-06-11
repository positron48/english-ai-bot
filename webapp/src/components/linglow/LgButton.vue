<template>
  <button
    class="lg-btn"
    :class="{ 'lg-btn--secondary': variant === 'secondary' }"
    :disabled="disabled"
    :type="type"
  >
    <slot />
  </button>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  disabled?: boolean
  variant?: 'primary' | 'secondary'
  type?: 'button' | 'submit'
}>(), { disabled: false, variant: 'primary', type: 'button' })
</script>

<style scoped>
.lg-btn {
  width: 100%;
  padding: 15px;
  border-radius: 50px;
  border: 1px solid var(--btn-border);
  cursor: pointer;
  background: var(--btn-gradient);
  color: white;
  font-size: 16px;
  font-weight: 600;
  box-shadow: var(--btn-shadow);
  transition: opacity .15s;
}
.lg-btn:disabled {
  border: none;
  cursor: default;
  background: var(--chip-bg);
  color: var(--subtext);
  box-shadow: none;
}
.lg-btn--secondary:not(:disabled) {
  background: var(--card-bg);
  color: var(--text);
  border: 1px solid var(--border);
  box-shadow: var(--shadow-soft);
}
</style>
