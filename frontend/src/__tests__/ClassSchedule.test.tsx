import { render, screen, waitFor } from '@testing-library/react'
import ClassSchedule, { sortClassesByDate } from '../pages/ClassSchedule'
import type { Class } from '../api/classes'

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

const classTypes = [
  { id: 1, name: 'Yoga', description: 'Yoga class', created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:00Z' },
  { id: 2, name: 'Karate', description: 'Karate class', created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:00Z' },
]

const classes: Class[] = [
  { id: 1, class_type_id: 2, instructor_id: 1, start_time: '2026-03-15T14:00:00Z', duration_minutes: 60, capacity: 20, created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:00Z' },
  { id: 2, class_type_id: 1, instructor_id: 1, start_time: '2026-03-10T09:00:00Z', duration_minutes: 45, capacity: 15, created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:00Z' },
  { id: 3, class_type_id: 1, instructor_id: 1, start_time: '2026-03-20T10:00:00Z', duration_minutes: 90, capacity: 10, created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:00Z' },
]

describe('sortClassesByDate', () => {
  it('sorts classes by start_time ascending', () => {
    const sorted = sortClassesByDate(classes)
    expect(sorted[0].id).toBe(2) // Mar 10
    expect(sorted[1].id).toBe(1) // Mar 15
    expect(sorted[2].id).toBe(3) // Mar 20
  })

  it('does not mutate the original array', () => {
    const original = [...classes]
    sortClassesByDate(classes)
    expect(classes).toEqual(original)
  })

  it('handles empty array', () => {
    expect(sortClassesByDate([])).toEqual([])
  })

  it('handles single element', () => {
    const single = [classes[0]]
    const sorted = sortClassesByDate(single)
    expect(sorted).toHaveLength(1)
    expect(sorted[0].id).toBe(1)
  })
})

describe('ClassSchedule', () => {
  it('shows loading state initially', () => {
    mockFetch.mockReturnValue(new Promise(() => {})) // never resolves
    render(<ClassSchedule />)
    expect(screen.getByText('Loading schedule...')).toBeInTheDocument()
  })

  it('renders classes sorted by date', async () => {
    mockFetch.mockImplementation((url: string) => {
      if (url.includes('/class-types')) return Promise.resolve(jsonResponse(classTypes))
      if (url.includes('/classes')) return Promise.resolve(jsonResponse(classes))
      return Promise.resolve(jsonResponse([], 404))
    })

    render(<ClassSchedule />)

    await waitFor(() => {
      expect(screen.queryByText('Loading schedule...')).not.toBeInTheDocument()
    })

    const rows = screen.getAllByRole('row')
    // header + 3 data rows
    expect(rows).toHaveLength(4)

    // Check class type names are resolved
    expect(screen.getAllByText('Yoga')).toHaveLength(2)
    expect(screen.getByText('Karate')).toBeInTheDocument()

    // Check durations render
    expect(screen.getByText('45 min')).toBeInTheDocument()
    expect(screen.getByText('60 min')).toBeInTheDocument()
    expect(screen.getByText('90 min')).toBeInTheDocument()

    // Check capacities render
    expect(screen.getByText('15')).toBeInTheDocument()
    expect(screen.getByText('20')).toBeInTheDocument()
    expect(screen.getByText('10')).toBeInTheDocument()

    // Verify sort order: first data row should be Mar 10 (Yoga, 45 min)
    // Second data row should be Mar 15 (Karate, 60 min)
    // Third data row should be Mar 20 (Yoga, 90 min)
    const cells = rows[1].querySelectorAll('td')
    expect(cells[2].textContent).toBe('Yoga')     // first by date is class_type_id=1 (Yoga)
    expect(cells[3].textContent).toBe('45 min')

    const cells2 = rows[2].querySelectorAll('td')
    expect(cells2[2].textContent).toBe('Karate')
    expect(cells2[3].textContent).toBe('60 min')

    const cells3 = rows[3].querySelectorAll('td')
    expect(cells3[2].textContent).toBe('Yoga')
    expect(cells3[3].textContent).toBe('90 min')
  })

  it('shows empty state when no classes', async () => {
    mockFetch.mockImplementation((url: string) => {
      if (url.includes('/class-types')) return Promise.resolve(jsonResponse([]))
      if (url.includes('/classes')) return Promise.resolve(jsonResponse([]))
      return Promise.resolve(jsonResponse([], 404))
    })

    render(<ClassSchedule />)

    await waitFor(() => {
      expect(screen.getByText('No classes scheduled.')).toBeInTheDocument()
    })
  })

  it('shows error on fetch failure', async () => {
    mockFetch.mockRejectedValue(new Error('Network error'))

    render(<ClassSchedule />)

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent('Failed to load schedule')
    })
  })

  it('shows Unknown for missing class type', async () => {
    const classWithUnknownType: Class[] = [
      { id: 1, class_type_id: 999, instructor_id: 1, start_time: '2026-03-15T14:00:00Z', duration_minutes: 60, capacity: 20, created_at: '2026-01-01T00:00:00Z', updated_at: '2026-01-01T00:00:00Z' },
    ]

    mockFetch.mockImplementation((url: string) => {
      if (url.includes('/class-types')) return Promise.resolve(jsonResponse([]))
      if (url.includes('/classes')) return Promise.resolve(jsonResponse(classWithUnknownType))
      return Promise.resolve(jsonResponse([], 404))
    })

    render(<ClassSchedule />)

    await waitFor(() => {
      expect(screen.getByText('Unknown')).toBeInTheDocument()
    })
  })
})
