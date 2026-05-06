const API_BASE = '/api/v1';

export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

export async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const resp = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options?.headers,
    },
  });

  if (!resp.ok) {
    let msg = `Request failed with status ${resp.status}`;
    try {
      const body = await resp.json();
      if (body.error) msg = body.error;
    } catch {
      // ignore parse errors
    }
    throw new ApiError(resp.status, msg);
  }

  if (resp.status === 204) return undefined as T;
  return resp.json();
}
