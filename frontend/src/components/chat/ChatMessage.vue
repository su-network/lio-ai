<template>
  <div
    :class="[
      'flex w-full',
      message.role === 'user' ? 'justify-end' : 'justify-start'
    ]"
  >
    <div
      :class="[
        'px-4 py-3 rounded-lg relative group',
        hasCodeBlock && message.role === 'assistant' 
          ? 'w-full max-w-full'
          : 'max-w-[85%] sm:max-w-2xl',
        message.role === 'user'
          ? 'bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 border border-gray-200 dark:border-gray-600'
          : 'bg-white dark:bg-gray-800 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-700'
      ]"
    >
      <div v-if="message.role === 'assistant'" class="flex items-center justify-between mb-3 pb-2 border-b border-gray-200 dark:border-gray-700">
        <div class="flex items-center space-x-2">
          <Bot class="w-4 h-4 flex-shrink-0" />
          <span class="text-xs font-medium text-gray-600 dark:text-gray-400">
            {{ chatStore.getModelDisplayName(message.model || chatStore.selectedModel) }}
          </span>
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
        <div class="flex items-center space-x-2">
          <!-- Code/Preview Toggle for code blocks - Always visible when code exists -->
          <div v-if="hasCodeBlock" class="flex items-center bg-gray-100 dark:bg-gray-700 rounded-lg p-1">
            <button
              @click="viewMode = 'code'"
              :class="[
                'px-3 py-1 text-xs font-medium rounded-md transition-all flex items-center space-x-1',
                viewMode === 'code'
                  ? 'bg-white dark:bg-gray-600 text-gray-900 dark:text-white shadow-sm'
                  : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white'
              ]"
              title="View code"
            >
              <Code2 class="w-3.5 h-3.5" />
              <span>Code</span>
            </button>
            <button
              @click="viewMode = 'preview'"
              :class="[
                'px-3 py-1 text-xs font-medium rounded-md transition-all flex items-center space-x-1',
                viewMode === 'preview'
                  ? 'bg-white dark:bg-gray-600 text-gray-900 dark:text-white shadow-sm'
                  : 'text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white'
              ]"
              title="View preview"
            >
              <Eye class="w-3.5 h-3.5" />
              <span>Preview</span>
            </button>
          </div>
          <!-- Action buttons - Show on hover -->
          <div class="flex items-center space-x-1 opacity-0 group-hover:opacity-100 transition-opacity">
            <button
              v-if="hasCodeBlock"
              @click="downloadCode"
              class="p-1.5 text-gray-400 hover:text-blue-600 dark:hover:text-blue-400 rounded hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
              title="Download code"
            >
              <Download class="w-3.5 h-3.5" />
            </button>
            <button
              @click="$emit('copy-message', message)"
              class="p-1.5 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 rounded hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
              title="Copy message"
            >
              <Copy class="w-3.5 h-3.5" />
            </button>
            <button
              @click="$emit('edit-message', message)"
              class="p-1.5 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 rounded hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
              title="Edit message"
            >
              <Edit class="w-3.5 h-3.5" />
            </button>
          </div>
        </div>
      </div>

      <!-- Content Area -->
      <div class="message-content">
        <!-- Code View -->
        <div v-if="viewMode === 'code' || !hasCodeBlock">
          <div 
            class="prose prose-sm dark:prose-invert max-w-none break-words"
            :class="hasCodeBlock ? 'prose-pre:max-w-full prose-pre:overflow-x-auto' : ''"
            v-html="chatStore.formatMessageContent(message.content)"
          ></div>
        </div>

        <!-- Preview View -->
        <div v-else-if="viewMode === 'preview'" class="space-y-3">
          <!-- Preview Controls -->
          <div class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-700">
            <div class="flex items-center space-x-2">
              <Eye class="w-4 h-4 text-gray-600 dark:text-gray-400" />
              <span class="text-sm font-medium text-gray-700 dark:text-gray-300">Live Preview</span>
            </div>
            <div class="flex items-center space-x-2">
              <select
                v-model="selectedLanguage"
                class="text-xs px-3 py-1.5 bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                <option value="html">HTML</option>
                <option value="javascript">JavaScript</option>
                <option value="typescript">TypeScript</option>
                <option value="python">Python</option>
                <option value="css">CSS</option>
                <option value="json">JSON</option>
              </select>
              <button
                @click="downloadCode"
                class="px-3 py-1.5 text-xs font-medium bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors flex items-center space-x-1.5 shadow-sm"
                title="Download code"
              >
                <Download class="w-3.5 h-3.5" />
                <span>Download</span>
              </button>
            </div>
          </div>

          <!-- Preview Content -->
          <div class="bg-white dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden">
            <!-- HTML/CSS Preview with iframe -->
            <div v-if="selectedLanguage === 'html' || selectedLanguage === 'css'" class="relative">
              <div class="absolute top-2 right-2 z-10">
                <span class="px-2 py-1 text-xs bg-gray-900/80 text-white rounded backdrop-blur-sm">
                  Live Preview
                </span>
              </div>
              <iframe
                :srcdoc="previewContent"
                class="w-full min-h-[400px] border-0 bg-white"
                sandbox="allow-scripts allow-modals allow-forms"
              ></iframe>
            </div>
            
            <!-- Code Preview for other languages -->
            <div v-else class="p-4 overflow-x-auto">
              <div class="flex items-center justify-between mb-2 pb-2 border-b border-gray-200 dark:border-gray-700">
                <span class="text-xs font-medium text-gray-600 dark:text-gray-400 uppercase">
                  {{ selectedLanguage }} Code
                </span>
                <span class="text-xs text-gray-500 dark:text-gray-500">
                  {{ extractedCode.split('\n').length }} lines
                </span>
              </div>
              <pre class="text-sm text-gray-800 dark:text-gray-200 overflow-x-auto"><code>{{ extractedCode }}</code></pre>
            </div>
          </div>

          <!-- Preview Info -->
          <div class="flex items-center justify-between px-3 py-2 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg">
            <div class="flex items-center space-x-2 text-xs text-blue-700 dark:text-blue-300">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span>{{ selectedLanguage === 'html' ? 'Interactive preview - changes render in real-time' : 'Switch to HTML for live preview' }}</span>
            </div>
          </div>
        </div>
      </div>

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
import { defineProps, defineEmits, ref, computed } from 'vue'
import { Bot, Copy, Edit, Clock, Zap, Code2, Eye, Download } from 'lucide-vue-next'
import type { Message } from '@/types'
import { useChatStore } from '@/stores/chat'

const props = defineProps<{
  message: Message
}>()

defineEmits(['copy-message', 'edit-message'])

const chatStore = useChatStore()
const viewMode = ref<'code' | 'preview'>('code')
const selectedLanguage = ref('html')

// Check if message contains code blocks
const hasCodeBlock = computed(() => {
  return props.message.content.includes('```')
})

// Extract code from code blocks
const extractedCode = computed(() => {
  const codeBlockRegex = /```[\w]*\n([\s\S]*?)```/g
  const matches = [...props.message.content.matchAll(codeBlockRegex)]
  if (matches.length > 0) {
    return matches.map(m => m[1]).join('\n\n')
  }
  return props.message.content
})

// Prepare preview content with proper HTML structure
const previewContent = computed(() => {
  const code = extractedCode.value
  
  if (selectedLanguage.value === 'html') {
    // Check if it's a complete HTML document
    if (code.includes('<!DOCTYPE') || code.includes('<html')) {
      return code
    }
    // Wrap fragment in basic HTML structure
    return `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <style>
    body { margin: 0; padding: 20px; font-family: system-ui, -apple-system, sans-serif; }
    * { box-sizing: border-box; }
  </style>
</head>
<body>
  ${code}
</body>
</html>`
  } else if (selectedLanguage.value === 'css') {
    // Wrap CSS in HTML with style tag
    return `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <style>
    ${code}
  </style>
</head>
<body>
  <div class="demo-container">
    <h1>CSS Preview</h1>
    <p>Add your HTML here to see the styles applied.</p>
  </div>
</body>
</html>`
  }
  
  return code
})

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

const downloadCode = () => {
  const blob = new Blob([extractedCode.value], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  
  const extensions: { [key: string]: string } = {
    html: 'html',
    css: 'css',
    javascript: 'js',
    typescript: 'ts',
    python: 'py',
    json: 'json',
    java: 'java',
    go: 'go',
    rust: 'rs',
    cpp: 'cpp'
  }
  
  const ext = extensions[selectedLanguage.value] || 'txt'
  const timestamp = new Date().toISOString().slice(0, 10)
  a.download = `generated-code-${timestamp}.${ext}`
  
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}
</script>

<style>
/* Global code block styling for chat messages */
.message-content pre {
  background-color: rgb(17 24 39);
  border-radius: 0.5rem;
  padding: 1rem;
  overflow-x: auto;
  border: 1px solid rgb(55 65 81);
  margin: 0.5rem 0;
}

.dark .message-content pre {
  background-color: rgb(3 7 18);
  border-color: rgb(31 41 55);
}

.message-content code {
  font-size: 0.875rem;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  color: rgb(243 244 246);
}

.message-content pre code {
  background: transparent;
  padding: 0;
}

/* Inline code */
.message-content p code,
.message-content li code {
  background-color: rgb(243 244 246);
  padding: 0.125rem 0.375rem;
  border-radius: 0.25rem;
  font-size: 0.875rem;
  color: rgb(31 41 55);
}

.dark .message-content p code,
.dark .message-content li code {
  background-color: rgb(55 65 81);
  color: rgb(229 231 235);
}

/* Prose styling */
.message-content .prose {
  color: rgb(31 41 55);
}

.dark .message-content .prose {
  color: rgb(229 231 235);
}

.message-content .prose h1,
.message-content .prose h2,
.message-content .prose h3 {
  color: rgb(17 24 39);
  font-weight: 600;
}

.dark .message-content .prose h1,
.dark .message-content .prose h2,
.dark .message-content .prose h3 {
  color: rgb(243 244 246);
}

.message-content .prose p {
  margin: 0.5rem 0;
}

.message-content .prose ul,
.message-content .prose ol {
  margin: 0.5rem 0;
}

.message-content .prose strong {
  font-weight: 600;
  color: rgb(17 24 39);
}

.dark .message-content .prose strong {
  color: rgb(243 244 246);
}

.message-content .prose a {
  color: rgb(37 99 235);
  text-decoration: none;
}

.message-content .prose a:hover {
  text-decoration: underline;
}

.dark .message-content .prose a {
  color: rgb(96 165 250);
}

/* Scrollbar styling for code blocks */
.message-content pre::-webkit-scrollbar {
  height: 8px;
}

.message-content pre::-webkit-scrollbar-track {
  background-color: rgb(31 41 55);
  border-radius: 0.25rem;
}

.message-content pre::-webkit-scrollbar-thumb {
  background-color: rgb(75 85 99);
  border-radius: 0.25rem;
}

.message-content pre::-webkit-scrollbar-thumb:hover {
  background-color: rgb(107 114 128);
}

/* Preview iframe styling */
iframe {
  opacity: 1;
  transition: opacity 0.3s;
}

/* Smooth transitions for view switching */
.message-content {
  transition: all 0.2s;
}
</style>