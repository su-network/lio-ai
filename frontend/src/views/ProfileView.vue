<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900 py-8">
    <div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="bg-white dark:bg-gray-800 rounded-lg shadow-md border border-gray-200 dark:border-gray-700 p-8">
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white mb-6">User Profile</h1>
        
        <div v-if="userStore.user" class="space-y-6">
          <div class="flex items-center space-x-6">
            <img 
              :src="userStore.user.avatar" 
              alt="User Avatar" 
              class="w-24 h-24 rounded-full border-4 border-blue-500 shadow-lg"
            >
            <div>
              <h2 class="text-2xl font-semibold text-gray-900 dark:text-white">{{ userStore.user.name }}</h2>
              <p class="text-gray-600 dark:text-gray-400">{{ userStore.user.email }}</p>
              <div class="mt-2 inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-200">
                <svg class="w-3 h-3 mr-1" fill="currentColor" viewBox="0 0 20 20">
                  <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"/>
                </svg>
                Active User
              </div>
            </div>
          </div>

          <div class="border-t border-gray-200 dark:border-gray-700 pt-6">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">Account Information</h3>
            <dl class="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div class="bg-gray-50 dark:bg-gray-700/30 px-4 py-3 rounded-lg border border-gray-200 dark:border-gray-700">
                <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">User ID</dt>
                <dd class="mt-1 text-sm text-gray-900 dark:text-white font-mono">{{ userStore.userId }}</dd>
              </div>
              <div class="bg-gray-50 dark:bg-gray-700/30 px-4 py-3 rounded-lg border border-gray-200 dark:border-gray-700">
                <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Account Type</dt>
                <dd class="mt-1 text-sm text-gray-900 dark:text-white">Free Tier</dd>
              </div>
              <div class="bg-gray-50 dark:bg-gray-700/30 px-4 py-3 rounded-lg border border-gray-200 dark:border-gray-700">
                <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Member Since</dt>
                <dd class="mt-1 text-sm text-gray-900 dark:text-white">{{ new Date().toLocaleDateString() }}</dd>
              </div>
              <div class="bg-gray-50 dark:bg-gray-700/30 px-4 py-3 rounded-lg border border-gray-200 dark:border-gray-700">
                <dt class="text-sm font-medium text-gray-500 dark:text-gray-400">Status</dt>
                <dd class="mt-1 text-sm">
                  <span class="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-200">
                    Active
                  </span>
                </dd>
              </div>
            </dl>
          </div>

          <div class="border-t border-gray-200 dark:border-gray-700 pt-6">
            <button 
              @click="userStore.logout" 
              class="w-full sm:w-auto px-6 py-3 bg-red-600 hover:bg-red-700 text-white font-medium rounded-lg border border-red-700 transition-colors"
            >
              <svg class="inline-block w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"/>
              </svg>
              Logout
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
</script>
