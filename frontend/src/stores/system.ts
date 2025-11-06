import { defineStore } from 'pinia'
import { ref } from 'vue'
import { apiService, type SystemMetrics, type Model } from '@/services/api'

export const useSystemStore = defineStore('system', () => {
  // State
  const metrics = ref<SystemMetrics | null>(null)
  const models = ref<Model[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Actions
  async function fetchMetrics() {
    try {
      loading.value = true
      error.value = null
      metrics.value = await apiService.getMetrics()
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch metrics'
      console.error('Error fetching metrics:', err)
    } finally {
      loading.value = false
    }
  }

  async function fetchAvailableModels() {
    try {
      loading.value = true
      error.value = null
      models.value = await apiService.getAvailableModels()
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch available models'
      console.error('Error fetching models:', err)
    } finally {
      loading.value = false
    }
  }

  function clearError() {
    error.value = null
  }

  return {
    // State
    metrics,
    models,
    loading,
    error,

    // Actions
    fetchMetrics,
    fetchAvailableModels,
    clearError
  }
})
