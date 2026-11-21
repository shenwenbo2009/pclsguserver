# pclsguserver

软件更新分发与插件市场托管服务端 — 单二进制部署，开箱即用。

## 功能概览

- **软件更新分发** — 版本检查、断点续传下载、客户端活跃度埋点
- **插件市场** — 搜索（内存倒排索引 + DB 降级）、审核流程、发布/下架生命周期
- **管理后台** — JWT 认证 + RBAC 权限、审计日志、统计指标、配置热更新
- **外部上传通道** — X-Upload-Token 认证，支持插件包/软件包/通用文件上传
- **API 文档** — Swagger UI + OpenAPI Spec + Markdown 导出
- **首次运行向导** — Web UI 引导完成数据库、管理员、JWT 密钥配置

## 技术栈

| 层 | 技术 |
|----|------|
| 后端 | Go 1.27 + Gin + GORM |
| 数据库 | SQLite（默认） / MySQL / PostgreSQL |
| 前端 | React 18 + TypeScript + Vite |
| 管理后台 UI | Ant Design |
| 插件市场 UI | TailwindCSS |
| 部署 | go:embed 单二进制 / Docker / systemd |

## 快速开始

### Docker

```bash
cp config.yaml.example config.yaml
# 编辑 config.yaml：修改 jwt.secret 和 admin.superPassword
docker-compose up -d
curl http://localhost:8080/health
```

### 从源码构建

```bash
# Windows
build.bat

# Linux / macOS
./build.sh
```

构建产物为单个可执行文件，前端资源通过 `go:embed` 内嵌。

### 首次运行

```bash
./pclsguserver -config config.yaml
```

首次启动自动进入初始化模式，浏览器访问 `http://localhost:8080/init` 按向导完成配置。

### 启动后访问

| 入口 | 地址 |
|------|------|
| 插件市场 | `http://localhost:8080/` |
| 管理后台 | `http://localhost:8080/admin` |
| API 文档 | `http://localhost:8080/swagger` |
| 健康检查 | `http://localhost:8080/health` |

## API 端点

### 公开接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | 健康检查 |
| GET | `/index.json` | 插件市场索引 |
| GET | `/api/update/check` | 软件更新检查 |
| GET | `/api/update/download` | 软件包下载（支持 Range） |
| POST | `/api/update/report` | 客户端事件上报 |
| GET | `/api/plugins` | 插件列表 |
| GET | `/api/plugins/search` | 插件搜索 |
| GET | `/api/plugins/:id` | 插件详情 |
| GET | `/api/plugins/:id/download` | 插件下载 |
| GET | `/files/*path` | 通用文件访问 |

### 外部上传接口（需 X-Upload-Token）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/upload/plugin` | 插件包上传 |
| POST | `/api/upload/software` | 软件包上传 |
| POST | `/api/upload/file` | 通用文件上传 |

### 管理后台接口（需 JWT 认证）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/admin/login` | 管理员登录 |
| GET/POST | `/api/admin/software/versions` | 软件版本管理 |
| GET/POST | `/api/admin/plugins` | 插件管理 |
| GET | `/api/admin/plugins/reviewing` | 待审核插件 |
| POST | `/api/admin/plugins/:id/review` | 审核插件 |
| GET | `/api/admin/audit` | 审计日志 |
| GET | `/api/admin/metrics` | 统计指标 |
| GET/POST | `/api/admin/upload-tokens` | 上传 Token 管理 |
| GET/PUT | `/api/admin/config` | 配置管理 |

## 项目结构

```
pclsguserver/
├── cmd/server/              # 主入口
├── internal/
│   ├── apidoc/              # API 文档生成
│   ├── audit/               # 审计日志服务
│   ├── config/              # 配置加载
│   ├── httpapi/             # HTTP 路由与 Handler
│   ├── index/               # 内存索引引擎
│   ├── init/                # 首次运行初始化向导
│   ├── metrics/             # 统计指标服务
│   ├── middleware/          # 中间件（认证、限流、CORS）
│   ├── model/               # 数据模型定义
│   ├── pkg/                 # 工具包（checksum、logger）
│   ├── repo/                # 数据库访问层
│   ├── service/             # 业务逻辑层
│   │   ├── auth/            # 认证服务
│   │   ├── config/          # 配置管理服务
│   │   ├── index/           # 索引管理服务
│   │   ├── plugin/          # 插件服务
│   │   ├── review/          # 审核服务
│   │   ├── update/          # 软件更新服务
│   │   ├── upload/          # 上传服务
│   │   └── user/            # 用户管理服务
│   └── store/               # 二进制文件存储
├── frontend/
│   ├── shared/              # 共享类型与 API 封装
│   ├── admin/               # 管理后台（Ant Design）
│   └── market/              # 插件市场（TailwindCSS）
├── web/                     # go:embed 前端资源 + Swagger UI
├── deploy/                  # 部署配置（nginx、systemd）
├── config.yaml.example      # 配置示例
├── Dockerfile               # Docker 构建
├── docker-compose.yml       # Docker Compose
├── build.sh / build.bat     # 构建脚本
└── INSTALL.md               # 详细安装文档
```

## 开发

### 环境要求

- Go 1.27+
- Node.js 20+

### 构建

```bash
# 构建前端
cd frontend/shared && npm install && cd ../..
cd frontend/admin && npm install && npm run build && cd ../..
cd frontend/market && npm install && npm run build && cd ../..

# 构建 Go 二进制
go build -o pclsguserver ./cmd/server
```

### 测试

```bash
go test ./...
```

### Go 代理配置（国内网络）

```bash
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GONOSUMDB=*
```

## 配置

完整配置项见 `config.yaml.example`，所有配置均可通过环境变量覆盖（前缀 `PCL_`）：

```bash
PCL_SERVER_PORT=9090
PCL_DB_DRIVER=postgres
PCL_JWT_SECRET=my-secret
```

详细配置说明见 [INSTALL.md](INSTALL.md)。

## 部署

### Docker Compose

```bash
docker-compose up -d
```

启动两个容器：app（主服务 8080）+ nginx（反向代理 80/443）。

### systemd

```bash
scp pclsguserver root@<your-domain>:/opt/pclsguserver/
scp deploy/pclsguserver.service root@<your-domain>:/etc/systemd/system/
systemctl enable --now pclsguserver
```

### Nginx 反向代理

```bash
cp deploy/nginx.conf /etc/nginx/conf.d/pclsguserver.conf
nginx -t && nginx -s reload
```

支持 HTTP→HTTPS 自动跳转、Gzip 压缩、500MB 上传 body。

## 许可证

Apache License 2.0