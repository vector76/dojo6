import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { AuthProvider } from '../hooks/useAuth'
import { ProtectedRoute } from '../components/ProtectedRoute'
import { Layout } from '../components/Layout'
import { MemberListPage } from '../pages/MemberListPage'
import { MemberDetailPage } from '../pages/MemberDetailPage'
import { AddMemberPage } from '../pages/AddMemberPage'
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
  name: 'Admin User',
  phone: '555-0100',
  role: 'admin' as const,
}

const instructorUser = {
  id: '2',
  email: 'instructor@test.com',
  name: 'Instructor',
  phone: '555-0200',
  role: 'instructor' as const,
}

const membersList = [
  {
    id: '1',
    name: 'Admin User',
    email: 'admin@test.com',
    phone: '555-0100',
    role: 'admin',
    membership_type: '',
    membership_status: '',
    emergency_contact: '',
    join_date: '',
    expected_balance: 0,
  },
  {
    id: '3',
    name: 'John Doe',
    email: 'john@test.com',
    phone: '555-0300',
    role: 'user',
    membership_type: 'monthly',
    membership_status: 'active',
    emergency_contact: 'Jane Doe',
    join_date: '2026-01-15',
    expected_balance: 50,
  },
  {
    id: '4',
    name: 'Jane Smith',
    email: 'jane@test.com',
    phone: '555-0400',
    role: 'user',
    membership_type: 'annual',
    membership_status: 'inactive',
    emergency_contact: '',
    join_date: '2025-06-01',
    expected_balance: 0,
  },
]

function renderWithAuth(initialRoute: string) {
  // First call: /auth/me, subsequent calls are page-specific
  setToken('valid')

  return render(
    <MemoryRouter initialEntries={[initialRoute]}>
      <AuthProvider>
        <Routes>
          <Route
            element={
              <ProtectedRoute>
                <Layout />
              </ProtectedRoute>
            }
          >
            <Route path="/members" element={<MemberListPage />} />
            <Route path="/members/new" element={<AddMemberPage />} />
            <Route path="/members/:id" element={<MemberDetailPage />} />
          </Route>
        </Routes>
      </AuthProvider>
    </MemoryRouter>,
  )
}

describe('MemberListPage', () => {
  it('renders member list', async () => {
    mockFetch
      .mockResolvedValueOnce(jsonResponse(adminUser))     // /auth/me
      .mockResolvedValueOnce(jsonResponse(membersList))    // /users

    renderWithAuth('/members')

    await waitFor(() => {
      expect(screen.getByText('John Doe')).toBeInTheDocument()
    })
    expect(screen.getByText('Jane Smith')).toBeInTheDocument()
    // Admin User appears in both nav and table; check the table link
    const table = screen.getByRole('table')
    expect(within(table).getByText('Admin User')).toBeInTheDocument()
  })

  it('filters by search text', async () => {
    mockFetch
      .mockResolvedValueOnce(jsonResponse(adminUser))
      .mockResolvedValueOnce(jsonResponse(membersList))

    renderWithAuth('/members')

    await waitFor(() => {
      expect(screen.getByText('John Doe')).toBeInTheDocument()
    })

    const user = userEvent.setup()
    await user.type(screen.getByLabelText('Search members'), 'john')

    expect(screen.getByText('John Doe')).toBeInTheDocument()
    expect(screen.queryByText('Jane Smith')).not.toBeInTheDocument()
  })

  it('filters by role', async () => {
    mockFetch
      .mockResolvedValueOnce(jsonResponse(adminUser))
      .mockResolvedValueOnce(jsonResponse(membersList))

    renderWithAuth('/members')

    await waitFor(() => {
      expect(screen.getByText('John Doe')).toBeInTheDocument()
    })

    const user = userEvent.setup()
    await user.selectOptions(screen.getByLabelText('Filter by role'), 'admin')

    const table = screen.getByRole('table')
    expect(within(table).getByText('Admin User')).toBeInTheDocument()
    expect(screen.queryByText('John Doe')).not.toBeInTheDocument()
  })

  it('filters by membership status', async () => {
    mockFetch
      .mockResolvedValueOnce(jsonResponse(adminUser))
      .mockResolvedValueOnce(jsonResponse(membersList))

    renderWithAuth('/members')

    await waitFor(() => {
      expect(screen.getByText('John Doe')).toBeInTheDocument()
    })

    const user = userEvent.setup()
    await user.selectOptions(screen.getByLabelText('Filter by status'), 'active')

    expect(screen.getByText('John Doe')).toBeInTheDocument()
    expect(screen.queryByText('Jane Smith')).not.toBeInTheDocument()
  })

  it('shows "No members found" when filter matches nothing', async () => {
    mockFetch
      .mockResolvedValueOnce(jsonResponse(adminUser))
      .mockResolvedValueOnce(jsonResponse(membersList))

    renderWithAuth('/members')

    await waitFor(() => {
      expect(screen.getByText('John Doe')).toBeInTheDocument()
    })

    const user = userEvent.setup()
    await user.type(screen.getByLabelText('Search members'), 'zzzznotfound')

    expect(screen.getByText('No members found')).toBeInTheDocument()
  })

  it('shows Add Member link for admin', async () => {
    mockFetch
      .mockResolvedValueOnce(jsonResponse(adminUser))
      .mockResolvedValueOnce(jsonResponse(membersList))

    renderWithAuth('/members')

    await waitFor(() => {
      expect(screen.getByText('Add Member')).toBeInTheDocument()
    })
  })

  it('hides Add Member link for instructor', async () => {
    mockFetch
      .mockResolvedValueOnce(jsonResponse(instructorUser))
      .mockResolvedValueOnce(jsonResponse(membersList))

    renderWithAuth('/members')

    await waitFor(() => {
      expect(screen.getByText('John Doe')).toBeInTheDocument()
    })
    expect(screen.queryByText('Add Member')).not.toBeInTheDocument()
  })

  it('shows error on API failure', async () => {
    mockFetch
      .mockResolvedValueOnce(jsonResponse(adminUser))
      .mockResolvedValueOnce(jsonResponse({ error: 'Server error' }, 500))

    renderWithAuth('/members')

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Server error')
    })
  })
})

describe('MemberDetailPage', () => {
  it('renders member detail form', async () => {
    mockFetch
      .mockResolvedValueOnce(jsonResponse(adminUser))     // /auth/me
      .mockResolvedValueOnce(jsonResponse(membersList[1])) // /users/3

    renderWithAuth('/members/3')

    await waitFor(() => {
      expect(screen.getByLabelText('Name')).toHaveValue('John Doe')
    })
    expect(screen.getByLabelText('Email')).toHaveValue('john@test.com')
    expect(screen.getByLabelText('Phone')).toHaveValue('555-0300')
  })

  it('admin can save changes', async () => {
    const updatedMember = { ...membersList[1], name: 'John Updated' }
    mockFetch
      .mockResolvedValueOnce(jsonResponse(adminUser))
      .mockResolvedValueOnce(jsonResponse(membersList[1]))
      .mockResolvedValueOnce(jsonResponse(updatedMember))

    renderWithAuth('/members/3')

    await waitFor(() => {
      expect(screen.getByLabelText('Name')).toHaveValue('John Doe')
    })

    const user = userEvent.setup()
    await user.clear(screen.getByLabelText('Name'))
    await user.type(screen.getByLabelText('Name'), 'John Updated')
    await user.click(screen.getByRole('button', { name: 'Save Changes' }))

    await waitFor(() => {
      expect(screen.getByRole('status')).toHaveTextContent('Saved successfully')
    })
  })

  it('admin sees delete button', async () => {
    mockFetch
      .mockResolvedValueOnce(jsonResponse(adminUser))
      .mockResolvedValueOnce(jsonResponse(membersList[1]))

    renderWithAuth('/members/3')

    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'Delete Member' })).toBeInTheDocument()
    })
  })

  it('admin sees password reset form', async () => {
    mockFetch
      .mockResolvedValueOnce(jsonResponse(adminUser))
      .mockResolvedValueOnce(jsonResponse(membersList[1]))

    renderWithAuth('/members/3')

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Reset Password' })).toBeInTheDocument()
    })
    expect(screen.getByLabelText('New Password')).toBeInTheDocument()
  })

  it('admin sees expected balance form', async () => {
    mockFetch
      .mockResolvedValueOnce(jsonResponse(adminUser))
      .mockResolvedValueOnce(jsonResponse(membersList[1]))

    renderWithAuth('/members/3')

    await waitFor(() => {
      expect(screen.getByText('Set Expected Balance')).toBeInTheDocument()
    })
    expect(screen.getByLabelText('Expected Balance')).toBeInTheDocument()
  })

  it('instructor cannot see admin-only actions', async () => {
    mockFetch
      .mockResolvedValueOnce(jsonResponse(instructorUser))
      .mockResolvedValueOnce(jsonResponse(membersList[1]))

    renderWithAuth('/members/3')

    await waitFor(() => {
      expect(screen.getByLabelText('Name')).toHaveValue('John Doe')
    })
    expect(screen.queryByText('Delete Member')).not.toBeInTheDocument()
    expect(screen.queryByText('Reset Password')).not.toBeInTheDocument()
    expect(screen.queryByText('Set Expected Balance')).not.toBeInTheDocument()
  })

  it('instructor cannot edit fields', async () => {
    mockFetch
      .mockResolvedValueOnce(jsonResponse(instructorUser))
      .mockResolvedValueOnce(jsonResponse(membersList[1]))

    renderWithAuth('/members/3')

    await waitFor(() => {
      expect(screen.getByLabelText('Name')).toBeDisabled()
    })
    expect(screen.getByLabelText('Email')).toBeDisabled()
  })

  it('shows error on load failure', async () => {
    mockFetch
      .mockResolvedValueOnce(jsonResponse(adminUser))
      .mockResolvedValueOnce(jsonResponse({ error: 'Not found' }, 404))

    renderWithAuth('/members/999')

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Not found')
    })
  })
})

describe('AddMemberPage', () => {
  it('renders add member form', async () => {
    mockFetch.mockResolvedValueOnce(jsonResponse(adminUser))

    renderWithAuth('/members/new')

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Add Member' })).toBeInTheDocument()
    })
    expect(screen.getByLabelText('Name')).toBeInTheDocument()
    expect(screen.getByLabelText('Email')).toBeInTheDocument()
    expect(screen.getByLabelText('Password')).toBeInTheDocument()
  })

  it('submits form and navigates to member detail', async () => {
    const newMember = {
      id: '5',
      name: 'New User',
      email: 'new@test.com',
      phone: '555-0500',
      role: 'user',
      membership_type: 'monthly',
      membership_status: 'active',
      emergency_contact: '',
      join_date: '',
      expected_balance: 0,
    }
    mockFetch
      .mockResolvedValueOnce(jsonResponse(adminUser))   // /auth/me
      .mockResolvedValueOnce(jsonResponse(newMember))     // POST /users
      .mockResolvedValueOnce(jsonResponse(newMember))     // GET /users/5 (detail page load)

    renderWithAuth('/members/new')

    await waitFor(() => {
      expect(screen.getByLabelText('Name')).toBeInTheDocument()
    })

    const user = userEvent.setup()
    await user.type(screen.getByLabelText('Name'), 'New User')
    await user.type(screen.getByLabelText('Email'), 'new@test.com')
    await user.type(screen.getByLabelText('Phone'), '555-0500')
    await user.type(screen.getByLabelText('Password'), 'password123')
    await user.click(screen.getByRole('button', { name: 'Create Member' }))

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Member Detail' })).toBeInTheDocument()
    })
  })

  it('shows error on create failure', async () => {
    mockFetch
      .mockResolvedValueOnce(jsonResponse(adminUser))
      .mockResolvedValueOnce(jsonResponse({ error: 'Email already exists' }, 409))

    renderWithAuth('/members/new')

    await waitFor(() => {
      expect(screen.getByLabelText('Name')).toBeInTheDocument()
    })

    const user = userEvent.setup()
    await user.type(screen.getByLabelText('Name'), 'Dup User')
    await user.type(screen.getByLabelText('Email'), 'existing@test.com')
    await user.type(screen.getByLabelText('Phone'), '555-0600')
    await user.type(screen.getByLabelText('Password'), 'password123')
    await user.click(screen.getByRole('button', { name: 'Create Member' }))

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Email already exists')
    })
  })
})
