import { useState } from 'react'
import { Steps, Form, Input, Button, Radio, Space, Card, message, Descriptions, Alert } from 'antd'
import { initApi } from '@pclsgu/shared'

const { Step } = Steps

interface InitConfigData {
  adminUsername: string
  adminPassword: string
  passwordConfirm: string
  jwtMode: string
  jwtSecret: string
  dbDriver: string
  sqlitePath: string
  dbHost: string
  dbPort: string
  dbName: string
  dbUser: string
  dbPassword: string
  storageDriver: string
  storagePath: string
  s3Bucket: string
  s3Region: string
  s3Endpoint: string
}

export default function InitWizard() {
  const [current, setCurrent] = useState(0)
  const [loading, setLoading] = useState(false)
  const [completed, setCompleted] = useState(false)
  const [config, setConfig] = useState<InitConfigData>({
    adminUsername: 'admin',
    adminPassword: '',
    passwordConfirm: '',
    jwtMode: 'auto',
    jwtSecret: '',
    dbDriver: 'sqlite',
    sqlitePath: './data/meta.db',
    dbHost: '127.0.0.1',
    dbPort: '5432',
    dbName: 'pclsgu',
    dbUser: '',
    dbPassword: '',
    storageDriver: 'fs',
    storagePath: './data',
    s3Bucket: '',
    s3Region: '',
    s3Endpoint: '',
  })

  const updateConfig = (patch: Partial<InitConfigData>) => {
    setConfig({ ...config, ...patch })
  }

  const handleNext = async () => {
    setLoading(true)
    try {
      const stepData = getStepData(current + 1, config)
      const result = await initApi.validate(current + 1, stepData)
      if (result.valid) {
        setCurrent(current + 1)
      } else {
        result.errors.forEach((e: string) => message.error(e))
      }
    } catch (err) {
      message.error(err instanceof Error ? err.message : '校验失败')
    } finally {
      setLoading(false)
    }
  }

  const handlePrev = () => {
    setCurrent(Math.max(0, current - 1))
  }

  const handleGenerateJwt = async () => {
    try {
      const result = await initApi.generateJwt()
      updateConfig({ jwtSecret: result.jwtSecret })
      message.success('密钥已生成')
    } catch (err) {
      message.error(err instanceof Error ? err.message : '生成失败')
    }
  }

  const handleTestDb = async () => {
    try {
      const result = await initApi.testDb({
        dbDriver: config.dbDriver,
        sqlitePath: config.sqlitePath,
        dbHost: config.dbHost,
        dbPort: config.dbPort,
        dbName: config.dbName,
        dbUser: config.dbUser,
        dbPassword: config.dbPassword,
      })
      if (result.success) {
        message.success('数据库连接成功')
      } else {
        message.error('连接失败')
      }
    } catch (err) {
      message.error(err instanceof Error ? err.message : '连接失败')
    }
  }

  const handleTestStorage = async () => {
    try {
      const result = await initApi.testStorage({
        storageDriver: config.storageDriver,
        storagePath: config.storagePath,
      })
      if (result.success) {
        message.success('存储路径可写')
      } else {
        message.error('路径不可写')
      }
    } catch (err) {
      message.error(err instanceof Error ? err.message : '校验失败')
    }
  }

  const handleComplete = async () => {
    setLoading(true)
    try {
      const result = await initApi.complete({
        adminUsername: config.adminUsername,
        adminPassword: config.adminPassword,
        jwtSecret: config.jwtSecret,
        dbDriver: config.dbDriver,
        sqlitePath: config.sqlitePath,
        dbHost: config.dbHost,
        dbPort: config.dbPort,
        dbName: config.dbName,
        dbUser: config.dbUser,
        dbPassword: config.dbPassword,
        storageDriver: config.storageDriver,
        storagePath: config.storagePath,
        s3Bucket: config.s3Bucket,
        s3Region: config.s3Region,
        s3Endpoint: config.s3Endpoint,
      })
      if (result.success) {
        setCompleted(true)
        message.success('初始化完成')
        setTimeout(() => {
          window.location.href = '/admin'
        }, 3000)
      }
    } catch (err) {
      message.error(err instanceof Error ? err.message : '初始化失败')
    } finally {
      setLoading(false)
    }
  }

  if (completed) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh', background: '#f0f2f5' }}>
        <Card style={{ width: 500, textAlign: 'center' }}>
          <Alert type="success" message="初始化完成" description="服务正在重启，3秒后自动跳转到管理后台..." showIcon />
        </Card>
      </div>
    )
  }

  const steps = [
    { title: '管理员账号' },
    { title: 'JWT 密钥' },
    { title: '数据库配置' },
    { title: '存储配置' },
    { title: '确认提交' },
  ]

  return (
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '100vh', background: '#f0f2f5', padding: '20px' }}>
      <Card style={{ width: 600 }}>
        <h2 style={{ textAlign: 'center', marginBottom: 24 }}>系统初始化向导</h2>
        <Steps current={current} style={{ marginBottom: 24 }}>
          {steps.map((s) => <Step key={s.title} title={s.title} />)}
        </Steps>

        {current === 0 && (
          <Form layout="vertical">
            <Form.Item label="管理员用户名" required>
              <Input value={config.adminUsername} onChange={(e) => updateConfig({ adminUsername: e.target.value })} placeholder="3-50字符，字母数字下划线" />
            </Form.Item>
            <Form.Item label="管理员密码" required>
              <Input.Password value={config.adminPassword} onChange={(e) => updateConfig({ adminPassword: e.target.value })} placeholder="至少8字符，含字母和数字" />
            </Form.Item>
            <Form.Item label="确认密码" required>
              <Input.Password value={config.passwordConfirm} onChange={(e) => updateConfig({ passwordConfirm: e.target.value })} placeholder="再次输入密码" />
            </Form.Item>
          </Form>
        )}

        {current === 1 && (
          <Form layout="vertical">
            <Form.Item label="JWT 密钥模式">
              <Radio.Group value={config.jwtMode} onChange={(e) => updateConfig({ jwtMode: e.target.value })}>
                <Radio value="auto">自动生成（推荐）</Radio>
                <Radio value="manual">手动输入</Radio>
              </Radio.Group>
            </Form.Item>
            {config.jwtMode === 'auto' ? (
              <Form.Item label="生成的密钥">
                <Space style={{ width: '100%' }}>
                  <Input value={config.jwtSecret} readOnly placeholder="点击下方按钮生成" style={{ flex: 1 }} />
                  <Button onClick={handleGenerateJwt}>生成密钥</Button>
                </Space>
              </Form.Item>
            ) : (
              <Form.Item label="JWT 密钥" required>
                <Input.Password value={config.jwtSecret} onChange={(e) => updateConfig({ jwtSecret: e.target.value })} placeholder="至少16字符" />
              </Form.Item>
            )}
          </Form>
        )}

        {current === 2 && (
          <Form layout="vertical">
            <Form.Item label="数据库类型">
              <Radio.Group value={config.dbDriver} onChange={(e) => {
                const driver = e.target.value
                if (driver === 'postgres') {
                  updateConfig({ dbDriver: driver, dbPort: '5432' })
                } else if (driver === 'mysql') {
                  updateConfig({ dbDriver: driver, dbPort: '3306' })
                } else {
                  updateConfig({ dbDriver: driver })
                }
              }}>
                <Radio value="sqlite">SQLite</Radio>
                <Radio value="postgres">PostgreSQL</Radio>
                <Radio value="mysql">MySQL / MariaDB</Radio>
              </Radio.Group>
            </Form.Item>
            {config.dbDriver === 'sqlite' && (
              <Form.Item label="SQLite 文件路径">
                <Input value={config.sqlitePath} onChange={(e) => updateConfig({ sqlitePath: e.target.value })} placeholder="./data/meta.db" />
              </Form.Item>
            )}
            {(config.dbDriver === 'postgres' || config.dbDriver === 'mysql') && (
              <>
                <Form.Item label="数据库地址" required>
                  <Input value={config.dbHost} onChange={(e) => updateConfig({ dbHost: e.target.value })} placeholder="127.0.0.1" />
                </Form.Item>
                <Form.Item label="端口" required>
                  <Input value={config.dbPort} onChange={(e) => updateConfig({ dbPort: e.target.value })} placeholder={config.dbDriver === 'postgres' ? '5432' : '3306'} />
                </Form.Item>
                <Form.Item label="数据库名" required>
                  <Input value={config.dbName} onChange={(e) => updateConfig({ dbName: e.target.value })} placeholder="pclsgu" />
                </Form.Item>
                <Form.Item label="用户名" required>
                  <Input value={config.dbUser} onChange={(e) => updateConfig({ dbUser: e.target.value })} placeholder={config.dbDriver === 'postgres' ? 'postgres' : 'root'} />
                </Form.Item>
                <Form.Item label="密码" required>
                  <Input.Password value={config.dbPassword} onChange={(e) => updateConfig({ dbPassword: e.target.value })} placeholder="数据库密码" />
                </Form.Item>
              </>
            )}
            <Button onClick={handleTestDb}>测试连接</Button>
          </Form>
        )}

        {current === 3 && (
          <Form layout="vertical">
            <Form.Item label="存储类型">
              <Radio.Group value={config.storageDriver} onChange={(e) => updateConfig({ storageDriver: e.target.value })}>
                <Radio value="fs">本地文件系统</Radio>
                <Radio value="s3">S3 对象存储</Radio>
              </Radio.Group>
            </Form.Item>
            {config.storageDriver === 'fs' ? (
              <Form.Item label="存储路径">
                <Input value={config.storagePath} onChange={(e) => updateConfig({ storagePath: e.target.value })} placeholder="./data" />
              </Form.Item>
            ) : (
              <>
                <Form.Item label="S3 Bucket">
                  <Input value={config.s3Bucket} onChange={(e) => updateConfig({ s3Bucket: e.target.value })} />
                </Form.Item>
                <Form.Item label="S3 Region">
                  <Input value={config.s3Region} onChange={(e) => updateConfig({ s3Region: e.target.value })} />
                </Form.Item>
                <Form.Item label="S3 Endpoint">
                  <Input value={config.s3Endpoint} onChange={(e) => updateConfig({ s3Endpoint: e.target.value })} />
                </Form.Item>
              </>
            )}
            <Button onClick={handleTestStorage}>测试路径</Button>
          </Form>
        )}

        {current === 4 && (
          <Descriptions column={1} bordered size="small">
            <Descriptions.Item label="管理员用户名">{config.adminUsername}</Descriptions.Item>
            <Descriptions.Item label="管理员密码">***</Descriptions.Item>
            <Descriptions.Item label="JWT 密钥">{config.jwtSecret ? '***' : '(未设置)'}</Descriptions.Item>
            <Descriptions.Item label="数据库">
              {config.dbDriver === 'sqlite' && `SQLite: ${config.sqlitePath}`}
              {config.dbDriver === 'postgres' && `PostgreSQL: ${config.dbHost}:${config.dbPort}/${config.dbName}`}
              {config.dbDriver === 'mysql' && `MySQL: ${config.dbHost}:${config.dbPort}/${config.dbName}`}
            </Descriptions.Item>
            <Descriptions.Item label="存储">{config.storageDriver === 'fs' ? `本地: ${config.storagePath}` : `S3: ${config.s3Bucket}`}</Descriptions.Item>
          </Descriptions>
        )}

        <div style={{ marginTop: 24, display: 'flex', justifyContent: 'space-between' }}>
          <Button onClick={handlePrev} disabled={current === 0}>上一步</Button>
          {current < 4 ? (
            <Button type="primary" onClick={handleNext} loading={loading}>下一步</Button>
          ) : (
            <Button type="primary" onClick={handleComplete} loading={loading}>完成初始化</Button>
          )}
        </div>
      </Card>
    </div>
  )
}

function getStepData(step: number, config: InitConfigData): Record<string, unknown> {
  switch (step) {
    case 1:
      return {
        adminUsername: config.adminUsername,
        adminPassword: config.adminPassword,
        passwordConfirm: config.passwordConfirm,
      }
    case 2:
      return { jwtMode: config.jwtMode, jwtSecret: config.jwtSecret }
    case 3:
      return {
        dbDriver: config.dbDriver,
        sqlitePath: config.sqlitePath,
        dbHost: config.dbHost,
        dbPort: config.dbPort,
        dbName: config.dbName,
        dbUser: config.dbUser,
        dbPassword: config.dbPassword,
      }
    case 4:
      return {
        storageDriver: config.storageDriver,
        storagePath: config.storagePath,
        s3Bucket: config.s3Bucket,
        s3Region: config.s3Region,
        s3Endpoint: config.s3Endpoint,
      }
    default:
      return {}
  }
}
