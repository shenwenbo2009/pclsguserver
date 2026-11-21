import { Layout, Menu, Button, Space } from 'antd'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { useAuth } from '../hooks/useAuth'
import {
  AppstoreOutlined,
  BlockOutlined,
  AuditOutlined,
  SearchOutlined,
  BarChartOutlined,
  FileSearchOutlined,
  TeamOutlined,
  SafetyOutlined,
  SettingOutlined,
  BookOutlined,
} from '@ant-design/icons'

const { Header, Sider, Content } = Layout

export default function AdminLayout() {
  const { permissions, logout } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()

  const hasPerm = (perm: string) => permissions.includes('*') || permissions.includes(perm)

  const menuItems = [
    hasPerm('software.publish') && { key: '/admin/software', icon: <AppstoreOutlined />, label: '软件版本' },
    hasPerm('plugin.manage') && { key: '/admin/plugins', icon: <BlockOutlined />, label: '插件管理' },
    hasPerm('plugin.review') && { key: '/admin/review', icon: <AuditOutlined />, label: '审核' },
    hasPerm('index.view') && { key: '/admin/index', icon: <SearchOutlined />, label: '索引管理' },
    hasPerm('metrics.view') && { key: '/admin/metrics', icon: <BarChartOutlined />, label: '统计' },
    hasPerm('audit.view') && { key: '/admin/audit', icon: <FileSearchOutlined />, label: '审计日志' },
    hasPerm('user.manage') && { key: '/admin/users', icon: <TeamOutlined />, label: '用户管理' },
    hasPerm('user.manage') && { key: '/admin/roles', icon: <SafetyOutlined />, label: '角色管理' },
    hasPerm('config.manage') && { key: '/admin/config', icon: <SettingOutlined />, label: '配置管理' },
    { key: '/admin/apidoc', icon: <BookOutlined />, label: 'API 文档' },
  ].filter(Boolean) as { key: string; icon: React.ReactNode; label: string }[]

  const handleLogout = async () => {
    await logout()
    navigate('/admin/login')
  }

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider collapsible>
        <div style={{ height: 48, margin: 8, color: '#fff', textAlign: 'center', lineHeight: '48px', fontSize: 16 }}>
          软件更新中心
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
        />
      </Sider>
      <Layout>
        <Header style={{ background: '#fff', padding: '0 16px', display: 'flex', justifyContent: 'flex-end', alignItems: 'center' }}>
          <Space>
            <Button type="text" onClick={handleLogout}>退出登录</Button>
          </Space>
        </Header>
        <Content style={{ margin: 16, padding: 24, background: '#fff', borderRadius: 8 }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}