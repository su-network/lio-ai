import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { apiService, type UsageQuota, type UsageSummary } from '@/services/api'
import { useUserStore } from './user'

export const useUsageStore = defineStore('usage', () => {
  const userStore = useUserStore()
  
  // State
  const quota = ref<UsageQuota | null>(null)
  const summary = ref<UsageSummary | null>(null)
  const history = ref<any[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  
  // Use user ID from user store
  const userId = computed(() => userStore.userId || 'anonymous')

  // Actions
  async function fetchQuotaStatus() {
    try {
      loading.value = true
      error.value = null
      quota.value = await apiService.getQuotaStatus()
    } catch (err: any) {
      // Silently handle if quota not found for new users
      if (err.response?.status === 404 || err.response?.status === 500) {
        console.warn('Quota not found for user (expected for new users)')
        quota.value = null
      } else {
        error.value = err.message || 'Failed to fetch quota status'
        console.error('Error fetching quota:', err)
      }
    } finally {
      loading.value = false
    }
  }

  async function fetchUsageSummary(period: 'daily' | 'monthly' | 'all_time' = 'monthly') {
    const apiPeriod = period === 'daily' ? 'day' : period === 'monthly' ? 'month' : 'month' as 'day' | 'week' | 'month'
    try {
      loading.value = true
      error.value = null
      summary.value = await apiService.getUsageSummary(apiPeriod)
    } catch (err: any) {
      // Silently handle if no usage data for new users
      if (err.response?.status === 404 || err.response?.status === 500) {
        console.warn('Usage summary not found for user (expected for new users)')
        summary.value = null
      } else {
        error.value = err.message || 'Failed to fetch usage summary'
        console.error('Error fetching usage summary:', err)
      }
    } finally {
      loading.value = false
    }
  }

  async function fetchUsageHistory(limit = 50, offset = 0) {
    try {
      loading.value = true
      error.value = null
      const response = await apiService.getUsageDashboard(limit, offset)
      history.value = response.data || []
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch usage history'
      console.error('Error fetching usage history:', err)
    } finally {
      loading.value = false
    }
  }

  async function checkQuota(tokensNeeded: number, modelName: string) {
    try {
      const response = await apiService.checkQuota(userId.value.toString(), tokensNeeded, modelName)
      return response.has_quota
    } catch (err: any) {
      error.value = err.message || 'Failed to check quota'
      console.error('Error checking quota:', err)
      return false
    }
  }

  async function trackUsage(data: {
    endpoint: string
    model_name: string
    tokens_input: number
    tokens_output: number
    tokens_total: number
    cost_usd: number
    duration_ms: number
    success: boolean
  }) {
    try {
      await apiService.trackUsage({
        user_id: userId.value.toString(),
        ...data
      })
    } catch (err: any) {
      console.error('Error tracking usage:', err)
    }
  }

  function clearError() {
    error.value = null
  }

  return {
    // State
    quota,
    summary,
    history,
    loading,
    error,
    userId,

    // Actions
    fetchQuotaStatus,
    fetchUsageSummary,
    fetchUsageHistory,
    checkQuota,
    trackUsage,
    clearError
  }
})
