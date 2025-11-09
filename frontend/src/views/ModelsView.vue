<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900 py-8">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <!-- Header -->
      <div class="mb-8">
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white mb-2">AI Models</h1>
        <p class="text-base text-gray-600 dark:text-gray-400">
          Explore our collection of AI models from leading providers including Google Gemini 2.5 Pro
        </p>
      </div>

      <!-- Quick Stats -->
      <div v-if="modelsStatus" class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
        <div class="bg-white dark:bg-gray-800 rounded-lg p-5 border border-gray-200 dark:border-gray-700">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-xs text-gray-600 dark:text-gray-400 mb-1">Total Models</p>
              <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ modelsStatus.total_models }}</p>
            </div>
            <div class="p-2.5 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
              <svg class="w-6 h-6 text-blue-600 dark:text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"/>
              </svg>
            </div>
          </div>
        </div>

        <div class="bg-white dark:bg-gray-800 rounded-lg p-5 border border-gray-200 dark:border-gray-700">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-xs text-gray-600 dark:text-gray-400 mb-1">Available</p>
              <p class="text-2xl font-bold text-green-600 dark:text-green-400">{{ modelsStatus.available }}</p>
            </div>
            <div class="p-2.5 bg-green-50 dark:bg-green-900/20 rounded-lg">
              <svg class="w-6 h-6 text-green-600 dark:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
              </svg>
            </div>
          </div>
        </div>

        <div class="bg-white dark:bg-gray-800 rounded-lg p-5 border border-gray-200 dark:border-gray-700">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-xs text-gray-600 dark:text-gray-400 mb-1">Need API Keys</p>
              <p class="text-2xl font-bold text-yellow-600 dark:text-yellow-400">{{ modelsStatus.no_api_key }}</p>
            </div>
            <div class="p-2.5 bg-yellow-50 dark:bg-yellow-900/20 rounded-lg">
              <svg class="w-6 h-6 text-yellow-600 dark:text-yellow-400" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clip-rule="evenodd"/>
              </svg>
            </div>
          </div>
        </div>

        <div class="bg-white dark:bg-gray-800 rounded-lg p-5 border border-gray-200 dark:border-gray-700">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-xs text-gray-600 dark:text-gray-400 mb-1">Providers</p>
              <p class="text-2xl font-bold text-purple-600 dark:text-purple-400">4</p>
            </div>
            <div class="p-2.5 bg-purple-50 dark:bg-purple-900/20 rounded-lg">
              <svg class="w-6 h-6 text-purple-600 dark:text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"/>
              </svg>
            </div>
          </div>
        </div>
      </div>

      <!-- Call to Action -->
      <div v-if="modelsStatus && modelsStatus.no_api_key > 0" class="mb-6 bg-gradient-to-r from-blue-600 to-blue-500 rounded-lg p-5 text-white">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3">
            <svg class="w-10 h-10 opacity-80" fill="currentColor" viewBox="0 0 20 20">
              <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd"/>
            </svg>
            <div>
              <h3 class="text-base font-semibold">Configure API Keys to Unlock Models</h3>
              <p class="text-blue-100 text-sm mt-0.5">{{ modelsStatus.no_api_key }} models are waiting for API keys. Add your keys in Settings to enable them.</p>
            </div>
          </div>
          <router-link
            to="/settings"
            class="px-5 py-2.5 bg-white text-blue-600 rounded-lg hover:bg-blue-50 transition-colors font-medium whitespace-nowrap"
          >
            Go to Settings →
          </router-link>
        </div>
      </div>

      <!-- Filter Tabs -->
      <div class="flex flex-wrap gap-2 mb-6">
        <button
          @click="filterStatus = 'all'"
          :class="[
            'px-4 py-2 rounded-lg text-sm font-medium transition-colors',
            filterStatus === 'all'
              ? 'bg-blue-600 text-white'
              : 'bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 border border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700'
          ]"
        >
          All Models
        </button>
        <button
          @click="filterStatus = 'available'"
          :class="[
            'px-4 py-2 rounded-lg text-sm font-medium transition-colors',
            filterStatus === 'available'
              ? 'bg-green-600 text-white'
              : 'bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 border border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700'
          ]"
        >
          Available
        </button>
        <button
          @click="filterStatus = 'locked'"
          :class="[
            'px-4 py-2 rounded-lg text-sm font-medium transition-colors',
            filterStatus === 'locked'
              ? 'bg-yellow-600 text-white'
              : 'bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 border border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700'
          ]"
        >
          Need API Key
        </button>

        <!-- Provider Filters -->
        <div class="ml-auto flex gap-2">
          <button
            v-for="provider in ['openai', 'anthropic', 'google', 'cohere']"
            :key="provider"
            @click="toggleProvider(provider)"
            :class="[
              'px-4 py-2 rounded-lg text-sm font-medium transition-colors capitalize',
              selectedProviders.includes(provider)
                ? 'bg-gray-900 dark:bg-gray-700 text-white'
                : 'bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 border border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700'
            ]"
          >
            {{ provider }}
          </button>
        </div>
      </div>

      <!-- Loading State -->
      <LoadingSpinner 
        v-if="systemStore.loading" 
        type="spinner" 
        :size="48" 
        color="blue" 
        text="Loading models..."
      />

      <!-- Models Grid -->
      <div v-else class="grid md:grid-cols-2 lg:grid-cols-3 gap-5">
        <div
          v-for="model in filteredModels"
          :key="model.id"
          class="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 hover:shadow-md transition-shadow duration-200 overflow-hidden group"
        >
          <!-- Model Header -->
          <div class="p-5 border-b border-gray-100 dark:border-gray-700">
            <div class="flex items-start justify-between mb-2.5">
              <div class="flex-1 min-w-0">
                <h3 class="text-lg font-semibold text-gray-900 dark:text-white group-hover:text-blue-600 dark:group-hover:text-blue-400 transition-colors truncate">
                  {{ model.name }}
                </h3>
                <p class="text-xs text-gray-500 dark:text-gray-400 font-mono mt-0.5 truncate">{{ model.model_name }}</p>
              </div>
              <span
                :class="[
                  'px-2.5 py-1 rounded-md text-xs font-medium capitalize flex-shrink-0 ml-2',
                  model.provider === 'openai' ? 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-300' :
                  model.provider === 'anthropic' ? 'bg-orange-100 dark:bg-orange-900/30 text-orange-700 dark:text-orange-300' :
                  model.provider === 'google' ? 'bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300' :
                  'bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-300'
                ]"
              >
                {{ model.provider }}
              </span>
            </div>

            <!-- Status Badge -->
            <div class="flex items-center gap-2">
              <span
                v-if="model.status === 'available'"
                class="inline-flex items-center gap-1.5 px-2.5 py-1 bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-300 rounded-md text-xs font-medium"
              >
                <span class="w-1.5 h-1.5 bg-green-600 rounded-full"></span>
                Available
              </span>
              <span
                v-else-if="model.status === 'no_api_key'"
                class="inline-flex items-center gap-1.5 px-2.5 py-1 bg-yellow-100 dark:bg-yellow-900/30 text-yellow-700 dark:text-yellow-300 rounded-md text-xs font-medium"
              >
                <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clip-rule="evenodd"/>
                </svg>
                Locked
              </span>
              <span
                v-else
                class="inline-flex items-center gap-1.5 px-2.5 py-1 bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 rounded-md text-xs font-medium"
              >
                Disabled
              </span>
            </div>
          </div>

          <!-- Model Details -->
          <div class="p-5">
            <!-- Metrics Grid -->
            <div class="grid grid-cols-2 gap-2.5 mb-4">
              <div class="bg-gray-50 dark:bg-gray-900/30 rounded-lg p-2.5">
                <p class="text-xs text-gray-500 dark:text-gray-400 mb-0.5">Context</p>
                <p class="text-sm font-semibold text-gray-900 dark:text-white">{{ formatNumber(model.capabilities.context_window) }}</p>
              </div>
              <div class="bg-gray-50 dark:bg-gray-900/30 rounded-lg p-2.5">
                <p class="text-xs text-gray-500 dark:text-gray-400 mb-0.5">Latency</p>
                <p class="text-sm font-semibold text-gray-900 dark:text-white">{{ model.metrics.average_latency_ms }}ms</p>
              </div>
              <div class="bg-gray-50 dark:bg-gray-900/30 rounded-lg p-2.5">
                <p class="text-xs text-gray-500 dark:text-gray-400 mb-0.5">Cost</p>
                <p class="text-sm font-semibold text-gray-900 dark:text-white">${{ model.metrics.cost_per_request.toFixed(3) }}</p>
              </div>
              <div class="bg-gray-50 dark:bg-gray-900/30 rounded-lg p-2.5">
                <p class="text-xs text-gray-500 dark:text-gray-400 mb-0.5">Success</p>
                <p class="text-sm font-semibold text-gray-900 dark:text-white">{{ (model.metrics.success_rate * 100).toFixed(0) }}%</p>
              </div>
            </div>

            <!-- Features -->
            <div class="space-y-2.5">
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400 mb-1.5">Languages</p>
                <div class="flex flex-wrap gap-1">
                  <span
                    v-for="lang in model.capabilities.languages.slice(0, 4)"
                    :key="lang"
                    class="px-2 py-0.5 text-xs bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-300 rounded"
                  >
                    {{ lang }}
                  </span>
                  <span
                    v-if="model.capabilities.languages.length > 4"
                    class="px-2 py-0.5 text-xs text-gray-500 dark:text-gray-400"
                  >
                    +{{ model.capabilities.languages.length - 4 }}
                  </span>
                </div>
              </div>

              <div v-if="model.capabilities.special_features?.length > 0">
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400 mb-1.5">Features</p>
                <div class="flex flex-wrap gap-1">
                  <span
                    v-for="feature in model.capabilities.special_features.slice(0, 3)"
                    :key="feature"
                    class="px-2 py-0.5 text-xs bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded"
                  >
                    {{ feature }}
                  </span>
                  <span
                    v-if="model.capabilities.special_features.length > 3"
                    class="px-2 py-0.5 text-xs text-gray-500 dark:text-gray-400"
                  >
                    +{{ model.capabilities.special_features.length - 3 }}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Empty State -->
      <div v-if="!systemStore.loading && filteredModels.length === 0" class="text-center py-16">
        <svg class="mx-auto h-16 w-16 text-gray-400 dark:text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z"/>
        </svg>
        <p class="mt-4 text-lg text-gray-500 dark:text-gray-400">No models match your filters</p>
        <button
          @click="clearFilters"
          class="mt-4 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
        >
          Clear Filters
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useSystemStore } from '@/stores/system'
import { apiService } from '@/services/api'
import LoadingSpinner from '@/components/LoadingSpinner.vue'
import type { Model } from '@/types'

const systemStore = useSystemStore()

const modelsStatus = ref<any>(null)
const filterStatus = ref<'all' | 'available' | 'locked'>('all')
const selectedProviders = ref<string[]>([])

const filteredModels = computed(() => {
  if (!modelsStatus.value?.models) return systemStore.models

  // Create status map
  const statusMap = new Map()
  modelsStatus.value.models.forEach((status: any) => {
    statusMap.set(status.model_id, status)
  })

  // Get all models with status
  let models = modelsStatus.value.models.map((status: any) => {
    const modelConfig = systemStore.models.find((m: any) => m.id === status.model_id)
    if (modelConfig) {
      return {
        ...modelConfig,
        status: status.status,
        reason: status.reason
      }
    }
    return null
  }).filter(Boolean)

  // Apply status filter
  if (filterStatus.value === 'available') {
    models = models.filter((m: any) => m.status === 'available')
  } else if (filterStatus.value === 'locked') {
    models = models.filter((m: any) => m.status === 'no_api_key')
  }

  // Apply provider filter
  if (selectedProviders.value.length > 0) {
    models = models.filter((m: any) => selectedProviders.value.includes(m.provider))
  }

  return models
})

const toggleProvider = (provider: string) => {
  const index = selectedProviders.value.indexOf(provider)
  if (index > -1) {
    selectedProviders.value.splice(index, 1)
  } else {
    selectedProviders.value.push(provider)
  }
}

const clearFilters = () => {
  filterStatus.value = 'all'
  selectedProviders.value = []
}

const formatNumber = (num: number) => {
  if (num >= 1000000) return `${(num / 1000000).toFixed(1)}M`
  if (num >= 1000) return `${(num / 1000).toFixed(0)}K`
  return num.toString()
}

const loadData = async () => {
  await Promise.allSettled([
    systemStore.fetchAvailableModels(),
    apiService.getModelsStatus().then(status => {
      modelsStatus.value = status
    })
  ])
}

onMounted(() => {
  loadData()
})
</script>
