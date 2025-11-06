<template>
  <div
    :class="[
      'flex',
      message.role === 'user' ? 'justify-end' : 'justify-start'
    ]"
  >
    <div
      :class="[
        'max-w-[85%] sm:max-w-2xl px-4 py-3 rounded-lg relative group break-words',
        message.role === 'user'
          ? 'bg-gray-800 dark:bg-gray-700 text-white border border-gray-700 dark:border-gray-600'
          : 'bg-white dark:bg-gray-800 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-700'
      ]"
    >
      <div v-if="message.role === 'assistant'" class="flex items-center justify-between mb-2">
        <div class="flex items-center space-x-2">
          <Bot class="w-4 h-4" />
          <span class="text-xs text-gray-500 dark:text-gray-400">{{ chatStore.getModelDisplayName(message.model || chatStore.selectedModel) }}</span>
          <div class="flex space-x-1">
            <span
              v-for="badge in chatStore.getModelBadges(message.model || chatStore.selectedModel)"
              :key="badge"
              class="px-1.5 py-0.5 text-xs rounded-full"
              :class="getBadgeClass(badge)"
            >
              {{ badge }}
            </span>
          </div>
        </div>
        <div class="flex items-center space-x-1 opacity-0 group-hover:opacity-100 transition-opacity">
          <button
            @click="$emit('copy-message', message)"
            class="p-1 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 rounded"
            title="Copy message"
          >
            <Copy class="w-3 h-3" />
          </button>
          <button
            @click="$emit('edit-message', message)"
            class="p-1 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 rounded"
            title="Edit message"
          >
            <Edit class="w-3 h-3" />
          </button>
        </div>
      </div>

      <div class="prose prose-sm dark:prose-invert max-w-none break-words whitespace-pre-wrap" v-html="chatStore.formatMessageContent(message.content)"></div>

      <div class="flex items-center justify-between mt-2 text-xs text-gray-400 dark:text-gray-500">
        <span>{{ formatTime(message.created_at || message.timestamp) }}</span>
        <div v-if="message.role === 'assistant'" class="flex items-center space-x-2">
          <span v-if="message.responseTime" class="flex items-center space-x-1">
            <Clock class="w-3 h-3" />
            <span>{{ message.responseTime }}ms</span>
          </span>
          <span v-if="message.tokens" class="flex items-center space-x-1">
            <Zap class="w-3 h-3" />
            <span>{{ message.tokens }} tokens</span>
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { defineProps, defineEmits } from 'vue'
import { Bot, Copy, Edit, Clock, Zap } from 'lucide-vue-next'
import type { Message } from '@/types'
import { useChatStore } from '@/stores/chat'

const props = defineProps<{
  message: Message
}>()

defineEmits(['copy-message', 'edit-message'])

const chatStore = useChatStore()

const getBadgeClass = (badge: string): string => {
  switch (badge) {
    case 'Pro':
      return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200'
    case 'Vision':
      return 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200'
    case 'New':
      return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
    default:
      return 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200'
  }
}

const formatTime = (date: Date | string | undefined): string => {
  if (!date) return ''
  const dateObj = date instanceof Date ? date : new Date(date)
  return dateObj.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}
</script>