<template>
  <div class="bg-white dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700 px-2 sm:px-4 py-3 sm:py-4">
    <!-- Attached files indicator -->
    <div v-if="attachedFiles.length > 0" class="flex flex-wrap gap-2 mb-3 px-1">
      <div 
        v-for="(file, index) in attachedFiles" 
        :key="index"
        class="flex items-center space-x-2 px-3 py-1 bg-gray-100 dark:bg-gray-700 rounded-full text-sm"
      >
        <Paperclip class="w-3 h-3" />
        <span class="text-gray-700 dark:text-gray-300 max-w-[200px] truncate">{{ file.name }}</span>
        <button 
          @click="removeFile(index)"
          class="text-gray-500 hover:text-gray-700 dark:hover:text-gray-300 ml-1"
        >
          <X class="w-3 h-3" />
        </button>
      </div>
    </div>

    <!-- Code Generation Options -->
    <div v-if="codeGenEnabled" class="mb-3 p-4 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-700 rounded-lg">
      <div class="flex items-center justify-between mb-3">
        <div class="flex items-center space-x-2">
          <Code class="w-4 h-4 text-blue-600 dark:text-blue-400" />
          <span class="text-sm font-medium text-blue-900 dark:text-blue-100">Code Generation Mode</span>
        </div>
        <button 
          @click="codeGenEnabled = false"
          class="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-200"
        >
          <X class="w-4 h-4" />
        </button>
      </div>
      
      <div class="grid grid-cols-2 gap-3">
        <div>
          <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
            Language
          </label>
          <select
            v-model="selectedLanguage"
            class="w-full px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
          >
            <option value="javascript">JavaScript</option>
            <option value="typescript">TypeScript</option>
            <option value="python">Python</option>
            <option value="java">Java</option>
            <option value="go">Go</option>
            <option value="rust">Rust</option>
            <option value="cpp">C++</option>
            <option value="csharp">C#</option>
          </select>
        </div>
        
        <div>
          <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
            Framework
          </label>
          <select
            v-model="selectedFramework"
            class="w-full px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
          >
            <option value="none">None</option>
            <option value="react">React</option>
            <option value="vue">Vue.js</option>
            <option value="angular">Angular</option>
            <option value="express">Express.js</option>
            <option value="django">Django</option>
            <option value="fastapi">FastAPI</option>
            <option value="spring">Spring Boot</option>
            <option value="dotnet">.NET</option>
          </select>
        </div>
      </div>
      
      <p class="text-xs text-gray-600 dark:text-gray-400 mt-2">
        Describe what you want to build, and AI will generate the code for you.
      </p>
    </div>

    <!-- Input area -->
    <div class="flex space-x-2 items-end">
      <!-- Options dropdown using Radix Popover -->
      <PopoverRoot v-model:open="showOptionsMenu">
        <PopoverTrigger as-child>
          <button
            class="p-3 text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white hover:bg-gray-50 dark:hover:bg-gray-800 rounded-lg transition-colors"
            style="min-height: 48px;"
          >
            <Settings class="w-5 h-5" />
          </button>
        </PopoverTrigger>
        
        <PopoverPortal>
          <PopoverContent
            side="top"
            :side-offset="8"
            class="w-64 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-600 rounded-lg shadow-lg z-50 p-3 space-y-3"
          >
            <div class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide">Options</div>
            
            <!-- Web Search Toggle using Radix Switch -->
            <div class="flex items-center justify-between py-2" :class="{ 'opacity-50': !supportsWebSearch }">
              <Label class="flex items-center space-x-2" :class="supportsWebSearch ? 'cursor-pointer' : 'cursor-not-allowed'">
                <Globe class="w-4 h-4 text-gray-600 dark:text-gray-400" />
                <span class="text-sm text-gray-900 dark:text-white">Web Search</span>
              </Label>
              <SwitchRoot
                v-model:checked="webSearchEnabled"
                :disabled="!supportsWebSearch"
                class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors data-[state=checked]:bg-gray-800 data-[state=unchecked]:bg-gray-200 dark:data-[state=checked]:bg-gray-600 dark:data-[state=unchecked]:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <SwitchThumb
                  class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform data-[state=checked]:translate-x-6 data-[state=unchecked]:translate-x-1"
                />
              </SwitchRoot>
            </div>
            <p v-if="!supportsWebSearch" class="text-xs text-gray-500 dark:text-gray-400 -mt-1 mb-1">
              Current model doesn't support web search
            </p>

            <!-- Code Generation Toggle -->
            <div class="flex items-center justify-between py-2" :class="{ 'opacity-50': !supportsCodeGeneration }">
              <Label class="flex items-center space-x-2" :class="supportsCodeGeneration ? 'cursor-pointer' : 'cursor-not-allowed'">
                <Code class="w-4 h-4 text-gray-600 dark:text-gray-400" />
                <span class="text-sm text-gray-900 dark:text-white">Code Generation</span>
              </Label>
              <SwitchRoot
                v-model:checked="codeGenEnabled"
                :disabled="!supportsCodeGeneration"
                class="relative inline-flex h-6 w-11 items-center rounded-full transition-colors data-[state=checked]:bg-blue-600 data-[state=unchecked]:bg-gray-200 dark:data-[state=checked]:bg-blue-500 dark:data-[state=unchecked]:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <SwitchThumb
                  class="inline-block h-4 w-4 transform rounded-full bg-white transition-transform data-[state=checked]:translate-x-6 data-[state=unchecked]:translate-x-1"
                />
              </SwitchRoot>
            </div>
            <p v-if="!supportsCodeGeneration" class="text-xs text-gray-500 dark:text-gray-400 -mt-1 mb-1">
              Current model doesn't support code generation
            </p>
            
            <!-- Attach File -->
            <button
              @click="handleAttachClick"
              :disabled="!supportsFileAttachment"
              class="flex items-center space-x-2 w-full py-2 px-2 rounded transition-colors"
              :class="supportsFileAttachment ? 'hover:bg-gray-50 dark:hover:bg-gray-700 cursor-pointer' : 'opacity-50 cursor-not-allowed'"
            >
              <Paperclip class="w-4 h-4 text-gray-600 dark:text-gray-400" />
              <span class="text-sm text-gray-900 dark:text-white">Attach File</span>
            </button>
            <p v-if="!supportsFileAttachment" class="text-xs text-gray-500 dark:text-gray-400 px-2 -mt-1">
              Current model doesn't support file attachments
            </p>
            <input 
              ref="fileInput" 
              type="file" 
              multiple 
              accept=".pdf,.doc,.docx,.txt,.csv,.xlsx,.xls"
              class="hidden"
              @change="handleFileUpload"
            />
          </PopoverContent>
        </PopoverPortal>
      </PopoverRoot>
      
      <div class="flex-1 relative">
        <textarea
          v-model="inputMessage"
          @keydown.enter.exact.prevent="sendMessage"
          @keydown.enter.shift.exact="inputMessage += '\n'"
          :placeholder="hasAvailableModels ? placeholder : 'Please configure API keys in Settings to start chatting'"
          :disabled="!hasAvailableModels"
          class="w-full resize-none rounded-lg border border-gray-300 dark:border-gray-600 px-4 py-3 bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:ring-2 focus:ring-gray-400 focus:border-transparent text-sm sm:text-base disabled:bg-gray-100 dark:disabled:bg-gray-800 disabled:cursor-not-allowed disabled:text-gray-500"
          rows="1"
          style="min-height: 48px; max-height: 120px;"
          ref="messageInput"
        ></textarea>
        <div class="absolute bottom-3 right-3 text-xs text-gray-400 hidden sm:block">
          {{ inputMessage.length }}/4000
        </div>
      </div>
      <button
        @click="sendMessage"
        :disabled="!inputMessage.trim() || isTyping || !hasAvailableModels"
        class="px-4 sm:px-5 py-3 bg-gray-800 dark:bg-gray-700 hover:bg-gray-900 dark:hover:bg-gray-600 disabled:bg-gray-400 disabled:cursor-not-allowed text-white rounded-lg transition-colors flex items-center space-x-2 flex-shrink-0"
        style="min-height: 48px;"
      >
        <Send class="w-5 h-5" />
        <span class="hidden sm:inline text-sm font-medium">Send</span>
      </button>
    </div>

    <div class="flex flex-wrap gap-2 mt-3 hidden sm:flex">
      <button
        v-for="prompt in quickPrompts"
        :key="prompt.id"
        @click="selectQuickPrompt(prompt)"
        class="px-3 py-1 text-xs text-gray-700 dark:text-gray-300 rounded-full border border-gray-200 dark:border-gray-700 transition-colors hover:bg-gray-50 dark:hover:bg-gray-800"
      >
        {{ prompt.label }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, defineProps, defineEmits, onMounted, watch, computed } from 'vue'
import { Send, X, Paperclip, Globe, Settings, Code } from 'lucide-vue-next'
import { 
  PopoverRoot, 
  PopoverTrigger, 
  PopoverContent, 
  PopoverPortal,
  SwitchRoot,
  SwitchThumb,
  Label
} from 'radix-vue'
import type { Message, QuickPrompt, Model } from '@/types'

const props = defineProps<{
  isTyping: boolean
  quickPrompts: QuickPrompt[]
  placeholder: string
  hasAvailableModels?: boolean
  selectedModel?: string
  availableModels?: Model[]
}>()

const emit = defineEmits(['sendMessage', 'selectQuickPrompt'])

const inputMessage = ref('')
const messageInput = ref<HTMLTextAreaElement>()
const fileInput = ref<HTMLInputElement>()
const webSearchEnabled = ref(false)
const codeGenEnabled = ref(false)
const selectedLanguage = ref('typescript')
const selectedFramework = ref('react')
const attachedFiles = ref<File[]>([])
const showOptionsMenu = ref(false)

// Compute model capabilities
const currentModel = computed(() => {
  if (!props.selectedModel || !props.availableModels) return null
  return props.availableModels.find(m => m.id === props.selectedModel)
})

const supportsWebSearch = computed(() => {
  const features = currentModel.value?.capabilities?.special_features || []
  return features.includes('web-search') || features.includes('grounding')
})

const supportsCodeGeneration = computed(() => {
  const features = currentModel.value?.capabilities?.special_features || []
  return features.includes('code-generation') || features.includes('code-analysis')
})

const supportsFileAttachment = computed(() => {
  const features = currentModel.value?.capabilities?.special_features || []
  return features.includes('multimodal') || features.includes('file-upload') || features.includes('vision')
})

const toggleWebSearch = () => {
  webSearchEnabled.value = !webSearchEnabled.value
}

const handleAttachClick = () => {
  if (supportsFileAttachment.value && fileInput.value) {
    fileInput.value.click()
    showOptionsMenu.value = false
  }
}

const handleFileUpload = (event: Event) => {
  const target = event.target as HTMLInputElement
  const files = target.files
  if (files && files.length > 0) {
    attachedFiles.value.push(...Array.from(files))
    // Reset input to allow selecting the same file again
    target.value = ''
  }
}

const removeFile = (index: number) => {
  attachedFiles.value.splice(index, 1)
}

const sendMessage = () => {
  const messageData: any = {
    content: inputMessage.value
  }
  
  // Add code generation metadata if enabled
  if (codeGenEnabled.value) {
    messageData.codeGeneration = {
      enabled: true,
      language: selectedLanguage.value,
      framework: selectedFramework.value
    }
  }
  
  // Add web search flag if enabled
  if (webSearchEnabled.value) {
    messageData.webSearch = true
  }
  
  // Add attached files if any
  if (attachedFiles.value.length > 0) {
    messageData.files = attachedFiles.value
  }
  
  emit('sendMessage', messageData)
  inputMessage.value = ''
  // Reset feature toggles and files after sending
  codeGenEnabled.value = false
  webSearchEnabled.value = false
  attachedFiles.value = []
  if (messageInput.value) {
    messageInput.value.style.height = '48px'
  }
}

const selectQuickPrompt = (prompt: QuickPrompt) => {
  inputMessage.value = prompt.prompt
  if (messageInput.value) {
    messageInput.value.focus()
  }
}

onMounted(() => {
  const textarea = messageInput.value
  if (textarea) {
    textarea.addEventListener('input', () => {
      textarea.style.height = '40px'
      textarea.style.height = Math.min(textarea.scrollHeight, 120) + 'px'
    })
  }
})
</script>