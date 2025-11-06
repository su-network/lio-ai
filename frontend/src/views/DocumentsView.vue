<template>
  <div class="container mx-auto px-4 py-8">
    <div class="max-w-6xl mx-auto">
      <div class="flex items-center justify-between mb-8">
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">Recent Chats</h1>
      </div>

      <div class="space-y-4">
        <div
          v-for="convo in chatStore.recentConversations"
          :key="convo.id"
          class="bg-white dark:bg-gray-800 p-4 rounded-lg shadow-md"
        >
          <h2 class="text-xl font-semibold text-gray-800 dark:text-gray-200 mb-2">
            {{ convo.title }}
          </h2>
          <div class="text-sm text-gray-500 dark:text-gray-400 mb-4">
            Last updated: {{ new Date(convo.lastUpdated).toLocaleString() }}
          </div>
          <div class="space-y-2">
            <div
              v-for="message in convo.messages.slice(0, 2)"
              :key="message.id"
              class="flex"
            >
              <div class="font-semibold w-20 flex-shrink-0">{{ message.role }}:</div>
              <div class="text-gray-700 dark:text-gray-300">{{ message.content }}</div>
            </div>
            <div
              v-if="convo.messages.length > 2"
              class="text-sm text-gray-500 dark:text-gray-400"
            >
              ... and {{ convo.messages.length - 2 }} more messages.
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useChatStore } from '@/stores/chat'

const chatStore = useChatStore()
</script>