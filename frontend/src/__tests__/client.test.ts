import {
  apiFetch,
  ApiError,
  clearToken,
  getToken,
  setOnUnauthorized,
  setToken,
} from '../api/client'

const mockFetch = vi.fn()
vi.stubGlobal('fetch', mockFetch)

beforeEach(() => {
  mockFetch.mockReset()
  localStorage.clear()
  setOnUnauthorized(null)
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

describe('token management', () => {
  it('stores and retrieves token', () => {
    expect(getToken()).toBeNull()
    setToken('abc123')
    expect(getToken()).toBe('abc123')
  })

  it('clears token', () => {
    setToken('abc123')
    clearToken()
    expect(getToken()).toBeNull()
  })
})

describe('apiFetch', () => {
  it('prepends /api to the path', async () => {
    mockFetch.mockResolvedValue(jsonResponse({ ok: true }))
    await apiFetch('/health')
    expect(mockFetch).toHaveBeenCalledWith(
      '/api/health',
      expect.any(Object),
    )
  })

  it('attaches Authorization header when token exists', async () => {
    setToken('mytoken')
    mockFetch.mockResolvedValue(jsonResponse({ ok: true }))
    await apiFetch('/health')

    const [, options] = mockFetch.mock.calls[0]
    const headers = new Headers(options.headers)
    expect(headers.get('Authorization')).toBe('Bearer mytoken')
  })

  it('does not attach Authorization header without token', async () => {
    mockFetch.mockResolvedValue(jsonResponse({ ok: true }))
    await apiFetch('/health')

    const [, options] = mockFetch.mock.calls[0]
    const headers = new Headers(options.headers)
    expect(headers.get('Authorization')).toBeNull()
  })

  it('sets Content-Type for JSON body', async () => {
    mockFetch.mockResolvedValue(jsonResponse({ ok: true }))
    await apiFetch('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email: 'a@b.com' }),
    })

    const [, options] = mockFetch.mock.calls[0]
    const headers = new Headers(options.headers)
    expect(headers.get('Content-Type')).toBe('application/json')
  })

  it('returns parsed JSON on success', async () => {
    mockFetch.mockResolvedValue(jsonResponse({ name: 'test' }))
    const result = await apiFetch<{ name: string }>('/test')
    expect(result).toEqual({ name: 'test' })
  })

  it('returns undefined for 204 responses', async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      status: 204,
      statusText: 'No Content',
      headers: new Headers(),
    } as Response)
    const result = await apiFetch('/test')
    expect(result).toBeUndefined()
  })

  it('throws ApiError on non-ok response', async () => {
    mockFetch.mockResolvedValue(
      jsonResponse({ error: 'Not found' }, 404),
    )
    await expect(apiFetch('/missing')).rejects.toThrow(ApiError)
    await expect(
      apiFetch('/missing').catch((e: ApiError) => {
        expect(e.status).toBe(404)
        expect(e.message).toBe('Not found')
        throw e
      }),
    ).rejects.toThrow()
  })

  it('clears token and calls onUnauthorized on 401', async () => {
    setToken('expired')
    const handler = vi.fn()
    setOnUnauthorized(handler)

    mockFetch.mockResolvedValue(
      jsonResponse({ error: 'Unauthorized' }, 401),
    )
    await expect(apiFetch('/protected')).rejects.toThrow(ApiError)

    expect(getToken()).toBeNull()
    expect(handler).toHaveBeenCalledOnce()
  })
})
