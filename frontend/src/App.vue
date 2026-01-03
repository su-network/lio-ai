<template>
  <div id="app" class="min-h-screen bg-gray-50 dark:bg-gray-900 flex">
    <!-- Mobile Menu Button -->
    <button
      v-if="userStore.isLoggedIn && isSidebarCollapsed"
      @click="isSidebarCollapsed = false"
      class="fixed top-4 left-4 z-40 p-2 bg-white dark:bg-gray-800 rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 md:hidden"
      aria-label="Open menu"
    >
      <Menu class="w-5 h-5 text-gray-600 dark:text-gray-400" />
    </button>

    <!-- Sidebar -->
    <aside v-if="userStore.isLoggedIn" :class="[
      'bg-white dark:bg-gray-800 shadow-lg border-r border-gray-200 dark:border-gray-700 transition-all duration-300 ease-in-out flex flex-col',
      isSidebarCollapsed ? 'w-16' : 'w-64',
      'fixed md:relative inset-y-0 left-0 z-50 md:z-auto',
      isSidebarCollapsed ? '-translate-x-full md:translate-x-0' : 'translate-x-0'
    ]">
      <!-- Logo and Toggle button at the very top -->
      <div class="p-4 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
        <!-- Logo -->
        <router-link
          to="/"
          v-if="!isSidebarCollapsed"
          class="flex items-center space-x-3 transition-all duration-300"
        >
          <div class="w-9 h-9 bg-gradient-to-br from-slate-600 to-slate-700 rounded-xl flex items-center justify-center shadow-lg flex-shrink-0">
            <span class="text-white font-bold text-base">L</span>
          </div>
          <span class="text-xl font-bold text-gray-900 dark:text-white tracking-tight whitespace-nowrap">Lio AI</span>
        </router-link>
        
        <!-- Logo when collapsed - clicking opens sidebar -->
        <button 
          v-if="isSidebarCollapsed"
          @click="toggleSidebar"
          class="w-9 h-9 bg-gradient-to-br from-slate-600 to-slate-700 rounded-xl flex items-center justify-center shadow-lg flex-shrink-0 mx-auto hover:scale-105 transition-transform"
          aria-label="Open sidebar"
        >
          <span class="text-white font-bold text-base">L</span>
        </button>

        <!-- Toggle button - only visible when expanded -->
        <button
          v-if="!isSidebarCollapsed"
          @click="toggleSidebar"
          class="p-2 rounded-lg transition-colors flex-shrink-0"
          aria-label="Toggle sidebar"
        >
          <ChevronLeft class="w-5 h-5 text-gray-600 dark:text-gray-400" />
        </button>
      </div>

      <div class="p-4 flex-1 flex flex-col overflow-hidden">
        <!-- New Chat Button -->
        <button
          @click="createNewChat"
          :class="[
            'flex items-center px-4 py-3 mb-4 rounded-xl font-medium text-sm transition-all duration-200 text-gray-700 dark:text-gray-300 border border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700/50',
            isSidebarCollapsed ? 'justify-center px-2' : 'space-x-3'
          ]"
        >
          <PenSquare class="w-5 h-5 flex-shrink-0" />
          <span
            :class="[
              'transition-all duration-300 whitespace-nowrap',
              isSidebarCollapsed ? 'opacity-0 w-0 overflow-hidden' : 'opacity-100'
            ]"
          >
            New Chat
          </span>
        </button>

        <!-- Navigation Menu -->
        <div class="space-y-1 mb-4">
          <!-- Chats with View All button -->
          <div class="flex items-center gap-1">
            <router-link
              to="/chat"
              :class="[
                'flex-1 flex items-center px-4 py-3 rounded-xl font-medium text-sm transition-all duration-200',
                isSidebarCollapsed ? 'justify-center px-2' : 'space-x-3',
                $route.path === '/chat' 
                  ? 'bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-300' 
                  : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
              ]"
            >
              <MessageSquare class="w-5 h-5 flex-shrink-0" />
              <span
                :class="[
                  'transition-all duration-300 whitespace-nowrap',
                  isSidebarCollapsed ? 'opacity-0 w-0 overflow-hidden' : 'opacity-100'
                ]"
              >
                Chats
              </span>
            </router-link>
            <button
              v-if="!isSidebarCollapsed"
              @click="showAllChatsModal = true"
              class="p-2 rounded-lg text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
              title="View All Chats"
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
              </svg>
            </button>
          </div>

          <!-- Models -->
          <router-link
            to="/models"
            :class="[
              'flex items-center px-4 py-3 rounded-xl font-medium text-sm transition-all duration-200',
              isSidebarCollapsed ? 'justify-center px-2' : 'space-x-3',
              $route.path === '/models' 
                ? 'bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-300' 
                : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
            ]"
          >
            <Layers class="w-5 h-5 flex-shrink-0" />
            <span
              :class="[
                'transition-all duration-300 whitespace-nowrap',
                isSidebarCollapsed ? 'opacity-0 w-0 overflow-hidden' : 'opacity-100'
              ]"
            >
              Models
            </span>
          </router-link>
        </div>

        <!-- Recent Chats List -->
        <div v-if="!isSidebarCollapsed" class="flex-1 overflow-y-auto">
          <div class="text-xs font-semibold text-gray-500 dark:text-gray-400 mb-2 px-2">RECENT</div>
          <ul class="space-y-1">
            <li v-for="convo in chatStore && Array.isArray(chatStore.recentConversations) ? chatStore.recentConversations.slice(0, 5) : []" :key="convo.id">
              <button
                @click="openRecentChat(convo.chat_uuid!)"
                class="w-full flex items-center px-3 py-2 rounded-lg text-sm transition-all duration-200 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 text-left"
                :class="{ 'bg-gray-100 dark:bg-gray-700': chatStore.currentConversationUUID === convo.chat_uuid }"
              >
                <MessageSquare class="w-4 h-4 flex-shrink-0 mr-2 text-gray-400" />
                <span class="truncate flex-1">{{ convo.title }}</span>
              </button>
            </li>
            <li v-if="(!chatStore || !Array.isArray(chatStore.recentConversations) || chatStore.recentConversations.length === 0)">
              <div class="text-gray-400 text-xs px-3 py-2">No chats yet</div>
            </li>
          </ul>
        </div>
       </div>
      
      <!-- User Profile and Settings -->
      <div class="p-4 border-t border-gray-200 dark:border-gray-700">
        <nav class="space-y-1">
          <router-link
            to="/settings"
            :class="[
              'flex items-center px-4 py-3 rounded-xl font-medium text-sm transition-all duration-200',
              isSidebarCollapsed ? 'justify-center px-2' : 'space-x-3',
              $route.path === '/settings'
                ? 'bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-300'
                : 'text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
            ]"
          >
            <Settings class="w-5 h-5 flex-shrink-0" />
            <span
              :class="[
                'transition-all duration-300 whitespace-nowrap',
                isSidebarCollapsed ? 'opacity-0 w-0 overflow-hidden' : 'opacity-100'
              ]"
            >
              Settings
            </span>
          </router-link>
        </nav>
        
        <router-link
          to="/profile"
          :class="[
            'flex items-center mt-4 rounded-xl transition-all duration-200',
            isSidebarCollapsed ? 'justify-center p-2' : 'px-4 py-2',
            $route.path === '/profile'
              ? 'bg-blue-50 dark:bg-blue-900/20'
              : 'hover:bg-gray-50 dark:hover:bg-gray-700/50'
          ]"
        >
          <img 
            :src="userStore.user?.avatar" 
            alt="User Avatar" 
            class="w-10 h-10 rounded-full ring-2 ring-gray-200 dark:ring-gray-700 flex-shrink-0"
          />
          <span 
            v-if="!isSidebarCollapsed" 
            class="ml-3 text-sm font-medium text-gray-900 dark:text-gray-100"
          >
            {{ userStore.user?.name }}
          </span>
        </router-link>
      </div>
    </aside>

    <!-- Main Content -->
    <main :class="[
      'flex-1 overflow-auto border-t border-gray-200 dark:border-gray-700',
      userStore.isLoggedIn && !isSidebarCollapsed ? 'md:ml-0' : '',
      userStore.isLoggedIn && isSidebarCollapsed ? 'md:ml-0' : ''
    ]">
      <router-view />
    </main>

    <!-- Toast Notifications -->
    <ToastNotification ref="toastRef" />
    
    <!-- Overlay for mobile when sidebar is open -->
        <!-- Mobile Overlay -->
    <div
      v-if="userStore.isLoggedIn && !isSidebarCollapsed"
      @click="isSidebarCollapsed = true"
      class="fixed inset-0 bg-black bg-opacity-50 z-40 md:hidden"
    ></div>
    
    <!-- Delete Confirmation Modal -->
    <AlertDialogRoot v-model:open="showDeleteModal">
      <AlertDialogPortal>
        <AlertDialogOverlay class="fixed inset-0 bg-black/50 z-50" />
        <AlertDialogContent class="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6 w-[90vw] max-w-md z-50">
          <AlertDialogTitle class="text-lg font-semibold text-gray-900 dark:text-white mb-2">
            Delete Conversation
          </AlertDialogTitle>
          <AlertDialogDescription class="text-sm text-gray-600 dark:text-gray-400 mb-6">
            Are you sure you want to delete "{{ deleteTarget?.title }}"? This action cannot be undone.
          </AlertDialogDescription>
          <div class="flex justify-end space-x-3">
            <AlertDialogCancel as-child>
              <button class="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors">
                Cancel
              </button>
            </AlertDialogCancel>
            <AlertDialogAction as-child>
              <button @click="executeDelete" class="px-4 py-2 text-sm font-medium text-white bg-red-600 hover:bg-red-700 rounded-lg transition-colors">
                Delete
              </button>
            </AlertDialogAction>
          </div>
        </AlertDialogContent>
      </AlertDialogPortal>
    </AlertDialogRoot>

    <!-- All Chats Modal -->
    <div 
      v-if="showAllChatsModal" 
      class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4"
      @click.self="showAllChatsModal = false"
    >
      <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-2xl w-full max-h-[80vh] flex flex-col border border-gray-200 dark:border-gray-700">
        <!-- Modal Header -->
        <div class="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
          <h2 class="text-xl font-semibold text-gray-900 dark:text-white">All Chats</h2>
          <button
            @click="showAllChatsModal = false"
            class="p-2 text-gray-500 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <!-- Modal Content -->
        <div class="flex-1 overflow-y-auto p-6">
          <div v-if="!chatStore || !Array.isArray(chatStore.recentConversations) || chatStore.recentConversations.length === 0" class="text-center py-12">
            <svg xmlns="http://www.w3.org/2000/svg" class="mx-auto h-12 w-12 text-gray-400 dark:text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
            </svg>
            <p class="mt-4 text-gray-500 dark:text-gray-400">No chats yet</p>
          </div>

          <div v-else class="space-y-2">
            <button
              v-for="chat in chatStore.recentConversations"
              :key="chat.id"
              @click="openChat(chat.chat_uuid!)"
              class="w-full flex items-center justify-between p-4 bg-gray-50 dark:bg-gray-700/30 rounded-lg border border-gray-200 dark:border-gray-700 hover:border-blue-300 dark:hover:border-blue-700 transition-colors group text-left"
              :class="{ 'border-blue-300 dark:border-blue-700 bg-blue-50 dark:bg-blue-900/20': chatStore.currentConversationUUID === chat.chat_uuid }"
            >
              <div class="flex items-start space-x-3 flex-1 min-w-0">
                <MessageSquare class="w-5 h-5 flex-shrink-0 text-gray-400 dark:text-gray-500 mt-0.5" />
                <div class="flex-1 min-w-0">
                  <div class="flex items-center space-x-2 mb-1">
                    <p class="text-sm font-medium text-gray-900 dark:text-white truncate">{{ chat.title }}</p>
                    <span 
                      v-if="chatStore.currentConversationId === chat.id"
                      class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-200"
                    >
                      Active
                    </span>
                  </div>
                  <p class="text-xs text-gray-500 dark:text-gray-400">
                    {{ formatChatDate(chat.updated_at) }}
                  </p>
                </div>
              </div>
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-gray-400 dark:text-gray-500 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
              </svg>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick, provide } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useChatStore } from '@/stores/chat'
import ToastNotification from '@/components/ToastNotification.vue'
import { ToastServiceKey, type ToastService, type ToastOptions } from '@/composables/useToast'
import {
  ChevronLeft,
  MessageSquare,
  Settings,
  Edit,
  Trash2,
  PenSquare,
  Menu,
  Check,
  X,
  Layers,
  Sparkles
} from 'lucide-vue-next'
import {
  AlertDialogRoot,
  AlertDialogPortal,
  AlertDialogOverlay,
  AlertDialogContent,
  AlertDialogTitle,
  AlertDialogDescription,
  AlertDialogCancel,
  AlertDialogAction
} from 'radix-vue'

const router = useRouter()
const userStore = useUserStore()
const chatStore = useChatStore()
const toastRef = ref<InstanceType<typeof ToastNotification> | null>(null)

// Provide toast service
const toastService: ToastService = {
  addToast: (options: ToastOptions) => {
    if (toastRef.value) {
      toastRef.value.addToast(options)
    }
  },
  success: (title: string, message?: string) => {
    if (toastRef.value) {
      toastRef.value.addToast({ type: 'success', title, message })
    }
  },
  error: (title: string, message?: string) => {
    if (toastRef.value) {
      toastRef.value.addToast({ type: 'error', title, message })
    }
  },
  info: (title: string, message?: string) => {
    if (toastRef.value) {
      toastRef.value.addToast({ type: 'info', title, message })
    }
  },
  warning: (title: string, message?: string) => {
    if (toastRef.value) {
      toastRef.value.addToast({ type: 'warning', title, message })
    }
  }
}

provide(ToastServiceKey, toastService)

const isSidebarCollapsed = ref(false)
const editingConvoId = ref<string | null>(null)
const editingTitle = ref('')
const editInput = ref<HTMLInputElement | null>(null)
const showDeleteModal = ref(false)
const deleteTarget = ref<any>(null)
const showAllChatsModal = ref(false)

// Toggle sidebar function
const toggleSidebar = () => {
  isSidebarCollapsed.value = !isSidebarCollapsed.value
}

// Create new chat
const createNewChat = () => {
  const newTitle = `New Chat ${chatStore.conversations.length + 1}`
  chatStore.createConversation(newTitle)
  router.push('/')
  // Close sidebar on mobile after creating chat
  if (window.innerWidth < 768) {
    isSidebarCollapsed.value = true
  }
}

// Start editing conversation inline
const startEdit = (convo: any) => {
  editingConvoId.value = convo.id
  editingTitle.value = convo.title
  nextTick(() => {
    if (editInput.value) {
      editInput.value.focus()
      editInput.value.select()
    }
  })
}

// Save edited title
const saveEdit = () => {
  if (editingTitle.value.trim() && editingConvoId.value) {
    const conversation = chatStore.conversations.find(c => c.id === editingConvoId.value)
    if (conversation) {
      conversation.title = editingTitle.value.trim()
    }
  }
  cancelEdit()
}

// Cancel editing
const cancelEdit = () => {
  editingConvoId.value = null
  editingTitle.value = ''
}

// Show delete confirmation modal
const confirmDelete = (convo: any) => {
  deleteTarget.value = convo
  showDeleteModal.value = true
}

// Execute delete
const executeDelete = () => {
  if (deleteTarget.value) {
    const index = chatStore.conversations.findIndex(c => c.id === deleteTarget.value.id)
    if (index !== -1) {
      chatStore.conversations.splice(index, 1)
      // If deleted conversation was active, switch to another one or create new
      if (chatStore.currentConversationId === deleteTarget.value.id) {
        if (chatStore.conversations.length > 0) {
          chatStore.selectConversation(chatStore.conversations[0].id)
        } else {
          createNewChat()
        }
      }
    }
  }
  showDeleteModal.value = false
  deleteTarget.value = null
}

// Handle theme switching
const isDarkMode = ref(false)

const toggleTheme = () => {
  isDarkMode.value = !isDarkMode.value
  if (isDarkMode.value) {
    document.documentElement.classList.add('dark')
    localStorage.setItem('theme', 'dark')
  } else {
    document.documentElement.classList.remove('dark')
    localStorage.setItem('theme', 'light')
  }
}

// Open chat from modal
const openChat = async (chat_uuid: string) => {
  await chatStore.selectConversation(chat_uuid)
  showAllChatsModal.value = false
  router.push(`/conversation/${chat_uuid}`)
}

// Open recent chat directly (hide secondary sidebar)
const openRecentChat = async (chat_uuid: string) => {
  console.log('Opening recent chat with UUID:', chat_uuid)
  await chatStore.selectConversation(chat_uuid)
  console.log('Chat selected, navigating to:', `/conversation/${chat_uuid}`)
  router.push(`/conversation/${chat_uuid}`)
}

// Format chat date
const formatChatDate = (dateString: string) => {
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

onMounted(() => {
  // Check for saved theme preference or default to light mode
  const savedTheme = localStorage.getItem('theme')
  if (savedTheme === 'dark' || (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
    isDarkMode.value = true
    document.documentElement.classList.add('dark')
  }
  
  // Auto-collapse sidebar on mobile
  if (window.innerWidth < 768) {
    isSidebarCollapsed.value = true
  }
  
  // Handle window resize
  window.addEventListener('resize', () => {
    if (window.innerWidth < 768 && !isSidebarCollapsed.value) {
      // Don't auto-collapse on mobile if user explicitly opened it
    }
  })
})
</script>