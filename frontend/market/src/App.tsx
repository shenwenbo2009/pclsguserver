import { useState, useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { api } from '@pclsgu/shared'
import type { Plugin, PluginListResult } from '@pclsgu/shared'

const CATEGORIES = ['全部', '工具', '主题', '扩展', '游戏', '其他']

export default function App() {
  const { id } = useParams<{ id: string }>()
  const [plugins, setPlugins] = useState<Plugin[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [keyword, setKeyword] = useState('')
  const [category, setCategory] = useState('全部')
  const [page, setPage] = useState(1)
  const [pageSize] = useState(12)
  const [selectedPlugin, setSelectedPlugin] = useState<Plugin | null>(null)

  const load = async () => {
    setLoading(true)
    try {
      const params: Record<string, string> = {
        page: String(page),
        pageSize: String(pageSize),
      }
      if (keyword) params.keyword = keyword
      if (category !== '全部') params.category = category
      const result = await api.get<PluginListResult>('/plugins', params)
      setPlugins(result.items || [])
      setTotal(result.total || 0)
    } catch {
      setPlugins([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (id) {
      api.get<Plugin>(`/plugins/${id}`).then(setSelectedPlugin).catch(() => setSelectedPlugin(null))
    } else {
      setSelectedPlugin(null)
      load()
    }
  }, [id, page, keyword, category])

  if (selectedPlugin) {
    return <PluginDetail plugin={selectedPlugin} onBack={() => { setSelectedPlugin(null); window.history.back() }} />
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white shadow-sm sticky top-0 z-10">
        <div className="max-w-6xl mx-auto px-4 py-4 flex items-center justify-between">
          <h1 className="text-xl font-bold text-gray-900">插件市场</h1>
          <div className="flex items-center gap-4">
            <input
              type="text"
              placeholder="搜索插件..."
              value={keyword}
              onChange={(e) => { setKeyword(e.target.value); setPage(1) }}
              className="px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 w-64"
            />
          </div>
        </div>
        <div className="max-w-6xl mx-auto px-4 pb-3 flex gap-2">
          {CATEGORIES.map((cat) => (
            <button
              key={cat}
              onClick={() => { setCategory(cat); setPage(1) }}
              className={`px-3 py-1 rounded-full text-sm transition ${
                category === cat ? 'bg-blue-500 text-white' : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
            >
              {cat}
            </button>
          ))}
        </div>
      </header>

      <main className="max-w-6xl mx-auto px-4 py-8">
        {loading ? (
          <div className="text-center py-20 text-gray-500">加载中...</div>
        ) : plugins.length === 0 ? (
          <div className="text-center py-20 text-gray-500">暂无插件</div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
            {plugins.map((p) => (
              <PluginCard key={p.id} plugin={p} onClick={() => setSelectedPlugin(p)} />
            ))}
          </div>
        )}

        {total > pageSize && (
          <div className="flex justify-center mt-8 gap-2">
            <button
              onClick={() => setPage(page - 1)}
              disabled={page <= 1}
              className="px-4 py-2 border rounded-lg disabled:opacity-50 hover:bg-gray-100"
            >
              上一页
            </button>
            <span className="px-4 py-2 text-gray-600">
              {page} / {Math.ceil(total / pageSize)}
            </span>
            <button
              onClick={() => setPage(page + 1)}
              disabled={page >= Math.ceil(total / pageSize)}
              className="px-4 py-2 border rounded-lg disabled:opacity-50 hover:bg-gray-100"
            >
              下一页
            </button>
          </div>
        )}
      </main>

      <footer className="text-center py-8 text-gray-400 text-sm">
        软件更新中心 - 插件市场
      </footer>
    </div>
  )
}

function PluginCard({ plugin, onClick }: { plugin: Plugin; onClick: () => void }) {
  return (
    <div
      onClick={onClick}
      className="bg-white rounded-lg shadow-sm border border-gray-200 p-4 cursor-pointer hover:shadow-md transition"
    >
      <div className="flex items-start justify-between mb-2">
        <h3 className="font-semibold text-gray-900 truncate">{plugin.name}</h3>
        <span className="text-xs text-gray-500 bg-gray-100 px-2 py-0.5 rounded">{plugin.category}</span>
      </div>
      <p className="text-sm text-gray-600 line-clamp-2 mb-3">{plugin.description}</p>
      <div className="flex items-center justify-between text-xs text-gray-500">
        <span>v{plugin.version} · {plugin.author}</span>
        <span>{plugin.downloadCount} 次下载</span>
      </div>
    </div>
  )
}

function PluginDetail({ plugin, onBack }: { plugin: Plugin; onBack: () => void }) {
  const formatSize = (bytes: number) => {
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    return `${(bytes / 1024 / 1024).toFixed(2)} MB`
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white shadow-sm">
        <div className="max-w-4xl mx-auto px-4 py-4">
          <button onClick={onBack} className="text-blue-500 hover:underline">← 返回列表</button>
        </div>
      </header>
      <main className="max-w-4xl mx-auto px-4 py-8">
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
          <div className="flex items-start justify-between mb-4">
            <div>
              <h1 className="text-2xl font-bold text-gray-900">{plugin.name}</h1>
              <p className="text-gray-500 mt-1">v{plugin.version} · {plugin.author}</p>
            </div>
            <span className="text-sm bg-blue-100 text-blue-700 px-3 py-1 rounded-full">{plugin.category}</span>
          </div>
          <p className="text-gray-700 mb-6">{plugin.description}</p>
          <div className="grid grid-cols-2 gap-4 mb-6 text-sm">
            <div className="bg-gray-50 rounded p-3">
              <div className="text-gray-500">文件大小</div>
              <div className="font-semibold">{formatSize(plugin.size)}</div>
            </div>
            <div className="bg-gray-50 rounded p-3">
              <div className="text-gray-500">下载次数</div>
              <div className="font-semibold">{plugin.downloadCount}</div>
            </div>
            <div className="bg-gray-50 rounded p-3">
              <div className="text-gray-500">校验码</div>
              <div className="font-mono text-xs break-all">{plugin.checksum}</div>
            </div>
            <div className="bg-gray-50 rounded p-3">
              <div className="text-gray-500">关键词</div>
              <div className="font-semibold">{plugin.keywords?.join(', ') || '-'}</div>
            </div>
          </div>
          <a
            href={plugin.downloadUrl}
            className="inline-block bg-blue-500 text-white px-6 py-3 rounded-lg font-semibold hover:bg-blue-600 transition"
          >
            下载插件
          </a>
        </div>
      </main>
    </div>
  )
}