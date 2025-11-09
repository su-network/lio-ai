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
}

export const ToastServiceKey: InjectionKey<ToastService> = Symbol('ToastService')

export function useToast(): ToastService {
  const toastService = inject(ToastServiceKey)
  
  if (!toastService) {
    // Fallback to console if toast service is not available
    return {
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
      }
    }
  }
  
  return toastService
}
