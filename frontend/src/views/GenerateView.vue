<template>
  <div class="container mx-auto px-4 py-8">
    <div class="max-w-6xl mx-auto">
      <div class="flex items-center justify-between mb-8">
        <div>
          <h1 class="text-3xl font-bold text-gray-900 dark:text-white">Generate Code</h1>
          <p class="text-gray-600 dark:text-gray-300 mt-2">Transform documents and prompts into functional code</p>
        </div>
        <div class="flex items-center space-x-4">
          <button
            @click="showTemplates = true"
            class="px-4 py-2 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
          >
            <Sparkles class="w-4 h-4 inline mr-2" />
            Templates
          </button>
        </div>
      </div>

      <div class="grid lg:grid-cols-3 gap-8">
        <!-- Input Section -->
        <div class="lg:col-span-2 space-y-6">
          <!-- Input Source -->
          <div class="bg-white dark:bg-gray-800 p-6 rounded-lg shadow-md">
            <h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-4">Input Source</h2>

            <!-- Tab Navigation -->
            <div class="flex space-x-1 mb-6">
              <button
                v-for="tab in tabs"
                :key="tab.id"
                @click="activeTab = tab.id"
                :class="[
                  'px-4 py-2 rounded-md text-sm font-medium transition-colors flex items-center space-x-2',
                  activeTab === tab.id
                    ? 'bg-blue-600 text-white'
                    : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-600'
                ]"
              >
                <component :is="tab.icon" class="w-4 h-4" />
                <span>{{ tab.name }}</span>
              </button>
            </div>

            <!-- File Upload Tab -->
            <div v-if="activeTab === 'upload'" class="space-y-4">
              <!-- Recently Accessed Files -->
              <div v-if="recentFiles.length > 0" class="mb-4">
                <h3 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Recent Files</h3>
                <div class="flex flex-wrap gap-2">
                  <button
                    v-for="file in recentFiles.slice(0, 5)"
                    :key="file.id"
                    @click="selectRecentFile(file)"
                    class="px-3 py-1 text-xs bg-blue-50 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 rounded-full hover:bg-blue-100 dark:hover:bg-blue-900/50 transition-colors"
                  >
                    {{ file.name }}
                  </button>
                </div>
              </div>

              <!-- Drag & Drop Zone -->
              <div
                @drop.prevent="onDrop"
                @dragover.prevent
                @dragenter.prevent
                :class="[
                  'border-2 border-dashed rounded-lg p-8 text-center transition-colors cursor-pointer relative',
                  isDragOver
                    ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
                    : 'border-gray-300 dark:border-gray-600 hover:border-blue-400 dark:hover:border-blue-500'
                ]"
                @click="$refs.fileInput.click()"
              >
                <Upload class="w-12 h-12 text-gray-400 mx-auto mb-4" />
                <p class="text-gray-600 dark:text-gray-300 mb-2 font-medium">
                  Drop files here or click to browse
                </p>
                <p class="text-sm text-gray-500 dark:text-gray-400">
                  Supports PDF, TXT, CODE, JSON, YAML, images with OCR
                </p>
                <div class="flex justify-center space-x-4 mt-4 text-xs text-gray-400">
                  <span class="flex items-center space-x-1">
                    <FileText class="w-3 h-3" />
                    <span>Documents</span>
                  </span>
                  <span class="flex items-center space-x-1">
                    <Code class="w-3 h-3" />
                    <span>Code</span>
                  </span>
                  <span class="flex items-center space-x-1">
                    <Image class="w-3 h-3" />
                    <span>Images</span>
                  </span>
                </div>
                <input
                  ref="fileInput"
                  type="file"
                  multiple
                  accept=".pdf,.txt,.js,.ts,.py,.java,.cpp,.c,.cs,.php,.rb,.go,.rs,.json,.yaml,.yml,.png,.jpg,.jpeg,.gif,.svg"
                  class="hidden"
                  @change="onFileSelect"
                />
              </div>

              <!-- File Preview -->
              <div v-if="uploadedFiles.length > 0" class="space-y-3">
                <h3 class="text-sm font-medium text-gray-700 dark:text-gray-300">Uploaded Files</h3>
                <div
                  v-for="file in uploadedFiles"
                  :key="file.name"
                  class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-700 rounded-lg"
                >
                  <div class="flex items-center space-x-3">
                    <div class="w-8 h-8 bg-blue-100 dark:bg-blue-900 rounded-lg flex items-center justify-center">
                      <component :is="getFileIcon(file.type)" class="w-4 h-4 text-blue-600 dark:text-blue-400" />
                    </div>
                    <div>
                      <p class="text-sm font-medium text-gray-900 dark:text-white">{{ file.name }}</p>
                      <div class="flex items-center space-x-2 text-xs text-gray-500 dark:text-gray-400">
                        <span>{{ formatFileSize(file.size) }}</span>
                        <span>•</span>
                        <span>{{ file.type.toUpperCase() }}</span>
                        <span>•</span>
                        <span>{{ formatDate(file.lastModified) }}</span>
                      </div>
                    </div>
                  </div>
                  <div class="flex items-center space-x-2">
                    <button
                      @click="previewFile(file)"
                      class="p-1 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                      title="Preview"
                    >
                      <Eye class="w-4 h-4" />
                    </button>
                    <button
                      @click="removeFile(file)"
                      class="p-1 text-red-400 hover:text-red-600"
                      title="Remove"
                    >
                      <X class="w-4 h-4" />
                    </button>
                  </div>
                </div>
              </div>

              <!-- Upload Progress -->
              <div v-if="isUploading" class="space-y-2">
                <div class="flex justify-between text-sm">
                  <span class="text-gray-600 dark:text-gray-300">Uploading...</span>
                  <span class="text-gray-600 dark:text-gray-300">{{ uploadProgress }}%</span>
                </div>
                <div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                  <div
                    class="bg-blue-600 h-2 rounded-full transition-all duration-300"
                    :style="{ width: uploadProgress + '%' }"
                  ></div>
                </div>
              </div>
            </div>

            <!-- Text Input Tab -->
            <div v-if="activeTab === 'text'" class="space-y-4">
              <div class="relative">
                <textarea
                  v-model="inputText"
                  placeholder="Describe what you want to generate... e.g., 'Create a React component for a user profile card with TypeScript'"
                  class="w-full h-40 p-4 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white resize-none focus:ring-2 focus:ring-blue-500 focus:border-transparent font-mono text-sm"
                  spellcheck="false"
                ></textarea>
                <div class="absolute bottom-2 right-2 text-xs text-gray-400">
                  {{ inputText.length }} characters
                </div>
              </div>

              <!-- Quick Actions -->
              <div class="flex flex-wrap gap-2">
                <button
                  v-for="template in quickTemplates"
                  :key="template.id"
                  @click="applyTemplate(template)"
                  class="px-3 py-1 text-xs bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-full hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
                >
                  {{ template.name }}
                </button>
              </div>
            </div>
          </div>

          <!-- Generation Options -->
          <div class="bg-white dark:bg-gray-800 p-6 rounded-lg shadow-md">
            <h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-4">Generation Options</h2>

            <div class="grid md:grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Programming Language
                </label>
                <select
                  v-model="selectedLanguage"
                  class="w-full p-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                >
                  <option value="javascript">JavaScript</option>
                  <option value="typescript">TypeScript</option>
                  <option value="python">Python</option>
                  <option value="java">Java</option>
                  <option value="csharp">C#</option>
                  <option value="go">Go</option>
                  <option value="rust">Rust</option>
                  <option value="cpp">C++</option>
                </select>
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Framework/Library
                </label>
                <select
                  v-model="selectedFramework"
                  class="w-full p-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                >
                  <option value="none">None</option>
                  <option value="react">React</option>
                  <option value="vue">Vue.js</option>
                  <option value="angular">Angular</option>
                  <option value="express">Express.js</option>
                  <option value="django">Django</option>
                  <option value="spring">Spring Boot</option>
                  <option value="dotnet">.NET</option>
                </select>
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Complexity Level
                </label>
                <select
                  v-model="complexityLevel"
                  class="w-full p-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                >
                  <option value="simple">Simple</option>
                  <option value="moderate">Moderate</option>
                  <option value="complex">Complex</option>
                </select>
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                  Include Tests
                </label>
                <div class="flex items-center space-x-2">
                  <input
                    v-model="includeTests"
                    type="checkbox"
                    class="rounded border-gray-300 dark:border-gray-600"
                  />
                  <span class="text-sm text-gray-600 dark:text-gray-300">Generate unit tests</span>
                </div>
              </div>
            </div>
          </div>

          <!-- AI Model Selection -->
          <div class="bg-white dark:bg-gray-800 p-6 rounded-lg shadow-md">
            <div class="flex items-center justify-between mb-4">
              <h2 class="text-xl font-semibold text-gray-900 dark:text-white">AI Models</h2>
              <button
                @click="fetchModels"
                :disabled="loadingModels"
                class="text-sm text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300"
              >
                {{ loadingModels ? 'Loading...' : 'Refresh' }}
              </button>
            </div>

            <div v-if="loadingModels" class="text-center py-4">
              <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
            </div>

            <div v-else-if="modelsError" class="text-red-600 dark:text-red-400 text-sm">
              {{ modelsError }}
            </div>

            <div v-else class="space-y-4">
              <!-- Model Strategy -->
              <div>
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Model Selection Strategy
                </label>
                <select
                  v-model="modelStrategy"
                  @change="onStrategyChange"
                  class="w-full p-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm"
                >
                  <option value="recommended">Recommended (Auto-select)</option>
                  <option value="manual">Manual Selection</option>
                  <option value="all">Use All Models</option>
                </select>
              </div>

              <!-- Recommended Models -->
              <div v-if="modelStrategy === 'recommended' && recommendedModels.length > 0">
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Recommended for {{ selectedLanguage }}
                </label>
                <div class="space-y-2">
                  <div
                    v-for="model in recommendedModels"
                    :key="model.id"
                    class="flex items-center justify-between p-3 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-md"
                  >
                    <div class="flex items-center space-x-3">
                      <div class="w-8 h-8 rounded-full flex items-center justify-center"
                           :class="getProviderColor(model.provider)">
                        <span class="text-xs font-bold text-white">{{ model.provider.slice(0, 2).toUpperCase() }}</span>
                      </div>
                      <div>
                        <p class="text-sm font-medium text-gray-900 dark:text-white">{{ model.name }}</p>
                        <p class="text-xs text-gray-600 dark:text-gray-400">{{ model.id }}</p>
                      </div>
                    </div>
                    <div class="flex items-center space-x-2">
                      <span class="px-2 py-1 text-xs bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 rounded">
                        Priority {{ model.priority }}
                      </span>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Manual Model Selection -->
              <div v-if="modelStrategy === 'manual'">
                <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Select Models (up to 3)
                </label>
                <div class="space-y-2 max-h-60 overflow-y-auto">
                  <div
                    v-for="model in availableModels"
                    :key="model.id"
                    @click="toggleModelSelection(model.id)"
                    class="flex items-center justify-between p-3 border rounded-md cursor-pointer transition-colors"
                    :class="selectedModels.includes(model.id)
                      ? 'bg-blue-50 dark:bg-blue-900/20 border-blue-500 dark:border-blue-400'
                      : 'bg-gray-50 dark:bg-gray-700 border-gray-200 dark:border-gray-600 hover:border-gray-300 dark:hover:border-gray-500'"
                  >
                    <div class="flex items-center space-x-3">
                      <input
                        type="checkbox"
                        :checked="selectedModels.includes(model.id)"
                        class="rounded border-gray-300 dark:border-gray-600"
                        @click.stop
                      />
                      <div class="w-8 h-8 rounded-full flex items-center justify-center"
                           :class="getProviderColor(model.provider)">
                        <span class="text-xs font-bold text-white">{{ model.provider.slice(0, 2).toUpperCase() }}</span>
                      </div>
                      <div>
                        <p class="text-sm font-medium text-gray-900 dark:text-white">{{ model.name }}</p>
                        <p class="text-xs text-gray-600 dark:text-gray-400">
                          {{ model.provider }} • Context: {{ formatNumber(model.capabilities.context_window) }}
                        </p>
                      </div>
                    </div>
                    <span class="text-xs text-gray-500 dark:text-gray-400">
                      {{ model.capabilities.max_complexity }}
                    </span>
                  </div>
                </div>
                <p class="text-xs text-gray-500 dark:text-gray-400 mt-2">
                  {{ selectedModels.length }} of 3 models selected
                </p>
              </div>

              <!-- All Models Info -->
              <div v-if="modelStrategy === 'all'" class="p-3 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-md">
                <p class="text-sm text-gray-700 dark:text-gray-300">
                  All {{ availableModels.length }} models will be used for generation.
                  Results will be compared for best output.
                </p>
              </div>

              <!-- Model Stats -->
              <div class="grid grid-cols-3 gap-2 pt-2 border-t border-gray-200 dark:border-gray-700">
                <div class="text-center">
                  <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ availableModels.length }}</p>
                  <p class="text-xs text-gray-600 dark:text-gray-400">Total Models</p>
                </div>
                <div class="text-center">
                  <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ uniqueProviders.length }}</p>
                  <p class="text-xs text-gray-600 dark:text-gray-400">Providers</p>
                </div>
                <div class="text-center">
                  <p class="text-2xl font-bold text-gray-900 dark:text-white">
                    {{ modelStrategy === 'manual' ? selectedModels.length : 
                       modelStrategy === 'recommended' ? recommendedModels.length : 
                       availableModels.length }}
                  </p>
                  <p class="text-xs text-gray-600 dark:text-gray-400">Active</p>
                </div>
              </div>
            </div>
          </div>

          <button
            @click="generateCode"
            :disabled="!canGenerate || isGenerating"
            class="w-full bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white py-3 px-6 rounded-lg font-medium transition-colors flex items-center justify-center space-x-2"
          >
            <Sparkles class="w-5 h-5" v-if="!isGenerating" />
            <div class="animate-spin rounded-full h-5 w-5 border-b-2 border-white" v-else></div>
            <span>{{ isGenerating ? 'Generating...' : 'Generate Code' }}</span>
          </button>
        </div>

        <!-- Output Section -->
        <div class="space-y-6">
          <div class="bg-white dark:bg-gray-800 p-6 rounded-lg shadow-md">
            <h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-4">Generated Code</h2>

            <div v-if="isGenerating" class="text-center py-8">
              <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
              <p class="text-gray-600 dark:text-gray-300">Analyzing input and generating code...</p>
              <div class="mt-4 text-sm text-gray-500 dark:text-gray-400">
                This may take a few moments for large files
              </div>
            </div>

            <div v-else-if="generatedCode" class="space-y-4">
              <div class="flex items-center justify-between">
                <div class="flex items-center space-x-2">
                  <div class="w-2 h-2 bg-green-500 rounded-full"></div>
                  <span class="text-sm text-gray-600 dark:text-gray-300">
                    Generated {{ formatDate(new Date()) }}
                  </span>
                </div>
                <div class="flex space-x-2">
                  <button
                    @click="copyToClipboard"
                    class="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 rounded"
                    title="Copy to clipboard"
                  >
                    <Copy class="w-4 h-4" />
                  </button>
                  <button
                    @click="downloadCode"
                    class="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 rounded"
                    title="Download"
                  >
                    <Download class="w-4 h-4" />
                  </button>
                </div>
              </div>

              <div class="relative">
                <pre class="bg-gray-50 dark:bg-gray-900 p-4 rounded-md overflow-x-auto text-sm font-mono"><code class="language-{{ selectedLanguage }}">{{ generatedCode }}</code></pre>
              </div>

              <div class="flex items-center justify-between text-sm text-gray-500 dark:text-gray-400">
                <span>{{ generatedCode.split('\n').length }} lines</span>
                <span>{{ selectedLanguage }} • {{ selectedFramework }}</span>
              </div>
            </div>

            <div v-else class="text-center py-12">
              <Code class="w-16 h-16 text-gray-400 mx-auto mb-4" />
              <h3 class="text-lg font-medium text-gray-900 dark:text-white mb-2">Ready to Generate</h3>
              <p class="text-gray-500 dark:text-gray-400 text-sm">
                Upload files or enter a prompt to get started
              </p>
            </div>
          </div>

          <!-- Generation History -->
          <div v-if="generationHistory.length > 0" class="bg-white dark:bg-gray-800 p-6 rounded-lg shadow-md">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">Recent Generations</h3>
            <div class="space-y-3">
              <div
                v-for="item in generationHistory.slice(0, 3)"
                :key="item.id"
                class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-700 rounded-lg cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-600 transition-colors"
                @click="loadGeneration(item)"
              >
                <div>
                  <p class="text-sm font-medium text-gray-900 dark:text-white">{{ item.language }} {{ item.framework }}</p>
                  <p class="text-xs text-gray-500 dark:text-gray-400">{{ formatDate(item.timestamp) }}</p>
                </div>
                <ChevronRight class="w-4 h-4 text-gray-400" />
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Templates Modal -->
      <div
        v-if="showTemplates"
        class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
        @click="showTemplates = false"
      >
        <div
          class="bg-white dark:bg-gray-800 p-6 rounded-lg shadow-xl max-w-2xl w-full mx-4 max-h-96 overflow-y-auto"
          @click.stop
        >
          <h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-4">Quick Templates</h2>
          <div class="grid md:grid-cols-2 gap-4">
            <button
              v-for="template in templates"
              :key="template.id"
              @click="selectTemplate(template)"
              class="p-4 border border-gray-200 dark:border-gray-600 rounded-lg hover:border-blue-500 dark:hover:border-blue-400 transition-colors text-left"
            >
              <h3 class="font-medium text-gray-900 dark:text-white mb-1">{{ template.name }}</h3>
              <p class="text-sm text-gray-600 dark:text-gray-300">{{ template.description }}</p>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import {
  Upload,
  FileText,
  Code,
  Image,
  Sparkles,
  Copy,
  Download,
  Eye,
  X,
  ChevronRight
} from 'lucide-vue-next'

// Types
interface UploadedFile {
  name: string
  size: number
  type: string
  file: File
  lastModified?: Date
  id?: string
}

interface RecentFile {
  id: string
  name: string
  type: string
  lastModified: Date
}

interface GenerationHistory {
  id: string
  language: string
  framework: string
  timestamp: Date
  code: string
}

interface Template {
  id: string
  name: string
  description: string
  prompt: string
}

// Reactive data
const activeTab = ref('upload')
const uploadedFiles = ref<UploadedFile[]>([])
const inputText = ref('')
const selectedLanguage = ref('typescript')
const selectedFramework = ref('react')
const complexityLevel = ref('moderate')
const includeTests = ref(false)
const generatedCode = ref('')
const isGenerating = ref(false)
const isUploading = ref(false)
const uploadProgress = ref(0)
const showTemplates = ref(false)
const isDragOver = ref(false)

// Model selection
const loadingModels = ref(false)
const modelsError = ref('')
const availableModels = ref<any[]>([])
const recommendedModels = ref<any[]>([])
const selectedModels = ref<string[]>([])
const modelStrategy = ref('recommended') // 'recommended', 'manual', 'all'

const fileInput = ref<HTMLInputElement>()

// Recent files (mock data - in real app, this would come from localStorage or API)
const recentFiles = ref<RecentFile[]>([
  { id: '1', name: 'user-api.js', type: 'js', lastModified: new Date(Date.now() - 86400000) },
  { id: '2', name: 'component.vue', type: 'vue', lastModified: new Date(Date.now() - 172800000) },
  { id: '3', name: 'requirements.txt', type: 'txt', lastModified: new Date(Date.now() - 259200000) },
])

// Generation history
const generationHistory = ref<GenerationHistory[]>([
  {
    id: '1',
    language: 'TypeScript',
    framework: 'React',
    timestamp: new Date(Date.now() - 3600000),
    code: '// Previous generation...'
  }
])

// Templates
const templates = ref<Template[]>([
  {
    id: '1',
    name: 'React Component',
    description: 'Create a reusable React component with TypeScript',
    prompt: 'Create a React component for a user profile card with TypeScript, including props validation and proper styling.'
  },
  {
    id: '2',
    name: 'API Endpoint',
    description: 'Generate a REST API endpoint with validation',
    prompt: 'Create an Express.js API endpoint for user management with input validation, error handling, and proper HTTP status codes.'
  },
  {
    id: '3',
    name: 'Database Model',
    description: 'Generate a database schema and model',
    prompt: 'Create a Sequelize model for a blog post with relationships, validations, and TypeScript types.'
  },
  {
    id: '4',
    name: 'Unit Tests',
    description: 'Generate comprehensive unit tests',
    prompt: 'Write Jest unit tests for a user authentication service including edge cases and mocking.'
  }
])

const quickTemplates = ref([
  { id: '1', name: 'CRUD API', prompt: 'Create a REST API with CRUD operations' },
  { id: '2', name: 'Login Form', prompt: 'Build a login form with validation' },
  { id: '3', name: 'Data Table', prompt: 'Create a sortable data table component' },
  { id: '4', name: 'File Upload', prompt: 'Implement a file upload component' }
])

// Computed properties
const tabs = computed(() => [
  { id: 'upload', name: 'Upload Files', icon: Upload },
  { id: 'text', name: 'Text Input', icon: FileText }
])

const canGenerate = computed(() => {
  if (activeTab.value === 'upload') {
    return uploadedFiles.value.length > 0
  } else {
    return inputText.value.trim().length > 0
  }
})

const uniqueProviders = computed(() => {
  return [...new Set(availableModels.value.map(m => m.provider))]
})

// Methods
const fetchModels = async () => {
  loadingModels.value = true
  modelsError.value = ''
  
  try {
    const response = await fetch('http://localhost:8000/api/v1/models')
    if (!response.ok) throw new Error('Failed to fetch models')
    
    const data = await response.json()
    availableModels.value = data.models || []
    
    // Fetch recommended models
    await fetchRecommendedModels()
  } catch (error) {
    console.error('Error fetching models:', error)
    modelsError.value = 'Failed to load models. Please try again.'
  } finally {
    loadingModels.value = false
  }
}

const fetchRecommendedModels = async () => {
  try {
    const response = await fetch(
      `http://localhost:8000/api/v1/models/recommend?language=${selectedLanguage.value}&complexity=${complexityLevel.value}&max_models=3`,
      { method: 'POST' }
    )
    if (!response.ok) throw new Error('Failed to fetch recommended models')
    
    const data = await response.json()
    recommendedModels.value = data.models || []
  } catch (error) {
    console.error('Error fetching recommended models:', error)
  }
}

const toggleModelSelection = (modelId: string) => {
  const index = selectedModels.value.indexOf(modelId)
  if (index > -1) {
    selectedModels.value.splice(index, 1)
  } else {
    if (selectedModels.value.length < 3) {
      selectedModels.value.push(modelId)
    }
  }
}

const onStrategyChange = () => {
  if (modelStrategy.value === 'manual') {
    selectedModels.value = []
  }
}

const getProviderColor = (provider: string): string => {
  const colors: { [key: string]: string } = {
    'openai': 'bg-green-500',
    'anthropic': 'bg-purple-500',
    'google': 'bg-blue-500',
    'cohere': 'bg-orange-500',
  }
  return colors[provider] || 'bg-gray-500'
}

const formatNumber = (num: number): string => {
  if (num >= 1000) {
    return (num / 1000).toFixed(0) + 'K'
  }
  return num.toString()
}

// Methods
const onDrop = (event: DragEvent) => {
  isDragOver.value = false
  const files = event.dataTransfer?.files
  if (files) {
    handleFiles(files)
  }
}

const onDragOver = () => {
  isDragOver.value = true
}

const onDragLeave = () => {
  isDragOver.value = false
}

const onFileSelect = (event: Event) => {
  const target = event.target as HTMLInputElement
  const files = target.files
  if (files) {
    handleFiles(files)
  }
}

const handleFiles = async (files: FileList) => {
  isUploading.value = true
  uploadProgress.value = 0

  for (let i = 0; i < files.length; i++) {
    const file = files[i]

    // Simulate upload progress
    for (let progress = 0; progress <= 100; progress += 10) {
      await new Promise(resolve => setTimeout(resolve, 50))
      uploadProgress.value = progress
    }

    const uploadedFile: UploadedFile = {
      name: file.name,
      size: file.size,
      type: getFileType(file),
      file,
      lastModified: new Date(file.lastModified),
      id: Date.now().toString() + i
    }

    uploadedFiles.value.push(uploadedFile)
  }

  isUploading.value = false
  uploadProgress.value = 0
}

const getFileType = (file: File): string => {
  const extension = file.name.split('.').pop()?.toLowerCase() || ''
  const typeMap: { [key: string]: string } = {
    'pdf': 'pdf',
    'txt': 'txt',
    'json': 'json',
    'yaml': 'yaml',
    'yml': 'yaml',
    'js': 'js',
    'ts': 'ts',
    'py': 'py',
    'java': 'java',
    'cpp': 'cpp',
    'c': 'c',
    'cs': 'cs',
    'php': 'php',
    'rb': 'rb',
    'go': 'go',
    'rs': 'rs',
    'png': 'png',
    'jpg': 'jpg',
    'jpeg': 'jpeg',
    'gif': 'gif',
    'svg': 'svg'
  }
  return typeMap[extension] || 'file'
}

const getFileIcon = (type: string) => {
  const iconMap: { [key: string]: any } = {
    'pdf': FileText,
    'txt': FileText,
    'json': Code,
    'yaml': Code,
    'js': Code,
    'ts': Code,
    'py': Code,
    'java': Code,
    'cpp': Code,
    'c': Code,
    'cs': Code,
    'php': Code,
    'rb': Code,
    'go': Code,
    'rs': Code,
    'png': Image,
    'jpg': Image,
    'jpeg': Image,
    'gif': Image,
    'svg': Image
  }
  return iconMap[type] || FileText
}

const removeFile = (file: UploadedFile) => {
  const index = uploadedFiles.value.indexOf(file)
  if (index > -1) {
    uploadedFiles.value.splice(index, 1)
  }
}

const selectRecentFile = (file: RecentFile) => {
  // In a real app, this would load the file content
  console.log('Selected recent file:', file.name)
}

const previewFile = (file: UploadedFile) => {
  // In a real app, this would open a file preview modal
  console.log('Previewing file:', file.name)
}

const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatDate = (date: Date): string => {
  return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

const applyTemplate = (template: { id: string, name: string, prompt: string }) => {
  inputText.value = template.prompt
  activeTab.value = 'text'
}

const selectTemplate = (template: Template) => {
  inputText.value = template.prompt
  activeTab.value = 'text'
  showTemplates.value = false
}

const generateCode = async () => {
  isGenerating.value = true

  try {
    // Simulate API call with more realistic delay
    await new Promise(resolve => setTimeout(resolve, 3000))

    // Mock generated code based on inputs
    let code = ''

    if (selectedLanguage.value === 'typescript' && selectedFramework.value === 'react') {
      code = `import React, { useState, useEffect } from 'react'

interface UserProfile {
  id: number
  name: string
  email: string
  avatar?: string
}

interface UserProfileCardProps {
  userId: number
  onUpdate?: (user: UserProfile) => void
}

const UserProfileCard: React.FC<UserProfileCardProps> = ({ userId, onUpdate }) => {
  const [user, setUser] = useState<UserProfile | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    fetchUserProfile()
  }, [userId])

  const fetchUserProfile = async () => {
    try {
      setLoading(true)
      // Simulated API call
      const response = await fetch('/api/users/' + userId)
      if (!response.ok) throw new Error('Failed to fetch user')
      const userData = await response.json()
      setUser(userData)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred')
    } finally {
      setLoading(false)
    }
  }

  const handleUpdate = (updates: Partial<UserProfile>) => {
    if (user) {
      const updatedUser = { ...user, ...updates }
      setUser(updatedUser)
      onUpdate?.(updatedUser)
    }
  }

  if (loading) return <div className="animate-pulse">Loading...</div>
  if (error) return <div className="text-red-500">Error: {error}</div>
  if (!user) return <div>User not found</div>

  return (
    <div className="bg-white rounded-lg shadow-md p-6 max-w-sm">
      <div className="flex items-center space-x-4">
        <img
          src={user.avatar || '/default-avatar.png'}
          alt={user.name}
          className="w-16 h-16 rounded-full"
        />
        <div>
          <h2 className="text-xl font-semibold text-gray-900">{user.name}</h2>
          <p className="text-gray-600">{user.email}</p>
        </div>
      </div>
      <button
        onClick={() => handleUpdate({ name: 'Updated Name' })}
        className="mt-4 bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
      >
        Update Profile
      </button>
    </div>
  )
}

export default UserProfileCard`
    } else if (selectedLanguage.value === 'python') {
      code = 'from typing import Optional, Dict, Any\\n' +
'from dataclasses import dataclass\\n' +
'from datetime import datetime\\n' +
'import requests\\n\\n' +
'@dataclass\\n' +
'class UserProfile:\\n' +
'    id: int\\n' +
'    name: str\\n' +
'    email: str\\n' +
'    avatar: Optional[str] = None\\n' +
'    created_at: Optional[datetime] = None\\n\\n' +
'class UserProfileService:\\n' +
'    def __init__(self, base_url: str = "http://localhost:8000"):\\n' +
'        self.base_url = base_url\\n\\n' +
'    def get_user_profile(self, user_id: int) -> Optional[UserProfile]:\\n' +
'        """Fetch user profile from API"""\\n' +
'        try:\\n' +
'            response = requests.get(f"{self.base_url}/api/users/{user_id}")\\n' +
'            response.raise_for_status()\\n' +
'            data = response.json()\\n\\n' +
'            return UserProfile(\\n' +
'                id=data["id"],\\n' +
'                name=data["name"],\\n' +
'                email=data["email"],\\n' +
'                avatar=data.get("avatar"),\\n' +
'                created_at=datetime.fromisoformat(data["created_at"]) if data.get("created_at") else None\\n' +
'            )\\n' +
'        except requests.RequestException as e:\\n' +
'            print(f"Error fetching user profile: {e}")\\n' +
'            return None\\n\\n' +
'    def update_user_profile(self, user_id: int, updates: Dict[str, Any]) -> bool:\\n' +
'        """Update user profile"""\\n' +
'        try:\\n' +
'            response = requests.patch(\\n' +
'                f"{self.base_url}/api/users/{user_id}",\\n' +
'                json=updates\\n' +
'            )\\n' +
'            response.raise_for_status()\\n' +
'            return True\\n' +
'        except requests.RequestException as e:\\n' +
'            print(f"Error updating user profile: {e}")\\n' +
'            return False\\n\\n' +
'# Usage example\\n' +
'if __name__ == "__main__":\\n' +
'    service = UserProfileService()\\n' +
'    user = service.get_user_profile(1)\\n\\n' +
'    if user:\\n' +
'        print(f"User: {user.name} ({user.email})")\\n' +
'        # Update user\\n' +
'        success = service.update_user_profile(1, {"name": "Updated Name"})\\n' +
'        print(f"Update successful: {success}")'
    } else {
      code = `// Generated ${selectedLanguage.value} code using ${selectedFramework.value}
// This is a placeholder - in a real implementation, this would be generated based on your input

console.log('Hello, World!');
console.log('Generated with ${selectedLanguage.value} and ${selectedFramework.value}');
console.log('Complexity: ${complexityLevel.value}');
console.log('Include tests: ${includeTests.value}');`
    }

    generatedCode.value = code

    // Add to history
    generationHistory.value.unshift({
      id: Date.now().toString(),
      language: selectedLanguage.value,
      framework: selectedFramework.value,
      timestamp: new Date(),
      code: code
    })

  } catch (error) {
    console.error('Generation failed:', error)
  } finally {
    isGenerating.value = false
  }
}

const copyToClipboard = async () => {
  try {
    await navigator.clipboard.writeText(generatedCode.value)
    // In a real app, show a toast notification
    console.log('Code copied to clipboard')
  } catch (error) {
    console.error('Failed to copy:', error)
  }
}

const downloadCode = () => {
  const blob = new Blob([generatedCode.value], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `generated-code.${getFileExtension(selectedLanguage.value)}`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

const getFileExtension = (language: string): string => {
  const extensions: { [key: string]: string } = {
    'javascript': 'js',
    'typescript': 'ts',
    'python': 'py',
    'java': 'java',
    'csharp': 'cs',
    'go': 'go',
    'rust': 'rs',
    'cpp': 'cpp'
  }
  return extensions[language] || 'txt'
}

const loadGeneration = (item: GenerationHistory) => {
  generatedCode.value = item.code
  selectedLanguage.value = item.language.toLowerCase()
  selectedFramework.value = item.framework.toLowerCase()
}

// Initialize
onMounted(() => {
  // Load models on mount
  fetchModels()
  // Load recent files from localStorage in a real app
  console.log('GenerateView mounted')
})

// Watch for language or complexity changes to update recommendations
watch([selectedLanguage, complexityLevel], () => {
  if (modelStrategy.value === 'recommended') {
    fetchRecommendedModels()
  }
})
</script>