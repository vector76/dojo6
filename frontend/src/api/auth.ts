import { apiFetch } from './client'

export interface User {
  id: string
  email: string
  name: string
  phone: string
  role: 'admin' | 'instructor' | 'user'
}

export interface LoginRequest {
  email: string
  password: string
}

export interface AuthResponse {
  token: string
  user: User
}

export interface SetupRequest {
  email: string
  password: string
  name: string
  phone: string
}

export interface SetupStatus {
  setupRequired: boolean
}

export function login(credentials: LoginRequest): Promise<AuthResponse> {
  return apiFetch<AuthResponse>('/auth/login', {
    method: 'POST',
    body: JSON.stringify(credentials),
  })
}

export function getMe(): Promise<User> {
  return apiFetch<User>('/auth/me')
}

export function getSetupStatus(): Promise<SetupStatus> {
  return apiFetch<SetupStatus>('/auth/setup-status')
}

export function setup(data: SetupRequest): Promise<AuthResponse> {
  return apiFetch<AuthResponse>('/auth/setup', {
    method: 'POST',
    body: JSON.stringify(data),
  })
}
