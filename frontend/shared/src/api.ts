const BASE_URL = '/api'

let accessToken: string | null = null
let refreshToken: string | null = null

export function setTokens(access: string, refresh: string) {
  accessToken = access
  refreshToken = refresh
  sessionStorage.setItem('access_token', access)
  sessionStorage.setItem('refresh_token', refresh)
}

export function getAccessToken(): string | null {
  if (accessToken) return accessToken
  accessToken = sessionStorage.getItem('access_token')
  return accessToken
}

export function getRefreshToken(): string | null {
  if (refreshToken) return refreshToken
  refreshToken = sessionStorage.getItem('refresh_token')
  return refreshToken
}

export function clearTokens() {
  accessToken = null
  refreshToken = null
  sessionStorage.removeItem('access_token')
  sessionStorage.removeItem('refresh_token')
}

export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}

async function request<T = unknown>(
  method: string,
  path: string,
  body?: unknown,
  params?: Record<string, string>,
): Promise<T> {
  const url = new URL(BASE_URL + path, window.location.origin)
  if (params) {
    for (const [k, v] of Object.entries(params)) {
      url.searchParams.set(k, v)
    }
  }

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }
  const token = getAccessToken()
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const resp = await fetch(url.toString(), {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  })

  if (resp.status === 401 && refreshToken) {
    const refreshResp = await fetch(`${BASE_URL}/admin/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refreshToken }),
    })
    if (refreshResp.ok) {
      const refreshData = await refreshResp.json()
      setTokens(refreshData.data.accessToken, refreshData.data.refreshToken)
      headers['Authorization'] = `Bearer ${getAccessToken()}`
      const retryResp = await fetch(url.toString(), {
        method,
        headers,
        body: body ? JSON.stringify(body) : undefined,
      })
      if (!retryResp.ok) throw new Error(`HTTP ${retryResp.status}`)
      const retryJson: ApiResponse<T> = await retryResp.json()
      if (retryJson.code !== 0) throw new Error(retryJson.message)
      return retryJson.data
    }
    clearTokens()
    window.location.href = '/admin'
    throw new Error('session expired')
  }

  if (!resp.ok) throw new Error(`HTTP ${resp.status}`)
  const json: ApiResponse<T> = await resp.json()
  if (json.code !== 0) throw new Error(json.message)
  return json.data
}

export const api = {
  get: <T = unknown>(path: string, params?: Record<string, string>) =>
    request<T>('GET', path, undefined, params),
  post: <T = unknown>(path: string, body?: unknown) =>
    request<T>('POST', path, body),
  put: <T = unknown>(path: string, body?: unknown) =>
    request<T>('PUT', path, body),
  delete: <T = unknown>(path: string) =>
    request<T>('DELETE', path),
  upload: async <T = unknown>(path: string, formData: FormData) => {
    const headers: Record<string, string> = {}
    const token = getAccessToken()
    if (token) headers['Authorization'] = `Bearer ${token}`
    const resp = await fetch(`${BASE_URL}${path}`, {
      method: 'POST',
      headers,
      body: formData,
    })
    if (!resp.ok) throw new Error(`HTTP ${resp.status}`)
    const json: ApiResponse<T> = await resp.json()
    if (json.code !== 0) throw new Error(json.message)
    return json.data
  },
}