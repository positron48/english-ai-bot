<template>
  <div class="admin-layout-wrapper">
    <AdminMenu />
    
    <div 
      class="admin-main-content"
      @touchstart="handleTouchStart"
      @touchmove="handleTouchMove"
      @touchend="handleTouchEnd"
      @mousedown="handleMouseDown"
      @mousemove="handleMouseMove"
      @mouseup="handleMouseUp"
      @mouseleave="handleMouseLeave"
    >
      <router-view />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import AdminMenu from '../components/AdminMenu.vue'

// Swipe gesture handling for layout
const touchStartX = ref(0)
const touchStartY = ref(0)
const touchStartTime = ref(0)
const SWIPE_THRESHOLD = 50
const SWIPE_EDGE_ZONE = 20
const SWIPE_MAX_VERTICAL = 100

const hasProcessedTouchSwipe = ref(false) // Флаг для предотвращения двойного срабатывания touch

const handleTouchStart = (e: TouchEvent) => {
  const touch = e.touches[0]
  if (!touch) return
  
  // Игнорируем если уже обрабатывается mouse событие
  if (isMouseDown.value) return
  
  hasProcessedTouchSwipe.value = false
  touchStartX.value = touch.clientX
  touchStartY.value = touch.clientY
  touchStartTime.value = Date.now()
  
  // Если начали от левого края - предотвращаем скролл
  if (touchStartX.value < SWIPE_EDGE_ZONE) {
    e.preventDefault()
  }
}

const handleTouchMove = (e: TouchEvent) => {
  // Предотвращаем скролл страницы при свайпе от края
  if (touchStartX.value > 0 && touchStartX.value < SWIPE_EDGE_ZONE) {
    e.preventDefault()
  }
}

const handleTouchEnd = (e: TouchEvent) => {
  if (!touchStartX.value || !touchStartY.value || hasProcessedTouchSwipe.value) {
    touchStartX.value = 0
    touchStartY.value = 0
    touchStartTime.value = 0
    hasProcessedTouchSwipe.value = false
    return
  }
  
  const touch = e.changedTouches[0]
  if (!touch) {
    touchStartX.value = 0
    touchStartY.value = 0
    touchStartTime.value = 0
    hasProcessedTouchSwipe.value = false
    return
  }
  
  const touchEndX = touch.clientX
  const touchEndY = touch.clientY
  const deltaX = touchEndX - touchStartX.value
  const deltaY = Math.abs(touchEndY - touchStartY.value)
  const deltaTime = Date.now() - touchStartTime.value
  
  // Если начали свайп от левого края - открываем меню
  if (touchStartX.value < SWIPE_EDGE_ZONE && deltaX > SWIPE_THRESHOLD && deltaY < SWIPE_MAX_VERTICAL && deltaTime < 500) {
    hasProcessedTouchSwipe.value = true
    // Открываем меню через событие на AdminMenu
    const event = new CustomEvent('openMenu')
    window.dispatchEvent(event)
  }
  
  // Сброс
  touchStartX.value = 0
  touchStartY.value = 0
  touchStartTime.value = 0
  hasProcessedTouchSwipe.value = false
}

// Mouse events для эмуляции браузера
const isMouseDown = ref(false)
const hasProcessedSwipe = ref(false) // Флаг для предотвращения двойного срабатывания

const handleMouseDown = (e: MouseEvent) => {
  // Игнорируем если уже обрабатывается touch событие
  if (touchStartX.value > 0) return
  
  isMouseDown.value = true
  hasProcessedSwipe.value = false
  touchStartX.value = e.clientX
  touchStartY.value = e.clientY
  touchStartTime.value = Date.now()
}

const handleMouseMove = (e: MouseEvent) => {
  if (!isMouseDown.value) return
  
  // Предотвращаем выделение текста при перетаскивании от края
  if (touchStartX.value < SWIPE_EDGE_ZONE) {
    e.preventDefault()
  }
}

const handleMouseUp = (e: MouseEvent) => {
  if (!isMouseDown.value || !touchStartX.value || !touchStartY.value || hasProcessedSwipe.value) {
    isMouseDown.value = false
    touchStartX.value = 0
    touchStartY.value = 0
    touchStartTime.value = 0
    hasProcessedSwipe.value = false
    return
  }
  
  const mouseEndX = e.clientX
  const mouseEndY = e.clientY
  const deltaX = mouseEndX - touchStartX.value
  const deltaY = Math.abs(mouseEndY - touchStartY.value)
  const deltaTime = Date.now() - touchStartTime.value
  
  // Если начали перетаскивание от левого края - открываем меню
  if (touchStartX.value < SWIPE_EDGE_ZONE && deltaX > SWIPE_THRESHOLD && deltaY < SWIPE_MAX_VERTICAL && deltaTime < 500) {
    hasProcessedSwipe.value = true
    const event = new CustomEvent('openMenu')
    window.dispatchEvent(event)
  }
  
  // Сброс
  isMouseDown.value = false
  touchStartX.value = 0
  touchStartY.value = 0
  touchStartTime.value = 0
  hasProcessedSwipe.value = false
}

const handleMouseLeave = () => {
  // Сброс при выходе курсора за пределы элемента
  isMouseDown.value = false
  touchStartX.value = 0
  touchStartY.value = 0
  touchStartTime.value = 0
  hasProcessedSwipe.value = false
}
</script>

<style scoped>
.admin-layout-wrapper {
  display: flex;
  min-height: 100vh;
  width: 100%;
  margin: 0;
  padding: 0;
}

.admin-main-content {
  flex: 1;
  margin-left: 240px; /* Ширина сайдбара на десктопе */
  min-height: 100vh;
  background: var(--bg-primary);
  padding: 20px;
  overflow-x: auto;
  width: calc(100% - 240px);
}

@media (max-width: 767px) {
  .admin-main-content {
    margin-left: 0;
    width: 100%;
    padding: 10px;
    margin-top: 0;
    touch-action: pan-y pan-x; /* Разрешаем вертикальный и горизонтальный скролл */
    -webkit-overflow-scrolling: touch;
  }
  
  .admin-layout-wrapper {
    touch-action: pan-y pan-x;
    -webkit-overflow-scrolling: touch;
  }
}
</style>
