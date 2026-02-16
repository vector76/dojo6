import { apiFetch } from './client'

export interface ClassType {
  id: number
  name: string
  description: string
  created_at: string
  updated_at: string
}

export interface CreateClassTypeRequest {
  name: string
  description: string
}

export interface Class {
  id: number
  class_type_id: number
  instructor_id: number
  start_time: string
  duration_minutes: number
  capacity: number
  created_at: string
  updated_at: string
}

export function listClassTypes(): Promise<ClassType[]> {
  return apiFetch<ClassType[]>('/class-types')
}

export function createClassType(data: CreateClassTypeRequest): Promise<ClassType> {
  return apiFetch<ClassType>('/class-types', {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export function updateClassType(id: number, data: CreateClassTypeRequest): Promise<ClassType> {
  return apiFetch<ClassType>(`/class-types/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
}

export function deleteClassType(id: number): Promise<void> {
  return apiFetch<void>(`/class-types/${id}`, {
    method: 'DELETE',
  })
}

export function listClasses(): Promise<Class[]> {
  return apiFetch<Class[]>('/classes')
}
