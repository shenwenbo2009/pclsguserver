import { useState, useEffect } from 'react'
import { Card, Button, Descriptions, message, Spin } from 'antd'
import { api } from '@pclsgu/shared'
import type { IndexStatus, RebuildResult } from '@pclsgu/shared'

export default function IndexManage() {
  const [status, setStatus] = useState<IndexStatus | null>(null)
  const [loading, setLoading] = useState(false)
  const [rebuilding, setRebuilding] = useState(false)

  const loadStatus = async () => {
    setLoading(true)
    try {
      const data = await api.get<IndexStatus>('/admin/index/status')
      setStatus(data)
    } catch (err) {
      message.error(err instanceof Error ? err.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { loadStatus() }, [])

  const handleRebuild = async () => {
    setRebuilding(true)
    try {
      const result = await api.post<RebuildResult>('/admin/index/rebuild')
      if (result.success) {
        message.success(`重建成功，共 ${result.total} 个插件，耗时 ${result.durationMs}ms`)
      } else {
        message.error(`重建失败: ${result.error}`)
      }
      loadStatus()
    } catch (err) {
      message.error(err instanceof Error ? err.message : '重建失败')
    } finally {
      setRebuilding(false)
    }
  }

  return (
    <div>
      <Spin spinning={loading}>
        <Card title="索引状态" style={{ marginBottom: 16 }}>
          <Descriptions column={1}>
            <Descriptions.Item label="最后构建时间">{status?.lastBuiltAt || '尚未构建'}</Descriptions.Item>
            <Descriptions.Item label="索引插件数">{status?.total ?? 0}</Descriptions.Item>
            <Descriptions.Item label="当前状态">{status?.rebuilding ? '重建中...' : '空闲'}</Descriptions.Item>
          </Descriptions>
        </Card>
      </Spin>
      <Button type="primary" loading={rebuilding} onClick={handleRebuild}>
        手动全量重建
      </Button>
    </div>
  )
}