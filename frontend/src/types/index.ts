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

export interface Model {
  id: string
  name: string
  description: string
  badges: string[]
  responseTime: string
  status: string
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
