import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { AuthProvider, useAuth } from '../hooks/useAuth'
import { setToken } from '../api/client'

const mockFetch = vi.fn()
vi.stubGlobal('fetch', mockFetch)

beforeEach(() => {
  mockFetch.mockReset()
  localStorage.clear()
})

function jsonResponse(body: unknown, status = 200): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    statusText: status === 200 ? 'OK' : 'Error',
    json: () => Promise.resolve(body),
    headers: new Headers(),
  } as Response
}

const testUser = {
  id: '1',
  email: 'admin@test.com',
  name: 'Admin',
  phone: '555-0100',
  role: 'admin' as const,
}

function TestConsumer() {
  const { user, loading, login, logout } = useAuth()
  return (
    <div>
      <div data-testid="loading">{String(loading)}</div>
      <div data-testid="user">{user ? user.name : 'none'}</div>
      <button onClick={() => login({ email: 'admin@test.com', password: 'pass' })}>
        Login
      </button>
      <button onClick={logout}>Logout</button>
    </div>
  )
}

describe('AuthProvider', () => {
  it('starts with no user and not loading when no token', () => {
    render(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>,
    )
    expect(screen.getByTestId('loading')).toHaveTextContent('false')
    expect(screen.getByTestId('user')).toHaveTextContent('none')
  })

  it('fetches user on mount when token exists', async () => {
    setToken('valid')
    mockFetch.mockResolvedValue(jsonResponse(testUser))

    render(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>,
    )

    expect(screen.getByTestId('loading')).toHaveTextContent('true')

    await waitFor(() => {
      expect(screen.getByTestId('user')).toHaveTextContent('Admin')
    })
    expect(screen.getByTestId('loading')).toHaveTextContent('false')
  })

  it('clears token and sets no user when /auth/me fails', async () => {
    setToken('expired')
    mockFetch.mockResolvedValue(
      jsonResponse({ error: 'Unauthorized' }, 401),
    )

    render(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>,
    )

    await waitFor(() => {
      expect(screen.getByTestId('loading')).toHaveTextContent('false')
    })
    expect(screen.getByTestId('user')).toHaveTextContent('none')
    expect(localStorage.getItem('auth_token')).toBeNull()
  })

  it('logs in and stores token', async () => {
    const authResponse = { token: 'new-token', user: testUser }
    mockFetch.mockResolvedValue(jsonResponse(authResponse))

    render(
      <AuthProvider>
        <TestConsumer />
      </AuthProvider>,
    )

    const user = userEvent.setup()
    await user.click(screen.getByText('Login'))

    await waitFor(() => {
      expect(screen.getByTestId('user')).toHaveTextContent('Admin')
    })
    expect(localStorage.getItem('auth_token')).toBe('new-token')
  })

  it('logs out, clears token, and calls onUnauthorized', async () => {
    const onUnauthorized = vi.fn()
    setToken('valid')
    mockFetch.mockResolvedValue(jsonResponse(testUser))

    render(
      <AuthProvider onUnauthorized={onUnauthorized}>
        <TestConsumer />
      </AuthProvider>,
    )

    await waitFor(() => {
      expect(screen.getByTestId('user')).toHaveTextContent('Admin')
    })

    const user = userEvent.setup()
    await user.click(screen.getByText('Logout'))

    expect(screen.getByTestId('user')).toHaveTextContent('none')
    expect(localStorage.getItem('auth_token')).toBeNull()
    expect(onUnauthorized).toHaveBeenCalledOnce()
  })
})

describe('useAuth outside provider', () => {
  it('throws when used outside AuthProvider', () => {
    const spy = vi.spyOn(console, 'error').mockImplementation(() => {})
    expect(() => render(<TestConsumer />)).toThrow(
      'useAuth must be used within an AuthProvider',
    )
    spy.mockRestore()
  })
})
