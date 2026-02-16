import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import ClassTypes from '../pages/ClassTypes'

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

const sampleClassTypes = [
  { id: 1, name: 'Yoga', description: 'Relaxing yoga', created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:00Z' },
  { id: 2, name: 'Karate', description: 'Traditional karate', created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:00Z' },
]

describe('ClassTypes', () => {
  it('shows loading state initially', () => {
    mockFetch.mockReturnValue(new Promise(() => {}))
    render(<ClassTypes />)
    expect(screen.getByText('Loading class types...')).toBeInTheDocument()
  })

  it('renders class types in a table', async () => {
    mockFetch.mockResolvedValue(jsonResponse(sampleClassTypes))

    render(<ClassTypes />)

    await waitFor(() => {
      expect(screen.queryByText('Loading class types...')).not.toBeInTheDocument()
    })

    expect(screen.getByText('Yoga')).toBeInTheDocument()
    expect(screen.getByText('Relaxing yoga')).toBeInTheDocument()
    expect(screen.getByText('Karate')).toBeInTheDocument()
    expect(screen.getByText('Traditional karate')).toBeInTheDocument()
  })

  it('shows empty state when no class types', async () => {
    mockFetch.mockResolvedValue(jsonResponse([]))

    render(<ClassTypes />)

    await waitFor(() => {
      expect(screen.getByText('No class types yet.')).toBeInTheDocument()
    })
  })

  it('shows error on fetch failure', async () => {
    mockFetch.mockRejectedValue(new Error('Network error'))

    render(<ClassTypes />)

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Failed to load class types')
    })
  })

  it('creates a new class type', async () => {
    const user = userEvent.setup()

    // First call: initial list (empty), second: POST create, third: refreshed list
    mockFetch
      .mockResolvedValueOnce(jsonResponse([]))
      .mockResolvedValueOnce(jsonResponse({ id: 1, name: 'Pilates', description: 'Core workout', created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:00Z' }))
      .mockResolvedValueOnce(jsonResponse([{ id: 1, name: 'Pilates', description: 'Core workout', created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:00Z' }]))

    render(<ClassTypes />)

    await waitFor(() => {
      expect(screen.getByText('No class types yet.')).toBeInTheDocument()
    })

    await user.type(screen.getByLabelText('Name'), 'Pilates')
    await user.type(screen.getByLabelText('Description'), 'Core workout')
    await user.click(screen.getByText('Create'))

    await waitFor(() => {
      expect(screen.getByText('Pilates')).toBeInTheDocument()
    })

    // Verify the POST call
    const postCall = mockFetch.mock.calls.find(
      (call: [string, RequestInit]) => {
        const opts = call[1] as RequestInit | undefined
        return opts?.method === 'POST'
      },
    )
    expect(postCall).toBeDefined()
    const body = JSON.parse(postCall![1].body as string)
    expect(body.name).toBe('Pilates')
    expect(body.description).toBe('Core workout')
  })

  it('edits an existing class type', async () => {
    const user = userEvent.setup()

    mockFetch
      .mockResolvedValueOnce(jsonResponse(sampleClassTypes))
      .mockResolvedValueOnce(jsonResponse({ ...sampleClassTypes[0], name: 'Hot Yoga', description: 'Heated yoga' }))
      .mockResolvedValueOnce(jsonResponse([{ ...sampleClassTypes[0], name: 'Hot Yoga', description: 'Heated yoga' }, sampleClassTypes[1]]))

    render(<ClassTypes />)

    await waitFor(() => {
      expect(screen.getByText('Yoga')).toBeInTheDocument()
    })

    // Click edit on Yoga
    const editButtons = screen.getAllByText('Edit')
    await user.click(editButtons[0])

    // Form should be populated
    expect(screen.getByLabelText('Name')).toHaveValue('Yoga')
    expect(screen.getByLabelText('Description')).toHaveValue('Relaxing yoga')
    expect(screen.getByText('Update')).toBeInTheDocument()
    expect(screen.getByText('Cancel')).toBeInTheDocument()

    // Clear and type new values
    await user.clear(screen.getByLabelText('Name'))
    await user.type(screen.getByLabelText('Name'), 'Hot Yoga')
    await user.clear(screen.getByLabelText('Description'))
    await user.type(screen.getByLabelText('Description'), 'Heated yoga')
    await user.click(screen.getByText('Update'))

    await waitFor(() => {
      expect(screen.getByText('Hot Yoga')).toBeInTheDocument()
    })
  })

  it('cancels editing and resets form', async () => {
    const user = userEvent.setup()

    mockFetch.mockResolvedValue(jsonResponse(sampleClassTypes))

    render(<ClassTypes />)

    await waitFor(() => {
      expect(screen.getByText('Yoga')).toBeInTheDocument()
    })

    const editButtons = screen.getAllByText('Edit')
    await user.click(editButtons[0])

    expect(screen.getByText('Update')).toBeInTheDocument()

    await user.click(screen.getByText('Cancel'))

    expect(screen.getByText('Create')).toBeInTheDocument()
    expect(screen.getByLabelText('Name')).toHaveValue('')
    expect(screen.getByLabelText('Description')).toHaveValue('')
  })

  it('deletes a class type', async () => {
    const user = userEvent.setup()

    mockFetch
      .mockResolvedValueOnce(jsonResponse(sampleClassTypes))
      .mockResolvedValueOnce(jsonResponse(undefined, 204))
      .mockResolvedValueOnce(jsonResponse([sampleClassTypes[1]]))

    render(<ClassTypes />)

    await waitFor(() => {
      expect(screen.getByText('Yoga')).toBeInTheDocument()
    })

    const deleteButtons = screen.getAllByText('Delete')
    await user.click(deleteButtons[0])

    await waitFor(() => {
      expect(screen.queryByText('Yoga')).not.toBeInTheDocument()
    })
    expect(screen.getByText('Karate')).toBeInTheDocument()
  })
})
