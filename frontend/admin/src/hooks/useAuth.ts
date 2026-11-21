import { useState, useEffect, useCallback } from 'react'
import { api, getAccessToken, clearTokens, setTokens } from '@pclsgu/shared'

export function useAuth() {
  const [isAuthed, setIsAuthed] = useState<boolean>(!!getAccessToken())
  const [permissions, setPermissions] = useState<string[]>([])

  const fetchPermissions = useCallback(async () => {
    try {
      const data = await api.get<{ permissions: string[] }>('/admin/auth/permissions')
      setPermissions(data.permissions || [])
    } catch {
      clearTokens()
      setIsAuthed(false)
    }
  }, [])

  useEffect(() => {
    if (isAuthed) {
      fetchPermissions()
    }
  }, [isAuthed, fetchPermissions])

  const login = useCallback(async (username: string, password: string) => {
    const data = await api.post<{ accessToken: string; refreshToken: string }>('/admin/login', {
      username,
      password,
    })
    setTokens(data.accessToken, data.refreshToken)
    setIsAuthed(true)
    await fetchPermissions()
  }, [fetchPermissions])

  const logout = useCallback(async () => {
    try {
      await api.post('/admin/logout')
    } catch {
      // ignore
    }
    clearTokens()
    setIsAuthed(false)
    setPermissions([])
  }, [])

  return { isAuthed, permissions, login, logout }
}