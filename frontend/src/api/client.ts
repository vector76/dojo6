const TOKEN_KEY = 'auth_token'

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken(): void {
  localStorage.removeItem(TOKEN_KEY)
}

let onUnauthorized: (() => void) | null = null

export function setOnUnauthorized(callback: (() => void) | null): void {
  onUnauthorized = callback
}

export class ApiError extends Error {
  status: number
  statusText: string
  body: unknown

  constructor(status: number, statusText: string, body: unknown) {
    const message =
      body && typeof body === 'object' && 'error' in body
        ? (body as { error: string }).error
        : statusText
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.statusText = statusText
    this.body = body
  }
}

export async function apiFetch<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const headers = new Headers(options.headers)

  const token = getToken()
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }

  if (
    options.body &&
    typeof options.body === 'string' &&
    !headers.has('Content-Type')
  ) {
    headers.set('Content-Type', 'application/json')
  }

  const response = await fetch(`/api${path}`, {
    ...options,
    headers,
  })

  if (response.status === 401) {
    clearToken()
    onUnauthorized?.()
  }

  if (!response.ok) {
    let body: unknown
    try {
      body = await response.json()
    } catch {
      body = null
    }
    throw new ApiError(response.status, response.statusText, body)
  }

  if (response.status === 204) {
    return undefined as T
  }

  return response.json() as Promise<T>
}
