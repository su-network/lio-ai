import { defineStore } from 'pinia'
import { ref, computed, reactive } from 'vue'
import type { Message, Model, ModelId, ModelParams } from '@/types'
import { apiService, type Chat } from '@/services/api'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { useUserStore } from './user'

interface ConversationWithMessages extends Chat {
  messages: Message[]
}

export const useChatStore = defineStore('chat', () => {
  const userStore = useUserStore()
  
  // State
  const conversations = ref<ConversationWithMessages[]>([])
  const currentConversationUUID = ref<string | null>(null)
  const isTyping = ref(false)
  const loadingChats = ref(false)
  const loadingMessages = ref(false)
  const sendingMessage = ref(false)
  const error = ref<string | null>(null)
  
  // Use user ID from user store
  const userId = computed(() => userStore.userId || 'anonymous')

  const selectedModel = ref<ModelId>('gpt-4-turbo')
  const availableModels = ref<Model[]>([])

  const modelParams = reactive<ModelParams>({
    temperature: 0.7,
    maxTokens: 1024,
    topP: 1,
    streaming: true
  })

  // Getters
  const currentConversation = computed(() => {
    return conversations.value.find(c => c.chat_uuid === currentConversationUUID.value)
  })

  const messages = computed(() => {
    return currentConversation.value?.messages || []
  })

  const recentConversations = computed(() => {
    return conversations.value.slice().sort((a, b) => 
      new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime()
    )
  })

  // Actions
  async function fetchChats() {
    loadingChats.value = true
    error.value = null
    try {
      const response = await apiService.getUserChats(userId.value, 50, 0)
      conversations.value = response.data.map(chat => ({
        ...chat,
        messages: []
      }))
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch chats'
      console.error('Error fetching chats:', err)
    } finally {
      loadingChats.value = false
    }
  }

  async function loadAvailableModels() {
    try {
      // Get models with status to filter only available ones
      const statusResponse = await apiService.getModelsStatus()
      
      if (statusResponse && statusResponse.models) {
        // Filter only models that are available (have API keys)
        const availableModelsList = statusResponse.models
          .filter((m: any) => m.status === 'available')
          .map((m: any) => {
            const modelConfig = m
            return {
              id: m.model_id,
              name: modelConfig.model_name || m.model_id.replace(/-/g, ' ').replace(/\b\w/g, (l: string) => l.toUpperCase()),
              description: `${m.provider} model`,
              status: 'Online',
              responseTime: `~${modelConfig.metrics?.average_latency_ms || 2000}ms`,
              badges: modelConfig.capabilities?.special_features || [],
              provider: m.provider,
              context_length: modelConfig.capabilities?.context_window || 4096
            }
          })
        
        if (availableModelsList.length > 0) {
          availableModels.value = availableModelsList
          // Set first available model as default if current selection not available
          const currentModelAvailable = availableModels.value.find(m => m.id === selectedModel.value)
          if (!currentModelAvailable && availableModels.value[0]) {
            selectedModel.value = availableModels.value[0].id
            console.log('Auto-selected first available model:', selectedModel.value)
          }
        } else {
          console.warn('No models with API keys available. Please configure API keys in Settings.')
          availableModels.value = []
        }
      }
    } catch (err: any) {
      console.error('Error loading available models:', err)
      // Keep empty array if API fails
      availableModels.value = []
    }
  }

  async function fetchMessages(chatId: number) {
    loadingMessages.value = true
    error.value = null
    try {
      const response = await apiService.getMessages(chatId, 100, 0)
      const conversation = conversations.value.find(c => c.id === chatId)
      if (conversation) {
        conversation.messages = response.data
      }
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch messages'
      console.error('Error fetching messages:', err)
    } finally {
      loadingMessages.value = false
    }
  }

  async function createConversation(title: string) {
    error.value = null
    try {
      const chat = await apiService.createChat(userId.value, title)
      const newConversation: ConversationWithMessages = {
        ...chat,
        messages: []
      }
      conversations.value.unshift(newConversation)
      currentConversationUUID.value = chat.chat_uuid || null
      return chat
    } catch (err: any) {
      error.value = err.message || 'Failed to create chat'
      console.error('Error creating chat:', err)
      return null
    }
  }

  async function selectConversation(chatUUID: string) {
    console.log('selectConversation called with UUID:', chatUUID)
    currentConversationUUID.value = chatUUID
    const conversation = conversations.value.find(c => c.chat_uuid === chatUUID)
    console.log('Found conversation:', conversation)
    if (conversation && conversation.messages.length === 0 && conversation.id) {
      console.log('Fetching messages for conversation:', conversation.id)
      await fetchMessages(conversation.id)
    } else if (conversation) {
      console.log('Conversation already has messages:', conversation.messages.length)
    } else {
      // If not found in list, try fetching by UUID
      await selectConversationByUUID(chatUUID)
    }
  }

  async function selectConversationByUUID(uuid: string) {
    console.log('selectConversationByUUID called with UUID:', uuid)
    loadingMessages.value = true
    error.value = null
    try {
      // Fetch chat by UUID
      const response = await apiService.getChatByUUID(uuid)
      
      // Response includes both chat fields and messages in flat structure
      const { messages: responseMessages, ...chatData } = response

      // Update or add conversation to the list
      const existingIndex = conversations.value.findIndex(c => c.chat_uuid === uuid)
      const conversationWithMessages = {
        ...chatData,
        messages: responseMessages || []
      }

      if (existingIndex !== -1) {
        conversations.value[existingIndex] = conversationWithMessages
      } else {
        conversations.value.push(conversationWithMessages)
      }

      currentConversationUUID.value = uuid
      console.log('Conversation loaded by UUID:', chatData)
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch conversation by UUID'
      console.error('Error fetching conversation by UUID:', err)
    } finally {
      loadingMessages.value = false
    }
  }

  // Helper function to get model display name
  function getModelDisplayName(modelId: string): string {
    const model = availableModels.value.find(m => m.id === modelId)
    return model?.name || modelId
  }

  // Helper function to get model badges
  function getModelBadges(modelId: string): string[] {
    const model = availableModels.value.find(m => m.id === modelId)
    return model?.badges || []
  }

  async function sendMessage(content: string, metadata?: any) {
    if (!availableModels.value || availableModels.value.length === 0) {
      error.value = 'No models available. Please configure API keys in Settings.'
      return
    }

    if (!currentConversationUUID.value) {
      // Generate title from first message (first 50 chars)
      const title = content.length > 50 ? content.substring(0, 47) + '...' : content
      const newChat = await createConversation(title)
      if (!newChat) return
    }

    const conversation = currentConversation.value
    if (!conversation) return

    // Add user message locally first
    const tempUserMessage: Message = {
      id: Date.now(),
      chat_id: conversation.id,
      role: 'user',
      content,
      created_at: new Date().toISOString()
    }
    conversation.messages.push(tempUserMessage)

    isTyping.value = true
    error.value = null

    try {
      // Save user message to API
      const savedUserMessage = await apiService.addMessage(
        conversation.id,
        'user',
        content
      )
      
      // Replace temp message with saved one
      const userMsgIndex = conversation.messages.findIndex(m => m.id === tempUserMessage.id)
      if (userMsgIndex !== -1) {
        conversation.messages[userMsgIndex] = savedUserMessage
      }

      let aiResponse = ''
      let usedModel = selectedModel.value
      
      try {
        // Call the AI generation API through the Go gateway
        const response = await fetch('/api/generate', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            prompt: content,
            language: 'python',
            complexity: 'simple',
            selected_models: [selectedModel.value],
            max_models: 1,
            timeout: 30,
            user_id: userId.value
          })
        })

        if (!response.ok) {
          const errorData = await response.json().catch(() => ({ error: 'Unknown error' }))
          throw new Error(errorData.error || `API Error: ${response.status}`)
        }

        const result = await response.json()
        
        if (result.status === 'success' && result.model_responses && result.model_responses.length > 0) {
          const modelResponse = result.model_responses[0]
          aiResponse = modelResponse.generated_code || modelResponse.response || 'No response generated'
          usedModel = modelResponse.model_id || selectedModel.value
          
          // Format the response nicely
          if (modelResponse.generated_code) {
            aiResponse = `Here's the generated code:\n\n\`\`\`python\n${modelResponse.generated_code}\n\`\`\`\n\n*Generated by ${getModelDisplayName(usedModel)}*`
          }
        } else if (result.consensus_code) {
          aiResponse = `\`\`\`python\n${result.consensus_code}\n\`\`\`\n\n*Generated by ${getModelDisplayName(result.best_model || selectedModel.value)}*`
          usedModel = result.best_model || selectedModel.value
        } else {
          throw new Error('No valid response from AI service')
        }
      } catch (apiError: any) {
        console.error('AI generation error:', apiError)
        // Fallback response
        aiResponse = `I apologize, but I encountered an error while processing your request: ${apiError.message}
        
Please make sure:
1. You have configured your API keys in Settings
2. The selected model (${getModelDisplayName(selectedModel.value)}) is available
3. Your API key has sufficient credits

You can check available models and configure API keys in the Settings page.`
      }

      // Save assistant message to API
      const assistantMessage = await apiService.addMessage(
        conversation.id,
        'assistant',
        aiResponse,
        usedModel
      )
      
      conversation.messages.push(assistantMessage)

      // Update conversation updated_at
      conversation.updated_at = new Date().toISOString()
      
    } catch (err: any) {
      error.value = err.message || 'Failed to send message'
      console.error('Error sending message:', err)
      // Remove the temporary user message on error
      const msgIndex = conversation.messages.findIndex(m => m.id === tempUserMessage.id)
      if (msgIndex !== -1) {
        conversation.messages.splice(msgIndex, 1)
      }
    } finally {
      isTyping.value = false
    }
  }

  async function updateMessage(messageId: number, newContent: string) {
    // TODO: Implement message update API endpoint
    const conversation = currentConversation.value
    if (conversation) {
      const message = conversation.messages.find(m => m.id === messageId)
      if (message) {
        message.content = newContent
      }
    }
  }

  async function clearChat() {
    const conversation = currentConversation.value
    if (!conversation) return

    try {
      // Delete and recreate the chat
      await apiService.deleteChat(conversation.id)
      const newChat = await apiService.createChat(userId.value, conversation.title)
      
      // Update local state
      const index = conversations.value.findIndex(c => c.id === conversation.id)
      if (index !== -1) {
        conversations.value[index] = {
          ...newChat,
          messages: []
        }
        currentConversationUUID.value = newChat.chat_uuid || null
      }
    } catch (err: any) {
      error.value = err.message || 'Failed to clear chat'
      console.error('Error clearing chat:', err)
    }
  }

  async function deleteConversation(id: number, uuid?: string) {
    try {
      await apiService.deleteChat(id)
      conversations.value = conversations.value.filter(c => c.id !== id)
      if (uuid && currentConversationUUID.value === uuid) {
        currentConversationUUID.value = null
      }
    } catch (err: any) {
      error.value = err.message || 'Failed to delete chat'
      console.error('Error deleting chat:', err)
    }
  }

  async function renameConversation(id: number, newTitle: string) {
    try {
      const updated = await apiService.updateChat(id, newTitle)
      const conversation = conversations.value.find(c => c.id === id)
      if (conversation) {
        conversation.title = updated.title
        conversation.updated_at = updated.updated_at
      }
    } catch (err: any) {
      error.value = err.message || 'Failed to rename chat'
      console.error('Error renaming chat:', err)
    }
  }

  function selectModel(modelId: ModelId) {
    const model = availableModels.value.find(m => m.id === modelId)
    if (model) {
      selectedModel.value = modelId
    }
  }

  function updateModelParams(params: Partial<ModelParams>) {
    Object.assign(modelParams, params)
  }

  function formatMessageContent(content: string) {
    const rawHtml = marked.parse(content)
    return DOMPurify.sanitize(rawHtml as string)
  }

  return {
    // State
    conversations,
    currentConversation,
    currentConversationUUID,
    messages,
    isTyping,
    loadingChats,
    loadingMessages,
    sendingMessage,
    error,
    selectedModel,
    availableModels,
    modelParams,
    recentConversations,
    
    // Actions
    fetchChats,
    fetchMessages,
    loadAvailableModels,
    createConversation,
    selectConversation,
    selectConversationByUUID,
    sendMessage,
    updateMessage,
    clearChat,
    deleteConversation,
    renameConversation,
    selectModel,
    updateModelParams,
    getModelDisplayName,
    getModelBadges,
    formatMessageContent,
  }
})
