import { describe, it, expect, vi, beforeEach } from 'vitest';
import { request, ApiError } from '../api/client';

describe('API client', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('makes a GET request and returns JSON', async () => {
    const mockData = { id: '1', name: 'Test' };
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve(mockData),
    }));

    const result = await request('/processes');
    expect(result).toEqual(mockData);
    expect(fetch).toHaveBeenCalledWith('/api/v1/processes', expect.objectContaining({
      headers: expect.objectContaining({ 'Content-Type': 'application/json' }),
    }));
  });

  it('returns undefined for 204 responses', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      status: 204,
    }));

    const result = await request('/processes/1', { method: 'DELETE' });
    expect(result).toBeUndefined();
  });

  it('throws ApiError on non-ok response', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: false,
      status: 404,
      json: () => Promise.resolve({ error: 'resource not found' }),
    }));

    await expect(request('/processes/999')).rejects.toThrow(ApiError);
    await expect(request('/processes/999')).rejects.toThrow('resource not found');
  });

  it('handles non-JSON error bodies', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: false,
      status: 500,
      json: () => Promise.reject(new Error('not json')),
    }));

    await expect(request('/processes')).rejects.toThrow('Request failed with status 500');
  });
});
