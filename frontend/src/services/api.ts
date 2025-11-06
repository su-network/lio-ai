import axios from 'axios'
import type { AxiosInstance } from 'axios'
import { deobfuscate, isObfuscated } from '@/utils/encryption'

// API Configuration
// Use empty string to make requests relative (go through Vite proxy)
// In production, set VITE_API_URL to the actual API URL
const API_URL = import.meta.env.VITE_API_URL || ''

// Enable response obfuscation (set to false to disable)
const ENABLE_OBFUSCATION = import.meta.env.VITE_ENABLE_OBFUSCATION !== 'false'

// Create axios instance
const apiClient: AxiosInstance = axios.create({
  baseURL: API_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
    'X-Obfuscated': ENABLE_OBFUSCATION ? 'true' : 'false'
  }
})

// Request interceptor - add auth token
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('auth_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor - handle errors and deobfuscation
apiClient.interceptors.response.use(
  (response) => {
    // Deobfuscate response if enabled and data is obfuscated
    if (ENABLE_OBFUSCATION && response.data && isObfuscated(response.data)) {
      const deobfuscatedData = deobfuscate(response.data)
      if (deobfuscatedData) {
        response.data = deobfuscatedData
      }
    }
    return response
  },
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('auth_token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// Type definitions
export interface Document {
  id: number
  title: string
  content: string
  created_at: string
  updated_at: string
}

export interface Chat {
  id: number
  user_id: string
  title: string
  chat_uuid?: string
  created_at: string
  updated_at: string
  message_count?: number
}

export interface Message {
  id: number
  chat_id: number
  role: 'user' | 'assistant' | 'system'
  content: string
  created_at: string
  tokens?: number
  model?: string
}

export interface UsageQuota {
  user_id: string
  tier: string
  monthly_limit: number
  current_usage: number
  remaining: number
  percentage_used: number
  reset_date: string
}

export interface UsageSummary {
  total_requests: number
  successful_requests: number
  failed_requests: number
  total_tokens: number
  total_cost: number
  period: string
  by_model: Record<string, {
    requests: number
    tokens: number
    cost: number
  }>
  by_endpoint: Record<string, {
    requests: number
    avg_latency: number
  }>
}

export interface SystemMetrics {
  total_users: number
  active_users: number
  total_chats: number
  total_documents: number
  total_requests: number
  successful_requests: number
  failed_requests: number
  total_tokens: number
  total_cost: number
  avg_latency: number
  endpoints: Array<{
    endpoint: string
    request_count: number
    avg_time: number
    error_rate: number
  }>
}

export interface Model {
  name: string
  provider: string
  description: string
  context_length: number
  pricing: {
    input: number
    output: number
  }
  capabilities: string[]
}

export interface CodeGenerationRequest {
  documentation: string
  language: string
  style_guide?: string
  context?: string
}

export interface CodeGenerationResponse {
  code: string
  language: string
  confidence: number
  warnings?: string[]
  explanation?: string
}

export interface HallucinationCheckRequest {
  code: string
  documentation: string
}

export interface HallucinationCheckResponse {
  has_hallucination: boolean
  confidence: number
  warnings?: string[]
}

export interface RAGQueryRequest {
  query: string
  top_k?: number
}

export interface RAGQueryResponse {
  results: Array<{
    content: string
    metadata?: Record<string, any>
    score?: number
  }>
}

export interface StatsResponse {
  total_requests: number
  avg_confidence: number
  models_loaded: number
  rag_documents_count: number
}

// API Service
export const apiService = {
  // Health check
  getHealth: async () => {
    const response = await apiClient.get('/health')
    return response.data
  },

  // Document endpoints
  createDocument: async (title: string, content: string): Promise<Document> => {
    const response = await apiClient.post('/api/v1/documents', { title, content })
    return response.data
  },

  getDocuments: async (): Promise<Document[]> => {
    const response = await apiClient.get('/api/v1/documents')
    return response.data
  },

  getDocument: async (id: number): Promise<Document> => {
    const response = await apiClient.get(`/api/v1/documents/${id}`)
    return response.data
  },

  updateDocument: async (id: number, title?: string, content?: string): Promise<Document> => {
    const response = await apiClient.put(`/api/v1/documents/${id}`, { title, content })
    return response.data
  },

  deleteDocument: async (id: number): Promise<void> => {
    await apiClient.delete(`/api/v1/documents/${id}`)
  },

  // Chat endpoints
  createChat: async (userId: string, title: string): Promise<Chat> => {
    const response = await apiClient.post('/api/v1/chats', { user_id: userId, title })
    return response.data
  },

  getUserChats: async (userId: string, limit = 20, offset = 0): Promise<{ data: Chat[], total: number }> => {
    const response = await apiClient.get('/api/v1/chats', {
      params: { user_id: userId, limit, offset }
    })
    return response.data
  },

  getChat: async (id: number): Promise<Chat> => {
    const response = await apiClient.get(`/api/v1/chats/${id}`)
    return response.data
  },

  getChatByUUID: async (uuid: string): Promise<Chat & { chat_uuid: string, messages: Message[] }> => {
    const response = await apiClient.get(`/api/v1/chats/uuid/${uuid}`)
    return response.data
  },

  updateChat: async (id: number, title: string): Promise<Chat> => {
    const response = await apiClient.put(`/api/v1/chats/${id}`, { title })
    return response.data
  },

  deleteChat: async (id: number): Promise<void> => {
    await apiClient.delete(`/api/v1/chats/${id}`)
  },

  addMessage: async (chatId: number, role: string, content: string, model?: string): Promise<Message> => {
    const response = await apiClient.post(`/api/v1/chats/${chatId}/messages`, { 
      role, content, model 
    })
    return response.data
  },

  getMessages: async (chatId: number, limit = 50, offset = 0): Promise<{ data: Message[], total: number }> => {
    const response = await apiClient.get(`/api/v1/chats/${chatId}/messages`, {
      params: { limit, offset }
    })
    return response.data
  },

  // Usage endpoints
  getQuotaStatus: async (userId: string): Promise<UsageQuota> => {
    const response = await apiClient.get('/api/v1/usage/quota', {
      params: { user_id: userId }
    })
    return response.data
  },

  getUsageSummary: async (userId: string, period: 'daily' | 'monthly' | 'all_time' = 'monthly'): Promise<UsageSummary> => {
    const response = await apiClient.get('/api/v1/usage/summary', {
      params: { user_id: userId, period }
    })
    return response.data
  },

  trackUsage: async (data: {
    user_id: string
    endpoint: string
    model_name: string
    tokens_input: number
    tokens_output: number
    tokens_total: number
    cost_usd: number
    duration_ms: number
    success: boolean
  }): Promise<void> => {
    await apiClient.post('/api/v1/usage/track', data)
  },

  checkQuota: async (userId: string, tokensNeeded: number, modelName: string): Promise<{ has_quota: boolean, remaining: number }> => {
    const response = await apiClient.post('/api/v1/usage/check-quota', {
      user_id: userId,
      tokens_needed: tokensNeeded,
      model_name: modelName
    })
    return response.data
  },

  getUsageHistory: async (userId: string, limit = 50, offset = 0): Promise<any> => {
    const response = await apiClient.get('/api/v1/usage/dashboard', {
      params: { user_id: userId, limit, offset }
    })
    return response.data
  },

  // System endpoints
  getMetrics: async (): Promise<SystemMetrics> => {
    const response = await apiClient.get('/api/v1/system/metrics')
    return response.data
  },

  getAvailableModels: async (): Promise<Model[]> => {
    try {
      const response = await apiClient.get('/api/v1/system/info')
      return response.data.data?.models || []
    } catch (error) {
      console.warn('Failed to load models, using empty array:', error)
      return []
    }
  },

  getSystemStats: async (): Promise<any> => {
    const response = await apiClient.get('/api/v1/system/stats')
    return response.data
  },

  // Code generation endpoints
  generateCode: async (request: CodeGenerationRequest): Promise<CodeGenerationResponse> => {
    const response = await apiClient.post('/api/v1/codegen/generate', request)
    return response.data
  },

  validateCode: async (request: HallucinationCheckRequest): Promise<HallucinationCheckResponse> => {
    const response = await apiClient.post('/api/v1/codegen/validate', request)
    return response.data
  },

  searchRAG: async (request: RAGQueryRequest): Promise<RAGQueryResponse> => {
    const response = await apiClient.post('/api/v1/codegen/rag/search', request)
    return response.data
  },

  getStats: async (): Promise<StatsResponse> => {
    const response = await apiClient.get('/api/v1/codegen/stats')
    return response.data
  }
}

export default apiService
