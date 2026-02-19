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
      const response = await apiService.getUserChats(50, 0)
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
      const [allModels, statusResponse] = await Promise.all([
        apiService.getAllModels(),
        apiService.getModelsStatus()
      ])

      const statusById = new Map<string, any>()
      if (statusResponse?.models && Array.isArray(statusResponse.models)) {
        for (const s of statusResponse.models) {
          if (s?.model_id) statusById.set(s.model_id, s)
        }
      }

      const modelsWithStatus = (allModels || []).map((m: any) => {
        const status = statusById.get(m.id)
        const availabilityStatus = status?.status || 'unknown'
        const isActive = availabilityStatus === 'available'

        return {
          ...m,
          availabilityStatus,
          availabilityReason: status?.reason || null,
          isActive,
          // Legacy fields used by components
          badges: m?.capabilities?.special_features || [],
          responseTime: m?.metrics?.average_latency_ms ? `~${m.metrics.average_latency_ms}ms` : undefined,
          status: isActive ? 'Online' : 'Offline',
          context_length: m?.capabilities?.context_window || 4096
        }
      })

      // Sort: active first, then priority (lower first), then name
      modelsWithStatus.sort((a: any, b: any) => {
        const activeDelta = (b.isActive ? 1 : 0) - (a.isActive ? 1 : 0)
        if (activeDelta !== 0) return activeDelta
        const prA = typeof a.priority === 'number' ? a.priority : 999
        const prB = typeof b.priority === 'number' ? b.priority : 999
        if (prA !== prB) return prA - prB
        return String(a.name || a.id).localeCompare(String(b.name || b.id))
      })

      availableModels.value = modelsWithStatus

      // Keep selectedModel on an active model when possible
      const selected = availableModels.value.find(m => m.id === selectedModel.value)
      const firstActive = availableModels.value.find(m => (m as any).isActive)
      if (selected && !(selected as any).isActive && firstActive) {
        selectedModel.value = firstActive.id
        console.log('Auto-selected first active model:', selectedModel.value)
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
      const chat = await apiService.createChat(title)
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
    
    // Trigger reactivity by replacing the conversations array
    const convIndex = conversations.value.findIndex(c => c.chat_uuid === currentConversationUUID.value)
    if (convIndex !== -1) {
      conversations.value[convIndex]!.messages = [...conversations.value[convIndex]!.messages, tempUserMessage]
    }

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
      if (convIndex !== -1) {
        const userMsgIndex = conversations.value[convIndex]!.messages.findIndex(m => m.id === tempUserMessage.id)
        if (userMsgIndex !== -1) {
          const updatedMessages = [...conversations.value[convIndex]!.messages]
          updatedMessages[userMsgIndex] = savedUserMessage
          conversations.value[convIndex]!.messages = updatedMessages
        }
      }

      // Call the chat completion API
      let aiResponse = ''
      let usedModel = selectedModel.value
      
      try {
        const result = await apiService.chatCompletion(
          conversation.id,
          content,
          selectedModel.value
        )
        
        aiResponse = result.content
        usedModel = selectedModel.value
        
        // Add assistant message
        const assistantMessage: Message = {
          id: result.message_id,
          chat_id: conversation.id,
          role: 'assistant',
          content: aiResponse,
          model: usedModel,
          tokens: result.tokens,
          created_at: new Date().toISOString()
        }
        
        // Trigger reactivity by replacing the messages array
        const convIndexAssist = conversations.value.findIndex(c => c.chat_uuid === currentConversationUUID.value)
        if (convIndexAssist !== -1) {
          conversations.value[convIndexAssist]!.messages = [...conversations.value[convIndexAssist]!.messages, assistantMessage]
        }
        
      } catch (apiError: any) {
        console.error('Chat completion error:', apiError)
        
        // Get current model to check if it's Ollama
        const currentModel = availableModels.value.find(m => m.id === selectedModel.value)
        const isOllama = currentModel?.provider === 'ollama'
        
        // Check if it's a rate limit error (429)
        const isRateLimitError = apiError.message?.includes('429') || 
                                  apiError.message?.toLowerCase().includes('rate limit') ||
                                  apiError.message?.toLowerCase().includes('too many requests')
        
        // Check if it's a timeout error
        const isTimeout = apiError.code === 'ECONNABORTED' || apiError.message?.toLowerCase().includes('timeout')
        
        // Create appropriate error response message
        if (isOllama && isTimeout) {
          aiResponse = `I apologize, but the local Ollama model timed out. This can happen when:

• The model is loading for the first time (can take 30-60 seconds)
• Your computer is under heavy load
• Ollama service is not running

**Solutions:**
1. Make sure Ollama is running: check if you see it in your system tray
2. Try again - first requests are slower while the model loads into memory
3. Try a smaller model like Phi3 (3.8B) instead of Llama3 (8.0B)
4. Visit https://ollama.ai if you need to install Ollama

Local models are free but require your computer's resources.`
        } else if (isRateLimitError && !isOllama) {
          aiResponse = `I apologize, but the ${getModelDisplayName(selectedModel.value)} model is currently rate-limited by the provider.

This typically happens when:
• You're using a free/trial API key with limited requests per minute
• Too many requests were made in a short time period

**Solutions:**
1. Wait 30-60 seconds before sending another message
2. Upgrade to a paid API tier for higher rate limits
3. Try switching to a different model temporarily
4. Use a local Ollama model (free, no API keys needed)

The system will automatically retry with exponential backoff, but you may need to wait a bit before trying again.`
        } else if (isOllama) {
          aiResponse = `I apologize, but I encountered an error with the local Ollama model: ${apiError.message}

**Troubleshooting for Ollama:**
1. Make sure Ollama is installed and running on your computer
2. Check if the model is downloaded: run 'ollama list' in terminal
3. Download the model if needed: 'ollama pull ${currentModel?.model_name || selectedModel.value}'
4. Visit https://ollama.ai for installation help

Note: Ollama models are free and run locally - no API keys needed!`
        } else {
          aiResponse = `I apologize, but I encountered an error while processing your request: ${apiError.message}

Please make sure:
1. You have configured your API keys in Settings (not needed for Ollama models)
2. The selected model (${getModelDisplayName(selectedModel.value)}) is available
3. Your API key has sufficient credits

You can check available models and configure API keys in the Settings page.

**Tip:** Try using a local Ollama model for free unlimited usage!`
        }
        
        // Save error message as assistant response
        const errorMessage = await apiService.addMessage(
          conversation.id,
          'assistant',
          aiResponse,
          usedModel
        )
        
        // Trigger reactivity by replacing the messages array
        const convIndexError = conversations.value.findIndex(c => c.chat_uuid === currentConversationUUID.value)
        if (convIndexError !== -1) {
          conversations.value[convIndexError]!.messages = [...conversations.value[convIndexError]!.messages, errorMessage]
        }
      }

      // Update conversation updated_at
      const convIndexFinal = conversations.value.findIndex(c => c.chat_uuid === currentConversationUUID.value)
      if (convIndexFinal !== -1) {
        conversations.value[convIndexFinal]!.updated_at = new Date().toISOString()
      }
      
    } catch (err: any) {
      error.value = err.message || 'Failed to send message'
      console.error('Error sending message:', err)
      // Remove the temporary user message on error
      const convIndexCleanup = conversations.value.findIndex(c => c.chat_uuid === currentConversationUUID.value)
      if (convIndexCleanup !== -1) {
        const msgIndex = conversations.value[convIndexCleanup]!.messages.findIndex(m => m.id === tempUserMessage.id)
        if (msgIndex !== -1) {
          conversations.value[convIndexCleanup]!.messages = conversations.value[convIndexCleanup]!.messages.filter(m => m.id !== tempUserMessage.id)
        }
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
      const newChat = await apiService.createChat(conversation.title)
      
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
