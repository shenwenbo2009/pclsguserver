import { useState, useEffect } from 'react'
import { Table, Input, DatePicker, Space } from 'antd'
import { api } from '@pclsgu/shared'
import type { AuditLog } from '@pclsgu/shared'
import dayjs from 'dayjs'

const { RangePicker } = DatePicker

export default function AuditLog() {
  const [data, setData] = useState<AuditLog[]>([])
  const [loading, setLoading] = useState(false)
  const [action, setAction] = useState('')
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs, dayjs.Dayjs]>([dayjs().subtract(7, 'day'), dayjs()])

  const load = async () => {
    setLoading(true)
    try {
      const params: Record<string, string> = {}
      if (action) params.action = action
      if (dateRange?.[0]) params.from = dateRange[0].format('YYYY-MM-DD')
      if (dateRange?.[1]) params.to = dateRange[1].format('YYYY-MM-DD')
      params.page = '1'
      params.size = '50'
      const result = await api.get<{ items: AuditLog[]; total: number }>('/admin/audit', params)
      setData(result.items || [])
    } catch {
      // ignore
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [action, dateRange])

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Input placeholder="操作类型" value={action} onChange={(e) => setAction(e.target.value)} style={{ width: 200 }} />
        <RangePicker
          value={dateRange}
          onChange={(dates) => {
            if (dates?.[0] && dates?.[1]) setDateRange([dates[0], dates[1]])
          }}
        />
      </Space>
      <Table
        loading={loading}
        dataSource={data}
        rowKey="id"
        columns={[
          { title: '操作', dataIndex: 'action' },
          { title: '操作人', dataIndex: 'operator' },
          { title: '对象类型', dataIndex: 'targetType' },
          { title: '对象ID', dataIndex: 'targetId' },
          { title: '详情', dataIndex: 'detail', ellipsis: true },
          { title: 'IP', dataIndex: 'ip' },
          { title: '时间', dataIndex: 'occurredAt' },
        ]}
      />
    </div>
  )
}