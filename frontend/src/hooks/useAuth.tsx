import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from 'react'
import type { ReactNode } from 'react'
import type { User } from '../api/auth'
import { getMe, login as apiLogin, setup as apiSetup } from '../api/auth'
import type { LoginRequest, SetupRequest } from '../api/auth'
import { clearToken, getToken, setOnUnauthorized, setToken } from '../api/client'

interface AuthState {
  user: User | null
  loading: boolean
}

interface AuthContextValue extends AuthState {
  login: (credentials: LoginRequest) => Promise<void>
  setup: (data: SetupRequest) => Promise<void>
  logout: () => void
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({
  children,
  onUnauthorized,
}: {
  children: ReactNode
  onUnauthorized?: () => void
}) {
  const [state, setState] = useState<AuthState>({
    user: null,
    loading: !!getToken(),
  })

  const handleUnauthorized = useCallback(() => {
    setState({ user: null, loading: false })
    onUnauthorized?.()
  }, [onUnauthorized])

  useEffect(() => {
    setOnUnauthorized(handleUnauthorized)
    return () => setOnUnauthorized(null)
  }, [handleUnauthorized])

  useEffect(() => {
    if (!getToken()) return

    let cancelled = false
    getMe()
      .then((user) => {
        if (!cancelled) setState({ user, loading: false })
      })
      .catch(() => {
        if (!cancelled) {
          clearToken()
          setState({ user: null, loading: false })
        }
      })
    return () => {
      cancelled = true
    }
  }, [])

  const login = useCallback(async (credentials: LoginRequest) => {
    const response = await apiLogin(credentials)
    setToken(response.token)
    setState({ user: response.user, loading: false })
  }, [])

  const setup = useCallback(async (data: SetupRequest) => {
    const response = await apiSetup(data)
    setToken(response.token)
    setState({ user: response.user, loading: false })
  }, [])

  const logout = useCallback(() => {
    clearToken()
    setState({ user: null, loading: false })
    onUnauthorized?.()
  }, [onUnauthorized])

  const value = useMemo(
    () => ({ ...state, login, setup, logout }),
    [state, login, setup, logout],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
