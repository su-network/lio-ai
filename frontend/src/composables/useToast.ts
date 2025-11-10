import { inject, type InjectionKey } from 'vue'

export interface ToastOptions {
  type: 'success' | 'error' | 'info' | 'warning'
  title: string
  message?: string
  duration?: number
}

export interface ToastService {
  addToast: (options: ToastOptions) => void
  success: (title: string, message?: string) => void
  error: (title: string, message?: string) => void
  info: (title: string, message?: string) => void
  warning: (title: string, message?: string) => void
  networkError: (operation: string) => void
  serviceUnavailable: (serviceName: string) => void
  noDataFound: (dataType: string) => void
}

export const ToastServiceKey: InjectionKey<ToastService> = Symbol('ToastService')

export function useToast(): ToastService {
  const toastService = inject(ToastServiceKey)
  
  if (!toastService) {
    // Enhanced fallback with better error handling
    const fallbackService: ToastService = {
      addToast: (options: ToastOptions) => {
        console.log(`[${options.type.toUpperCase()}] ${options.title}${options.message ? ': ' + options.message : ''}`)
      },
      success: (title: string, message?: string) => {
        console.log(`[SUCCESS] ${title}${message ? ': ' + message : ''}`)
      },
      error: (title: string, message?: string) => {
        console.error(`[ERROR] ${title}${message ? ': ' + message : ''}`)
      },
      info: (title: string, message?: string) => {
        console.info(`[INFO] ${title}${message ? ': ' + message : ''}`)
      },
      warning: (title: string, message?: string) => {
        console.warn(`[WARNING] ${title}${message ? ': ' + message : ''}`)
      },
      networkError: (operation: string) => {
        console.error(`[NETWORK ERROR] Failed to ${operation}. Please check if the server is running.`)
      },
      serviceUnavailable: (serviceName: string) => {
        console.error(`[SERVICE ERROR] ${serviceName} is currently unavailable.`)
      },
      noDataFound: (dataType: string) => {
        const suggestion = getNoDataSuggestion(dataType)
        console.info(`[NO DATA] No ${dataType} found. ${suggestion}`)
      }
    }
    
    return fallbackService
  }
  
  // Extend the service with additional methods
  return {
    ...toastService,
    networkError: (operation: string) => {
      toastService.error(
        'Connection Error',
        `Failed to ${operation}. Please check if the server is running and try again.`
      )
    },
    serviceUnavailable: (serviceName: string) => {
      toastService.error(
        `${serviceName} Unavailable`,
        `The ${serviceName} service is currently unavailable. Please try again in a moment.`
      )
    },
    noDataFound: (dataType: string) => {
      const suggestion = getNoDataSuggestion(dataType)
      toastService.info(
        'No Data Found',
        `No ${dataType} found. ${suggestion}`
      )
    }
  }
}

function getNoDataSuggestion(dataType: string): string {
  const suggestions: Record<string, string> = {
    'models': 'Please add API keys in Settings to enable AI models.',
    'conversations': 'Start your first conversation by sending a message.',
    'api keys': 'Add your first API key to enable AI models.',
    'documents': 'Upload your first document to get started.',
    'usage data': 'Usage data will appear after you start using AI models.'
  }
  
  return suggestions[dataType.toLowerCase()] || 'Please refresh the page or try again.'
}
