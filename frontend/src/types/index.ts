export interface Message {
  id: number
  chat_id?: number
  role: 'user' | 'assistant' | 'system'
  content: string
  created_at: string
  model?: string
  responseTime?: number
  tokens?: number
  // Legacy support
  timestamp?: Date
}

export interface ModelCapabilities {
  languages: string[]
  frameworks: string[]
  max_complexity: 'simple' | 'intermediate' | 'advanced'
  special_features: string[]
  context_window: number
  output_limit: number
}

export interface ModelMetrics {
  average_latency_ms: number
  success_rate: number
  cost_per_request: number
  requests_per_minute: number
  uptime: number
}

export interface ModelConfig {
  temperature: number
  top_p?: number
  top_k?: number
  max_tokens?: number
  presence_penalty?: number
  frequency_penalty?: number
}

export interface Model {
  id: string
  name: string
  provider: 'openai' | 'anthropic' | 'google' | 'cohere' | 'mistral'
  model_name: string
  enabled: boolean
  priority: number
  capabilities: ModelCapabilities
  metrics: ModelMetrics
  config: ModelConfig
  // Legacy support
  description?: string
  badges?: string[]
  responseTime?: string
  status?: string
  context_length?: number

  // Availability/status support (from /api/v1/models/status)
  availabilityStatus?: string
  availabilityReason?: string | null
  isActive?: boolean
}

export type ModelId = string

export interface ModelParams {
  temperature: number
  maxTokens: number
  topP: number
  streaming: boolean
}

export interface Conversation {
  id: string
  title: string
  messagesCount: number
  lastUpdated: Date
}

export interface QuickPrompt {
  id: string
  label: string
  prompt: string
}
