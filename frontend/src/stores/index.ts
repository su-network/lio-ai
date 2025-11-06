import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface ChatMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  timestamp: Date
  model?: string
}

export interface Document {
  id: string
  name: string
  type: string
  size: number
  uploadDate: Date
  tags: string[]
  content?: string
}

export const useChatStore = defineStore('chat', () => {
  const messages = ref<ChatMessage[]>([
    {
      id: '1',
      role: 'assistant',
      content: 'Hello! I\'m your AI coding assistant. How can I help you with code generation today?',
      timestamp: new Date(),
      model: 'gpt-4',
    },
  ])

  const selectedModel = ref('gpt-4')
  const isTyping = ref(false)

  const addMessage = (message: Omit<ChatMessage, 'id' | 'timestamp'>) => {
    const newMessage: ChatMessage = {
      ...message,
      id: Date.now().toString(),
      timestamp: new Date(),
    }
    messages.value.push(newMessage)
  }

  const clearMessages = () => {
    messages.value = [
      {
        id: '1',
        role: 'assistant',
        content: 'Hello! I\'m your AI coding assistant. How can I help you with code generation today?',
        timestamp: new Date(),
        model: selectedModel.value,
      },
    ]
  }

  const setTyping = (typing: boolean) => {
    isTyping.value = typing
  }

  return {
    messages,
    selectedModel,
    isTyping,
    addMessage,
    clearMessages,
    setTyping,
  }
})

export const useDocumentStore = defineStore('documents', () => {
  const documents = ref<Document[]>([])
  const searchQuery = ref('')
  const selectedType = ref('')
  const sortBy = ref('date')

  const filteredDocuments = computed(() => {
    let filtered = documents.value.filter(doc => {
      const matchesSearch = doc.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
                           doc.tags.some(tag => tag.toLowerCase().includes(searchQuery.value.toLowerCase()))
      const matchesType = !selectedType.value || doc.type === selectedType.value
      return matchesSearch && matchesType
    })

    // Sort
    filtered.sort((a, b) => {
      switch (sortBy.value) {
        case 'name':
          return a.name.localeCompare(b.name)
        case 'size':
          return b.size - a.size
        case 'date':
        default:
          return b.uploadDate.getTime() - a.uploadDate.getTime()
      }
    })

    return filtered
  })

  const addDocument = (doc: Omit<Document, 'id' | 'uploadDate'>) => {
    const newDoc: Document = {
      ...doc,
      id: Date.now().toString() + Math.random().toString(36).substr(2, 9),
      uploadDate: new Date(),
    }
    documents.value.unshift(newDoc)
  }

  const removeDocument = (id: string) => {
    const index = documents.value.findIndex(doc => doc.id === id)
    if (index > -1) {
      documents.value.splice(index, 1)
    }
  }

  const updateDocumentTags = (id: string, tags: string[]) => {
    const doc = documents.value.find(d => d.id === id)
    if (doc) {
      doc.tags = tags
    }
  }

  return {
    documents,
    searchQuery,
    selectedType,
    sortBy,
    filteredDocuments,
    addDocument,
    removeDocument,
    updateDocumentTags,
  }
})

export const useGenerationStore = defineStore('generation', () => {
  const uploadedFiles = ref<File[]>([])
  const inputText = ref('')
  const selectedLanguage = ref('typescript')
  const selectedFramework = ref('react')
  const generatedCode = ref('')
  const isGenerating = ref(false)

  const canGenerate = computed(() => {
    return uploadedFiles.value.length > 0 || inputText.value.trim().length > 0
  })

  const addFile = (file: File) => {
    uploadedFiles.value.push(file)
  }

  const removeFile = (file: File) => {
    const index = uploadedFiles.value.indexOf(file)
    if (index > -1) {
      uploadedFiles.value.splice(index, 1)
    }
  }

  const clearFiles = () => {
    uploadedFiles.value = []
  }

  const setGeneratedCode = (code: string) => {
    generatedCode.value = code
  }

  const setGenerating = (generating: boolean) => {
    isGenerating.value = generating
  }

  return {
    uploadedFiles,
    inputText,
    selectedLanguage,
    selectedFramework,
    generatedCode,
    isGenerating,
    canGenerate,
    addFile,
    removeFile,
    clearFiles,
    setGeneratedCode,
    setGenerating,
  }
})