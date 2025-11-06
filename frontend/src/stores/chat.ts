import { defineStore } from 'pinia'
import { ref, computed, reactive } from 'vue'
import type { Message, Model, ModelId, ModelParams } from '@/types'
import { marked } from 'marked'
import DOMPurify from 'dompurify'

interface Conversation {
  id: string
  title: string
  messages: Message[]
  lastUpdated: Date
}

export const useChatStore = defineStore('chat', () => {
  // State
  const conversations = ref<Conversation[]>([])
  const currentConversationId = ref<string | null>(null)
  const isTyping = ref(false)

  const selectedModel = ref<ModelId>('claude-3-opus')
  const availableModels = ref<Model[]>([
    {
      id: 'claude-3-opus',
      name: 'Claude 3 Opus',
      description: 'Most powerful model for complex tasks.',
      status: 'Online',
      responseTime: '~2s',
      badges: ['Pro', 'Vision']
    },
    {
      id: 'gpt-4-turbo',
      name: 'GPT-4 Turbo',
      description: 'High-end model with large context.',
      status: 'Online',
      responseTime: '~1.5s',
      badges: ['Vision']
    },
    {
      id: 'gemini-1.5-pro',
      name: 'Gemini 1.5 Pro',
      description: 'Large context window, multimodal.',
      status: 'Maintenance',
      responseTime: '~2.5s',
      badges: ['New']
    }
  ])

  const modelParams = reactive<ModelParams>({
    temperature: 0.7,
    maxTokens: 1024,
    topP: 1,
    streaming: true
  })

  // Getters
  const currentConversation = computed(() => {
    return conversations.value.find(c => c.id === currentConversationId.value)
  })

  const messages = computed(() => {
    return currentConversation.value?.messages || []
  })

  const recentConversations = computed(() => {
    return conversations.value.slice().sort((a, b) => b.lastUpdated.getTime() - a.lastUpdated.getTime())
  })

  const getModelDisplayName = (modelId: ModelId) => {
    return availableModels.value.find(m => m.id === modelId)?.name || 'Unknown Model'
  }

  const getModelBadges = (modelId: ModelId) => {
    return availableModels.value.find(m => m.id === modelId)?.badges || []
  }

  // Actions
  function addMessage(message: Message) {
    const conversation = currentConversation.value
    if (conversation) {
      conversation.messages.push(message)
      conversation.lastUpdated = new Date()
    }
  }

  async function sendMessage(content: string) {
    addMessage({
      id: Date.now().toString(),
      role: 'user',
      content,
      timestamp: new Date()
    })

    isTyping.value = true

    // Simulate API call
    await new Promise(resolve => setTimeout(resolve, 1500))

    const responseContent = `This is a simulated response from **${getModelDisplayName(selectedModel.value)}**. 
You said: *"${content}"*.

Here's a code block:
\`\`\`javascript
console.log("Hello, world!");
\`\`\`
`
    addMessage({
      id: Date.now().toString(),
      role: 'assistant',
      content: responseContent,
      timestamp: new Date(),
      model: selectedModel.value,
      responseTime: 1350,
      tokens: 128
    })

    isTyping.value = false
  }

  function updateMessage(updatedMessage: Message) {
    const conversation = currentConversation.value
    if (conversation) {
      const index = conversation.messages.findIndex(m => m.id === updatedMessage.id)
      if (index !== -1) {
        conversation.messages[index] = updatedMessage
        conversation.lastUpdated = new Date()
      }
    }
  }

  function createConversation(title: string) {
    const newConversation: Conversation = {
      id: Date.now().toString(),
      title,
      messages: [],
      lastUpdated: new Date(),
    }
    conversations.value.unshift(newConversation)
    currentConversationId.value = newConversation.id
  }

  function selectConversation(id: string) {
    currentConversationId.value = id
  }

  function clearChat() {
    const conversation = currentConversation.value
    if (conversation) {
      conversation.messages = []
      conversation.lastUpdated = new Date()
    }
  }

  function selectModel(modelId: ModelId) {
    const model = availableModels.value.find(m => m.id === modelId)
    if (model && model.status === 'Online') {
      selectedModel.value = modelId
    }
  }

  function updateModelParams(params: Partial<ModelParams>) {
    Object.assign(modelParams, params)
  }

  function formatMessageContent(content: string) {
    const rawHtml = marked(content)
    return DOMPurify.sanitize(rawHtml)
  }

  return {
    conversations,
    currentConversation,
    messages,
    addMessage,
    createConversation,
    selectConversation,
    recentConversations,
    isTyping,
    selectedModel,
    availableModels,
    modelParams,
    sendMessage,
    updateMessage,
    clearChat,
    selectModel,
    updateModelParams,
    getModelDisplayName,
    getModelBadges,
    formatMessageContent,
  }
})
