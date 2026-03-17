const BASE = '/api/v1'

function getToken(): string | null {
  return localStorage.getItem('access_token')
}

export class ApiError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

// Deduplicate concurrent refresh calls — all 401s share one in-flight refresh
let refreshPromise: Promise<string | null> | null = null

async function doRefresh(): Promise<string | null> {
  const refreshToken = localStorage.getItem('refresh_token')
  if (!refreshToken) return null
  try {
    const res = await fetch(`${BASE}/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    })
    if (!res.ok) return null
    const json = await res.json()
    const { access_token, refresh_token } = json.data
    localStorage.setItem('access_token', access_token)
    localStorage.setItem('refresh_token', refresh_token)
    return access_token
  } catch {
    return null
  }
}

function refreshOnce(): Promise<string | null> {
  if (!refreshPromise) {
    refreshPromise = doRefresh().finally(() => { refreshPromise = null })
  }
  return refreshPromise
}

async function request<T>(path: string, options: RequestInit = {}, retry = true): Promise<T> {
  const token = getToken()
  const isFormData = options.body instanceof FormData

  const headers: Record<string, string> = {
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...(isFormData ? {} : { 'Content-Type': 'application/json' }),
    ...(options.headers as Record<string, string> | undefined),
  }

  const res = await fetch(`${BASE}${path}`, { ...options, headers })

  if (res.status === 401 && retry && path !== '/auth/refresh') {
    const newToken = await refreshOnce()
    if (newToken) return request<T>(path, options, false)
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
  }

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: 'Request failed' }))
    throw new ApiError(res.status, body.error ?? 'Request failed')
  }

  if (res.status === 204) return undefined as T
  const json = await res.json()
  return json.data as T
}

export const api = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: 'POST', body: body ? JSON.stringify(body) : undefined }),
  patch: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: 'PATCH', body: body ? JSON.stringify(body) : undefined }),
  put: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: 'PUT', body: body ? JSON.stringify(body) : undefined }),
  putForm: <T>(path: string, form: FormData) =>
    request<T>(path, { method: 'PUT', body: form }),
  postForm: <T>(path: string, form: FormData) =>
    request<T>(path, { method: 'POST', body: form }),
  delete: <T>(path: string) => request<T>(path, { method: 'DELETE' }),
}
