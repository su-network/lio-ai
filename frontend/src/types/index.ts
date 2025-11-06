export interface Message {
  id: string
  role: 'user' | 'assistant'
  content: string
  timestamp: Date
  model?: string
  responseTime?: number
  tokens?: number
}

export interface Model {
  id: string
  name: string
  description: string
  badges: string[]
  responseTime: string
  status: string
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
