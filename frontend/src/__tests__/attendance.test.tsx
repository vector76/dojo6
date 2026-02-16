import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { AuthProvider } from '../hooks/useAuth'
import { AttendancePage } from '../pages/AttendancePage'
import { MyAttendancePage } from '../pages/MyAttendancePage'
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

const adminUser = {
  id: '1',
  email: 'admin@test.com',
  name: 'Admin',
  phone: '555-0100',
  role: 'admin' as const,
}

const memberUser = {
  id: '5',
  email: 'member@test.com',
  name: 'Member',
  phone: '555-0200',
  role: 'user' as const,
}

const sampleClassTypes = [
  { id: 1, name: 'Yoga', description: 'Yoga class', created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:00Z' },
]

const sampleClasses = [
  { id: 10, class_type_id: 1, instructor_id: 1, start_time: '2026-03-15T14:00:00Z', duration_minutes: 60, capacity: 20, created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:00Z' },
]

const sampleAttendance = [
  { id: 1, class_id: 10, user_id: 5, checked_in_at: '2026-03-15T14:05:00Z' },
  { id: 2, class_id: 10, user_id: 6, checked_in_at: '2026-03-15T14:10:00Z' },
]

function setupAttendancePageMocks(overrides?: Record<string, unknown>) {
  setToken('valid')
  mockFetch.mockImplementation((url: string) => {
    if (url.includes('/auth/me')) return Promise.resolve(jsonResponse(adminUser))
    if (url.includes('/class-types')) return Promise.resolve(jsonResponse(sampleClassTypes))
    if (url.includes('/classes') && !url.includes('/attendance')) return Promise.resolve(jsonResponse(sampleClasses))
    if (overrides && url in overrides) return Promise.resolve(jsonResponse(overrides[url]))
    return Promise.resolve(jsonResponse([]))
  })
}

function renderAttendancePage() {
  setupAttendancePageMocks()
  return render(
    <MemoryRouter>
      <AuthProvider>
        <AttendancePage />
      </AuthProvider>
    </MemoryRouter>,
  )
}

function renderMyAttendancePage() {
  setToken('valid')
  mockFetch.mockImplementation((url: string) => {
    if (url.includes('/auth/me')) return Promise.resolve(jsonResponse(memberUser))
    if (url.includes('/attendance')) return Promise.resolve(jsonResponse(sampleAttendance))
    return Promise.resolve(jsonResponse([]))
  })
  return render(
    <MemoryRouter>
      <AuthProvider>
        <MyAttendancePage />
      </AuthProvider>
    </MemoryRouter>,
  )
}

describe('AttendancePage', () => {
  it('renders check-in form with class selector', async () => {
    renderAttendancePage()

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Record Attendance' })).toBeInTheDocument()
    })
    expect(screen.getByLabelText('Class')).toBeInTheDocument()
    expect(screen.getByLabelText('Member ID')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Check In' })).toBeInTheDocument()
  })

  it('populates class dropdown with class type names', async () => {
    renderAttendancePage()

    await waitFor(() => {
      const select = screen.getByLabelText('Class') as HTMLSelectElement
      expect(select.options.length).toBeGreaterThan(1)
    })

    const select = screen.getByLabelText('Class') as HTMLSelectElement
    // Default option + 1 class
    expect(select.options).toHaveLength(2)
    expect(select.options[1].textContent).toContain('Yoga')
  })

  it('loads attendance when a class is selected', async () => {
    setupAttendancePageMocks()
    // After initial load, override to return attendance when class is selected
    const originalImpl = mockFetch.getMockImplementation()!
    mockFetch.mockImplementation((url: string, opts?: RequestInit) => {
      if (url.includes('/classes/10/attendance') && !opts?.method) {
        return Promise.resolve(jsonResponse(sampleAttendance))
      }
      return originalImpl(url, opts)
    })

    render(
      <MemoryRouter>
        <AuthProvider>
          <AttendancePage />
        </AuthProvider>
      </MemoryRouter>,
    )
    const user = userEvent.setup()

    await waitFor(() => {
      const select = screen.getByLabelText('Class') as HTMLSelectElement
      expect(select.options.length).toBeGreaterThan(1)
    })

    await user.selectOptions(screen.getByLabelText('Class'), '10')

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Class Attendance' })).toBeInTheDocument()
    })

    const rows = screen.getAllByRole('row')
    // header + 2 attendance records
    expect(rows).toHaveLength(3)
    expect(screen.getByText('5')).toBeInTheDocument()
    expect(screen.getByText('6')).toBeInTheDocument()
  })

  it('shows empty attendance state', async () => {
    renderAttendancePage()
    const user = userEvent.setup()

    await waitFor(() => {
      const select = screen.getByLabelText('Class') as HTMLSelectElement
      expect(select.options.length).toBeGreaterThan(1)
    })

    await user.selectOptions(screen.getByLabelText('Class'), '10')

    await waitFor(() => {
      expect(screen.getByText('No attendance recorded for this class.')).toBeInTheDocument()
    })
  })

  it('records attendance and shows success', async () => {
    let postCalled = false
    setToken('valid')
    mockFetch.mockImplementation((url: string, opts?: RequestInit) => {
      if (url.includes('/auth/me')) return Promise.resolve(jsonResponse(adminUser))
      if (url.includes('/class-types')) return Promise.resolve(jsonResponse(sampleClassTypes))
      if (url.includes('/classes/10/attendance')) {
        if (opts?.method === 'POST') {
          postCalled = true
          return Promise.resolve(jsonResponse({ id: 3, class_id: 10, user_id: 7, checked_in_at: '2026-03-15T14:15:00Z' }))
        }
        // GET attendance: return updated list after post
        if (postCalled) {
          return Promise.resolve(jsonResponse([{ id: 3, class_id: 10, user_id: 7, checked_in_at: '2026-03-15T14:15:00Z' }]))
        }
        return Promise.resolve(jsonResponse([]))
      }
      if (url.includes('/classes')) return Promise.resolve(jsonResponse(sampleClasses))
      return Promise.resolve(jsonResponse([]))
    })

    render(
      <MemoryRouter>
        <AuthProvider>
          <AttendancePage />
        </AuthProvider>
      </MemoryRouter>,
    )
    const user = userEvent.setup()

    await waitFor(() => {
      const select = screen.getByLabelText('Class') as HTMLSelectElement
      expect(select.options.length).toBeGreaterThan(1)
    })

    await user.selectOptions(screen.getByLabelText('Class'), '10')

    await waitFor(() => {
      expect(screen.getByText('No attendance recorded for this class.')).toBeInTheDocument()
    })

    await user.type(screen.getByLabelText('Member ID'), '7')
    await user.click(screen.getByRole('button', { name: 'Check In' }))

    await waitFor(() => {
      expect(screen.getByRole('status')).toHaveTextContent('Attendance recorded successfully')
    })

    // Verify the POST call was made
    const postCall = (mockFetch.mock.calls as [string, RequestInit][]).find(
      (call) => call[1]?.method === 'POST',
    )
    expect(postCall).toBeDefined()
    expect(postCall![0]).toContain('/classes/10/attendance')
    const body = JSON.parse(postCall![1].body as string)
    expect(body.user_id).toBe(7)
  })

  it('shows error on check-in failure', async () => {
    setToken('valid')
    mockFetch.mockImplementation((url: string, opts?: RequestInit) => {
      if (url.includes('/auth/me')) return Promise.resolve(jsonResponse(adminUser))
      if (url.includes('/class-types')) return Promise.resolve(jsonResponse(sampleClassTypes))
      if (url.includes('/classes/10/attendance')) {
        if (opts?.method === 'POST') {
          return Promise.resolve(jsonResponse({ error: 'Already checked in' }, 409))
        }
        return Promise.resolve(jsonResponse([]))
      }
      if (url.includes('/classes')) return Promise.resolve(jsonResponse(sampleClasses))
      return Promise.resolve(jsonResponse([]))
    })

    render(
      <MemoryRouter>
        <AuthProvider>
          <AttendancePage />
        </AuthProvider>
      </MemoryRouter>,
    )
    const user = userEvent.setup()

    await waitFor(() => {
      const select = screen.getByLabelText('Class') as HTMLSelectElement
      expect(select.options.length).toBeGreaterThan(1)
    })

    await user.selectOptions(screen.getByLabelText('Class'), '10')

    await waitFor(() => {
      expect(screen.getByText('No attendance recorded for this class.')).toBeInTheDocument()
    })

    await user.type(screen.getByLabelText('Member ID'), '5')
    await user.click(screen.getByRole('button', { name: 'Check In' }))

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Already checked in')
    })
  })

  it('shows error when classes fail to load', async () => {
    setToken('valid')
    mockFetch.mockImplementation((url: string) => {
      if (url.includes('/auth/me')) return Promise.resolve(jsonResponse(adminUser))
      return Promise.reject(new Error('Network error'))
    })

    render(
      <MemoryRouter>
        <AuthProvider>
          <AttendancePage />
        </AuthProvider>
      </MemoryRouter>,
    )

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Failed to load classes')
    })
  })
})

describe('MyAttendancePage', () => {
  it('loads and displays attendance history on mount', async () => {
    renderMyAttendancePage()

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'My Attendance' })).toBeInTheDocument()
    })

    await waitFor(() => {
      const rows = screen.getAllByRole('row')
      // header + 2 records
      expect(rows).toHaveLength(3)
    })

    expect(screen.getAllByText('10')).toHaveLength(2)
  })

  it('shows empty state when no attendance', async () => {
    setToken('valid')
    mockFetch.mockImplementation((url: string) => {
      if (url.includes('/auth/me')) return Promise.resolve(jsonResponse(memberUser))
      if (url.includes('/attendance')) return Promise.resolve(jsonResponse([]))
      return Promise.resolve(jsonResponse([]))
    })

    render(
      <MemoryRouter>
        <AuthProvider>
          <MyAttendancePage />
        </AuthProvider>
      </MemoryRouter>,
    )

    await waitFor(() => {
      expect(screen.getByText('No attendance history found.')).toBeInTheDocument()
    })
  })

  it('shows error on load failure', async () => {
    setToken('valid')
    mockFetch.mockImplementation((url: string) => {
      if (url.includes('/auth/me')) return Promise.resolve(jsonResponse(memberUser))
      if (url.includes('/attendance')) return Promise.resolve(jsonResponse({ error: 'Forbidden' }, 403))
      return Promise.resolve(jsonResponse([]))
    })

    render(
      <MemoryRouter>
        <AuthProvider>
          <MyAttendancePage />
        </AuthProvider>
      </MemoryRouter>,
    )

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Forbidden')
    })
  })
})
