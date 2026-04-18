import { create } from 'zustand'
import { authApi, clearSession, getStoredUser, saveSession } from '@/lib/auth'
import type { LoginRequest, RegisterRequest } from '@/types/api'

interface AuthUser {
  id: string
  clinic_id: string
  email: string
  first_name: string
  last_name: string
  role: string
}

interface AuthState {
  user: AuthUser | null
  isLoading: boolean
  error: string | null

  login: (req: LoginRequest) => Promise<void>
  register: (req: RegisterRequest) => Promise<void>
  logout: () => void
  init: () => void
}

// Global auth store — accessible anywhere without context providers
export const useAuth = create<AuthState>((set) => ({
  user: null,
  isLoading: false,
  error: null,

  init() {
    const stored = getStoredUser()
    if (stored) set({ user: stored })
  },

  async login(req) {
    set({ isLoading: true, error: null })
    try {
      const resp = await authApi.login(req)
      saveSession(resp)
      set({ user: resp.user, isLoading: false })
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Ошибка входа'
      set({ error: msg, isLoading: false })
      throw err
    }
  },

  async register(req) {
    set({ isLoading: true, error: null })
    try {
      const resp = await authApi.register(req)
      saveSession(resp)
      set({ user: resp.user, isLoading: false })
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Ошибка регистрации'
      set({ error: msg, isLoading: false })
      throw err
    }
  },

  logout() {
    clearSession()
    set({ user: null })
  },
}))
