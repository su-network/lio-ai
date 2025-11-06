<template>
  <div class="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 px-3 sm:px-4 py-3">
    <div class="flex items-center justify-between">
      <!-- Model Selector using Radix Dropdown Menu -->
      <DropdownMenuRoot v-if="availableModels && availableModels.length > 0">
        <DropdownMenuTrigger as-child>
          <button
            class="flex items-center space-x-2 px-3 sm:px-4 py-2 rounded-lg border border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600 transition-colors"
          >
            <span class="text-sm sm:text-base font-medium text-gray-900 dark:text-white">
              {{ getModelDisplayName(selectedModel) }}
            </span>
            <ChevronDown class="w-4 h-4 text-gray-500 dark:text-gray-400" />
          </button>
        </DropdownMenuTrigger>

        <DropdownMenuPortal>
          <DropdownMenuContent
            :side-offset="8"
            class="w-full min-w-[240px] sm:w-80 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-600 rounded-lg shadow-lg z-50 max-h-96 overflow-y-auto p-2"
          >
            <DropdownMenuLabel class="text-xs font-semibold text-gray-500 dark:text-gray-400 mb-2 px-3 uppercase tracking-wide">
              Select Model
            </DropdownMenuLabel>
            <DropdownMenuItem
              v-for="model in availableModels"
              :key="model.id"
              @click="selectModel(model)"
              class="flex items-center justify-between p-3 rounded-lg cursor-pointer transition-colors outline-none focus:bg-gray-100 dark:focus:bg-gray-700 data-[highlighted]:bg-gray-100 dark:data-[highlighted]:bg-gray-700"
              :class="{ 'bg-gray-50 dark:bg-gray-750': selectedModel === model.id }"
            >
              <span class="text-sm font-medium text-gray-900 dark:text-white">{{ model.name }}</span>
              <div v-if="selectedModel === model.id" class="w-2 h-2 rounded-full bg-gray-900 dark:bg-gray-100"></div>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenuPortal>
      </DropdownMenuRoot>
      <div v-else class="text-sm text-gray-500 dark:text-gray-400">
        <!-- Empty space when no models -->
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

const getModelDisplayName = (modelId: string): string => {
  const model = props.availableModels.find(m => m.id === modelId)
  return model?.name || modelId
}

const getModelStatusColor = (modelId: string): string => {
  const model = props.availableModels.find(m => m.id === modelId)
  if (model?.status === 'Online') return 'bg-green-500'
  if (model?.status === 'Beta') return 'bg-yellow-500'
  return 'bg-gray-500'
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