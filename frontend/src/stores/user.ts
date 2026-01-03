import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { apiService } from '@/services/api'

interface User {
  id: number // Changed from string to number for JWT user_id
  name: string
  email: string
  avatar?: string
}

export const useUserStore = defineStore('user', () => {
  const router = useRouter()
  const user = ref<User | null>(null)
  const userId = computed(() => user.value?.id || 0)
  const isAuthenticated = ref(false)
  const isInitialized = ref(false)

  // Initialize user from httpOnly cookie (auth_token)
  const initializeUser = async () => {
    // The JWT token is in httpOnly cookie, so we just try to fetch the profile
    // If the cookie exists and is valid, the request will succeed
    try {
      const profile = await apiService.getProfile()
      user.value = {
        id: profile.id,
        name: profile.name,
        email: profile.email,
        avatar: profile.avatar || `https://api.dicebear.com/8.x/avataaars/svg?seed=${profile.email}`
      }
      isAuthenticated.value = true
      
      // Trigger API key sync on login to ensure models are available
      try {
        await apiService.syncApiKeys()
        console.log('✓ API keys synced on user initialization')
      } catch (syncError) {
        console.warn('Failed to sync API keys:', syncError)
        // Don't fail initialization if sync fails
      }
    } catch (error) {
      console.error('Failed to load user profile:', error)
      // No valid cookie, user needs to log in
      isAuthenticated.value = false
      user.value = null
    }
    isInitialized.value = true
  }

  const isLoggedIn = computed(() => isAuthenticated.value && !!user.value)

  async function register(username: string, email: string, password: string, name: string) {
    try {
      const response = await apiService.register(username, email, password, name)
      
      // Token is stored in httpOnly cookie by the server
      // No need to store in localStorage
      
      user.value = {
        id: response.user.id,
        name: response.user.name,
        email: response.user.email,
        avatar: `https://api.dicebear.com/8.x/avataaars/svg?seed=${email}`
      }
      isAuthenticated.value = true
      router.push('/')
      return { success: true }
    } catch (error: any) {
      return { 
        success: false, 
        error: error.response?.data?.error || 'Registration failed' 
      }
    }
  }

  async function login(email: string, password: string) {
    try {
      const response = await apiService.login(email, password)
      
      // Token is stored in httpOnly cookie by the server
      // No need to store in localStorage
      
      user.value = {
        id: response.user.id,
        name: response.user.name,
        email: response.user.email,
        avatar: `https://api.dicebear.com/8.x/avataaars/svg?seed=${email}`
      }
      isAuthenticated.value = true
      router.push('/')
      return { success: true }
    } catch (error: any) {
      return { 
        success: false, 
        error: error.response?.data?.error || 'Login failed' 
      }
    }
  }

  async function logout() {
    try {
      await apiService.logout()
    } catch (error) {
      console.error('Logout error:', error)
    } finally {
      user.value = null
      isAuthenticated.value = false
      // Cookie will be cleared by the server on logout
      router.push('/login')
    }
  }

  function updateProfile(newName: string, newAvatarSeed: string) {
    if (user.value) {
      user.value.name = newName
      user.value.avatar = `https://api.dicebear.com/8.x/avataaars/svg?seed=${encodeURIComponent(newAvatarSeed)}`
    }
  }

  return {
    user,
    userId,
    isLoggedIn,
    isAuthenticated,
    isInitialized,
    register,
    login,
    logout,
    updateProfile,
    initializeUser
  }
})
