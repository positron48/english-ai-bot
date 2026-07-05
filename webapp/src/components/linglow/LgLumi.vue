<template>
  <img :src="src" :style="imgStyle" alt="Lumi" class="lg-lumi" />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import lumiHappy from '../../assets/linglow/art/lumi-happy.png'
import lumiWelcome from '../../assets/linglow/art/lumi-welcome.png'
import lumiBook from '../../assets/linglow/art/lumi-reading-book.png'
import lumiWrites from '../../assets/linglow/art/lumi-writes.png'
import lumiPencil from '../../assets/linglow/art/lumi-with-pencil-and-checklist.png'
import lumiMap from '../../assets/linglow/art/lumi-map.png'
import lumiTeacher from '../../assets/linglow/art/lumi-teacher.png'
import lumiThumbsUp from '../../assets/linglow/art/lumi-thumbs-up.png'
import lumiClapping from '../../assets/linglow/art/lumi-clapping.png'
import lumiHugging from '../../assets/linglow/art/lumi-hugging.png'
import lumiProud from '../../assets/linglow/art/lumi-proud-himself.png'
import lumiDreaming from '../../assets/linglow/art/lumi-dreaming.png'
import lumiQuizzical from '../../assets/linglow/art/lumi-quizzical.png'
import lumiHearing from '../../assets/linglow/art/lumi-hearing.png'
import lumiPoint from '../../assets/linglow/art/lumi-point.png'
import lumiWithPointer from '../../assets/linglow/art/lumi-with-pointer.png'
import lumiWithStar from '../../assets/linglow/art/lumi-with-star.png'
import lumiWithHeart from '../../assets/linglow/art/lumi-with-heart.png'
import lumiWithCup from '../../assets/linglow/art/lumi-with-reward-cup.png'
import lumiWithSpeech from '../../assets/linglow/art/lumi-with-speech-baloon.png'
import lumiMessage from '../../assets/linglow/art/lumi-message-baloon.png'
import lumiWithFire from '../../assets/linglow/art/lumi-with-fire-card.png'
import lumiWithBookPointer from '../../assets/linglow/art/lumi-with-book-and-pointer.png'
import lumiWithWordCard from '../../assets/linglow/art/lumi-with-word-card-casa.png'

const props = withDefaults(defineProps<{
  size?: number
  pose?: 'default' | 'welcome' | 'book' | 'pencil' | 'writes' | 'map' | 'teacher'
    | 'thumbs-up' | 'clapping' | 'hugging' | 'proud' | 'dreaming' | 'quizzical'
    | 'hearing' | 'point' | 'pointer' | 'star' | 'heart' | 'cup' | 'speech'
    | 'message' | 'fire' | 'book-pointer' | 'word-card'
}>(), { size: 80, pose: 'welcome' })

// PNG files have ~20% whitespace on each side; render 1.6x larger and pull in with negative margin
const imgStyle = computed(() => {
  const rendered = Math.round(props.size * 1.6)
  const offset = Math.round((rendered - props.size) / 2)
  return {
    width: `${rendered}px`,
    height: 'auto',
    margin: `-${offset}px`,
    display: 'block',
    objectFit: 'contain',
    flexShrink: '0',
  }
})

const src = computed(() => {
  switch (props.pose) {
    case 'welcome':      return lumiWelcome
    case 'book':         return lumiBook
    case 'writes':       return lumiWrites
    case 'pencil':       return lumiPencil
    case 'map':          return lumiMap
    case 'teacher':      return lumiTeacher
    case 'thumbs-up':    return lumiThumbsUp
    case 'clapping':     return lumiClapping
    case 'hugging':      return lumiHugging
    case 'proud':        return lumiProud
    case 'dreaming':     return lumiDreaming
    case 'quizzical':    return lumiQuizzical
    case 'hearing':      return lumiHearing
    case 'point':        return lumiPoint
    case 'pointer':      return lumiWithPointer
    case 'star':         return lumiWithStar
    case 'heart':        return lumiWithHeart
    case 'cup':          return lumiWithCup
    case 'speech':       return lumiWithSpeech
    case 'message':      return lumiMessage
    case 'fire':         return lumiWithFire
    case 'book-pointer': return lumiWithBookPointer
    case 'word-card':    return lumiWithWordCard
    case 'happy':        return lumiHappy
    default:             return lumiWelcome
  }
})
</script>

<style scoped>
.lg-lumi { display: block; }
</style>
