<template>
  <svg :width="w ?? s" :height="h ?? s" viewBox="0 0 24 24" fill="none">
    <template v-if="name === 'home'">
      <path d="M3 12L12 4L21 12V20C21 20.6 20.6 21 20 21H15V16H9V21H4C3.4 21 3 20.6 3 20V12Z" :fill="c" />
    </template>
    <template v-else-if="name === 'map-pin' || name === 'pin'">
      <circle cx="12" cy="10" r="3" :stroke="c" stroke-width="2" />
      <path d="M12 2C8.1 2 5 5.1 5 9c0 5.3 7 13 7 13s7-7.7 7-13c0-3.9-3.1-7-7-7z" :stroke="c" stroke-width="2" fill="none" />
    </template>
    <template v-else-if="name === 'book'">
      <rect x="4" y="3" width="16" height="18" rx="2" :stroke="c" stroke-width="2" fill="none" />
      <path d="M8 7H16M8 11H16M8 15H12" :stroke="c" stroke-width="2" stroke-linecap="round" />
    </template>
    <template v-else-if="name === 'refresh'">
      <path d="M1 4v6h6" :stroke="c" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
      <path d="M23 20v-6h-6" :stroke="c" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
      <path d="M20.5 9A9 9 0 005.6 5.6L1 10M23 14l-4.6 4.4A9 9 0 013.5 15" :stroke="c" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
    </template>
    <template v-else-if="name === 'send'">
      <path d="M22 2L11 13" :stroke="c" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
      <path d="M22 2L15 22L11 13L2 9L22 2Z" :stroke="c" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
    </template>
    <template v-else-if="name === 'mic'">
      <rect x="9" y="2" width="6" height="11" rx="3" :stroke="c" stroke-width="2" fill="none" />
      <path d="M19 10a7 7 0 01-14 0M12 19v3M8 22h8" :stroke="c" stroke-width="2" stroke-linecap="round" />
    </template>
    <template v-else-if="name === 'hamburger'">
      <path d="M3 6H21M3 12H21M3 18H21" :stroke="c" stroke-width="2" stroke-linecap="round" />
    </template>
    <template v-else-if="name === 'clock'">
      <circle cx="12" cy="12" r="9" :stroke="c" stroke-width="2" fill="none" />
      <path d="M12 7v5l3 3" :stroke="c" stroke-width="2" stroke-linecap="round" />
    </template>
    <template v-else-if="name === 'dumbbell'">
      <rect x="1" y="10" width="3" height="4" rx="1" :fill="c" />
      <rect x="4" y="8" width="3" height="8" rx="1.5" :fill="c" />
      <rect x="7" y="11" width="10" height="2" rx="1" :fill="c" />
      <rect x="17" y="8" width="3" height="8" rx="1.5" :fill="c" />
      <rect x="20" y="10" width="3" height="4" rx="1" :fill="c" />
    </template>
    <template v-else-if="name === 'bar'">
      <rect x="3" y="12" width="4" height="9" rx="1" :fill="c" />
      <rect x="10" y="7" width="4" height="14" rx="1" :fill="c" />
      <rect x="17" y="3" width="4" height="18" rx="1" :fill="c" />
    </template>
    <template v-else-if="name === 'user'">
      <circle cx="12" cy="8" r="4" :stroke="c" stroke-width="2" fill="none" />
      <path d="M4 20C4 16.7 7.6 14 12 14C16.4 14 20 16.7 20 20" :stroke="c" stroke-width="2" stroke-linecap="round" fill="none" />
    </template>
    <template v-else-if="name === 'chevron-right'">
      <path d="M9 6L15 12L9 18" :stroke="c" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
    </template>
    <template v-else-if="name === 'chevron-left'">
      <path d="M15 6L9 12L15 18" :stroke="c" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
    </template>
    <template v-else-if="name === 'chevron-down'">
      <path d="M6 9L12 15L18 9" :stroke="c" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
    </template>
    <template v-else-if="name === 'x'">
      <path d="M6 6L18 18M6 18L18 6" :stroke="c" stroke-width="2" stroke-linecap="round" />
    </template>
    <template v-else-if="name === 'check'">
      <path d="M5 12L10 17L19 7" :stroke="c" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" />
    </template>
    <template v-else-if="name === 'sound'">
      <path d="M11 5L6 9H2V15H6L11 19V5Z" :fill="c" />
      <path d="M19.1 4.9A10 10 0 0 1 19.1 19.1M15.5 8.5A5 5 0 0 1 15.5 15.5" :stroke="c" stroke-width="2" fill="none" stroke-linecap="round" />
    </template>
    <template v-else-if="name === 'moon'">
      <path d="M21 12.8A9 9 0 1 1 11.2 3a7 7 0 0 0 9.8 9.8z" :fill="c" />
    </template>
    <!-- Glyphs below are adapted from Lucide (https://lucide.dev, ISC license). -->
    <template v-else-if="name === 'lock'">
      <rect x="3" y="11" width="18" height="11" rx="2" :stroke="c" stroke-width="2" fill="none" />
      <path d="M7 11V7a5 5 0 0 1 10 0v4" :stroke="c" stroke-width="2" fill="none" stroke-linecap="round" />
    </template>
    <template v-else-if="name === 'flame'">
      <path d="M8.5 14.5A2.5 2.5 0 0 0 11 12c0-1.38-.5-2-1-3-1.072-2.143-.224-4.054 2-6 .5 2.5 2 4.9 4 6.5 2 1.6 3 3.5 3 5.5a7 7 0 1 1-14 0c0-1.153.433-2.294 1-3a2.5 2.5 0 0 0 2.5 2.5z" :stroke="c" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round" />
    </template>
    <template v-else-if="name === 'book-open'">
      <path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z" :stroke="c" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round" />
      <path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z" :stroke="c" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round" />
    </template>
    <template v-else-if="name === 'sprout'">
      <path d="M7 20h10" :stroke="c" stroke-width="2" stroke-linecap="round" />
      <path d="M12 20c0-4.5 1.5-8 6-9.5.5 3-1 8.5-6 9.5z" :stroke="c" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round" />
      <path d="M12 20c0-5.5-2-9.5-7-11-.5 3.5.5 10 7 11z" :stroke="c" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round" />
      <path d="M12 9c0-2.5 1-4.5 3-6" :stroke="c" stroke-width="2" fill="none" stroke-linecap="round" />
    </template>
    <template v-else-if="name === 'building'">
      <rect x="4" y="2" width="16" height="20" rx="2" :stroke="c" stroke-width="2" fill="none" />
      <path d="M9 22v-4h6v4M8 6h.01M16 6h.01M12 6h.01M12 10h.01M12 14h.01M16 10h.01M16 14h.01M8 10h.01M8 14h.01" :stroke="c" stroke-width="2" stroke-linecap="round" />
    </template>
    <template v-else-if="name === 'crown'">
      <path d="M11.56 3.7a.5.5 0 0 1 .88 0l2.61 4.98a1 1 0 0 0 1.46.36l3.6-2.55a.5.5 0 0 1 .78.52l-1.92 9.02a1 1 0 0 1-.98.8H5.99a1 1 0 0 1-.97-.8L3.1 6.99a.5.5 0 0 1 .78-.5l3.6 2.55a1 1 0 0 0 1.46-.36z" :stroke="c" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round" />
      <path d="M5 21h14" :stroke="c" stroke-width="2" stroke-linecap="round" />
    </template>
    <template v-else-if="name === 'party-popper'">
      <path d="M5.8 11.3 2 22l10.7-3.79" :stroke="c" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round" />
      <path d="M4 3h.01M22 8h.01M15 2h.01M22 20h.01" :stroke="c" stroke-width="2" stroke-linecap="round" />
      <path d="m22 2-2.24.75a2.9 2.9 0 0 0-1.96 3.12c.1.86-.57 1.63-1.45 1.63h-.38c-.86 0-1.6.6-1.76 1.44L14 10" :stroke="c" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round" />
      <path d="m22 13-.82-.33c-.86-.34-1.82.2-1.98 1.11c-.11.7-.72 1.22-1.43 1.22H17" :stroke="c" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round" />
      <path d="m11 2 .33.82c.34.86-.2 1.82-1.11 1.98C9.52 4.9 9 5.52 9 6.23V7" :stroke="c" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round" />
      <path d="M11 13c1.93 1.93 2.83 4.17 2 5-.83.83-3.07-.07-5-2-1.93-1.93-2.83-4.17-2-5 .83-.83 3.07.07 5 2Z" :stroke="c" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round" />
    </template>
    <template v-else-if="name === 'lightbulb'">
      <path d="M15 14c.2-1 .7-1.7 1.5-2.5 1-.9 1.5-2.2 1.5-3.5A6 6 0 0 0 6 8c0 1 .2 2.2 1.5 3.5.7.7 1.3 1.5 1.5 2.5" :stroke="c" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round" />
      <path d="M9 18h6M10 22h4" :stroke="c" stroke-width="2" stroke-linecap="round" />
    </template>
    <template v-else-if="name === 'message-circle'">
      <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8z" :stroke="c" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round" />
    </template>
    <template v-else-if="name === 'sparkles'">
      <path d="M12 3l1.9 5.8a2 2 0 0 0 1.3 1.3L21 12l-5.8 1.9a2 2 0 0 0-1.3 1.3L12 21l-1.9-5.8a2 2 0 0 0-1.3-1.3L3 12l5.8-1.9a2 2 0 0 0 1.3-1.3z" :stroke="c" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round" />
      <path d="M5 3v4M3 5h4M19 17v4M17 19h4" :stroke="c" stroke-width="2" stroke-linecap="round" />
    </template>
    <template v-else-if="name === 'sparkle'">
      <path d="M12 3l1.9 5.8a2 2 0 0 0 1.3 1.3L21 12l-5.8 1.9a2 2 0 0 0-1.3 1.3L12 21l-1.9-5.8a2 2 0 0 0-1.3-1.3L3 12l5.8-1.9a2 2 0 0 0 1.3-1.3z" :stroke="c" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round" />
    </template>
    <template v-else-if="name === 'circle'">
      <circle cx="12" cy="12" r="9" :stroke="c" stroke-width="2" fill="none" />
    </template>
    <template v-else-if="name === 'frown'">
      <circle cx="12" cy="12" r="9" :stroke="c" stroke-width="2" fill="none" />
      <path d="M16 16s-1.5-2-4-2-4 2-4 2" :stroke="c" stroke-width="2" fill="none" stroke-linecap="round" />
      <path d="M9 9h.01M15 9h.01" :stroke="c" stroke-width="2" stroke-linecap="round" />
    </template>
    <template v-else-if="name === 'star'">
      <path d="M11.53 3.05a.5.5 0 0 1 .94 0l2 5.05a.5.5 0 0 0 .44.32l5.42.36a.5.5 0 0 1 .29.89l-4.18 3.47a.5.5 0 0 0-.16.52l1.34 5.26a.5.5 0 0 1-.76.55L12.28 17a.5.5 0 0 0-.56 0l-4.58 2.92a.5.5 0 0 1-.76-.55l1.34-5.26a.5.5 0 0 0-.16-.52L3.38 9.67a.5.5 0 0 1 .29-.89l5.42-.36a.5.5 0 0 0 .44-.32z" :stroke="c" stroke-width="2" fill="none" stroke-linejoin="round" />
    </template>
    <template v-else-if="name === 'star-filled'">
      <path d="M11.53 3.05a.5.5 0 0 1 .94 0l2 5.05a.5.5 0 0 0 .44.32l5.42.36a.5.5 0 0 1 .29.89l-4.18 3.47a.5.5 0 0 0-.16.52l1.34 5.26a.5.5 0 0 1-.76.55L12.28 17a.5.5 0 0 0-.56 0l-4.58 2.92a.5.5 0 0 1-.76-.55l1.34-5.26a.5.5 0 0 0-.16-.52L3.38 9.67a.5.5 0 0 1 .29-.89l5.42-.36a.5.5 0 0 0 .44-.32z" :fill="c" />
    </template>
    <template v-else-if="name === 'sun'">
      <circle cx="12" cy="12" r="5" :stroke="c" stroke-width="2" fill="none" />
      <path d="M12 2V4M12 20V22M2 12H4M20 12H22M5.6 5.6L7 7M17 17L18.4 18.4M18.4 5.6L17 7M7 17L5.6 18.4" :stroke="c" stroke-width="2" stroke-linecap="round" />
    </template>
    <template v-else-if="name === 'palette'">
      <path d="M12 2C6.5 2 2 6.5 2 12s4.5 10 10 10c.926 0 1.648-.746 1.648-1.688 0-.437-.18-.835-.437-1.125-.29-.289-.438-.652-.438-1.125a1.64 1.64 0 0 1 1.668-1.668h1.996c3.051 0 5.555-2.503 5.555-5.554C21.965 6.012 17.461 2 12 2z" :stroke="c" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round" />
      <circle cx="6.5" cy="12.5" r="1" :fill="c" />
      <circle cx="8.5" cy="7.5" r="1" :fill="c" />
      <circle cx="13.5" cy="6.5" r="1" :fill="c" />
      <circle cx="17.5" cy="10.5" r="1" :fill="c" />
    </template>
    <template v-else-if="name === 'ruler'">
      <path d="M21.3 15.3a2.4 2.4 0 0 1 0 3.4l-2.6 2.6a2.4 2.4 0 0 1-3.4 0L2.7 8.7a2.41 2.41 0 0 1 0-3.4l2.6-2.6a2.41 2.41 0 0 1 3.4 0Z" :stroke="c" stroke-width="2" fill="none" stroke-linecap="round" stroke-linejoin="round" />
      <path d="m14.5 12.5 2-2M11.5 9.5l2-2M8.5 6.5l2-2M17.5 15.5l2-2" :stroke="c" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
    </template>
    <template v-else-if="name === 'hash'">
      <path d="M4 9h16M4 15h16M10 3 8 21M16 3l-2 18" :stroke="c" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
    </template>
  </svg>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  name: string
  s?: number
  w?: number
  h?: number
  c?: string
}>(), { s: 22, c: 'currentColor' })
</script>
