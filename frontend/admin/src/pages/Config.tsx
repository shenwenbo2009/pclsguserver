import { useState, useEffect } from 'react'
import { Table, Button, Modal, Input, Tag, Space, message } from 'antd'
import { api } from '@pclsgu/shared'
import type { SystemConfig } from '@pclsgu/shared'

export default function ConfigPage() {
  const [data, setData] = useState<SystemConfig[]>([])
  const [loading, setLoading] = useState(false)
  const [editTarget, setEditTarget] = useState<SystemConfig | null>(null)
  const [editValue, setEditValue] = useState('')
  const [confirmValue, setConfirmValue] = useState('')

  const load = async () => {
    setLoading(true)
    try {
      const result = await api.get<{ items: SystemConfig[] }>('/admin/config')
      setData(result.items || [])
    } catch (err) {
      message.error(err instanceof Error ? err.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  const handleEdit = async () => {
    if (!editTarget) return
    const body: Record<string, unknown> = { value: editValue }
    if (editTarget.sensitive) {
      body.confirm = confirmValue
    }
    try {
      await api.put(`/admin/config/${editTarget.key}`, body)
      message.success('修改成功')
      setEditTarget(null)
      setEditValue('')
      setConfirmValue('')
      load()
    } catch (err) {
      message.error(err instanceof Error ? err.message : '修改失败')
    }
  }

  return (
    <div>
      <Table
        loading={loading}
        dataSource={data}
        rowKey="key"
        columns={[
          { title: '配置项', dataIndex: 'key' },
          { title: '当前值', dataIndex: 'value', render: (v: string, record: SystemConfig) => record.sensitive ? '******' : v },
          { title: '敏感', dataIndex: 'sensitive', render: (s: boolean) => s ? <Tag color="red">敏感</Tag> : null },
          { title: '热更新', dataIndex: 'hotReloadable', render: (h: boolean) => h ? <Tag color="green">可热更新</Tag> : null },
          { title: '描述', dataIndex: 'description' },
          { title: '更新时间', dataIndex: 'updatedAt' },
          {
            title: '操作',
            render: (_, record: SystemConfig) => (
              <Button size="small" onClick={() => { setEditTarget(record); setEditValue(''); setConfirmValue('') }}>修改</Button>
            ),
          },
        ]}
      />
      <Modal
        title={`修改配置: ${editTarget?.key || ''}`}
        open={!!editTarget}
        onCancel={() => setEditTarget(null)}
        onOk={handleEdit}
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          {editTarget?.sensitive && (
            <>
              <Input.Password placeholder="输入新值" value={editValue} onChange={(e) => setEditValue(e.target.value)} />
              <Input.Password placeholder="再次确认新值" value={confirmValue} onChange={(e) => setConfirmValue(e.target.value)} />
            </>
          )}
          {!editTarget?.sensitive && (
            <Input value={editValue} onChange={(e) => setEditValue(e.target.value)} placeholder="输入新值" />
          )}
        </Space>
      </Modal>
    </div>
  )
}