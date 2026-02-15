import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { AuthProvider } from '../hooks/useAuth'
import { ProtectedRoute } from '../components/ProtectedRoute'
import { Layout } from '../components/Layout'
import { LoginPage } from '../pages/LoginPage'
import { SetupPage } from '../pages/SetupPage'
import { DashboardPage } from '../pages/DashboardPage'
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
  name: 'Admin User',
  phone: '555-0100',
  role: 'admin' as const,
}

function renderApp(initialRoute = '/') {
  return render(
    <MemoryRouter initialEntries={[initialRoute]}>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/setup" element={<SetupPage />} />
          <Route
            element={
              <ProtectedRoute>
                <Layout />
              </ProtectedRoute>
            }
          >
            <Route path="/" element={<DashboardPage />} />
          </Route>
        </Routes>
      </AuthProvider>
    </MemoryRouter>,
  )
}

describe('redirect on unauthenticated', () => {
  it('redirects to /login when not authenticated', () => {
    renderApp('/')
    expect(screen.getByRole('heading', { name: 'Login' })).toBeInTheDocument()
  })
})

describe('authenticated access', () => {
  it('shows dashboard when authenticated', async () => {
    setToken('valid')
    mockFetch.mockResolvedValue(jsonResponse(testUser))

    renderApp('/')

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Dashboard' })).toBeInTheDocument()
    })
    expect(screen.getByText('Welcome, Admin User!')).toBeInTheDocument()
  })

  it('shows role-aware navigation for admin', async () => {
    setToken('valid')
    mockFetch.mockResolvedValue(jsonResponse(testUser))

    renderApp('/')

    const nav = await screen.findByRole('navigation')
    expect(nav).toHaveTextContent('Dashboard')
    expect(nav).toHaveTextContent('Members')
    expect(nav).toHaveTextContent('Classes')
    expect(nav).toHaveTextContent('Payments')
    expect(nav).not.toHaveTextContent('My Attendance')
    expect(nav).not.toHaveTextContent('My Payments')
  })

  it('shows role-aware navigation for user role', async () => {
    const memberUser = { ...testUser, role: 'user' as const }
    setToken('valid')
    mockFetch.mockResolvedValue(jsonResponse(memberUser))

    renderApp('/')

    await waitFor(() => {
      expect(screen.getByText('My Attendance')).toBeInTheDocument()
    })
    expect(screen.getByText('My Payments')).toBeInTheDocument()
    expect(screen.queryByText('Members')).not.toBeInTheDocument()
    expect(screen.queryByText('Payments')).not.toBeInTheDocument()
  })
})

describe('login form submission', () => {
  it('submits login and shows dashboard on success', async () => {
    const authResponse = { token: 'new-token', user: testUser }
    mockFetch.mockResolvedValue(jsonResponse(authResponse))

    renderApp('/login')

    const user = userEvent.setup()
    await user.type(screen.getByLabelText('Email'), 'admin@test.com')
    await user.type(screen.getByLabelText('Password'), 'password123')
    await user.click(screen.getByRole('button', { name: 'Login' }))

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Dashboard' })).toBeInTheDocument()
    })
  })

  it('shows error on login failure', async () => {
    mockFetch.mockResolvedValue(
      jsonResponse({ error: 'Invalid credentials' }, 401),
    )

    renderApp('/login')

    const user = userEvent.setup()
    await user.type(screen.getByLabelText('Email'), 'bad@test.com')
    await user.type(screen.getByLabelText('Password'), 'wrong')
    await user.click(screen.getByRole('button', { name: 'Login' }))

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Invalid credentials')
    })
  })
})

describe('setup flow', () => {
  it('submits setup form and navigates to dashboard', async () => {
    const authResponse = { token: 'admin-token', user: testUser }
    mockFetch.mockResolvedValue(jsonResponse(authResponse))

    renderApp('/setup')

    expect(screen.getByRole('heading', { name: 'Initial Setup' })).toBeInTheDocument()

    const user = userEvent.setup()
    await user.type(screen.getByLabelText('Name'), 'Admin User')
    await user.type(screen.getByLabelText('Email'), 'admin@test.com')
    await user.type(screen.getByLabelText('Phone'), '555-0100')
    await user.type(screen.getByLabelText('Password'), 'password123')
    await user.click(screen.getByRole('button', { name: 'Create Admin Account' }))

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Dashboard' })).toBeInTheDocument()
    })
  })

  it('shows error on setup failure', async () => {
    mockFetch.mockResolvedValue(
      jsonResponse({ error: 'Users already exist' }, 409),
    )

    renderApp('/setup')

    const user = userEvent.setup()
    await user.type(screen.getByLabelText('Name'), 'Admin')
    await user.type(screen.getByLabelText('Email'), 'admin@test.com')
    await user.type(screen.getByLabelText('Phone'), '555-0100')
    await user.type(screen.getByLabelText('Password'), 'pass')
    await user.click(screen.getByRole('button', { name: 'Create Admin Account' }))

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Users already exist')
    })
  })
})

describe('logout', () => {
  it('logs out and redirects to login', async () => {
    setToken('valid')
    mockFetch.mockResolvedValue(jsonResponse(testUser))

    renderApp('/')

    await waitFor(() => {
      expect(screen.getByText('Admin User')).toBeInTheDocument()
    })

    const user = userEvent.setup()
    await user.click(screen.getByRole('button', { name: 'Logout' }))

    expect(screen.getByRole('heading', { name: 'Login' })).toBeInTheDocument()
  })
})
