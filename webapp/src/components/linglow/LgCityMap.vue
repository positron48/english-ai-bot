<template>
  <div ref="wrapRef" class="lg-city-map">
    <img ref="imgRef" class="lg-city-map-img" :src="mapSrc" alt="" @load="draw" />
    <canvas
      v-if="hasPolygons"
      ref="canvasRef"
      class="lg-city-map-canvas"
      @click="onClick"
      @mousemove="onMove"
      @mouseleave="hoverCode = ''"
    />
    <div class="lg-city-map-overlay">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import type { CourseMapDistrict, CourseProgressDistrict } from '../../api/courseClient'

const props = defineProps<{
  mapSrc: string
  districts: CourseMapDistrict[]
  progressByCode: Record<string, CourseProgressDistrict>
}>()

const emit = defineEmits<{ (e: 'select', districtCode: string): void }>()

const wrapRef = ref<HTMLElement | null>(null)
const imgRef = ref<HTMLImageElement | null>(null)
const canvasRef = ref<HTMLCanvasElement | null>(null)
const hoverCode = ref('')

interface Poly { code: string; points: Array<[number, number]>; progress: number; locked: boolean }

const polys = computed<Poly[]>(() => {
  return props.districts
    .filter(d => Array.isArray(d.metadata?.polygon) && (d.metadata!.polygon!.length ?? 0) >= 3)
    .map(d => ({
      code: d.code,
      points: d.metadata!.polygon!,
      progress: props.progressByCode[d.code]?.progress_percent ?? 0,
      locked: d.status === 'locked',
    }))
})

const hasPolygons = computed(() => polys.value.length > 0)

function polyPath(ctx: CanvasRenderingContext2D, poly: Poly, w: number, h: number) {
  ctx.beginPath()
  poly.points.forEach(([x, y], i) => {
    const px = (x / 100) * w
    const py = (y / 100) * h
    if (i === 0) ctx.moveTo(px, py)
    else ctx.lineTo(px, py)
  })
  ctx.closePath()
}

function draw() {
  const canvas = canvasRef.value
  const img = imgRef.value
  if (!canvas || !img || !hasPolygons.value) return
  const rect = img.getBoundingClientRect()
  if (rect.width === 0) return
  const dpr = window.devicePixelRatio || 1
  canvas.width = rect.width * dpr
  canvas.height = rect.height * dpr
  canvas.style.width = `${rect.width}px`
  canvas.style.height = `${rect.height}px`
  const ctx = canvas.getContext('2d')
  if (!ctx) return
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
  ctx.clearRect(0, 0, rect.width, rect.height)

  for (const poly of polys.value) {
    polyPath(ctx, poly, rect.width, rect.height)
    const isHover = poly.code === hoverCode.value
    if (poly.locked) {
      ctx.fillStyle = 'rgba(40, 40, 40, 0.45)'
    } else {
      // Progress tints the district from sand to green
      const alpha = isHover ? 0.45 : 0.28
      const g = Math.round(110 + (poly.progress / 100) * 60)
      ctx.fillStyle = `rgba(63, ${g}, 63, ${alpha * (0.35 + (poly.progress / 100) * 0.65)})`
      if (isHover) ctx.fillStyle = `rgba(217, 168, 63, 0.35)`
    }
    ctx.fill()
    ctx.strokeStyle = isHover ? 'rgba(217,168,63,0.9)' : 'rgba(255,249,237,0.65)'
    ctx.lineWidth = isHover ? 2.5 : 1.5
    ctx.stroke()
  }
}

function hitTest(e: MouseEvent): string {
  const canvas = canvasRef.value
  const img = imgRef.value
  if (!canvas || !img) return ''
  const rect = img.getBoundingClientRect()
  const x = e.clientX - rect.left
  const y = e.clientY - rect.top
  const ctx = canvas.getContext('2d')
  if (!ctx) return ''
  for (const poly of polys.value) {
    polyPath(ctx, poly, rect.width, rect.height)
    if (ctx.isPointInPath(x * (window.devicePixelRatio || 1), y * (window.devicePixelRatio || 1))) return poly.code
  }
  return ''
}

function onClick(e: MouseEvent) {
  const code = hitTest(e)
  if (code) emit('select', code)
}

function onMove(e: MouseEvent) {
  const code = hitTest(e)
  if (code !== hoverCode.value) {
    hoverCode.value = code
    draw()
  }
  if (canvasRef.value) canvasRef.value.style.cursor = code ? 'pointer' : 'default'
}

watch(polys, draw, { deep: true })
let resizeObserver: ResizeObserver | null = null
onMounted(() => {
  draw()
  if (typeof ResizeObserver !== 'undefined' && wrapRef.value) {
    resizeObserver = new ResizeObserver(draw)
    resizeObserver.observe(wrapRef.value)
  }
})
onBeforeUnmount(() => resizeObserver?.disconnect())
</script>

<style scoped>
.lg-city-map { position: relative; }
.lg-city-map-img { display: block; width: 100%; height: 100%; object-fit: cover; }
.lg-city-map-canvas { position: absolute; inset: 0; z-index: 2; }
.lg-city-map-overlay { position: absolute; inset: 0; z-index: 3; pointer-events: none; }
</style>
