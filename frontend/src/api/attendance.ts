import { apiFetch } from './client'

export interface Attendance {
  id: number
  class_id: number
  user_id: number
  checked_in_at: string
}

export interface RecordAttendanceRequest {
  user_id: number
}

export function recordAttendance(classId: number, data: RecordAttendanceRequest): Promise<Attendance> {
  return apiFetch<Attendance>(`/classes/${classId}/attendance`, {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export function getClassAttendance(classId: number): Promise<Attendance[]> {
  return apiFetch<Attendance[]>(`/classes/${classId}/attendance`)
}

export function getUserAttendance(userId: number): Promise<Attendance[]> {
  return apiFetch<Attendance[]>(`/users/${userId}/attendance`)
}
