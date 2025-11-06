import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'

interface User {
  id: string
  name: string
  email: string
  avatar: string
}

/**
 * Generate a unique user ID based on browser fingerprint
 */
function generateUserId(): string {
  // Check if we already have a user ID in localStorage
  const existingId = localStorage.getItem('lioai_user_id')
  if (existingId) return existingId

  // Generate a unique ID based on browser characteristics
  const nav = navigator as any
  const screen = window.screen
  
  const fingerprint = [
    nav.userAgent,
    nav.language,
    screen.width,
    screen.height,
    screen.colorDepth,
    new Date().getTimezoneOffset(),
    nav.hardwareConcurrency || 'unknown',
    nav.platform
  ].join('|')

  // Create a simple hash
  let hash = 0
  for (let i = 0; i < fingerprint.length; i++) {
    const char = fingerprint.charCodeAt(i)
    hash = ((hash << 5) - hash) + char
    hash = hash & hash // Convert to 32bit integer
  }

  // Create a user ID with timestamp for uniqueness
  const userId = `user_${Math.abs(hash).toString(36)}_${Date.now().toString(36)}`
  
  // Save to localStorage
  localStorage.setItem('lioai_user_id', userId)
  
  return userId
}

export const useUserStore = defineStore('user', () => {
  const router = useRouter()
  const user = ref<User | null>(null)
  const userId = ref<string>('')

  // Initialize user ID
  if (typeof window !== 'undefined') {
    userId.value = generateUserId()
    
    // Load user from localStorage on store init
    const saved = localStorage.getItem('lioai_user')
    if (saved) {
      try {
        const parsedUser = JSON.parse(saved)
        user.value = {
          ...parsedUser,
          id: userId.value // Ensure ID is always set
        }
      } catch {}
    } else {
      // Create a default anonymous user
      user.value = {
        id: userId.value,
        name: 'Anonymous User',
        email: `${userId.value}@lio-ai.local`,
        avatar: `https://api.dicebear.com/8.x/avataaars/svg?seed=${userId.value}`
      }
      localStorage.setItem('lioai_user', JSON.stringify(user.value))
    }
  }

  const isLoggedIn = computed(() => !!user.value)

  async function login(email: string, name: string) {
    // Simulate an API call
    await new Promise(resolve => setTimeout(resolve, 500))
    user.value = {
      id: userId.value,
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
      // Don't remove user_id, keep for tracking
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
    userId,
    isLoggedIn,
    login,
    logout,
    updateProfile
  }
})
