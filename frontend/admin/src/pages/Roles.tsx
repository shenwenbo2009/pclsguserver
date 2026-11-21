import { useState, useEffect } from 'react'
import { Table, Button, Modal, Form, Input, Tag, Space, message } from 'antd'
import { api } from '@pclsgu/shared'
import type { Role } from '@pclsgu/shared'

export default function Roles() {
  const [data, setData] = useState<Role[]>([])
  const [loading, setLoading] = useState(false)
  const [createOpen, setCreateOpen] = useState(false)
  const [form] = Form.useForm()

  const load = async () => {
    setLoading(true)
    try {
      const result = await api.get<{ items: Role[] }>('/admin/roles')
      setData(result.items || [])
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
      await api.post('/admin/roles', {
        name: values.name,
        permissions: values.permissions?.split(',').map((s: string) => s.trim()).filter(Boolean) || [],
        description: values.description || '',
      })
      message.success('创建成功')
      setCreateOpen(false)
      form.resetFields()
      load()
    } catch (err) {
      message.error(err instanceof Error ? err.message : '创建失败')
    }
  }

  const handleDelete = async (id: string) => {
    Modal.confirm({
      title: '确认删除',
      content: '删除角色将影响关联用户权限，确认操作？',
      onOk: async () => {
        await api.delete(`/admin/roles/${id}`)
        message.success('已删除')
        load()
      },
    })
  }

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button type="primary" onClick={() => setCreateOpen(true)}>新增角色</Button>
      </Space>
      <Table
        loading={loading}
        dataSource={data}
        rowKey="id"
        columns={[
          { title: '角色名', dataIndex: 'name' },
          { title: '权限', dataIndex: 'permissions', render: (ps: string[]) => ps?.map((p) => <Tag key={p}>{p}</Tag>) },
          { title: '内置', dataIndex: 'builtin', render: (b: boolean) => b ? <Tag color="blue">内置</Tag> : <Tag>自定义</Tag> },
          { title: '描述', dataIndex: 'description' },
          {
            title: '操作',
            render: (_, record: Role) => (
              !record.builtin ? <Button size="small" danger onClick={() => handleDelete(record.id)}>删除</Button> : null
            ),
          },
        ]}
      />
      <Modal title="新增角色" open={createOpen} onCancel={() => setCreateOpen(false)} onOk={handleCreate}>
        <Form form={form} layout="vertical">
          <Form.Item name="name" label="角色名" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="permissions" label="权限标识（逗号分隔）">
            <Input placeholder="如: software.publish, plugin.manage" />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}