import { getAccessToken } from '@pclsgu/shared'

export async function fetchOpenApiSpec(): Promise<Record<string, unknown>> {
  const resp = await fetch('/swagger/doc.json')
  if (!resp.ok) throw new Error(`HTTP ${resp.status}`)
  return resp.json()
}

export async function fetchMarkdownDoc(): Promise<string> {
  const resp = await fetch('/docs/api.md')
  if (!resp.ok) throw new Error(`HTTP ${resp.status}`)
  return resp.text()
}

export async function regenerateDocs(): Promise<{ success: boolean; pathCount: number }> {
  const token = getAccessToken()
  const resp = await fetch('/api/admin/docs/regenerate', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
  })
  if (!resp.ok) throw new Error(`HTTP ${resp.status}`)
  const json = await resp.json()
  return json.data
}

export interface TryApiResult {
  status: number
  body: string
  latencyMs: number
}

export async function tryApiCall(
  method: string,
  path: string,
  params?: Record<string, string>,
  body?: unknown,
): Promise<TryApiResult> {
  const token = getAccessToken()
  const url = new URL(path, window.location.origin)
  if (params) {
    for (const [k, v] of Object.entries(params)) {
      url.searchParams.set(k, v)
    }
  }
  const headers: Record<string, string> = {}
  if (token) headers['Authorization'] = `Bearer ${token}`
  if (body) headers['Content-Type'] = 'application/json'

  const start = performance.now()
  const resp = await fetch(url.toString(), {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  })
  const latencyMs = Math.round(performance.now() - start)
  const text = await resp.text()
  return { status: resp.status, body: text, latencyMs }
}