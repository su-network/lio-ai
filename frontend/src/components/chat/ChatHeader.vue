<template>
  <div class="h-[57px] bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 px-3 sm:px-4 flex items-center">
    <div class="flex items-center justify-between w-full gap-3">
      <!-- Left side - Model Selector -->
      <div class="flex items-center gap-2">
        <!-- Show message when no models available -->
        <div v-if="!availableModels || availableModels.length === 0" class="flex items-center gap-2 px-3 py-2 rounded-lg bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-700">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-amber-600 dark:text-amber-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
          <span class="text-sm font-medium text-amber-800 dark:text-amber-200">No models available</span>
          <router-link to="/settings" class="text-sm text-blue-600 dark:text-blue-400 hover:underline ml-1">
            Configure API keys →
          </router-link>
        </div>

        <!-- Model dropdown when models are available -->
        <DropdownMenuRoot v-else>
          <DropdownMenuTrigger as-child>
            <button
              class="flex items-center space-x-2 px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors"
            >
              <!-- Online status indicator -->
              <span class="flex h-2 w-2 relative">
                <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
                <span class="relative inline-flex rounded-full h-2 w-2 bg-green-500"></span>
              </span>
              <span class="text-sm font-medium text-gray-900 dark:text-white truncate max-w-[150px] sm:max-w-none">
                {{ getModelDisplayName(selectedModel) }}
              </span>
              <ChevronDown class="w-4 h-4 text-gray-500 dark:text-gray-400 flex-shrink-0" />
            </button>
          </DropdownMenuTrigger>

          <DropdownMenuPortal>
            <DropdownMenuContent
              :side-offset="8"
              :align="'start'"
              class="min-w-[280px] sm:w-96 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-600 rounded-lg shadow-lg z-50 max-h-96 overflow-y-auto p-2"
            >
              <DropdownMenuLabel class="text-xs font-semibold text-gray-500 dark:text-gray-400 mb-2 px-3 uppercase tracking-wide">
                Available Models
              </DropdownMenuLabel>
              <DropdownMenuItem
                v-for="model in availableModels"
                :key="model.id"
                @click="selectModel(model)"
                class="flex flex-col items-start p-3 rounded-lg cursor-pointer transition-colors outline-none focus:bg-gray-100 dark:focus:bg-gray-700 data-[highlighted]:bg-gray-100 dark:data-[highlighted]:bg-gray-700"
                :class="{ 'bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-700': selectedModel === model.id }"
              >
                <div class="flex items-center justify-between w-full mb-1">
                  <div class="flex items-center gap-2">
                    <!-- Online indicator -->
                    <span class="flex h-2 w-2 rounded-full bg-green-500"></span>
                    <span class="text-sm font-medium text-gray-900 dark:text-white">{{ model.name }}</span>
                  </div>
                  <div v-if="selectedModel === model.id" class="flex items-center">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-blue-600 dark:text-blue-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                    </svg>
                  </div>
                </div>
                <div class="flex items-center gap-2 flex-wrap">
                  <span v-if="model.provider" class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400">
                    {{ model.provider }}
                  </span>
                  <span v-if="model.context_length" class="text-xs text-gray-500 dark:text-gray-400">
                    {{ formatContextLength(model.context_length) }}
                  </span>
                </div>
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenuPortal>
        </DropdownMenuRoot>
      </div>

      <!-- Right side - Actions (only show if models available) -->
      <div v-if="availableModels && availableModels.length > 0" class="flex items-center gap-2">
        <!-- Temperature Control -->
        <DropdownMenuRoot>
          <DropdownMenuTrigger as-child>
            <button
              class="hidden sm:flex items-center space-x-1.5 px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors"
              title="Temperature"
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-gray-600 dark:text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
              </svg>
              <span class="text-sm text-gray-700 dark:text-gray-300">{{ modelParams.temperature }}</span>
            </button>
          </DropdownMenuTrigger>

          <DropdownMenuPortal>
            <DropdownMenuContent
              :side-offset="8"
              :align="'end'"
              class="w-64 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-600 rounded-lg shadow-lg z-50 p-4"
            >
              <div class="space-y-3">
                <div>
                  <label class="text-xs font-semibold text-gray-700 dark:text-gray-300 mb-2 block">
                    Temperature: {{ modelParams.temperature }}
                  </label>
                  <input
                    type="range"
                    :value="modelParams.temperature"
                    @input="updateTemperature($event)"
                    min="0"
                    max="2"
                    step="0.1"
                    class="w-full h-2 bg-gray-200 dark:bg-gray-700 rounded-lg appearance-none cursor-pointer accent-blue-600"
                  />
                  <div class="flex justify-between text-xs text-gray-500 dark:text-gray-400 mt-1">
                    <span>Precise</span>
                    <span>Creative</span>
                  </div>
                </div>
                
                <div>
                  <label class="text-xs font-semibold text-gray-700 dark:text-gray-300 mb-2 block">
                    Max Tokens: {{ modelParams.max_tokens }}
                  </label>
                  <input
                    type="range"
                    :value="modelParams.max_tokens"
                    @input="updateMaxTokens($event)"
                    min="256"
                    max="4096"
                    step="256"
                    class="w-full h-2 bg-gray-200 dark:bg-gray-700 rounded-lg appearance-none cursor-pointer accent-blue-600"
                  />
                  <div class="flex justify-between text-xs text-gray-500 dark:text-gray-400 mt-1">
                    <span>256</span>
                    <span>4096</span>
                  </div>
                </div>
              </div>
            </DropdownMenuContent>
          </DropdownMenuPortal>
        </DropdownMenuRoot>

        <!-- Settings Button (Mobile - shows all params) -->
        <DropdownMenuRoot>
          <DropdownMenuTrigger as-child>
            <button
              class="sm:hidden p-2 rounded-lg border border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors"
              title="Settings"
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-gray-600 dark:text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
            </button>
          </DropdownMenuTrigger>

          <DropdownMenuPortal>
            <DropdownMenuContent
              :side-offset="8"
              :align="'end'"
              class="w-64 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-600 rounded-lg shadow-lg z-50 p-4"
            >
              <div class="space-y-3">
                <div>
                  <label class="text-xs font-semibold text-gray-700 dark:text-gray-300 mb-2 block">
                    Temperature: {{ modelParams.temperature }}
                  </label>
                  <input
                    type="range"
                    :value="modelParams.temperature"
                    @input="updateTemperature($event)"
                    min="0"
                    max="2"
                    step="0.1"
                    class="w-full h-2 bg-gray-200 dark:bg-gray-700 rounded-lg appearance-none cursor-pointer accent-blue-600"
                  />
                </div>
                
                <div>
                  <label class="text-xs font-semibold text-gray-700 dark:text-gray-300 mb-2 block">
                    Max Tokens: {{ modelParams.max_tokens }}
                  </label>
                  <input
                    type="range"
                    :value="modelParams.max_tokens"
                    @input="updateMaxTokens($event)"
                    min="256"
                    max="4096"
                    step="256"
                    class="w-full h-2 bg-gray-200 dark:bg-gray-700 rounded-lg appearance-none cursor-pointer accent-blue-600"
                  />
                </div>
              </div>
            </DropdownMenuContent>
          </DropdownMenuPortal>
        </DropdownMenuRoot>

        <!-- Clear Chat Button -->
        <button
          @click="clearChat"
          class="p-2 rounded-lg border border-gray-200 dark:border-gray-700 hover:border-red-300 dark:hover:border-red-700 hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors group"
          title="Clear Chat"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-gray-600 dark:text-gray-400 group-hover:text-red-600 dark:group-hover:text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
          </svg>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, defineProps, defineEmits } from 'vue'
import { ChevronDown } from 'lucide-vue-next'
import { 
  DropdownMenuRoot,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuPortal,
  DropdownMenuItem,
  DropdownMenuLabel
} from 'radix-vue'
import type { Model } from '@/types'

const props = defineProps<{
  selectedModel: string
  availableModels: Model[]
  modelParams: any
}>()

const emit = defineEmits(['update:selectedModel', 'update:modelParams', 'clearChat'])

const selectModel = (model: Model) => {
  emit('update:selectedModel', model.id)
}

const clearChat = () => {
  emit('clearChat')
}

const updateTemperature = (event: Event) => {
  const target = event.target as HTMLInputElement
  emit('update:modelParams', {
    ...props.modelParams,
    temperature: parseFloat(target.value)
  })
}

const updateMaxTokens = (event: Event) => {
  const target = event.target as HTMLInputElement
  emit('update:modelParams', {
    ...props.modelParams,
    max_tokens: parseInt(target.value)
  })
}

const getModelDisplayName = (modelId: string): string => {
  const model = props.availableModels.find(m => m.id === modelId)
  return model?.name || modelId
}

const formatContextLength = (length: number): string => {
  if (length >= 1000000) {
    return `${(length / 1000000).toFixed(1)}M context`
  } else if (length >= 1000) {
    return `${(length / 1000).toFixed(0)}K context`
  }
  return `${length} tokens`
}

const getModelStatusColor = (modelId: string): string => {
  // All models shown are available (have API keys), so always show green
  return 'bg-green-500'
}

const getBadgeClass = (badge: string): string => {
  if (badge.includes('⚡')) return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
  if (badge.includes('🧠')) return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200'
  if (badge.includes('📝')) return 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200'
  if (badge.includes('🔍')) return 'bg-indigo-100 text-indigo-800 dark:bg-indigo-900 dark:text-indigo-200'
  if (badge.includes('💻')) return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200'
  return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200'
}
</script>