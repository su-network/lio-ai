<template>
  <div class="p-4" v-if="userStore.user">
    <h1 class="text-2xl font-bold">Profile</h1>
    <div class="mt-4">
      <img :src="userStore.user.avatar" alt="User Avatar" class="w-24 h-24 rounded-full">
      <p class="mt-2"><strong>Name:</strong> {{ userStore.user.name }}</p>
      <p><strong>Email:</strong> {{ userStore.user.email }}</p>
      <button @click="userStore.logout" class="mt-4 px-4 py-2 bg-red-500 text-white rounded">Logout</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
</script>
