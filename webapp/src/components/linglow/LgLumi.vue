<template>
  <img :src="src" :width="size" alt="Lumi" class="lg-lumi" />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import lumiDefault from '../../assets/linglow/lumi.png'
import lumiBook from '../../assets/linglow/lumi_book.png'
import lumiPencil from '../../assets/linglow/lumi_pencil.png'
import lumiMap from '../../assets/linglow/lumi_map.png'

const props = withDefaults(defineProps<{
  size?: number
  pose?: 'default' | 'book' | 'pencil' | 'map'
}>(), { size: 80, pose: 'default' })

const src = computed(() => {
  switch (props.pose) {
    case 'book': return lumiBook
    case 'pencil': return lumiPencil
    case 'map': return lumiMap
    default: return lumiDefault
  }
})
</script>

<style scoped>
.lg-lumi {
  height: auto;
  flex-shrink: 0;
  display: block;
  object-fit: contain;
}
</style>
