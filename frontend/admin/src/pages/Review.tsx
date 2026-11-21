import { useState, useEffect } from 'react'
import { Table, Button, Tag, Space, Modal, Input, message } from 'antd'
import { api } from '@pclsgu/shared'
import type { Plugin } from '@pclsgu/shared'

export default function Review() {
  const [data, setData] = useState<Plugin[]>([])
  const [loading, setLoading] = useState(false)
  const [reviewTarget, setReviewTarget] = useState<Plugin | null>(null)
  const [reviewForm, setReviewForm] = useState({ pass: true, reason: '', comment: '' })

  const load = async () => {
    setLoading(true)
    try {
      const result = await api.get<{ items: Plugin[] }>('/admin/plugins/reviewing')
      setData(result.items || [])
    } catch (err) {
      message.error(err instanceof Error ? err.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  const handleReview = async () => {
    if (!reviewTarget) return
    try {
      await api.post(`/admin/plugins/${reviewTarget.id}/review`, {
        pass: reviewForm.pass,
        reason: reviewForm.reason,
        comment: reviewForm.comment,
      })
      message.success(reviewForm.pass ? '审核通过' : '已驳回')
      setReviewTarget(null)
      setReviewForm({ pass: true, reason: '', comment: '' })
      load()
    } catch (err) {
      message.error(err instanceof Error ? err.message : '操作失败')
    }
  }

  const handleResubmit = async (id: string) => {
    try {
      await api.post(`/admin/plugins/${id}/resubmit`)
      message.success('已重新提交')
      load()
    } catch (err) {
      message.error(err instanceof Error ? err.message : '操作失败')
    }
  }

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
          { title: '描述', dataIndex: 'description', ellipsis: true },
          { title: '状态', dataIndex: 'status', render: (s: string) => <Tag color={s === 'reviewing' ? 'orange' : 'red'}>{s}</Tag> },
          {
            title: '操作',
            render: (_, record: Plugin) => (
              <Space>
                <Button size="small" type="primary" onClick={() => { setReviewTarget(record); setReviewForm({ pass: true, reason: '', comment: '' }) }}>
                  审核
                </Button>
                {record.status === 'rejected' && (
                  <Button size="small" onClick={() => handleResubmit(record.id)}>重新提交</Button>
                )}
              </Space>
            ),
          },
        ]}
      />
      <Modal
        title={`审核: ${reviewTarget?.name || ''}`}
        open={!!reviewTarget}
        onCancel={() => setReviewTarget(null)}
        onOk={handleReview}
      >
        <Space direction="vertical" style={{ width: '100%' }}>
          <Space>
            <Button type={reviewForm.pass ? 'primary' : 'default'} onClick={() => setReviewForm({ ...reviewForm, pass: true })}>
              通过
            </Button>
            <Button type={!reviewForm.pass ? 'primary' : 'default'} danger onClick={() => setReviewForm({ ...reviewForm, pass: false })}>
              驳回
            </Button>
          </Space>
          {!reviewForm.pass && (
            <Input placeholder="驳回原因" value={reviewForm.reason} onChange={(e) => setReviewForm({ ...reviewForm, reason: e.target.value })} />
          )}
          <Input.TextArea placeholder="审核意见" value={reviewForm.comment} onChange={(e) => setReviewForm({ ...reviewForm, comment: e.target.value })} rows={3} />
        </Space>
      </Modal>
    </div>
  )
}