import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import './assets/style.css'
import { useUserStore } from './stores/user'
import { useChatStore } from './stores/chat'

const pinia = createPinia()
const app = createApp(App)

app.use(pinia)
app.use(router)

// Initialize user store before mounting
const userStore = useUserStore()
userStore.initializeUser().finally(async () => {
  // If user is authenticated, load available models
  if (userStore.isAuthenticated) {
    const chatStore = useChatStore()
    await chatStore.loadAvailableModels()
    console.log('✓ Models loaded on startup')
  }
  app.mount('#app')
})