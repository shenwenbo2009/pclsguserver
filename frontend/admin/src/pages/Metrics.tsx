import { useState, useEffect } from 'react'
import { Table, Select, DatePicker, Space, Card } from 'antd'
import { api } from '@pclsgu/shared'
import type { MetricsEvent } from '@pclsgu/shared'
import dayjs from 'dayjs'

const { RangePicker } = DatePicker

export default function Metrics() {
  const [data, setData] = useState<MetricsEvent[]>([])
  const [loading, setLoading] = useState(false)
  const [dimension, setDimension] = useState('software_download')
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs, dayjs.Dayjs]>([dayjs().subtract(7, 'day'), dayjs()])

  const load = async () => {
    setLoading(true)
    try {
      const params: Record<string, string> = {
        dimension,
        granularity: 'day',
      }
      if (dateRange?.[0]) params.from = dateRange[0].format('YYYY-MM-DD')
      if (dateRange?.[1]) params.to = dateRange[1].format('YYYY-MM-DD')
      const result = await api.get<{ items: MetricsEvent[] }>('/admin/metrics', params)
      setData(result.items || [])
    } catch {
      // ignore
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [dimension, dateRange])

  return (
    <div>
      <Card style={{ marginBottom: 16 }}>
        <Space>
          <Select
            value={dimension}
            onChange={setDimension}
            style={{ width: 200 }}
            options={[
              { value: 'software_download', label: '软件下载' },
              { value: 'software_activity', label: '软件查询' },
              { value: 'plugin_download', label: '插件下载' },
              { value: 'plugin_search', label: '插件搜索' },
            ]}
          />
          <RangePicker
            value={dateRange}
            onChange={(dates) => {
              if (dates?.[0] && dates?.[1]) {
                setDateRange([dates[0], dates[1]])
              }
            }}
          />
        </Space>
      </Card>
      <Table
        loading={loading}
        dataSource={data}
        rowKey={(r) => `${r.dimension}-${r.target}-${r.date}`}
        columns={[
          { title: '维度', dataIndex: 'dimension' },
          { title: '目标', dataIndex: 'target' },
          { title: '独立用户数', dataIndex: 'uniqueCount' },
          { title: '总次数', dataIndex: 'totalCount' },
          { title: '日期', dataIndex: 'date' },
        ]}
      />
    </div>
  )
}