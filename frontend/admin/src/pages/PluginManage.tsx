import { useState, useEffect } from 'react'
import { Table, Button, Modal, Tag, Space, message } from 'antd'
import { api } from '@pclsgu/shared'
import type { Plugin } from '@pclsgu/shared'

export default function PluginManage() {
  const [data, setData] = useState<Plugin[]>([])
  const [loading, setLoading] = useState(false)

  const load = async () => {
    setLoading(true)
    try {
      const result = await api.get<{ items: Plugin[] }>('/admin/plugins')
      setData(result.items || [])
    } catch (err) {
      message.error(err instanceof Error ? err.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  const handleOffline = async (id: string) => {
    Modal.confirm({
      title: '确认下架',
      content: '下架后该插件将不再在市场展示，确认操作？',
      onOk: async () => {
        await api.put(`/admin/plugins/${id}/offline`)
        message.success('已下架')
        load()
      },
    })
  }

  const handleDelete = async (id: string) => {
    Modal.confirm({
      title: '确认删除',
      content: '删除操作不可恢复，确认删除该插件？',
      onOk: async () => {
        await api.delete(`/admin/plugins/${id}`)
        message.success('已删除')
        load()
      },
    })
  }

  const statusColor: Record<string, string> = { published: 'green', reviewing: 'orange', rejected: 'red', offline: 'gray', draft: 'blue' }

  return (
    <div>
      <Table
        loading={loading}
        dataSource={data}
        rowKey="id"
        columns={[
          { title: '名称', dataIndex: 'name' },
          { title: '版本', dataIndex: 'version' },
          { title: '作者', dataIndex: 'author' },
          { title: '分类', dataIndex: 'category' },
          { title: '下载次数', dataIndex: 'downloadCount' },
          { title: '状态', dataIndex: 'status', render: (s: string) => <Tag color={statusColor[s]}>{s}</Tag> },
          { title: '提交时间', dataIndex: 'createdAt' },
          {
            title: '操作',
            render: (_, record: Plugin) => (
              <Space>
                {record.status === 'published' && (
                  <Button size="small" onClick={() => handleOffline(record.id)}>下架</Button>
                )}
                <Button size="small" danger onClick={() => handleDelete(record.id)}>删除</Button>
              </Space>
            ),
          },
        ]}
      />
    </div>
  )
}