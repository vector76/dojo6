import { apiFetch } from './client'

export interface Payment {
  id: number
  user_id: number
  amount: number
  date: string
  note: string
  recorded_by: number
  created_at: string
  updated_at: string
}

export interface RecordPaymentRequest {
  user_id: number
  amount: number
  date: string
  note: string
}

export interface BalanceInfo {
  user_id: number
  expected_balance: number
  total_payments: number
  outstanding_balance: number
}

export function recordPayment(data: RecordPaymentRequest): Promise<Payment> {
  return apiFetch<Payment>('/payments', {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export function getPaymentHistory(userId: number): Promise<Payment[]> {
  return apiFetch<Payment[]>(`/users/${userId}/payments`)
}

export function getBalance(userId: number): Promise<BalanceInfo> {
  return apiFetch<BalanceInfo>(`/users/${userId}/balance`)
}
