/**
 * Simple XOR-based obfuscation for API responses
 * Note: This is NOT strong encryption, but adds a layer of obfuscation
 * For production, use proper encryption libraries like crypto-js
 */

const SECRET_KEY = import.meta.env.VITE_ENCRYPTION_KEY || 'lio-ai-secret-key-2025'

/**
 * Simple XOR cipher for obfuscation
 */
function xorCipher(text: string, key: string): string {
  let result = ''
  for (let i = 0; i < text.length; i++) {
    result += String.fromCharCode(text.charCodeAt(i) ^ key.charCodeAt(i % key.length))
  }
  return result
}

/**
 * Encode to Base64
 */
function toBase64(str: string): string {
  return btoa(unescape(encodeURIComponent(str)))
}

/**
 * Decode from Base64
 */
function fromBase64(str: string): string {
  return decodeURIComponent(escape(atob(str)))
}

/**
 * Obfuscate data (encrypt and encode)
 */
export function obfuscate(data: any): string {
  const jsonString = JSON.stringify(data)
  const ciphered = xorCipher(jsonString, SECRET_KEY)
  return toBase64(ciphered)
}

/**
 * Deobfuscate data (decode and decrypt)
 */
export function deobfuscate(obfuscatedData: string): any {
  try {
    const ciphered = fromBase64(obfuscatedData)
    const jsonString = xorCipher(ciphered, SECRET_KEY)
    return JSON.parse(jsonString)
  } catch (error) {
    console.error('Deobfuscation failed:', error)
    return null
  }
}

/**
 * Check if response is obfuscated
 */
export function isObfuscated(data: any): boolean {
  if (typeof data === 'string') {
    try {
      // Try to decode from base64 and check if it looks like obfuscated data
      fromBase64(data)
      return true
    } catch {
      return false
    }
  }
  return false
}
