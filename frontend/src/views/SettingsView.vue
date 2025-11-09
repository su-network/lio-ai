<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900">
    <div class="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
      <!-- Page Header -->
      <div class="mb-8">
        <h1 class="text-4xl font-bold text-gray-900 dark:text-white mb-2">Settings</h1>
        <p class="text-lg text-gray-600 dark:text-gray-400">
          Configure your API keys to unlock AI models
        </p>
      </div>

      <!-- Main Settings Card -->
      <div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 overflow-hidden">
        <!-- Card Header -->
        <div class="bg-gray-50 dark:bg-gray-800/50 px-8 py-6 border-b border-gray-200 dark:border-gray-700">
          <div class="flex items-center justify-between">
            <div>
              <h2 class="text-xl font-semibold text-gray-900 dark:text-white">API Keys</h2>
              <p class="text-sm text-gray-600 dark:text-gray-400 mt-1">Manage your provider API keys</p>
            </div>
            <button
              @click="openModal"
              class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium flex items-center gap-2"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
              </svg>
              Add API Key
            </button>
          </div>
        </div>

        <!-- Card Body -->
        <div class="p-8">
          <!-- Loading -->
          <div v-if="loading" class="text-center py-12">
            <div class="inline-block animate-spin rounded-full h-10 w-10 border-b-2 border-blue-600"></div>
            <p class="mt-3 text-sm text-gray-600 dark:text-gray-400">Loading...</p>
          </div>

          <!-- Keys List -->
          <div v-else-if="keys.length > 0" class="space-y-3">
            <div
              v-for="key in keys"
              :key="key.id"
              class="flex items-center justify-between p-5 bg-gray-50 dark:bg-gray-900/30 rounded-lg border border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600 transition-colors"
            >
              <div class="flex items-center gap-4 flex-1">
                <!-- Icon -->
                <div
                  class="w-12 h-12 rounded-lg flex items-center justify-center"
                  :class="getProviderColor(key.provider)"
                >
                  <svg class="w-6 h-6" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clip-rule="evenodd"/>
                  </svg>
                </div>

                <!-- Info -->
                <div class="flex-1">
                  <div class="flex items-center gap-2 mb-1">
                    <h3 class="text-lg font-semibold text-gray-900 dark:text-white capitalize">
                      {{ key.provider }}
                    </h3>
                    <span class="px-2 py-0.5 bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-300 text-xs font-medium rounded">
                      Active
                    </span>
                  </div>
                  <p class="text-sm text-gray-600 dark:text-gray-400">
                    {{ key.models_enabled?.length || 'All' }} models • 
                    {{ key.last_used_at ? 'Last used ' + formatDate(key.last_used_at) : 'Not used yet' }}
                  </p>
                </div>
              </div>

              <!-- Delete Button -->
              <button
                @click="openDeleteModal(key.provider)"
                class="px-3 py-2 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors text-sm font-medium flex items-center gap-2"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
                </svg>
                Delete
              </button>
            </div>
          </div>

          <!-- Empty State -->
          <div v-else class="text-center py-12">
            <div class="w-16 h-16 mx-auto mb-4 bg-gray-100 dark:bg-gray-800 rounded-full flex items-center justify-center">
              <svg class="w-8 h-8 text-gray-400 dark:text-gray-600" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clip-rule="evenodd"/>
              </svg>
            </div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">No API Keys</h3>
            <p class="text-sm text-gray-600 dark:text-gray-400 mb-6 max-w-md mx-auto">
              Add your first API key to start using AI models from different providers
            </p>
            <button
              @click="openModal"
              class="px-6 py-2.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium inline-flex items-center gap-2"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
              </svg>
              Add Your First API Key
            </button>
          </div>
        </div>
      </div>

      <!-- Info Card -->
      <div class="mt-6 bg-blue-50 dark:bg-blue-900/10 rounded-xl border border-blue-200 dark:border-blue-900/50 p-6">
        <div class="flex gap-4">
          <div class="flex-shrink-0">
            <div class="w-10 h-10 bg-blue-600 rounded-lg flex items-center justify-center">
              <svg class="w-5 h-5 text-white" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd"/>
              </svg>
            </div>
          </div>
          <div class="flex-1">
            <h4 class="text-base font-semibold text-blue-900 dark:text-blue-300 mb-3">Security & Privacy</h4>
            <ul class="space-y-2 text-sm text-blue-800 dark:text-blue-300">
              <li class="flex items-start gap-2">
                <svg class="w-5 h-5 flex-shrink-0 mt-0.5 text-blue-600 dark:text-blue-400" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"/>
                </svg>
                <span><strong>Encrypted Storage:</strong> All keys are encrypted with AES-256</span>
              </li>
              <li class="flex items-start gap-2">
                <svg class="w-5 h-5 flex-shrink-0 mt-0.5 text-blue-600 dark:text-blue-400" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"/>
                </svg>
                <span><strong>Private & Secure:</strong> Never shared with third parties</span>
              </li>
              <li class="flex items-start gap-2">
                <svg class="w-5 h-5 flex-shrink-0 mt-0.5 text-blue-600 dark:text-blue-400" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"/>
                </svg>
                <span><strong>Get Keys:</strong> 
                  <a href="https://platform.openai.com" target="_blank" class="underline hover:text-blue-700">OpenAI</a> • 
                  <a href="https://console.anthropic.com" target="_blank" class="underline hover:text-blue-700">Anthropic</a> • 
                  <a href="https://makersuite.google.com/app/apikey" target="_blank" class="underline hover:text-blue-700">Google</a> • 
                  <a href="https://dashboard.cohere.com" target="_blank" class="underline hover:text-blue-700">Cohere</a>
                </span>
              </li>
            </ul>
          </div>
        </div>
      </div>
    </div>

    <!-- Modal -->
    <ApiKeyModal
      :isOpen="showModal"
      :allModels="[]"
      @close="closeModal"
      @saved="handleSaved"
    />

    <!-- Delete Confirmation Modal -->
    <DeleteConfirmModal
      :isOpen="showDeleteModal"
      :title="`Delete ${providerToDelete} API Key?`"
      :message="`Are you sure you want to delete your ${providerToDelete} API key? This will disable all ${providerToDelete} models.`"
      :warningMessage="`Models from ${providerToDelete} will no longer be available in the chat interface.`"
      confirmText="Delete API Key"
      :loading="deleting"
      @confirm="handleDeleteConfirm"
      @cancel="closeDeleteModal"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import { useChatStore } from '@/stores/chat'
import { useToast } from '@/composables/useToast'
import { apiService } from '@/services/api'
import ApiKeyModal from '@/components/ApiKeyModal.vue'
import DeleteConfirmModal from '@/components/DeleteConfirmModal.vue'

const userStore = useUserStore()
const chatStore = useChatStore()
const toast = useToast()
const loading = ref(false)
const keys = ref<any[]>([])
const showModal = ref(false)
const showDeleteModal = ref(false)
const providerToDelete = ref('')
const deleting = ref(false)

const getProviderColor = (provider: string) => {
  const colors: Record<string, string> = {
    openai: 'bg-green-100 dark:bg-green-900/30 text-green-600 dark:text-green-400',
    anthropic: 'bg-orange-100 dark:bg-orange-900/30 text-orange-600 dark:text-orange-400',
    google: 'bg-blue-100 dark:bg-blue-900/30 text-blue-600 dark:text-blue-400',
    cohere: 'bg-purple-100 dark:bg-purple-900/30 text-purple-600 dark:text-purple-400'
  }
  return colors[provider] || 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-400'
}

const formatDate = (dateString: string) => {
  const date = new Date(dateString)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))

  if (diffDays === 0) return 'today'
  if (diffDays === 1) return 'yesterday'
  if (diffDays < 7) return `${diffDays} days ago`
  if (diffDays < 30) return `${Math.floor(diffDays / 7)} weeks ago`
  return date.toLocaleDateString()
}

const loadKeys = async () => {
  loading.value = true
  try {
    keys.value = await apiService.getProviderKeys(userStore.userId)
  } catch (error) {
    console.error('Failed to load API keys:', error)
  } finally {
    loading.value = false
  }
}

const openDeleteModal = (provider: string) => {
  providerToDelete.value = provider
  showDeleteModal.value = true
}

const closeDeleteModal = () => {
  showDeleteModal.value = false
  providerToDelete.value = ''
}

const handleDeleteConfirm = async () => {
  deleting.value = true
  try {
    await apiService.deleteProviderKey(userStore.userId, providerToDelete.value)
    
    // Reload keys in settings
    await loadKeys()
    
    // Sync with chat store - reload available models
    await chatStore.loadAvailableModels()
    
    // Show success toast
    toast.success(
      'API Key Deleted',
      `${providerToDelete.value} API key has been removed successfully.`
    )
    
    closeDeleteModal()
  } catch (error) {
    console.error('Failed to delete key:', error)
    toast.error(
      'Delete Failed',
      'Could not delete API key. Please try again.'
    )
  } finally {
    deleting.value = false
  }
}

const openModal = () => {
  showModal.value = true
}

const closeModal = () => {
  showModal.value = false
}

const handleSaved = async () => {
  showModal.value = false
  
  // Reload keys in settings
  await loadKeys()
  
  // Sync with chat store - reload available models
  await chatStore.loadAvailableModels()
  
  // Show success toast
  toast.success(
    'API Key Saved',
    'Your API key has been saved and models are now available.'
  )
}

onMounted(() => {
  loadKeys()
})
</script>
