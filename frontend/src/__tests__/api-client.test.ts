import { describe, it, expect, vi, beforeEach } from 'vitest'
import { apiFetch, ApiError } from '../api/client'

describe('apiFetch', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
  })

  it('returns parsed JSON on success', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ id: '123' }),
    })
    const result = await apiFetch<{ id: string }>('/api/v1/test')
    expect(result).toEqual({ id: '123' })
  })

  it('throws ApiError on non-2xx with error body', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 404,
      json: async () => ({ error: { code: 'NOT_FOUND', message: 'not found' } }),
    })
    await expect(apiFetch('/api/v1/missing')).rejects.toMatchObject({
      name: 'ApiError',
      code: 'NOT_FOUND',
      message: 'not found',
    })
  })

  it('throws ApiError on non-2xx without body', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 500,
      json: async () => { throw new Error('no body') },
    })
    await expect(apiFetch('/api/v1/crash')).rejects.toBeInstanceOf(ApiError)
  })

  it('returns undefined for 204', async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 204,
    })
    const result = await apiFetch('/api/v1/delete')
    expect(result).toBeUndefined()
  })
})
