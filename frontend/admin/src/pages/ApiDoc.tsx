import { useState, useEffect, useMemo } from 'react'
import { Card, Input, Select, List, Tag, Space, Button, Table, Descriptions, message, Spin, Empty } from 'antd'
import { SearchOutlined, ReloadOutlined, ThunderboltOutlined } from '@ant-design/icons'
import { fetchOpenApiSpec, tryApiCall, regenerateDocs } from '../api/apidoc'

interface OpenAPISpec {
  openapi: string
  info: { title: string; version: string; description: string }
  paths: Record<string, PathItem>
  tags: { name: string; description: string }[]
  components: { schemas: Record<string, unknown>; securitySchemes: Record<string, unknown> }
}

interface PathItem {
  get?: Operation
  post?: Operation
  put?: Operation
  delete?: Operation
}

interface Operation {
  tags: string[]
  summary: string
  description?: string
  parameters?: Parameter[]
  requestBody?: RequestBody
  responses: Record<string, { description: string; content?: Record<string, { schema: Schema }> }>
  security?: unknown[]
}

interface Parameter {
  name: string
  in: string
  required: boolean
  schema: { type: string; description?: string }
  description?: string
}

interface RequestBody {
  content: Record<string, { schema: Schema }>
  required?: boolean
}

interface Schema {
  type?: string
  properties?: Record<string, Schema>
  required?: string[]
  description?: string
  $ref?: string
}

interface RouteEntry {
  method: string
  path: string
  op: Operation
}

const methodColors: Record<string, string> = {
  GET: 'green',
  POST: 'blue',
  PUT: 'orange',
  DELETE: 'red',
}

export default function ApiDoc() {
  const [spec, setSpec] = useState<OpenAPISpec | null>(null)
  const [loading, setLoading] = useState(true)
  const [selectedTag, setSelectedTag] = useState<string | undefined>(undefined)
  const [searchKeyword, setSearchKeyword] = useState('')
  const [selectedRoute, setSelectedRoute] = useState<RouteEntry | null>(null)
  const [debugVisible, setDebugVisible] = useState(false)
  const [debugParams, setDebugParams] = useState<Record<string, string>>({})
  const [debugBody, setDebugBody] = useState('')
  const [debugResult, setDebugResult] = useState<{ status: number; body: string; latencyMs: number } | null>(null)
  const [debugLoading, setDebugLoading] = useState(false)

  useEffect(() => {
    loadSpec()
  }, [])

  const loadSpec = async () => {
    setLoading(true)
    try {
      const data = await fetchOpenApiSpec()
      setSpec(data as OpenAPISpec)
    } catch {
      message.error('文档加载失败，请重试')
    } finally {
      setLoading(false)
    }
  }

  const allRoutes = useMemo<RouteEntry[]>(() => {
    if (!spec) return []
    const routes: RouteEntry[] = []
    for (const [path, item] of Object.entries(spec.paths)) {
      if (item.get) routes.push({ method: 'GET', path, op: item.get })
      if (item.post) routes.push({ method: 'POST', path, op: item.post })
      if (item.put) routes.push({ method: 'PUT', path, op: item.put })
      if (item.delete) routes.push({ method: 'DELETE', path, op: item.delete })
    }
    return routes
  }, [spec])

  const filteredRoutes = useMemo(() => filterRoutes(allRoutes, selectedTag, searchKeyword), [allRoutes, selectedTag, searchKeyword])

  const handleRegenerate = async () => {
    try {
      const result = await regenerateDocs()
      message.success(`重新生成成功，共 ${result.pathCount} 个接口`)
      await loadSpec()
    } catch {
      message.error('重新生成失败')
    }
  }

  const handleDebug = async () => {
    if (!selectedRoute) return
    setDebugLoading(true)
    try {
      let body: unknown = undefined
      if (debugBody.trim()) {
        body = JSON.parse(debugBody)
      }
      const result = await tryApiCall(selectedRoute.method, selectedRoute.path, debugParams, body)
      setDebugResult(result)
    } catch (err) {
      setDebugResult({ status: 0, body: err instanceof Error ? err.message : '请求失败', latencyMs: 0 })
    } finally {
      setDebugLoading(false)
    }
  }

  if (loading) {
    return <div style={{ display: 'flex', justifyContent: 'center', padding: 100 }}><Spin size="large" /></div>
  }

  if (!spec) {
    return (
      <div style={{ textAlign: 'center', padding: 100 }}>
        <Empty description="文档加载失败" />
        <Button icon={<ReloadOutlined />} onClick={loadSpec} style={{ marginTop: 16 }}>重试</Button>
      </div>
    )
  }

  return (
    <div style={{ display: 'flex', gap: 16, height: 'calc(100vh - 120px)' }}>
      <Card style={{ width: 360, overflow: 'auto' }} title="接口列表" extra={<Button size="small" icon={<ReloadOutlined />} onClick={handleRegenerate}>刷新</Button>}>
        <Space direction="vertical" style={{ width: '100%', marginBottom: 12 }}>
          <Select
            style={{ width: '100%' }}
            placeholder="按模块筛选"
            allowClear
            value={selectedTag}
            onChange={setSelectedTag}
            options={spec.tags.map(t => ({ label: t.name, value: t.name }))}
          />
          <Input
            prefix={<SearchOutlined />}
            placeholder="搜索接口路径或名称"
            value={searchKeyword}
            onChange={e => setSearchKeyword(e.target.value)}
            allowClear
          />
        </Space>
        <List
          size="small"
          dataSource={filteredRoutes}
          renderItem={(route) => (
            <List.Item
              onClick={() => { setSelectedRoute(route); setDebugVisible(false); setDebugResult(null); setDebugParams({}); setDebugBody('') }}
              style={{ cursor: 'pointer', background: selectedRoute?.path === route.path && selectedRoute?.method === route.method ? '#e6f4ff' : undefined }}
            >
              <Space>
                <Tag color={methodColors[route.method]} style={{ width: 50, textAlign: 'center' }}>{route.method}</Tag>
                <div>
                  <div style={{ fontSize: 12 }}>{route.path}</div>
                  <div style={{ fontSize: 11, color: '#999' }}>{route.op.summary}</div>
                </div>
              </Space>
            </List.Item>
          )}
        />
      </Card>

      <Card style={{ flex: 1, overflow: 'auto' }} title={selectedRoute ? `${selectedRoute.method} ${selectedRoute.path}` : '接口详情'}>
        {!selectedRoute ? (
          <Empty description="请选择一个接口" />
        ) : (
          <RouteDetail
            route={selectedRoute}
            debugVisible={debugVisible}
            setDebugVisible={setDebugVisible}
            debugParams={debugParams}
            setDebugParams={setDebugParams}
            debugBody={debugBody}
            setDebugBody={setDebugBody}
            debugResult={debugResult}
            debugLoading={debugLoading}
            onDebug={handleDebug}
          />
        )}
      </Card>
    </div>
  )
}

function filterRoutes(routes: RouteEntry[], tag: string | undefined, keyword: string): RouteEntry[] {
  let result = routes
  if (tag) {
    result = result.filter(r => r.op.tags?.includes(tag))
  }
  if (keyword) {
    const kw = keyword.toLowerCase()
    result = result.filter(r => r.path.toLowerCase().includes(kw) || r.op.summary.toLowerCase().includes(kw))
  }
  return result
}

function RouteDetail({ route, debugVisible, setDebugVisible, debugParams, setDebugParams, debugBody, setDebugBody, debugResult, debugLoading, onDebug }: {
  route: RouteEntry
  debugVisible: boolean
  setDebugVisible: (v: boolean) => void
  debugParams: Record<string, string>
  setDebugParams: (v: Record<string, string>) => void
  debugBody: string
  setDebugBody: (v: string) => void
  debugResult: { status: number; body: string; latencyMs: number } | null
  debugLoading: boolean
  onDebug: () => void
}) {
  const { op } = route
  const requiresAuth = !!op.security?.length

  return (
    <div>
      <Descriptions column={1} size="small" bordered>
        <Descriptions.Item label="摘要">{op.summary}</Descriptions.Item>
        {op.description && <Descriptions.Item label="描述">{op.description}</Descriptions.Item>}
        <Descriptions.Item label="认证">{requiresAuth ? <Tag color="orange">需要 Bearer Token</Tag> : <Tag color="green">公开</Tag>}</Descriptions.Item>
      </Descriptions>

      {op.parameters && op.parameters.length > 0 && (
        <div style={{ marginTop: 16 }}>
          <h4>参数</h4>
          <Table size="small" pagination={false} dataSource={op.parameters} rowKey="name" columns={[
            { title: '参数名', dataIndex: 'name', width: 120 },
            { title: '位置', dataIndex: 'in', width: 80 },
            { title: '类型', dataIndex: ['schema', 'type'], width: 80 },
            { title: '必填', dataIndex: 'required', width: 60, render: (v: boolean) => v ? '是' : '否' },
            { title: '说明', dataIndex: 'description' },
          ]} />
        </div>
      )}

      {op.requestBody && (
        <div style={{ marginTop: 16 }}>
          <h4>请求体</h4>
          {Object.entries(op.requestBody.content).map(([contentType, media]) => (
            <div key={contentType}>
              <Tag>{contentType}</Tag>
              {media.schema?.properties && (
                <Table size="small" pagination={false} style={{ marginTop: 8 }} dataSource={Object.entries(media.schema.properties).map(([name, schema]) => ({ name, ...schema, required: media.schema?.required?.includes(name) }))} rowKey="name" columns={[
                  { title: '字段', dataIndex: 'name', width: 120 },
                  { title: '类型', dataIndex: 'type', width: 80 },
                  { title: '必填', dataIndex: 'required', width: 60, render: (v: boolean) => v ? '是' : '否' },
                  { title: '说明', dataIndex: 'description' },
                ]} />
              )}
            </div>
          ))}
        </div>
      )}

      <div style={{ marginTop: 16 }}>
        <h4>响应</h4>
        <Table size="small" pagination={false} dataSource={Object.entries(op.responses).map(([status, resp]) => ({ status, description: resp.description }))} rowKey="status" columns={[
          { title: '状态码', dataIndex: 'status', width: 80 },
          { title: '说明', dataIndex: 'description' },
        ]} />
      </div>

      <div style={{ marginTop: 16 }}>
        <Button type="primary" icon={<ThunderboltOutlined />} onClick={() => setDebugVisible(!debugVisible)}>
          {debugVisible ? '收起调试' : '在线调试'}
        </Button>
      </div>

      {debugVisible && (
        <div style={{ marginTop: 16, padding: 16, background: '#fafafa', borderRadius: 8 }}>
          <h4>调试参数</h4>
          {op.parameters?.map(p => (
            <div key={p.name} style={{ marginBottom: 8 }}>
              <span style={{ display: 'inline-block', width: 120 }}>{p.name} ({p.in})</span>
              <Input
                style={{ width: 300 }}
                value={debugParams[p.name] || ''}
                onChange={e => setDebugParams({ ...debugParams, [p.name]: e.target.value })}
                placeholder={p.schema?.type || ''}
              />
            </div>
          ))}
          {op.requestBody && (
            <div style={{ marginBottom: 8 }}>
              <span style={{ display: 'inline-block', width: 120 }}>请求体 (JSON)</span>
              <Input.TextArea
                style={{ width: '100%', maxWidth: 500 }}
                rows={4}
                value={debugBody}
                onChange={e => setDebugBody(e.target.value)}
                placeholder='{"key": "value"}'
              />
            </div>
          )}
          <Button type="primary" loading={debugLoading} onClick={onDebug} style={{ marginTop: 8 }}>发送请求</Button>

          {debugResult && (
            <div style={{ marginTop: 16 }}>
              <h4>响应结果</h4>
              <Descriptions column={2} size="small" bordered>
                <Descriptions.Item label="状态码">{debugResult.status}</Descriptions.Item>
                <Descriptions.Item label="耗时">{debugResult.latencyMs}ms</Descriptions.Item>
              </Descriptions>
              <pre style={{ marginTop: 8, padding: 12, background: '#fff', borderRadius: 4, maxHeight: 300, overflow: 'auto', fontSize: 12 }}>
                {formatJson(debugResult.body)}
              </pre>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

function formatJson(text: string): string {
  try {
    return JSON.stringify(JSON.parse(text), null, 2)
  } catch {
    return text
  }
}