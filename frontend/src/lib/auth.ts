import { api } from './api'
import type { User } from '../types/user'

export interface AuthResponse {
  user: User
  access_token: string
  refresh_token: string
}

export interface TokenPair {
  access_token: string
  refresh_token: string
}

export const authApi = {
  register: (username: string, email: string, password: string) =>
    api.post<AuthResponse>('/auth/register', { username, email, password }),

  login: (email: string, password: string) =>
    api.post<AuthResponse>('/auth/login', { email, password }),

  refresh: (refresh_token: string) =>
    api.post<TokenPair>('/auth/refresh', { refresh_token }),

  logout: (refresh_token: string) =>
    api.post<void>('/auth/logout', { refresh_token }),

  me: () =>
    api.get<User>('/auth/me'),
}
