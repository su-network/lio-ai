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

  const getModelDisplayName = (modelId: ModelId) => {
    return availableModels.value.find(m => m.id === modelId)?.name || 'Unknown Model'
  }

  const getModelBadges = (modelId: ModelId) => {
    return availableModels.value.find(m => m.id === modelId)?.badges || []
  }

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
      const models = await apiService.getAvailableModels()
      if (models && models.length > 0) {
        availableModels.value = models.map(m => ({
          id: m.name.toLowerCase().replace(/\s+/g, '-'),
          name: m.name,
          description: m.description,
          status: 'Online',
          responseTime: '~2s',
          badges: m.capabilities || []
        }))
        // Set first model as default if current selection not found
        if (!availableModels.value.find(m => m.id === selectedModel.value)) {
          selectedModel.value = availableModels.value[0]?.id || 'gpt-4-turbo'
        }
      }
    } catch (err: any) {
      console.error('Error loading models:', err)
      // Keep empty array if API fails
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

  async function sendMessage(content: string) {
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

      // TODO: Call AI model API endpoint here
      // For now, simulate AI response
      await new Promise(resolve => setTimeout(resolve, 1500))

      const aiResponse = `This is a response from **${getModelDisplayName(selectedModel.value)}**. 

You said: *"${content}"*

Here's an example code block:
\`\`\`javascript
console.log("Hello from the integrated API!");
\`\`\`

This message is now stored in the database via the Go API!`

      // Save assistant message to API
      const assistantMessage = await apiService.addMessage(
        conversation.id,
        'assistant',
        aiResponse,
        selectedModel.value
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
    if (model && model.status === 'Online') {
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
