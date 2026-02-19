 like <template>
  <div class="flex h-screen bg-gray-50 dark:bg-gray-900">
    <!-- Chat History Sidebar - Chat list only -->
    <div class="w-64 bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700 flex flex-col">
      <!-- Sidebar Header -->
      <div class="flex-none h-[57px] px-4 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
        <h2 class="font-semibold text-gray-900 dark:text-white">Chats</h2>
        <button
          @click="handleNewChat"
          class="p-2 text-blue-600 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded-lg transition-colors"
          title="New Chat"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
        </button>
      </div>

      <!-- Chat List with inline actions -->
      <div class="flex-1 overflow-y-auto p-2">
        <div v-if="chatStore.loadingChats" class="flex flex-col items-center justify-center py-8">
          <div class="relative w-10 h-10">
            <div class="absolute inset-0 rounded-full border-2 border-gray-200 dark:border-gray-700"></div>
            <div class="absolute inset-0 rounded-full border-2 border-transparent animate-spin border-t-blue-600 border-r-blue-600"></div>
          </div>
          <p class="mt-3 text-xs text-gray-500 dark:text-gray-400">Loading chats...</p>
        </div>
        
        <div v-else-if="chatStore.recentConversations.length === 0" class="text-center py-8 px-4">
          <p class="text-sm text-gray-500 dark:text-gray-400">No chats yet</p>
        </div>

        <div v-else class="space-y-1">
          <div
            v-for="chat in chatStore.recentConversations"
            :key="chat.id"
            :class="[
              'w-full rounded-lg transition-colors group',
              chatStore.currentConversationUUID === chat.chat_uuid
                ? 'bg-blue-50 dark:bg-blue-900/20'
                : 'hover:bg-gray-100 dark:hover:bg-gray-700'
            ]"
          >
            <div class="flex items-center justify-between p-3">
              <!-- Edit mode -->
              <div v-if="editingChatId === chat.id" class="flex items-center space-x-2 flex-1">
                <input
                  v-model="editingTitle"
                  @keydown.enter="saveEdit"
                  @keydown.esc="cancelChatEdit"
                  ref="editInput"
                  class="flex-1 px-2 py-1.5 text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-white rounded border border-blue-500 dark:border-blue-400 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="Chat title"
                />
                <button
                  @click="saveEdit"
                  class="p-1.5 text-green-600 dark:text-green-400 hover:bg-green-50 dark:hover:bg-green-900/20 rounded transition-colors"
                  title="Save"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                  </svg>
                </button>
                <button
                  @click="cancelChatEdit"
                  class="p-1.5 text-gray-600 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-600 rounded transition-colors"
                  title="Cancel"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>

              <!-- Normal mode -->
              <template v-else>
                <button
                  @click="selectChat(chat.chat_uuid!)"
                  class="flex-1 text-left min-w-0"
                >
                  <p :class="[
                    'text-sm font-medium truncate',
                    chatStore.currentConversationUUID === chat.chat_uuid
                      ? 'text-blue-700 dark:text-blue-300'
                      : 'text-gray-700 dark:text-gray-300'
                  ]">
                    {{ chat.title }}
                  </p>
                  <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
                    {{ formatDate(chat.updated_at) }}
                  </p>
                </button>

                <div class="flex items-center space-x-1 ml-2 opacity-0 group-hover:opacity-100 transition-opacity">
                  <button
                    @click.stop="startEdit(chat)"
                    class="p-1.5 text-gray-600 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-600 rounded transition-colors"
                    title="Edit"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                    </svg>
                  </button>
                  <button
                    @click.stop="confirmDelete(chat)"
                    class="p-1.5 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 rounded transition-colors"
                    title="Delete"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </button>
                </div>
              </template>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <div 
      v-if="chatToDelete" 
      class="fixed inset-0 bg-black/50 z-[60] flex items-center justify-center p-4"
      @click.self="chatToDelete = null"
    >
      <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-md w-full border border-gray-200 dark:border-gray-700">
        <div class="p-6">
          <div class="flex items-center mb-4">
            <div class="flex-shrink-0 w-10 h-10 rounded-full bg-red-100 dark:bg-red-900/30 flex items-center justify-center">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-red-600 dark:text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
            </div>
            <h3 class="ml-3 text-lg font-medium text-gray-900 dark:text-white">Delete Chat</h3>
          </div>
          <p class="text-sm text-gray-500 dark:text-gray-400 mb-1">
            Are you sure you want to delete this chat?
          </p>
          <p class="text-sm font-medium text-gray-900 dark:text-white mb-6">
            "{{ chatToDelete.title }}"
          </p>
          <div class="flex justify-end space-x-3">
            <button
              @click="chatToDelete = null"
              class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-600 transition-colors"
            >
              Cancel
            </button>
            <button
              @click="deleteChat(chatToDelete.id, chatToDelete.chat_uuid)"
              class="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-lg hover:bg-red-700 border border-red-700 transition-colors"
            >
              Delete
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Main Chat Area -->
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
        :quick-prompts="[]"
        :placeholder="'Type your message...'"
        :has-available-models="chatStore.availableModels.length > 0"
        :selected-model="chatStore.selectedModel"
        :available-models="chatStore.availableModels"
        @send-message="sendMessage"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick, onMounted } from 'vue'
import { useChatStore } from '@/stores/chat'
import type { Message } from '@/types'
import ChatHeader from '@/components/chat/ChatHeader.vue'
import ChatMessage from '@/components/chat/ChatMessage.vue'
import ChatInput from '@/components/chat/ChatInput.vue'
import { Bot, MessageSquare } from 'lucide-vue-next'

const chatStore = useChatStore()

const messagesContainer = ref<HTMLElement | null>(null)

// Modal state for delete confirmation
const chatToDelete = ref<any>(null)

// Inline edit state
const editingChatId = ref<number | null>(null)
const editingTitle = ref('')
const editInput = ref<HTMLInputElement | null>(null)

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

watch(() => chatStore.messages, scrollToBottom, { deep: true })

const sendMessage = (messageData: any) => {
  // Handle both string and object formats
  if (typeof messageData === 'string') {
    chatStore.sendMessage(messageData)
  } else {
    chatStore.sendMessage(messageData.content, messageData)
  }
}

const copyMessage = (message: Message) => {
  navigator.clipboard.writeText(message.content)
}

const selectChat = async (chatUUID: string) => {
  await chatStore.selectConversation(chatUUID)
}

const handleNewChat = () => {
  // Clear current conversation to start fresh
  // Title will be auto-generated from first message
  chatStore.currentConversationUUID = null
}

const confirmDelete = (chat: any) => {
  chatToDelete.value = chat
}

const deleteChat = async (chatId: number, chatUUID?: string) => {
  try {
    await chatStore.deleteConversation(chatId, chatUUID)
    chatToDelete.value = null
    
    // If deleted chat was selected, clear selection
    if (chatUUID && chatStore.currentConversationUUID === chatUUID) {
      chatStore.currentConversationUUID = null
    }
  } catch (error) {
    console.error('Error deleting chat:', error)
  }
}

const startEdit = (chat: any) => {
  editingChatId.value = chat.id
  editingTitle.value = chat.title
  nextTick(() => {
    if (editInput.value) {
      editInput.value.focus()
      editInput.value.select()
    }
  })
}

const saveEdit = async () => {
  if (!editingChatId.value || !editingTitle.value.trim()) {
    cancelChatEdit()
    return
  }

  try {
    await chatStore.renameConversation(editingChatId.value, editingTitle.value.trim())
    editingChatId.value = null
    editingTitle.value = ''
  } catch (error) {
    console.error('Error updating chat title:', error)
    alert('Failed to update chat title')
  }
}

const cancelChatEdit = () => {
  editingChatId.value = null
  editingTitle.value = ''
}

const formatDate = (dateString: string) => {
  const date = new Date(dateString)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))

  if (days === 0) {
    return date.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })
  } else if (days === 1) {
    return 'Yesterday'
  } else if (days < 7) {
    return `${days} days ago`
  } else {
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
  }
}

// Handle keyboard shortcuts
const handleKeyDown = (e: KeyboardEvent) => {
  if (e.key === 'Escape' && chatToDelete.value) {
    chatToDelete.value = null
  }
}

onMounted(() => {
  // Fetch chats and models without awaiting to not block UI
  chatStore.fetchChats()
  chatStore.loadAvailableModels()
  
  // Add keyboard event listener
  window.addEventListener('keydown', handleKeyDown)
})
</script>