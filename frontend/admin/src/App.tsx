import { useState, useEffect } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { initApi } from '@pclsgu/shared'
import { useAuth } from './hooks/useAuth'
import Login from './pages/Login'
import InitWizard from './pages/InitWizard'
import AdminLayout from './layouts/AdminLayout'
import SoftwareVersions from './pages/SoftwareVersions'
import PluginManage from './pages/PluginManage'
import Review from './pages/Review'
import IndexManage from './pages/IndexManage'
import Metrics from './pages/Metrics'
import AuditLog from './pages/AuditLog'
import Users from './pages/Users'
import Roles from './pages/Roles'
import ConfigPage from './pages/Config'
import ApiDoc from './pages/ApiDoc'

export default function App() {
  const [initChecked, setInitChecked] = useState(false)
  const [isFirstRun, setIsFirstRun] = useState(false)
  const { isAuthed, permissions } = useAuth()

  useEffect(() => {
    initApi.getStatus()
      .then((data) => {
        setIsFirstRun(data.isFirstRun)
        setInitChecked(true)
      })
      .catch(() => {
        setInitChecked(true)
      })
  }, [])

  if (!initChecked) {
    return <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh' }}>加载中...</div>
  }

  if (isFirstRun) {
    return (
      <BrowserRouter>
        <Routes>
          <Route path="/init" element={<InitWizard />} />
          <Route path="*" element={<Navigate to="/init" replace />} />
        </Routes>
      </BrowserRouter>
    )
  }

  if (!isAuthed) {
    return (
      <BrowserRouter>
        <Routes>
          <Route path="/admin/login" element={<Login />} />
          <Route path="*" element={<Navigate to="/admin/login" replace />} />
        </Routes>
      </BrowserRouter>
    )
  }

  const hasPerm = (perm: string) => permissions.includes('*') || permissions.includes(perm)

  return (
    <BrowserRouter>
      <Routes>
        <Route path="/admin" element={<AdminLayout />}>
          <Route index element={<Navigate to="/admin/software" replace />} />
          {hasPerm('software.publish') && <Route path="software" element={<SoftwareVersions />} />}
          {hasPerm('plugin.manage') && <Route path="plugins" element={<PluginManage />} />}
          {hasPerm('plugin.review') && <Route path="review" element={<Review />} />}
          {hasPerm('index.view') && <Route path="index" element={<IndexManage />} />}
          {hasPerm('metrics.view') && <Route path="metrics" element={<Metrics />} />}
          {hasPerm('audit.view') && <Route path="audit" element={<AuditLog />} />}
          {hasPerm('user.manage') && <Route path="users" element={<Users />} />}
          {hasPerm('user.manage') && <Route path="roles" element={<Roles />} />}
          {hasPerm('config.manage') && <Route path="config" element={<ConfigPage />} />}
          <Route path="apidoc" element={<ApiDoc />} />
          <Route path="*" element={<Navigate to="/admin/software" replace />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}