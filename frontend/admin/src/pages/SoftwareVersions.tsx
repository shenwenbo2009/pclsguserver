import { useState, useEffect } from 'react'
import { Table, Button, Modal, Form, Input, Upload, message, Tag, Space } from 'antd'
import { UploadOutlined } from '@ant-design/icons'
import { api } from '@pclsgu/shared'
import type { SoftwareVersion } from '@pclsgu/shared'

export default function SoftwareVersions() {
  const [data, setData] = useState<SoftwareVersion[]>([])
  const [loading, setLoading] = useState(false)
  const [publishOpen, setPublishOpen] = useState(false)
  const [form] = Form.useForm()

  const load = async () => {
    setLoading(true)
    try {
      const result = await api.get<{ items: SoftwareVersion[] }>('/admin/software/versions')
      setData(result.items || [])
    } catch (err) {
      message.error(err instanceof Error ? err.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  const handlePublish = async () => {
    const values = await form.validateFields()
    const formData = new FormData()
    formData.append('swCode', values.swCode)
    formData.append('version', values.version)
    formData.append('swUrl', values.swUrl || '')
    formData.append('releaseNotes', values.releaseNotes || '')
    if (values.package?.fileList?.[0]?.originFileObj) {
      formData.append('package', values.package.fileList[0].originFileObj)
    }
    try {
      await api.upload('/admin/software/versions', formData)
      message.success('发布成功')
      setPublishOpen(false)
      form.resetFields()
      load()
    } catch (err) {
      message.error(err instanceof Error ? err.message : '发布失败')
    }
  }

  const handleOffline = async (id: string) => {
    Modal.confirm({
      title: '确认下架',
      content: '下架后该版本将不再对客户端可见，确认操作？',
      onOk: async () => {
        await api.put(`/admin/software/versions/${id}/offline`)
        message.success('已下架')
        load()
      },
    })
  }

  const handleDelete = async (id: string) => {
    Modal.confirm({
      title: '确认删除',
      content: '删除操作不可恢复，确认删除该版本？',
      onOk: async () => {
        await api.delete(`/admin/software/versions/${id}`)
        message.success('已删除')
        load()
      },
    })
  }

  const statusColor: Record<string, string> = { published: 'green', offline: 'gray', draft: 'blue' }

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button type="primary" onClick={() => setPublishOpen(true)}>发布新版本</Button>
      </Space>
      <Table
        loading={loading}
        dataSource={data}
        rowKey="id"
        columns={[
          { title: '软件编码', dataIndex: 'swCode' },
          { title: '版本', dataIndex: 'version' },
          { title: '大小', dataIndex: 'size', render: (v: number) => `${(v / 1024 / 1024).toFixed(2)} MB` },
          { title: '下载次数', dataIndex: 'downloadCount' },
          { title: '状态', dataIndex: 'status', render: (s: string) => <Tag color={statusColor[s]}>{s}</Tag> },
          { title: '创建时间', dataIndex: 'createdAt' },
          {
            title: '操作',
            render: (_, record: SoftwareVersion) => (
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
      <Modal title="发布软件版本" open={publishOpen} onCancel={() => setPublishOpen(false)} onOk={handlePublish} width={500}>
        <Form form={form} layout="vertical">
          <Form.Item name="swCode" label="软件编码" rules={[{ required: true }]}>
            <Input placeholder="如: my-app" />
          </Form.Item>
          <Form.Item name="version" label="版本号" rules={[{ required: true }]}>
            <Input placeholder="如: 1.2.0" />
          </Form.Item>
          <Form.Item name="swUrl" label="下载URL（可选）">
            <Input placeholder="外部下载链接" />
          </Form.Item>
          <Form.Item name="releaseNotes" label="更新说明">
            <Input.TextArea rows={3} />
          </Form.Item>
          <Form.Item name="package" label="上传安装包">
            <Upload maxCount={1} beforeUpload={() => false}>
              <Button icon={<UploadOutlined />}>选择文件</Button>
            </Upload>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}