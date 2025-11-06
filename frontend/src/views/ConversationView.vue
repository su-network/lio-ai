<template>
  <div class="flex h-screen bg-gray-50 dark:bg-gray-900">
    <!-- Main Chat Area - No secondary sidebar -->
    <div class="flex-1 flex flex-col">
      <ChatHeader
        :selected-model="chatStore.selectedModel"
        :available-models="chatStore.availableModels"
        :model-params="chatStore.modelParams"
        @update:selected-model="chatStore.selectModel($event)"
        @update:model-params="chatStore.updateModelParams($event)"
        @clear-chat="chatStore.clearChat"
      />

      <div class="flex-1 overflow-y-auto px-2 sm:px-4 md:px-6 py-4 space-y-4" ref="messagesContainer">
        <!-- Empty state -->
        <div v-if="!chatStore.messages || chatStore.messages.length === 0" class="flex items-center justify-center h-full">
          <div class="text-center max-w-md px-4">
            <div class="w-16 h-16 bg-gray-200 dark:bg-gray-700 rounded-full flex items-center justify-center mx-auto mb-4">
              <MessageSquare class="w-8 h-8 text-gray-400 dark:text-gray-500" />
            </div>
            <h3 class="text-lg font-medium text-gray-900 dark:text-gray-100 mb-2">Start a conversation</h3>
            <p class="text-sm text-gray-500 dark:text-gray-400">Send a message to begin chatting with the AI assistant.</p>
          </div>
        </div>

        <!-- Messages -->
        <ChatMessage
          v-for="message in chatStore.messages"
          :key="message.id"
          :message="message"
          :selected-model="chatStore.selectedModel"
          @edit-message="editMessage"
          @copy-message="copyMessage"
        />

        <div v-if="chatStore.isTyping" class="flex justify-start">
          <div class="bg-white dark:bg-gray-800 px-3 sm:px-4 py-3 rounded-lg border border-gray-200 dark:border-gray-700">
            <div class="flex items-center space-x-2 sm:space-x-3">
              <Bot class="w-4 h-4 flex-shrink-0" />
              <div class="flex items-center space-x-2">
                <div class="w-2 h-2 rounded-full bg-gray-500 animate-bounce"></div>
                <div class="w-2 h-2 rounded-full bg-gray-500 animate-bounce" style="animation-delay: 0.1s"></div>
                <div class="w-2 h-2 rounded-full bg-gray-500 animate-bounce" style="animation-delay: 0.2s"></div>
              </div>
              <span class="text-sm text-gray-500 dark:text-gray-400 hidden sm:inline">{{ chatStore.getModelDisplayName(chatStore.selectedModel) }} is thinking...</span>
            </div>
          </div>
        </div>
      </div>

      <ChatInput
        :is-typing="chatStore.isTyping"
        :editing-message="editingMessage"
        :quick-prompts="[]"
        :placeholder="'Type your message...'"
        @send-message="sendMessage"
        @cancel-edit="cancelEdit"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useChatStore } from '@/stores/chat'
import type { Message } from '@/types'
import ChatHeader from '@/components/chat/ChatHeader.vue'
import ChatMessage from '@/components/chat/ChatMessage.vue'
import ChatInput from '@/components/chat/ChatInput.vue'
import { Bot, MessageSquare } from 'lucide-vue-next'

const route = useRoute()
const chatStore = useChatStore()

const messagesContainer = ref<HTMLElement | null>(null)
const editingMessage = ref<Message | null>(null)
const editContent = ref('')

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

watch(() => chatStore.messages, scrollToBottom, { deep: true })

const sendMessage = (content: string) => {
  if (editingMessage.value) {
    chatStore.updateMessage(editingMessage.value.id, content)
    editingMessage.value = null
    editContent.value = ''
  } else {
    chatStore.sendMessage(content)
  }
}

const editMessage = (message: Message) => {
  editingMessage.value = message
  editContent.value = message.content
}

const cancelEdit = () => {
  editingMessage.value = null
  editContent.value = ''
}

const copyMessage = (message: Message) => {
  navigator.clipboard.writeText(message.content)
}

onMounted(async () => {
  // Get chat UUID from route params
  const uuid = route.params.uuid as string
  console.log('ConversationView mounted with UUID:', uuid)
  
  if (uuid) {
    try {
      // First, ensure chats are loaded
      if (!chatStore.recentConversations || chatStore.recentConversations.length === 0) {
        console.log('Loading chats first...')
        await chatStore.fetchChats()
      }
      
      // Load by UUID
      console.log('Loading conversation by UUID:', uuid)
      await chatStore.selectConversation(uuid)
      
      console.log('Conversation loaded:', {
        currentConversationUUID: chatStore.currentConversationUUID,
        messagesCount: chatStore.messages?.length,
        messages: chatStore.messages
      })
    } catch (error) {
      console.error('Error loading conversation:', error)
    }
  }
  
  // Fetch models
  chatStore.loadAvailableModels()
})
</script>
