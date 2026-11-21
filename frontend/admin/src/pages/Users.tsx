import { useState, useEffect } from 'react'
import { Table, Button, Modal, Form, Input, Select, Tag, Space, message } from 'antd'
import { api } from '@pclsgu/shared'
import type { AdminUser, Role } from '@pclsgu/shared'

export default function Users() {
  const [data, setData] = useState<AdminUser[]>([])
  const [roles, setRoles] = useState<Role[]>([])
  const [loading, setLoading] = useState(false)
  const [createOpen, setCreateOpen] = useState(false)
  const [roleAssignTarget, setRoleAssignTarget] = useState<AdminUser | null>(null)
  const [form] = Form.useForm()
  const [roleForm] = Form.useForm()

  const load = async () => {
    setLoading(true)
    try {
      const [usersResult, rolesResult] = await Promise.all([
        api.get<{ items: AdminUser[] }>('/admin/users'),
        api.get<{ items: Role[] }>('/admin/roles'),
      ])
      setData(usersResult.items || [])
      setRoles(rolesResult.items || [])
    } catch (err) {
      message.error(err instanceof Error ? err.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  const handleCreate = async () => {
    const values = await form.validateFields()
    try {
      await api.post('/admin/users', {
        username: values.username,
        password: values.password,
      })
      message.success('创建成功')
      setCreateOpen(false)
      form.resetFields()
      load()
    } catch (err) {
      message.error(err instanceof Error ? err.message : '创建失败')
    }
  }

  const handleToggleStatus = async (id: string, currentStatus: string) => {
    const newStatus = currentStatus === 'active' ? 'disabled' : 'active'
    try {
      await api.put(`/admin/users/${id}`, { status: newStatus })
      message.success('操作成功')
      load()
    } catch (err) {
      message.error(err instanceof Error ? err.message : '操作失败')
    }
  }

  const handleAssignRoles = async () => {
    if (!roleAssignTarget) return
    const values = await roleForm.validateFields()
    try {
      await api.put(`/admin/users/${roleAssignTarget.id}/roles`, { roleIds: values.roleIds })
      message.success('角色分配成功')
      setRoleAssignTarget(null)
      roleForm.resetFields()
      load()
    } catch (err) {
      message.error(err instanceof Error ? err.message : '操作失败')
    }
  }

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button type="primary" onClick={() => setCreateOpen(true)}>新增管理员</Button>
      </Space>
      <Table
        loading={loading}
        dataSource={data}
        rowKey="id"
        columns={[
          { title: '用户名', dataIndex: 'username' },
          { title: '状态', dataIndex: 'status', render: (s: string) => <Tag color={s === 'active' ? 'green' : 'red'}>{s}</Tag> },
          { title: '角色', dataIndex: 'roles', render: (rs: Role[]) => rs?.map((r) => <Tag key={r.id}>{r.name}</Tag>) },
          { title: '创建时间', dataIndex: 'createdAt' },
          {
            title: '操作',
            render: (_, record: AdminUser) => (
              <Space>
                <Button size="small" onClick={() => { setRoleAssignTarget(record); roleForm.setFieldsValue({ roleIds: record.roles?.map((r) => r.id) || [] }) }}>
                  分配角色
                </Button>
                <Button size="small" onClick={() => handleToggleStatus(record.id, record.status)}>
                  {record.status === 'active' ? '禁用' : '启用'}
                </Button>
              </Space>
            ),
          },
        ]}
      />
      <Modal title="新增管理员" open={createOpen} onCancel={() => setCreateOpen(false)} onOk={handleCreate}>
        <Form form={form} layout="vertical">
          <Form.Item name="username" label="用户名" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="password" label="密码" rules={[{ required: true }]}>
            <Input.Password />
          </Form.Item>
        </Form>
      </Modal>
      <Modal title={`分配角色: ${roleAssignTarget?.username || ''}`} open={!!roleAssignTarget} onCancel={() => setRoleAssignTarget(null)} onOk={handleAssignRoles}>
        <Form form={roleForm} layout="vertical">
          <Form.Item name="roleIds" label="角色" rules={[{ required: true }]}>
            <Select mode="multiple" options={roles.map((r) => ({ value: r.id, label: r.name }))} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}