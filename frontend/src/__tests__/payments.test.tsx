import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { AuthProvider } from '../hooks/useAuth'
import { PaymentsPage } from '../pages/PaymentsPage'
import { MyPaymentsPage } from '../pages/MyPaymentsPage'
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

const samplePayments = [
  {
    id: 2,
    user_id: 5,
    amount: 75.0,
    date: '2026-02-10',
    note: 'Drop-in',
    recorded_by: 1,
    created_at: '2026-02-10T14:20:00Z',
    updated_at: '2026-02-10T14:20:00Z',
  },
  {
    id: 1,
    user_id: 5,
    amount: 50.0,
    date: '2026-02-01',
    note: 'Monthly dues',
    recorded_by: 1,
    created_at: '2026-02-01T09:00:00Z',
    updated_at: '2026-02-01T09:00:00Z',
  },
]

const sampleBalance = {
  user_id: 5,
  expected_balance: 200.0,
  total_payments: 125.0,
  outstanding_balance: 75.0,
}

function getById(id: string): HTMLElement {
  const el = document.getElementById(id)
  if (!el) throw new Error(`Element #${id} not found`)
  return el
}

function renderPaymentsPage() {
  setToken('valid')
  mockFetch.mockResolvedValueOnce(jsonResponse(adminUser))
  return render(
    <MemoryRouter>
      <AuthProvider>
        <PaymentsPage />
      </AuthProvider>
    </MemoryRouter>,
  )
}

function renderMyPaymentsPage() {
  setToken('valid')
  mockFetch.mockResolvedValueOnce(jsonResponse(memberUser))
  mockFetch.mockResolvedValueOnce(jsonResponse(samplePayments))
  mockFetch.mockResolvedValueOnce(jsonResponse(sampleBalance))
  return render(
    <MemoryRouter>
      <AuthProvider>
        <MyPaymentsPage />
      </AuthProvider>
    </MemoryRouter>,
  )
}

describe('PaymentsPage', () => {
  it('renders record payment form', async () => {
    renderPaymentsPage()
    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'Record Payment' })).toBeInTheDocument()
    })
    expect(getById('pay-user-id')).toBeInTheDocument()
    expect(screen.getByLabelText('Amount')).toBeInTheDocument()
    expect(screen.getByLabelText('Date')).toBeInTheDocument()
    expect(screen.getByLabelText('Note')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Record Payment' })).toBeInTheDocument()
  })

  it('submits payment and shows success', async () => {
    renderPaymentsPage()
    const newPayment = {
      id: 3,
      user_id: 5,
      amount: 100.0,
      date: '2026-02-15',
      note: 'Test',
      recorded_by: 1,
      created_at: '2026-02-15T10:00:00Z',
      updated_at: '2026-02-15T10:00:00Z',
    }
    mockFetch.mockResolvedValueOnce(jsonResponse(newPayment, 200))

    const user = userEvent.setup()
    await user.type(getById('pay-user-id'), '5')
    await user.clear(screen.getByLabelText('Amount'))
    await user.type(screen.getByLabelText('Amount'), '100')
    await user.type(screen.getByLabelText('Note'), 'Test')
    await user.click(screen.getByRole('button', { name: 'Record Payment' }))

    await waitFor(() => {
      expect(screen.getByRole('status')).toHaveTextContent('Payment recorded successfully')
    })
  })

  it('shows error on payment failure', async () => {
    renderPaymentsPage()
    mockFetch.mockResolvedValueOnce(
      jsonResponse({ error: 'User not found' }, 404),
    )

    const user = userEvent.setup()
    await user.type(getById('pay-user-id'), '999')
    await user.clear(screen.getByLabelText('Amount'))
    await user.type(screen.getByLabelText('Amount'), '50')
    await user.click(screen.getByRole('button', { name: 'Record Payment' }))

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('User not found')
    })
  })

  it('displays payment history and balance for a member', async () => {
    renderPaymentsPage()
    mockFetch.mockResolvedValueOnce(jsonResponse(samplePayments))
    mockFetch.mockResolvedValueOnce(jsonResponse(sampleBalance))

    const user = userEvent.setup()
    await user.type(getById('hist-user-id'), '5')
    await user.click(screen.getByRole('button', { name: 'View History' }))

    await waitFor(() => {
      expect(screen.getByTestId('balance-summary')).toBeInTheDocument()
    })
    expect(screen.getByText('Expected: $200.00')).toBeInTheDocument()
    expect(screen.getByText('Total Paid: $125.00')).toBeInTheDocument()
    expect(screen.getByText('Outstanding: $75.00')).toBeInTheDocument()

    expect(screen.getByText('$75.00')).toBeInTheDocument()
    expect(screen.getByText('Drop-in')).toBeInTheDocument()
    expect(screen.getByText('$50.00')).toBeInTheDocument()
    expect(screen.getByText('Monthly dues')).toBeInTheDocument()
  })
})

describe('MyPaymentsPage', () => {
  it('loads and displays payment history on mount', async () => {
    renderMyPaymentsPage()

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: 'My Payments' })).toBeInTheDocument()
    })

    await waitFor(() => {
      expect(screen.getByTestId('balance-summary')).toBeInTheDocument()
    })
    expect(screen.getByText('Expected: $200.00')).toBeInTheDocument()
    expect(screen.getByText('Outstanding: $75.00')).toBeInTheDocument()

    expect(screen.getByText('Drop-in')).toBeInTheDocument()
    expect(screen.getByText('Monthly dues')).toBeInTheDocument()
  })

  it('shows empty state when no payments', async () => {
    setToken('valid')
    mockFetch.mockResolvedValueOnce(jsonResponse(memberUser))
    mockFetch.mockResolvedValueOnce(jsonResponse([]))
    mockFetch.mockResolvedValueOnce(
      jsonResponse({ user_id: 5, expected_balance: 0, total_payments: 0, outstanding_balance: 0 }),
    )

    render(
      <MemoryRouter>
        <AuthProvider>
          <MyPaymentsPage />
        </AuthProvider>
      </MemoryRouter>,
    )

    await waitFor(() => {
      expect(screen.getByText('No payment history found.')).toBeInTheDocument()
    })
  })

  it('shows error on load failure', async () => {
    setToken('valid')
    mockFetch.mockResolvedValueOnce(jsonResponse(memberUser))
    // Both parallel requests need mocks; first fails, second also fails
    mockFetch.mockResolvedValueOnce(jsonResponse({ error: 'Forbidden' }, 403))
    mockFetch.mockResolvedValueOnce(jsonResponse({ error: 'Forbidden' }, 403))

    render(
      <MemoryRouter>
        <AuthProvider>
          <MyPaymentsPage />
        </AuthProvider>
      </MemoryRouter>,
    )

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Forbidden')
    })
  })
})
