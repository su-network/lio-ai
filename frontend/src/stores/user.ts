import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'

interface User {
  name: string
  email: string
  avatar: string
}

export const useUserStore = defineStore('user', () => {
  const router = useRouter()
  const user = ref<User | null>(null)

  // Load user from localStorage on store init
  if (typeof window !== 'undefined') {
    const saved = localStorage.getItem('lioai_user')
    if (saved) {
      try {
        user.value = JSON.parse(saved)
      } catch {}
    }
  }

  const isLoggedIn = computed(() => !!user.value)

  async function login(email: string, name: string) {
    // Simulate an API call
    await new Promise(resolve => setTimeout(resolve, 500))
    user.value = {
      name,
      email,
      avatar: `https://api.dicebear.com/8.x/avataaars/svg?seed=${email}`
    }
    // Save to localStorage
    if (typeof window !== 'undefined') {
      localStorage.setItem('lioai_user', JSON.stringify(user.value))
    }
    router.push('/')
  }

  function logout() {
    user.value = null
    if (typeof window !== 'undefined') {
      localStorage.removeItem('lioai_user')
    }
    router.push('/login')
  }

  function updateProfile(newName: string, newAvatarSeed: string) {
    if (user.value) {
      user.value.name = newName
      user.value.avatar = `https://api.dicebear.com/8.x/avataaars/svg?seed=${encodeURIComponent(newAvatarSeed)}`
      // Save to localStorage
      if (typeof window !== 'undefined') {
        localStorage.setItem('lioai_user', JSON.stringify(user.value))
      }
    }
  }

  return {
    user,
    isLoggedIn,
    login,
    logout,
    updateProfile
  }
})
