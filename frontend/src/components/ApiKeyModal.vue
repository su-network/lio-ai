<template>
  <teleport to="body">
    <div v-if="isOpen" class="fixed inset-0 z-[9999] overflow-y-auto">
      <div class="flex items-center justify-center min-h-screen px-4 py-8">
        <!-- Background overlay -->
        <div class="fixed inset-0 bg-gray-900/20 dark:bg-gray-900/40 backdrop-blur-sm transition-opacity" @click="close"></div>

        <!-- Modal panel -->
        <div class="relative bg-white dark:bg-gray-800 rounded-2xl shadow-2xl transform transition-all w-full max-w-3xl z-[10000] border border-gray-200 dark:border-gray-700">
          <!-- Header -->
          <div class="bg-gradient-to-r from-blue-600 via-purple-600 to-blue-600 px-8 py-6 rounded-t-2xl">
            <div class="flex items-center justify-between">
              <div>
                <h3 class="text-2xl font-bold text-white">Configure API Keys</h3>
                <p class="mt-2 text-sm text-blue-100">Add your API keys to enable AI models from different providers</p>
              </div>
              <button @click="close" class="text-white hover:text-gray-200 transition-colors p-2 hover:bg-white/10 rounded-lg">
                <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                </svg>
              </button>
            </div>
          </div>

          <!-- Body -->
          <div class="px-8 py-6">
            <form @submit.prevent="saveApiKey">
              <!-- Provider Selection -->
              <div class="mb-6">
                <label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">
                  Select Provider
                </label>
                <div class="grid grid-cols-2 gap-3">
                  <button
                    v-for="provider in providers"
                    :key="provider.id"
                    type="button"
                    @click="selectProvider(provider.id)"
                    :class="[
                      'p-4 rounded-xl border-2 transition-all text-left',
                      form.provider === provider.id
                        ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
                        : 'border-gray-200 dark:border-gray-700 hover:border-blue-300 dark:hover:border-blue-700 bg-white dark:bg-gray-900/50'
                    ]"
                  >
                    <div class="flex items-center gap-3">
                      <div :class="[
                        'w-12 h-12 rounded-lg flex items-center justify-center',
                        provider.color
                      ]">
                        <svg class="w-6 h-6" fill="currentColor" viewBox="0 0 20 20">
                          <path fill-rule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clip-rule="evenodd"/>
                        </svg>
                      </div>
                      <div class="flex-1">
                        <p class="font-bold text-gray-900 dark:text-white">{{ provider.name }}</p>
                        <p class="text-xs text-gray-500 dark:text-gray-400">{{ provider.models }}</p>
                      </div>
                      <div v-if="form.provider === provider.id" class="text-blue-600">
                        <svg class="w-6 h-6" fill="currentColor" viewBox="0 0 20 20">
                          <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"/>
                        </svg>
                      </div>
                    </div>
                  </button>
                </div>
              </div>

              <!-- Provider Info -->
              <div v-if="form.provider" class="mb-6 p-5 bg-gradient-to-r from-blue-50 to-purple-50 dark:from-blue-900/20 dark:to-purple-900/20 border-2 border-blue-200 dark:border-blue-800 rounded-xl">
                <div class="flex items-start gap-3">
                  <div class="flex-shrink-0 w-10 h-10 bg-blue-600 rounded-lg flex items-center justify-center">
                    <svg class="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 20 20">
                      <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd"/>
                    </svg>
                  </div>
                  <div class="flex-1">
                    <p class="text-sm font-bold text-blue-900 dark:text-blue-300">{{ providerInfo[form.provider]?.name }}</p>
                    <p class="text-xs text-blue-800 dark:text-blue-400 mt-1">{{ providerInfo[form.provider]?.description }}</p>
                    <a :href="providerInfo[form.provider]?.url" target="_blank" class="inline-flex items-center gap-1 text-xs font-semibold text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 mt-2">
                      Get API Key
                      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"/>
                      </svg>
                    </a>
                  </div>
                </div>
              </div>

              <!-- API Key Input -->
              <div class="mb-6">
                <label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">
                  API Key
                </label>
                <div class="relative">
                  <input
                    v-model="form.apiKey"
                    :type="showKey ? 'text' : 'password'"
                    placeholder="sk-..."
                    class="w-full px-4 py-3 pr-12 border-2 border-gray-300 dark:border-gray-600 rounded-xl bg-white dark:bg-gray-900 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500 font-mono text-sm transition-all"
                    required
                  />
                  <button
                    type="button"
                    @click="showKey = !showKey"
                    class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                  >
                    <svg v-if="!showKey" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/>
                    </svg>
                    <svg v-else class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21"/>
                    </svg>
                  </button>
                </div>
                <div class="flex items-center gap-2 mt-2 text-xs text-gray-500 dark:text-gray-400">
                  <svg class="w-4 h-4 text-green-600 dark:text-green-400" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M2.166 4.999A11.954 11.954 0 0010 1.944 11.954 11.954 0 0017.834 5c.11.65.166 1.32.166 2.001 0 5.225-3.34 9.67-8 11.317C5.34 16.67 2 12.225 2 7c0-.682.057-1.35.166-2.001zm11.541 3.708a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"/>
                  </svg>
                  <span>Your API key is encrypted with AES-256 and stored securely</span>
                </div>
              </div>

              <!-- Models Selection -->
              <div v-if="form.provider && availableModelsForProvider.length > 0" class="mb-6">
                <label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">
                  Enable Specific Models (Optional)
                </label>
                <div class="space-y-2 max-h-64 overflow-y-auto border-2 border-gray-200 dark:border-gray-700 rounded-xl p-4 bg-gray-50 dark:bg-gray-900/50">
                  <label
                    v-for="model in availableModelsForProvider"
                    :key="model.id"
                    class="flex items-center space-x-3 p-3 hover:bg-white dark:hover:bg-gray-800 rounded-lg cursor-pointer transition-all border border-transparent hover:border-blue-200 dark:hover:border-blue-800"
                  >
                    <input
                      type="checkbox"
                      :value="model.id"
                      v-model="form.modelsEnabled"
                      class="w-5 h-5 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                    />
                    <div class="flex-1">
                      <span class="text-sm font-semibold text-gray-900 dark:text-white">{{ model.name }}</span>
                      <span class="text-xs text-gray-500 dark:text-gray-400 ml-2">({{ model.capabilities.context_window.toLocaleString() }} tokens)</span>
                    </div>
                  </label>
                </div>
                <p class="text-xs text-gray-500 dark:text-gray-400 mt-2 flex items-center gap-2">
                  <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd"/>
                  </svg>
                  Leave unchecked to enable all models from this provider
                </p>
              </div>

              <!-- Action Buttons -->
              <div class="flex justify-end space-x-3 pt-6 border-t-2 border-gray-200 dark:border-gray-700">
                <button
                  type="button"
                  @click="close"
                  class="px-6 py-3 text-sm font-semibold text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 border-2 border-gray-300 dark:border-gray-600 rounded-xl hover:bg-gray-50 dark:hover:bg-gray-600 transition-all shadow-sm hover:shadow-md"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  :disabled="loading || !form.provider || !form.apiKey"
                  class="px-6 py-3 text-sm font-semibold text-white bg-gradient-to-r from-blue-600 to-purple-600 rounded-xl hover:from-blue-700 hover:to-purple-700 disabled:opacity-50 disabled:cursor-not-allowed transition-all shadow-lg hover:shadow-xl flex items-center gap-2"
                >
                  <svg v-if="loading" class="animate-spin w-5 h-5" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  <svg v-else class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
                  </svg>
                  {{ loading ? 'Saving...' : 'Save API Key' }}
                </button>
              </div>
            </form>

            <!-- Error Message -->
            <div v-if="error" class="mt-4 p-4 bg-red-50 dark:bg-red-900/20 border-2 border-red-200 dark:border-red-800 rounded-xl flex items-start gap-3">
              <svg class="w-5 h-5 text-red-600 dark:text-red-400 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd"/>
              </svg>
              <div>
                <p class="text-sm font-semibold text-red-800 dark:text-red-400">Error</p>
                <p class="text-sm text-red-700 dark:text-red-300">{{ error }}</p>
              </div>
            </div>

            <!-- Success Message -->
            <div v-if="success" class="mt-4 p-4 bg-green-50 dark:bg-green-900/20 border-2 border-green-200 dark:border-green-800 rounded-xl flex items-start gap-3">
              <svg class="w-5 h-5 text-green-600 dark:text-green-400 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"/>
              </svg>
              <div>
                <p class="text-sm font-semibold text-green-800 dark:text-green-400">Success!</p>
                <p class="text-sm text-green-700 dark:text-green-300">API key saved successfully!</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { apiService } from '@/services/api'
import { useUserStore } from '@/stores/user'
import type { Model } from '@/types'

interface Props {
  isOpen: boolean
  allModels?: Model[]
}

const props = withDefaults(defineProps<Props>(), {
  allModels: () => []
})
const emit = defineEmits(['close', 'saved'])

const userStore = useUserStore()

const providers = [
  {
    id: 'openai',
    name: 'OpenAI',
    models: 'GPT-4, GPT-3.5',
    color: 'bg-green-100 dark:bg-green-900/30 text-green-600 dark:text-green-400'
  },
  {
    id: 'anthropic',
    name: 'Anthropic',
    models: 'Claude 3',
    color: 'bg-orange-100 dark:bg-orange-900/30 text-orange-600 dark:text-orange-400'
  },
  {
    id: 'google',
    name: 'Google AI',
    models: 'Gemini',
    color: 'bg-blue-100 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400'
  },
  {
    id: 'cohere',
    name: 'Cohere',
    models: 'Command R+',
    color: 'bg-purple-100 dark:bg-purple-900/30 text-purple-600 dark:text-purple-400'
  },
  {
    id: 'ollama',
    name: 'Ollama',
    models: 'Local Models',
    color: 'bg-gray-100 dark:bg-gray-900/30 text-gray-600 dark:text-gray-400'
  },
  {
    id: 'bedrock',
    name: 'AWS Bedrock',
    models: 'Claude, Llama',
    color: 'bg-yellow-100 dark:bg-yellow-900/30 text-yellow-600 dark:text-yellow-400'
  },
  {
    id: 'azure',
    name: 'Azure OpenAI',
    models: 'GPT-4, GPT-3.5',
    color: 'bg-cyan-100 dark:bg-cyan-900/30 text-cyan-600 dark:text-cyan-400'
  }
]

const form = ref({
  provider: '',
  apiKey: '',
  modelsEnabled: [] as string[]
})

const loading = ref(false)
const error = ref('')
const success = ref(false)
const showKey = ref(false)

const providerInfo: Record<string, any> = {
  openai: {
    name: 'OpenAI',
    description: 'Access GPT-4, GPT-4 Turbo, and GPT-3.5 Turbo models',
    url: 'https://platform.openai.com/api-keys'
  },
  anthropic: {
    name: 'Anthropic',
    description: 'Access Claude 3 Opus, Sonnet, and Haiku models',
    url: 'https://console.anthropic.com/'
  },
  google: {
    name: 'Google AI',
    description: 'Access Gemini Pro and Gemini 1.5 Pro models',
    url: 'https://makersuite.google.com/app/apikey'
  },
  cohere: {
    name: 'Cohere',
    description: 'Access Command R+ and other Cohere models',
    url: 'https://dashboard.cohere.com/api-keys'
  },
  ollama: {
    name: 'Ollama',
    description: 'Run open-source models locally (Llama 3, Mistral, Code Llama, etc.)',
    url: 'https://ollama.ai/'
  },
  bedrock: {
    name: 'AWS Bedrock',
    description: 'Access Claude, Llama, and other models via AWS Bedrock',
    url: 'https://console.aws.amazon.com/bedrock/'
  },
  azure: {
    name: 'Azure OpenAI',
    description: 'Access OpenAI models through Microsoft Azure',
    url: 'https://portal.azure.com/#create/Microsoft.CognitiveServicesOpenAI'
  }
}

const availableModelsForProvider = computed(() => {
  if (!form.value.provider || !props.allModels) return []
  return props.allModels.filter(m => m.provider === form.value.provider)
})

const selectProvider = (providerId: string) => {
  form.value.provider = providerId
}

watch(() => form.value.provider, () => {
  form.value.modelsEnabled = []
  error.value = ''
  success.value = false
})

const saveApiKey = async () => {
  try {
    loading.value = true
    error.value = ''
    success.value = false

    await apiService.createProviderKey(
      form.value.provider,
      form.value.apiKey,
      form.value.modelsEnabled
    )

    success.value = true
    setTimeout(() => {
      emit('saved')
      close()
    }, 1500)

  } catch (err: any) {
    console.error('Save API Key Error:', err)
    console.error('Response data:', err.response?.data)
    
    // Show detailed error message
    const errorMsg = err.response?.data?.error || 'Failed to save API key'
    const details = err.response?.data?.details
    error.value = details ? `${errorMsg}: ${details}` : errorMsg
  } finally {
    loading.value = false
  }
}

const close = () => {
  form.value = {
    provider: '',
    apiKey: '',
    modelsEnabled: []
  }
  error.value = ''
  success.value = false
  showKey.value = false
  emit('close')
}
</script>
