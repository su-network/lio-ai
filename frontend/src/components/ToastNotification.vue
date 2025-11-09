<template>
  <teleport to="body">
    <div class="fixed top-4 right-4 z-[10000] space-y-2">
      <transition-group name="toast">
        <div
          v-for="toast in toasts"
          :key="toast.id"
          :class="[
            'max-w-md px-6 py-4 rounded-xl shadow-2xl border-2 backdrop-blur-sm',
            'transform transition-all duration-300',
            toast.type === 'success' && 'bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800',
            toast.type === 'error' && 'bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800',
            toast.type === 'info' && 'bg-blue-50 dark:bg-blue-900/20 border-blue-200 dark:border-blue-800',
            toast.type === 'warning' && 'bg-yellow-50 dark:bg-yellow-900/20 border-yellow-200 dark:border-yellow-800'
          ]"
        >
          <div class="flex items-start gap-3">
            <!-- Icon -->
            <div class="flex-shrink-0">
              <svg
                v-if="toast.type === 'success'"
                class="w-6 h-6 text-green-600 dark:text-green-400"
                fill="currentColor"
                viewBox="0 0 20 20"
              >
                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"/>
              </svg>
              <svg
                v-else-if="toast.type === 'error'"
                class="w-6 h-6 text-red-600 dark:text-red-400"
                fill="currentColor"
                viewBox="0 0 20 20"
              >
                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd"/>
              </svg>
              <svg
                v-else-if="toast.type === 'warning'"
                class="w-6 h-6 text-yellow-600 dark:text-yellow-400"
                fill="currentColor"
                viewBox="0 0 20 20"
              >
                <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd"/>
              </svg>
              <svg
                v-else
                class="w-6 h-6 text-blue-600 dark:text-blue-400"
                fill="currentColor"
                viewBox="0 0 20 20"
              >
                <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd"/>
              </svg>
            </div>

            <!-- Content -->
            <div class="flex-1 min-w-0">
              <p
                :class="[
                  'text-sm font-semibold',
                  toast.type === 'success' && 'text-green-900 dark:text-green-200',
                  toast.type === 'error' && 'text-red-900 dark:text-red-200',
                  toast.type === 'info' && 'text-blue-900 dark:text-blue-200',
                  toast.type === 'warning' && 'text-yellow-900 dark:text-yellow-200'
                ]"
              >
                {{ toast.title }}
              </p>
              <p
                v-if="toast.message"
                :class="[
                  'text-xs mt-1',
                  toast.type === 'success' && 'text-green-800 dark:text-green-300',
                  toast.type === 'error' && 'text-red-800 dark:text-red-300',
                  toast.type === 'info' && 'text-blue-800 dark:text-blue-300',
                  toast.type === 'warning' && 'text-yellow-800 dark:text-yellow-300'
                ]"
              >
                {{ toast.message }}
              </p>
            </div>

            <!-- Close button -->
            <button
              @click="removeToast(toast.id)"
              :class="[
                'flex-shrink-0 rounded-lg p-1 transition-colors',
                toast.type === 'success' && 'text-green-600 dark:text-green-400 hover:bg-green-100 dark:hover:bg-green-900/40',
                toast.type === 'error' && 'text-red-600 dark:text-red-400 hover:bg-red-100 dark:hover:bg-red-900/40',
                toast.type === 'info' && 'text-blue-600 dark:text-blue-400 hover:bg-blue-100 dark:hover:bg-blue-900/40',
                toast.type === 'warning' && 'text-yellow-600 dark:text-yellow-400 hover:bg-yellow-100 dark:hover:bg-yellow-900/40'
              ]"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
              </svg>
            </button>
          </div>
        </div>
      </transition-group>
    </div>
  </teleport>
</template>

<script setup lang="ts">
import { ref } from 'vue'

interface Toast {
  id: number
  type: 'success' | 'error' | 'info' | 'warning'
  title: string
  message?: string
  duration?: number
}

const toasts = ref<Toast[]>([])
let nextId = 1

const addToast = (toast: Omit<Toast, 'id'>) => {
  const id = nextId++
  const duration = toast.duration || 3000
  
  toasts.value.push({ ...toast, id })
  
  setTimeout(() => {
    removeToast(id)
  }, duration)
}

const removeToast = (id: number) => {
  const index = toasts.value.findIndex(t => t.id === id)
  if (index !== -1) {
    toasts.value.splice(index, 1)
  }
}

defineExpose({ addToast })
</script>

<style scoped>
.toast-enter-active,
.toast-leave-active {
  transition: all 0.3s ease;
}

.toast-enter-from {
  opacity: 0;
  transform: translateX(100px);
}

.toast-leave-to {
  opacity: 0;
  transform: translateX(100px) scale(0.8);
}
</style>
