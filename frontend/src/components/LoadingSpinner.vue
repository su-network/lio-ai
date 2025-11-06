<template>
  <div class="flex items-center justify-center" :class="containerClass">
    <div class="relative" :style="{ width: size + 'px', height: size + 'px' }">
      <!-- Modern 3-dot bouncing loader -->
      <div v-if="type === 'dots'" class="flex items-center justify-center space-x-2">
        <div 
          v-for="i in 3" 
          :key="i"
          class="rounded-full animate-bounce"
          :class="color"
          :style="{ 
            width: dotSize + 'px', 
            height: dotSize + 'px',
            animationDelay: (i - 1) * 0.15 + 's'
          }"
        ></div>
      </div>

      <!-- Modern spinner -->
      <div v-else-if="type === 'spinner'" class="relative">
        <div 
          class="absolute inset-0 rounded-full border-2 border-gray-200 dark:border-gray-700"
        ></div>
        <div 
          class="absolute inset-0 rounded-full border-2 border-transparent animate-spin"
          :class="borderColor"
          style="border-top-color: currentColor; border-right-color: currentColor;"
        ></div>
      </div>

      <!-- Pulsing circle -->
      <div v-else-if="type === 'pulse'" class="relative">
        <div 
          class="absolute inset-0 rounded-full animate-ping opacity-75"
          :class="color"
        ></div>
        <div 
          class="relative rounded-full"
          :class="color"
          :style="{ width: size + 'px', height: size + 'px' }"
        ></div>
      </div>

      <!-- Bars loader -->
      <div v-else-if="type === 'bars'" class="flex items-end justify-center space-x-1">
        <div 
          v-for="i in 4" 
          :key="i"
          class="rounded-sm animate-pulse"
          :class="color"
          :style="{ 
            width: '4px',
            height: (barHeights[i - 1]) + 'px',
            animationDelay: (i - 1) * 0.1 + 's',
            animationDuration: '0.8s'
          }"
        ></div>
      </div>
    </div>
    
    <!-- Optional loading text -->
    <p v-if="text" class="ml-3 text-sm font-medium" :class="textColor">
      {{ text }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  type?: 'dots' | 'spinner' | 'pulse' | 'bars'
  size?: number
  color?: 'blue' | 'green' | 'purple' | 'gray'
  text?: string
  containerClass?: string
}

const props = withDefaults(defineProps<Props>(), {
  type: 'spinner',
  size: 40,
  color: 'blue',
  text: '',
  containerClass: 'py-12'
})

const dotSize = computed(() => Math.floor(props.size / 6))
const barHeights = computed(() => [20, 30, 25, 35])

const colorClasses = {
  blue: 'bg-blue-600',
  green: 'bg-green-600',
  purple: 'bg-purple-600',
  gray: 'bg-gray-600'
}

const borderColorClasses = {
  blue: 'text-blue-600',
  green: 'text-green-600',
  purple: 'text-purple-600',
  gray: 'text-gray-600'
}

const textColorClasses = {
  blue: 'text-blue-600 dark:text-blue-400',
  green: 'text-green-600 dark:text-green-400',
  purple: 'text-purple-600 dark:text-purple-400',
  gray: 'text-gray-600 dark:text-gray-400'
}

const color = computed(() => colorClasses[props.color])
const borderColor = computed(() => borderColorClasses[props.color])
const textColor = computed(() => textColorClasses[props.color])
</script>

<style scoped>
@keyframes bounce {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-8px);
  }
}

.animate-bounce {
  animation: bounce 0.6s infinite;
}
</style>
