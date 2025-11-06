
<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-900 py-8">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <h1 class="text-3xl font-bold text-gray-900 dark:text-white mb-8">Settings & Dashboard</h1>
      
      <!-- Tab Navigation -->
      <div class="flex space-x-1 mb-6 border-b border-gray-200 dark:border-gray-700">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          @click="activeTab = tab.id"
          :class="[
            'px-4 py-2 text-sm font-medium transition-colors border-b-2',
            activeTab === tab.id
              ? 'border-blue-600 text-blue-600 dark:text-blue-400'
              : 'border-transparent text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200'
          ]"
        >
          {{ tab.name }}
        </button>
      </div>

      <!-- User Settings Tab -->
      <div v-if="activeTab === 'settings'" class="grid lg:grid-cols-2 gap-6">
        <div class="bg-white dark:bg-gray-800 rounded-lg shadow-md border border-gray-200 dark:border-gray-700 p-6">
          <h2 class="text-xl font-semibold mb-4 text-gray-900 dark:text-white">User Settings</h2>
          <div v-if="userStore.user" class="space-y-4">
            <div>
              <label class="block text-sm font-medium mb-1 text-gray-700 dark:text-gray-300">Name</label>
              <input v-model="name" class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-transparent" />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1 text-gray-700 dark:text-gray-300">Avatar Seed</label>
              <input v-model="avatarSeed" class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-transparent" />
              <div class="mt-2">
                <img :src="avatarUrl" alt="Avatar" class="w-16 h-16 rounded-full border-2 border-gray-300 dark:border-gray-600" />
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium mb-1 text-gray-700 dark:text-gray-300">Theme</label>
              <select v-model="theme" class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:ring-2 focus:ring-blue-500 focus:border-transparent">
                <option value="light">Light</option>
                <option value="dark">Dark</option>
              </select>
            </div>
            <button @click="saveSettings" class="w-full px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 border border-blue-700 transition-colors">
              Save Settings
            </button>
          </div>
        </div>

        <div class="bg-white dark:bg-gray-800 rounded-lg shadow-md border border-gray-200 dark:border-gray-700 p-6">
          <h2 class="text-xl font-semibold mb-4 text-gray-900 dark:text-white">API Configuration</h2>
          <div class="space-y-3 text-sm">
            <div class="flex justify-between py-2 border-b border-gray-200 dark:border-gray-700">
              <span class="text-gray-600 dark:text-gray-400">API URL:</span>
              <span class="font-mono text-gray-900 dark:text-white">{{ apiUrl }}</span>
            </div>
            <div class="flex justify-between py-2 border-b border-gray-200 dark:border-gray-700">
              <span class="text-gray-600 dark:text-gray-400">Environment:</span>
              <span class="font-mono text-gray-900 dark:text-white">{{ environment }}</span>
            </div>
            <div class="flex justify-between py-2 border-b border-gray-200 dark:border-gray-700">
              <span class="text-gray-600 dark:text-gray-400">Gateway Status:</span>
              <span class="flex items-center">
                <span class="w-2 h-2 bg-green-500 rounded-full mr-2 border border-green-600"></span>
                <span class="text-green-600 dark:text-green-400">Connected</span>
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- Usage Tab -->
      <div v-if="activeTab === 'usage'" class="space-y-6">
        <div class="flex justify-between items-center mb-4">
          <div>
            <h2 class="text-2xl font-bold text-gray-900 dark:text-white">Usage & Quota</h2>
            <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">User ID: {{ userStore.userId }}</p>
          </div>
          <button 
            @click="loadUsageData" 
            :disabled="usageStore.loading"
            class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 border border-blue-700 transition-colors"
          >
            {{ usageStore.loading ? 'Loading...' : 'Refresh' }}
          </button>
        </div>

        <!-- Loading State with Modern Loader -->
        <LoadingSpinner 
          v-if="usageStore.loading" 
          type="spinner" 
          :size="48" 
          color="blue" 
          text="Loading usage data..."
        />

        <!-- Empty State -->
        <div v-else-if="!usageStore.loading && !usageStore.quota && !usageStore.summary" class="bg-white dark:bg-gray-800 rounded-lg shadow-md border border-gray-200 dark:border-gray-700 p-12 text-center">
          <svg xmlns="http://www.w3.org/2000/svg" class="mx-auto h-16 w-16 text-gray-400 dark:text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
          </svg>
          <h3 class="mt-4 text-lg font-medium text-gray-900 dark:text-white">No Usage Data Available</h3>
          <p class="mt-2 text-gray-500 dark:text-gray-400">Start using the application to see your usage statistics.</p>
          <button 
            @click="loadUsageData" 
            class="mt-4 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 border border-blue-700 transition-colors"
          >
            Load Data
          </button>
        </div>

        <!-- Quota Status -->
        <div v-else-if="usageStore.quota" class="bg-white dark:bg-gray-800 rounded-lg shadow-md border border-gray-200 dark:border-gray-700 p-6">
          <h2 class="text-xl font-semibold mb-4 text-gray-900 dark:text-white">Usage Quota</h2>
          
          <!-- Quota Visual Chart -->
          <div class="mb-6 border border-gray-200 dark:border-gray-700 rounded-lg p-4">
            <div class="flex justify-center">
              <PieChart
                chart-id="quota-chart"
                :data="quotaChartData"
              />
            </div>
          </div>

          <div class="space-y-4">
            <div class="flex justify-between items-center p-4 bg-gray-50 dark:bg-gray-700/30 rounded-lg border border-gray-200 dark:border-gray-700">
              <div>
                <p class="text-sm text-gray-600 dark:text-gray-400">Current Tier</p>
                <p class="text-2xl font-bold text-gray-900 dark:text-white capitalize">{{ usageStore.quota.tier }}</p>
              </div>
              <div class="text-right">
                <p class="text-sm text-gray-600 dark:text-gray-400">Monthly Limit</p>
                <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ usageStore.quota.monthly_limit.toLocaleString() }}</p>
              </div>
            </div>

            <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
              <div class="flex justify-between text-sm mb-2">
                <span class="text-gray-600 dark:text-gray-400">Usage Progress</span>
                <span class="font-medium text-gray-900 dark:text-white">
                  {{ usageStore.quota.current_usage.toLocaleString() }} / {{ usageStore.quota.monthly_limit.toLocaleString() }} tokens
                </span>
              </div>
              <div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-3 border border-gray-300 dark:border-gray-600">
                <div 
                  class="h-3 rounded-full transition-all"
                  :class="getQuotaColor(usageStore.quota.percentage_used)"
                  :style="{ width: Math.min(usageStore.quota.percentage_used, 100) + '%' }"
                ></div>
              </div>
              <div class="flex justify-between mt-2 text-xs text-gray-500 dark:text-gray-400">
                <span>{{ usageStore.quota.remaining.toLocaleString() }} tokens remaining</span>
                <span>{{ usageStore.quota.percentage_used.toFixed(1) }}% used</span>
              </div>
            </div>

            <div class="pt-4 border-t border-gray-200 dark:border-gray-700">
              <p class="text-sm text-gray-600 dark:text-gray-400">
                Quota resets on: <span class="font-medium text-gray-900 dark:text-white">{{ formatDate(usageStore.quota.reset_date) }}</span>
              </p>
            </div>
          </div>
        </div>

        <!-- Usage Summary -->
        <div v-if="!usageStore.loading && usageStore.summary" class="bg-white dark:bg-gray-800 rounded-lg shadow-md border border-gray-200 dark:border-gray-700 p-6">
          <div class="flex justify-between items-center mb-4">
            <h2 class="text-xl font-semibold text-gray-900 dark:text-white">Usage Summary</h2>
            <select 
              v-model="usagePeriod" 
              @change="loadUsageSummary"
              class="px-3 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
            >
              <option value="daily">Daily</option>
              <option value="monthly">Monthly</option>
              <option value="all_time">All Time</option>
            </select>
          </div>

          <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
            <div class="bg-blue-50 dark:bg-blue-900/20 p-4 rounded-lg border border-blue-200 dark:border-blue-800 text-center">
              <p class="text-3xl font-bold text-blue-600 dark:text-blue-400">{{ usageStore.summary.total_requests }}</p>
              <p class="text-xs text-gray-600 dark:text-gray-400 mt-1">Total Requests</p>
            </div>
            <div class="bg-green-50 dark:bg-green-900/20 p-4 rounded-lg border border-green-200 dark:border-green-800 text-center">
              <p class="text-3xl font-bold text-green-600 dark:text-green-400">{{ usageStore.summary.successful_requests }}</p>
              <p class="text-xs text-gray-600 dark:text-gray-400 mt-1">Successful</p>
            </div>
            <div class="bg-purple-50 dark:bg-purple-900/20 p-4 rounded-lg border border-purple-200 dark:border-purple-800 text-center">
              <p class="text-3xl font-bold text-purple-600 dark:text-purple-400">{{ formatTokens(usageStore.summary.total_tokens) }}</p>
              <p class="text-xs text-gray-600 dark:text-gray-400 mt-1">Tokens Used</p>
            </div>
            <div class="bg-orange-50 dark:bg-orange-900/20 p-4 rounded-lg border border-orange-200 dark:border-orange-800 text-center">
              <p class="text-3xl font-bold text-orange-600 dark:text-orange-400">${{ usageStore.summary.total_cost.toFixed(2) }}</p>
              <p class="text-xs text-gray-600 dark:text-gray-400 mt-1">Total Cost</p>
            </div>
          </div>

          <!-- Usage Charts -->
          <div v-if="Object.keys(usageStore.summary.by_model || {}).length > 0" class="grid md:grid-cols-2 gap-6 mb-6">
            <!-- Requests by Model Chart -->
            <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-4">
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4 text-center">Requests by Model</h3>
              <PieChart
                chart-id="requests-chart"
                :data="requestsChartData"
              />
            </div>

            <!-- Tokens by Model Chart -->
            <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-4">
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4 text-center">Tokens by Model</h3>
              <PieChart
                chart-id="tokens-chart"
                :data="tokensChartData"
              />
            </div>
          </div>

          <!-- Usage by Model Table -->
          <div v-if="Object.keys(usageStore.summary.by_model || {}).length > 0" class="mt-6 border border-gray-200 dark:border-gray-700 rounded-lg overflow-hidden">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-3 px-4 pt-4">Usage by Model</h3>
            <div class="space-y-0">
              <div 
                v-for="(data, model) in usageStore.summary.by_model" 
                :key="model"
                class="flex items-center justify-between p-4 border-t border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors"
              >
                <div class="flex-1">
                  <p class="font-medium text-gray-900 dark:text-white">{{ model }}</p>
                  <p class="text-xs text-gray-500 dark:text-gray-400">{{ data.requests }} requests</p>
                </div>
                <div class="text-right">
                  <p class="text-sm font-semibold text-gray-900 dark:text-white">{{ formatTokens(data.tokens) }} tokens</p>
                  <p class="text-xs text-gray-500 dark:text-gray-400">${{ data.cost.toFixed(2) }}</p>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Loading State -->
        <div v-if="usageStore.loading && !usageStore.quota" class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-12 flex justify-center">
          <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
        </div>

        <!-- Empty State -->
        <div v-if="!usageStore.loading && !usageStore.quota && !usageStore.error" class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-12 text-center">
          <p class="text-gray-500 dark:text-gray-400">No usage data available</p>
          <button @click="loadUsageData" class="mt-4 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700">
            Load Usage Data
          </button>
        </div>

        <!-- Error State -->
        <div v-if="usageStore.error" class="bg-red-50 dark:bg-red-900/20 rounded-lg p-4">
          <p class="text-red-600 dark:text-red-400">{{ usageStore.error }}</p>
        </div>
      </div>

      <!-- Models Tab -->
      <div v-if="activeTab === 'models'" class="space-y-6">
        <div class="flex justify-between items-center mb-4">
          <h2 class="text-2xl font-bold text-gray-900 dark:text-white">Models & Configuration</h2>
          <button 
            @click="loadModelData" 
            :disabled="systemStore.loading"
            class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 border border-blue-700 transition-colors"
          >
            {{ systemStore.loading ? 'Loading...' : 'Refresh' }}
          </button>
        </div>

        <!-- Loading State -->
        <LoadingSpinner 
          v-if="systemStore.loading" 
          type="dots" 
          :size="48" 
          color="purple" 
          text="Loading model data..."
        />

        <!-- Model Statistics -->
        <div v-else-if="!systemStore.loading && codegenStore.stats" class="bg-white dark:bg-gray-800 rounded-lg shadow-md border border-gray-200 dark:border-gray-700 p-6">
          <h2 class="text-xl font-semibold mb-4 text-gray-900 dark:text-white">Code Generation Statistics</h2>
          <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div class="text-center p-4 bg-blue-50 dark:bg-blue-900/20 rounded-lg border border-blue-200 dark:border-blue-800">
              <p class="text-3xl font-bold text-blue-600 dark:text-blue-400">{{ codegenStore.stats?.total_requests || 0 }}</p>
              <p class="text-xs text-gray-600 dark:text-gray-400 mt-1">Total Requests</p>
            </div>
            <div class="text-center p-4 bg-green-50 dark:bg-green-900/20 rounded-lg border border-green-200 dark:border-green-800">
              <p class="text-3xl font-bold text-green-600 dark:text-green-400">{{ ((codegenStore.stats?.avg_confidence || 0) * 100).toFixed(1) }}%</p>
              <p class="text-xs text-gray-600 dark:text-gray-400 mt-1">Avg Confidence</p>
            </div>
            <div class="text-center p-4 bg-purple-50 dark:bg-purple-900/20 rounded-lg border border-purple-200 dark:border-purple-800">
              <p class="text-3xl font-bold text-purple-600 dark:text-purple-400">{{ codegenStore.stats?.models_loaded || 0 }}</p>
              <p class="text-xs text-gray-600 dark:text-gray-400 mt-1">Models Loaded</p>
            </div>
            <div class="text-center p-4 bg-orange-50 dark:bg-orange-900/20 rounded-lg border border-orange-200 dark:border-orange-800">
              <p class="text-3xl font-bold text-orange-600 dark:text-orange-400">{{ codegenStore.stats?.rag_documents_count || 0 }}</p>
              <p class="text-xs text-gray-600 dark:text-gray-400 mt-1">RAG Documents</p>
            </div>
          </div>
        </div>

        <!-- Empty State for Stats -->
        <div v-else-if="!systemStore.loading && !codegenStore.stats" class="bg-white dark:bg-gray-800 rounded-lg shadow-md border border-gray-200 dark:border-gray-700 p-12 text-center">
          <svg xmlns="http://www.w3.org/2000/svg" class="mx-auto h-16 w-16 text-gray-400 dark:text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          <h3 class="mt-4 text-lg font-medium text-gray-900 dark:text-white">AI Service Unavailable</h3>
          <p class="mt-2 text-gray-500 dark:text-gray-400">The Python AI service is not running. Statistics will be available when the service starts.</p>
          <p class="mt-2 text-sm text-gray-400 dark:text-gray-500">Expected endpoint: http://localhost:8000</p>
        </div>

        <!-- Available Models -->
        <div v-if="!systemStore.loading" class="bg-white dark:bg-gray-800 rounded-lg shadow-md border border-gray-200 dark:border-gray-700 p-6">
          <h2 class="text-xl font-semibold mb-4 text-gray-900 dark:text-white">Available Models</h2>
          
          <!-- Empty State -->
          <div v-if="systemStore.models.length === 0" class="text-center py-8">
            <svg xmlns="http://www.w3.org/2000/svg" class="mx-auto h-12 w-12 text-gray-400 dark:text-gray-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
            </svg>
            <p class="mt-4 text-gray-500 dark:text-gray-400">No models found</p>
            <p class="text-sm text-gray-400 dark:text-gray-500 mt-2">Configure models in the backend API</p>
          </div>

          <!-- Models Grid -->
          <div v-else class="grid md:grid-cols-2 gap-4">
            <div 
              v-for="model in systemStore.models" 
              :key="model.name"
              class="border border-gray-200 dark:border-gray-700 rounded-lg p-4 hover:shadow-lg transition-shadow"
            >
              <div class="flex justify-between items-start mb-2">
                <h3 class="font-semibold text-lg text-gray-900 dark:text-white">{{ model.name }}</h3>
                <span class="px-2 py-1 text-xs bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-300 rounded">
                  {{ model.provider }}
                </span>
              </div>
              <p class="text-sm text-gray-600 dark:text-gray-400 mb-3">{{ model.description }}</p>
              <div class="grid grid-cols-2 gap-2 text-xs">
                <div class="bg-gray-50 dark:bg-gray-700/50 p-2 rounded">
                  <p class="text-gray-500 dark:text-gray-400">Context Length</p>
                  <p class="font-semibold text-gray-900 dark:text-white">{{ model.context_length.toLocaleString() }}</p>
                </div>
                <div class="bg-gray-50 dark:bg-gray-700/50 p-2 rounded">
                  <p class="text-gray-500 dark:text-gray-400">Pricing</p>
                  <p class="font-semibold text-gray-900 dark:text-white">${{ model.pricing.input }}/${{ model.pricing.output }}</p>
                </div>
              </div>
              <div class="mt-3 flex flex-wrap gap-1">
                <span 
                  v-for="cap in model.capabilities" 
                  :key="cap"
                  class="px-2 py-1 text-xs bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded"
                >
                  {{ cap }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- Loading State -->
        <div v-if="systemStore.loading && systemStore.models.length === 0" class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-12 flex justify-center">
          <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
        </div>
      </div>

      <!-- System Tab -->
      <div v-if="activeTab === 'system'" class="space-y-6">
        <div class="flex justify-between items-center mb-4">
          <h2 class="text-2xl font-bold text-gray-900 dark:text-white">System Metrics</h2>
          <button 
            @click="loadSystemData" 
            :disabled="systemStore.loading"
            class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 border border-blue-700 transition-colors"
          >
            {{ systemStore.loading ? 'Loading...' : 'Refresh' }}
          </button>
        </div>

        <!-- Loading State -->
        <LoadingSpinner 
          v-if="systemStore.loading" 
          type="pulse" 
          :size="48" 
          color="green" 
          text="Loading system metrics..."
        />

        <!-- System Metrics -->
        <div v-else-if="systemStore.metrics" class="bg-white dark:bg-gray-800 rounded-lg shadow-md border border-gray-200 dark:border-gray-700 p-6">
          <h2 class="text-xl font-semibold mb-4 text-gray-900 dark:text-white">System Metrics</h2>
          
          <div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
            <div class="bg-blue-50 dark:bg-blue-900/20 p-4 rounded-lg text-center">
              <p class="text-3xl font-bold text-blue-600 dark:text-blue-400">{{ systemStore.metrics.total_users }}</p>
              <p class="text-xs text-gray-600 dark:text-gray-400 mt-1">Total Users</p>
            </div>
            <div class="bg-green-50 dark:bg-green-900/20 p-4 rounded-lg text-center">
              <p class="text-3xl font-bold text-green-600 dark:text-green-400">{{ systemStore.metrics.active_users }}</p>
              <p class="text-xs text-gray-600 dark:text-gray-400 mt-1">Active Users</p>
            </div>
            <div class="bg-purple-50 dark:bg-purple-900/20 p-4 rounded-lg text-center">
              <p class="text-3xl font-bold text-purple-600 dark:text-purple-400">{{ systemStore.metrics.total_chats }}</p>
              <p class="text-xs text-gray-600 dark:text-gray-400 mt-1">Total Chats</p>
            </div>
            <div class="bg-orange-50 dark:bg-orange-900/20 p-4 rounded-lg text-center">
              <p class="text-3xl font-bold text-orange-600 dark:text-orange-400">{{ systemStore.metrics.total_documents }}</p>
              <p class="text-xs text-gray-600 dark:text-gray-400 mt-1">Documents</p>
            </div>
          </div>

          <div class="grid md:grid-cols-3 gap-4">
            <div class="bg-gray-50 dark:bg-gray-700/50 p-4 rounded-lg">
              <p class="text-sm text-gray-600 dark:text-gray-400">Total Requests</p>
              <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ systemStore.metrics.total_requests.toLocaleString() }}</p>
            </div>
            <div class="bg-gray-50 dark:bg-gray-700/50 p-4 rounded-lg">
              <p class="text-sm text-gray-600 dark:text-gray-400">Avg Latency</p>
              <p class="text-2xl font-bold text-gray-900 dark:text-white">{{ systemStore.metrics.avg_latency.toFixed(0) }}ms</p>
            </div>
            <div class="bg-gray-50 dark:bg-gray-700/50 p-4 rounded-lg">
              <p class="text-sm text-gray-600 dark:text-gray-400">Success Rate</p>
              <p class="text-2xl font-bold text-gray-900 dark:text-white">
                {{ ((systemStore.metrics.successful_requests / systemStore.metrics.total_requests) * 100).toFixed(1) }}%
              </p>
            </div>
          </div>

          <!-- Endpoint Performance -->
          <div v-if="systemStore.metrics.endpoints && systemStore.metrics.endpoints.length > 0" class="mt-6">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-3">Endpoint Performance</h3>
            <div class="overflow-x-auto">
              <table class="w-full text-sm">
                <thead class="text-xs text-gray-700 dark:text-gray-300 uppercase bg-gray-50 dark:bg-gray-700/50">
                  <tr>
                    <th class="px-4 py-3 text-left">Endpoint</th>
                    <th class="px-4 py-3 text-right">Requests</th>
                    <th class="px-4 py-3 text-right">Avg Time</th>
                    <th class="px-4 py-3 text-right">Error Rate</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-gray-200 dark:divide-gray-700">
                  <tr 
                    v-for="endpoint in systemStore.metrics.endpoints.slice(0, 10)" 
                    :key="endpoint.endpoint"
                    class="hover:bg-gray-50 dark:hover:bg-gray-700/50"
                  >
                    <td class="px-4 py-3 text-gray-900 dark:text-white font-mono text-xs">{{ endpoint.endpoint }}</td>
                    <td class="px-4 py-3 text-right text-gray-900 dark:text-white">{{ endpoint.request_count }}</td>
                    <td class="px-4 py-3 text-right text-gray-900 dark:text-white">{{ endpoint.avg_time.toFixed(0) }}ms</td>
                    <td class="px-4 py-3 text-right">
                      <span :class="endpoint.error_rate > 5 ? 'text-red-600 dark:text-red-400' : 'text-green-600 dark:text-green-400'">
                        {{ endpoint.error_rate.toFixed(1) }}%
                      </span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <!-- Loading State -->
        <div v-if="systemStore.loading && !systemStore.metrics" class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-12 flex justify-center">
          <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import { useUsageStore } from '@/stores/usage'
import { useSystemStore } from '@/stores/system'
import { useCodegenStore } from '@/stores/codegen'
import PieChart from '@/components/charts/PieChart.vue'
import LoadingSpinner from '@/components/LoadingSpinner.vue'

const userStore = useUserStore()
const usageStore = useUsageStore()
const systemStore = useSystemStore()
const codegenStore = useCodegenStore()

const name = ref(userStore.user?.name || '')
const avatarSeed = ref(userStore.user?.email || '')
const theme = ref(localStorage.getItem('theme') || 'light')
const activeTab = ref('settings')
const usagePeriod = ref<'daily' | 'monthly' | 'all_time'>('monthly')

const apiUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080'
const environment = import.meta.env.VITE_ENV || 'development'

const tabs = [
  { id: 'settings', name: 'Settings' },
  { id: 'usage', name: 'Usage & Quota' },
  { id: 'models', name: 'Models' },
  { id: 'system', name: 'System Metrics' }
]

const avatarUrl = computed(() =>
  `https://api.dicebear.com/8.x/avataaars/svg?seed=${encodeURIComponent(avatarSeed.value)}`
)

// Chart data for quota visualization
const quotaChartData = computed(() => {
  if (!usageStore.quota) return { labels: [], datasets: [] }
  
  const used = usageStore.quota.current_usage
  const remaining = usageStore.quota.remaining
  
  return {
    labels: ['Used Tokens', 'Remaining Tokens'],
    datasets: [{
      data: [used, remaining],
      backgroundColor: [
        used / usageStore.quota.monthly_limit >= 0.9 ? '#EF4444' :
        used / usageStore.quota.monthly_limit >= 0.7 ? '#F59E0B' :
        '#10B981',
        '#E5E7EB'
      ],
      borderColor: ['#1F2937', '#1F2937'],
      borderWidth: 2
    }]
  }
})

// Chart data for requests by model
const requestsChartData = computed(() => {
  if (!usageStore.summary?.by_model) return { labels: [], datasets: [] }
  
  const models = Object.keys(usageStore.summary.by_model)
  const requests = models.map(model => usageStore.summary.by_model[model].requests)
  
  const colors = [
    '#3B82F6', '#10B981', '#F59E0B', '#EF4444', 
    '#8B5CF6', '#EC4899', '#14B8A6', '#F97316'
  ]
  
  return {
    labels: models,
    datasets: [{
      data: requests,
      backgroundColor: colors.slice(0, models.length),
      borderColor: Array(models.length).fill('#1F2937'),
      borderWidth: 2
    }]
  }
})

// Chart data for tokens by model
const tokensChartData = computed(() => {
  if (!usageStore.summary?.by_model) return { labels: [], datasets: [] }
  
  const models = Object.keys(usageStore.summary.by_model)
  const tokens = models.map(model => usageStore.summary.by_model[model].tokens)
  
  const colors = [
    '#8B5CF6', '#EC4899', '#F59E0B', '#10B981', 
    '#3B82F6', '#EF4444', '#14B8A6', '#F97316'
  ]
  
  return {
    labels: models,
    datasets: [{
      data: tokens,
      backgroundColor: colors.slice(0, models.length),
      borderColor: Array(models.length).fill('#1F2937'),
      borderWidth: 2
    }]
  }
})

watch(theme, (val) => {
  if (val === 'dark') {
    document.documentElement.classList.add('dark')
  } else {
    document.documentElement.classList.remove('dark')
  }
  localStorage.setItem('theme', val)
})

const saveSettings = () => {
  userStore.updateProfile(name.value, avatarSeed.value)
}

const loadUsageData = async () => {
  try {
    await Promise.allSettled([
      usageStore.fetchQuotaStatus(),
      usageStore.fetchUsageSummary(usagePeriod.value)
    ])
  } catch (err) {
    console.error('Error loading usage data:', err)
  }
}

const loadUsageSummary = async () => {
  try {
    await usageStore.fetchUsageSummary(usagePeriod.value)
  } catch (err) {
    console.error('Error loading usage summary:', err)
  }
}

const loadModelData = async () => {
  try {
    await Promise.allSettled([
      systemStore.fetchAvailableModels(),
      codegenStore.fetchStats()
    ])
  } catch (err) {
    console.error('Error loading model data:', err)
  }
}

const loadSystemData = async () => {
  try {
    await systemStore.fetchMetrics()
  } catch (err) {
    console.error('Error loading system data:', err)
  }
}

const getQuotaColor = (percentage: number) => {
  if (percentage >= 90) return 'bg-red-500'
  if (percentage >= 70) return 'bg-yellow-500'
  return 'bg-green-500'
}

const formatTokens = (tokens: number) => {
  if (tokens >= 1000000) return `${(tokens / 1000000).toFixed(1)}M`
  if (tokens >= 1000) return `${(tokens / 1000).toFixed(1)}K`
  return tokens.toString()
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  })
}

onMounted(() => {
  // Load initial data
  loadUsageData()
  loadModelData()
  loadSystemData()
})
</script>
