export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}

export interface SoftwareVersion {
  id: string
  swCode: string
  version: string
  swUrl: string
  checksum: string
  size: number
  releaseNotes: string
  status: string
  downloadCount: number
  createdAt: string
  updatedAt: string
}

export interface UpdateCheckResult {
  hasUpdate: boolean
  latestVersion: string
  swUrl: string
  checksum: string
  size: number
  releaseNotes: string
}

export interface Plugin {
  id: string
  name: string
  version: string
  author: string
  category: string
  keywords: string[]
  description: string
  downloadUrl: string
  checksum: string
  size: number
  downloadCount: number
  status: string
  submittedBy: string
  reviewComment: string
  createdAt: string
  updatedAt: string
}

export interface PluginListResult {
  items: Plugin[]
  total: number
  page: number
  pageSize: number
}

export interface ReviewRecord {
  id: string
  pluginId: string
  reviewer: string
  pass: boolean
  reason: string
  comment: string
  createdAt: string
}

export interface AdminUser {
  id: string
  username: string
  status: string
  roles: Role[]
  createdAt: string
  updatedAt: string
}

export interface Role {
  id: string
  name: string
  permissions: string[]
  builtin: boolean
  description: string
}

export interface AuditLog {
  id: string
  action: string
  operator: string
  targetType: string
  targetId: string
  detail: string
  ip: string
  occurredAt: string
}

export interface MetricsEvent {
  dimension: string
  target: string
  uniqueCount: number
  totalCount: number
  date: string
}

export interface SystemConfig {
  key: string
  value: string
  sensitive: boolean
  hotReloadable: boolean
  description: string
  updatedAt: string
}

export interface IndexStatus {
  lastBuiltAt: string
  total: number
  rebuilding: boolean
}

export interface LoginResult {
  accessToken: string
  refreshToken: string
  expiresIn: number
}

export interface RebuildResult {
  success: boolean
  total: number
  durationMs: number
  error?: string
}