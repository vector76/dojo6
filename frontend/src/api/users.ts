import { apiFetch } from './client'
import type { User } from './auth'

export interface Member extends User {
  membership_type: string
  membership_status: string
  emergency_contact: string
  join_date: string
  expected_balance: number
}

export interface CreateMemberRequest {
  name: string
  email: string
  phone: string
  password: string
  role: 'admin' | 'instructor' | 'user'
  membership_type: string
  membership_status: string
  emergency_contact: string
  join_date: string
}

export interface UpdateMemberRequest {
  name: string
  email: string
  phone: string
  role: 'admin' | 'instructor' | 'user'
  membership_type: string
  membership_status: string
  emergency_contact: string
  join_date: string
}

export function listUsers(): Promise<Member[]> {
  return apiFetch<Member[]>('/users')
}

export function getUser(id: string): Promise<Member> {
  return apiFetch<Member>(`/users/${id}`)
}

export function createUser(data: CreateMemberRequest): Promise<Member> {
  return apiFetch<Member>('/users', {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export function updateUser(id: string, data: UpdateMemberRequest): Promise<Member> {
  return apiFetch<Member>(`/users/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
}

export function deleteUser(id: string): Promise<void> {
  return apiFetch<void>(`/users/${id}`, {
    method: 'DELETE',
  })
}

export function resetPassword(id: string, password: string): Promise<void> {
  return apiFetch<void>(`/users/${id}/password`, {
    method: 'PUT',
    body: JSON.stringify({ password }),
  })
}

export function setExpectedBalance(id: string, expected_balance: number): Promise<Member> {
  return apiFetch<Member>(`/users/${id}/balance`, {
    method: 'PUT',
    body: JSON.stringify({ expected_balance }),
  })
}
