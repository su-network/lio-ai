
<template>
  <div class="p-4 max-w-lg mx-auto">
    <h1 class="text-2xl font-bold mb-6">Settings</h1>
    <div v-if="userStore.user" class="space-y-6">
      <div>
        <label class="block text-sm font-medium mb-1">Name</label>
        <input v-model="name" class="w-full px-3 py-2 border rounded" />
      </div>
      <div>
        <label class="block text-sm font-medium mb-1">Avatar Seed</label>
        <input v-model="avatarSeed" class="w-full px-3 py-2 border rounded" />
        <div class="mt-2">
          <img :src="avatarUrl" alt="Avatar" class="w-16 h-16 rounded-full border" />
        </div>
      </div>
      <div>
        <label class="block text-sm font-medium mb-1">Theme</label>
        <select v-model="theme" class="w-full px-3 py-2 border rounded">
          <option value="light">Light</option>
          <option value="dark">Dark</option>
        </select>
      </div>
      <button @click="saveSettings" class="px-4 py-2 bg-blue-600 text-white rounded">Save Settings</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const name = ref(userStore.user?.name || '')
const avatarSeed = ref(userStore.user?.email || '')
const theme = ref(localStorage.getItem('theme') || 'light')

const avatarUrl = computed(() =>
  `https://api.dicebear.com/8.x/avataaars/svg?seed=${encodeURIComponent(avatarSeed.value)}`
)

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
</script>
