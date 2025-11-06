import { defineStore } from 'pinia'
import { ref } from 'vue'
import { apiService, type CodeGenerationRequest, type StatsResponse } from '@/services/api'

export const useCodegenStore = defineStore('codegen', () => {
  // State
  const generatedCode = ref<string>('')
  const codeLanguage = ref<string>('')
  const confidence = ref<number>(0)
  const warnings = ref<string[]>([])
  const explanation = ref<string>('')
  const loading = ref(false)
  const error = ref<string | null>(null)
  const stats = ref<StatsResponse | null>(null)

  // Actions
  async function generateCode(request: CodeGenerationRequest) {
    try {
      loading.value = true
      error.value = null
      const response = await apiService.generateCode(request)
      
      generatedCode.value = response.code
      codeLanguage.value = response.language
      confidence.value = response.confidence
      warnings.value = response.warnings || []
      explanation.value = response.explanation || ''
    } catch (err: any) {
      error.value = err.message || 'Failed to generate code'
      console.error('Error generating code:', err)
    } finally {
      loading.value = false
    }
  }

  async function validateCode(code: string, documentation: string) {
    try {
      loading.value = true
      error.value = null
      const response = await apiService.validateCode({ code, documentation })
      return response
    } catch (err: any) {
      error.value = err.message || 'Failed to validate code'
      console.error('Error validating code:', err)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function searchDocuments(query: string, topK = 5) {
    try {
      loading.value = true
      error.value = null
      const response = await apiService.searchRAG({ query, top_k: topK })
      return response.results
    } catch (err: any) {
      error.value = err.message || 'Failed to search documents'
      console.error('Error searching documents:', err)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function fetchStats() {
    try {
      loading.value = true
      error.value = null
      stats.value = await apiService.getStats()
    } catch (err: any) {
      // Silently fail if AI service is not available (502 error)
      if (err.response?.status === 502) {
        console.warn('AI service unavailable (expected if not running)')
        stats.value = null
      } else {
        error.value = err.message || 'Failed to fetch stats'
        console.error('Error fetching stats:', err)
      }
    } finally {
      loading.value = false
    }
  }

  function clearCode() {
    generatedCode.value = ''
    codeLanguage.value = ''
    confidence.value = 0
    warnings.value = []
    explanation.value = ''
  }

  function clearError() {
    error.value = null
  }

  return {
    // State
    generatedCode,
    codeLanguage,
    confidence,
    warnings,
    explanation,
    loading,
    error,
    stats,

    // Actions
    generateCode,
    validateCode,
    searchDocuments,
    fetchStats,
    clearCode,
    clearError
  }
})
