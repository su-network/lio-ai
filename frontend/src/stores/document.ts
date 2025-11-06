import { defineStore } from 'pinia'
import { ref } from 'vue'
import { apiService, type Document } from '@/services/api'

export const useDocumentStore = defineStore('document', () => {
  const documents = ref<Document[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const fetchDocuments = async () => {
    loading.value = true
    error.value = null
    try {
      documents.value = await apiService.getDocuments()
    } catch (err) {
      error.value = 'Failed to fetch documents'
      console.error(err)
    } finally {
      loading.value = false
    }
  }

  const getDocument = async (id: number): Promise<Document | null> => {
    try {
      return await apiService.getDocument(id)
    } catch (err) {
      error.value = 'Failed to fetch document'
      console.error(err)
      return null
    }
  }

  const createDocument = async (title: string, content: string): Promise<Document | null> => {
    loading.value = true
    error.value = null
    try {
      const doc = await apiService.createDocument(title, content)
      documents.value.unshift(doc)
      return doc
    } catch (err) {
      error.value = 'Failed to create document'
      console.error(err)
      return null
    } finally {
      loading.value = false
    }
  }

  const updateDocument = async (id: number, title?: string, content?: string): Promise<Document | null> => {
    loading.value = true
    error.value = null
    try {
      const doc = await apiService.updateDocument(id, title, content)
      const index = documents.value.findIndex(d => d.id === id)
      if (index !== -1) {
        documents.value[index] = doc
      }
      return doc
    } catch (err) {
      error.value = 'Failed to update document'
      console.error(err)
      return null
    } finally {
      loading.value = false
    }
  }

  const deleteDocument = async (id: number): Promise<boolean> => {
    loading.value = true
    error.value = null
    try {
      await apiService.deleteDocument(id)
      documents.value = documents.value.filter(d => d.id !== id)
      return true
    } catch (err) {
      error.value = 'Failed to delete document'
      console.error(err)
      return false
    } finally {
      loading.value = false
    }
  }

  return {
    documents,
    loading,
    error,
    fetchDocuments,
    getDocument,
    createDocument,
    updateDocument,
    deleteDocument
  }
})
