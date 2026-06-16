export class ApiError extends Error {
  code: string
  details: unknown

  constructor(code: string, message: string, details?: unknown) {
    super(message)
    this.name = 'ApiError'
    this.code = code
    this.details = details
  }
}

export async function apiFetch<T>(path: string, opts?: RequestInit): Promise<T> {
  const resp = await fetch(path, {
    ...opts,
    headers: {
      'Content-Type': 'application/json',
      ...opts?.headers,
    },
  })

  if (!resp.ok) {
    let code = 'INTERNAL'
    let message = `HTTP ${resp.status}`
    let details: unknown
    try {
      const body = await resp.json()
      if (body?.error) {
        code = body.error.code ?? code
        message = body.error.message ?? message
        details = body.error.details
      }
    } catch { /* ignore */ }
    throw new ApiError(code, message, details)
  }

  if (resp.status === 204) return undefined as T

  return resp.json()
}
