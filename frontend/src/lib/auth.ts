import { post } from './api'
import type { AuthResponse, LoginRequest, RegisterRequest } from '@/types/api'

export const authApi = {
  login: (req: LoginRequest) => post<AuthResponse>('/auth/login', req),
  register: (req: RegisterRequest) => post<AuthResponse>('/auth/register', req),
  me: () => post<AuthResponse['user']>('/auth/me'),
}

// ─── Token helpers ────────────────────────────────────────────────────────────

export function saveSession(resp: AuthResponse) {
  localStorage.setItem('dd_token', resp.token)
  localStorage.setItem('dd_user', JSON.stringify(resp.user))
}

export function clearSession() {
  localStorage.removeItem('dd_token')
  localStorage.removeItem('dd_user')
}

export function getStoredUser(): AuthResponse['user'] | null {
  const raw = localStorage.getItem('dd_user')
  if (!raw) return null
  try {
    return JSON.parse(raw) as AuthResponse['user']
  } catch {
    return null
  }
}

export function isAuthenticated(): boolean {
  return !!localStorage.getItem('dd_token')
}
