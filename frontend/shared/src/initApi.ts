const BASE_URL = '/api/init'

async function initRequest<T = unknown>(
  method: string,
  path: string,
  body?: unknown,
): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }
  const resp = await fetch(BASE_URL + path, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  })
  const json = await resp.json()
  if (json.code !== 'OK') throw new Error(json.message)
  return json.data
}

export const initApi = {
  getStatus: () => initRequest<{ isFirstRun: boolean; initStatus: string }>('GET', '/status'),
  validate: (step: number, data: Record<string, unknown>) =>
    initRequest<{ valid: boolean; errors: string[] }>('POST', '/validate', { step, data }),
  generateJwt: () => initRequest<{ jwtSecret: string }>('POST', '/generate-jwt'),
  testDb: (data: { dbDriver: string; sqlitePath?: string; dbHost?: string; dbPort?: string; dbName?: string; dbUser?: string; dbPassword?: string }) =>
    initRequest<{ success: boolean }>('POST', '/test-db', data),
  testStorage: (data: { storageDriver: string; storagePath?: string }) =>
    initRequest<{ success: boolean }>('POST', '/test-storage', data),
  complete: (config: Record<string, unknown>) =>
    initRequest<{ success: boolean; message: string }>('POST', '/complete', config),
}