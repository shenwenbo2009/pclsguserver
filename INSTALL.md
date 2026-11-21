# pclsguserver 安装文档

pclsguserver 是一个软件更新分发 + 插件市场托管服务端，提供软件版本更新检查、断点续传下载、插件市场搜索/审核、管理后台等功能。后端使用 Go 编写，前端使用 React + TypeScript，通过 `go:embed` 打包为单二进制部署。

---

## 1. 环境要求

### 构建环境

| 工具 | 最低版本 | 说明 |
|------|----------|------|
| Go | 1.27 | 后端编译 |
| Node.js | 20 | 前端构建 |
| npm | 10 | 前端依赖管理 |

### 运行环境

- **操作系统**: Linux / Windows / macOS
- **数据库**: SQLite（默认，零配置）/ MySQL 8.0+ / PostgreSQL 14+
- **磁盘空间**: 至少 1GB（数据 + 日志）

---

## 2. Docker 快速部署（推荐）

### 2.1 使用 Docker Compose

```bash
# 1. 准备配置文件
cp config.yaml.example config.yaml

# 2. 修改 JWT 密钥和管理员密码
# 编辑 config.yaml，将 jwt.secret 和 admin.superPassword 改为安全值

# 3. 启动服务
docker-compose up -d

# 4. 检查服务状态
docker-compose ps
curl http://localhost:8080/health
```

Docker Compose 会启动两个容器：
- **app**: pclsguserver 主服务（端口 8080）
- **nginx**: 反向代理（端口 80/443，可选）

### 2.2 使用 Docker 单容器

```bash
docker build -t pclsguserver .

docker run -d \
  --name pclsguserver \
  -p 8080:8080 \
  -v ./data:/app/data \
  -v ./config.yaml:/app/config.yaml:ro \
  -e PCL_JWT_SECRET=your-secret-key \
  -e PCL_ADMIN_SUPERUSERNAME=admin \
  -e PCL_ADMIN_SUPERPASSWORD=your-password \
  pclsguserver
```

---

## 3. 从源码构建

### 3.1 Windows

```bat
# 1. 确保已安装 Go 1.27+ 和 Node.js 20+
go version
node -v

# 2. 运行构建脚本
build.bat

# 3. 生成可执行文件 pclsguserver.exe
```

### 3.2 Linux / macOS

```bash
# 1. 确保已安装 Go 1.27+ 和 Node.js 20+
go version
node -v

# 2. 运行构建脚本
chmod +x build.sh
./build.sh

# 3. 生成可执行文件 pclsguserver
```

### 3.3 手动构建步骤

如果构建脚本不可用，可手动执行：

```bash
# 1. 构建前端
cd frontend/shared && npm install && cd ../..
cd frontend/admin && npm install && npm run build && cd ../..
cd frontend/market && npm install && npm run build && cd ../..

# 2. 构建 Go 二进制
go build -o pclsguserver ./cmd/server

# 3. 交叉编译（Linux amd64）
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o pclsguserver-linux ./cmd/server
```

前端构建产物会自动输出到 `web/admin/` 和 `web/market/` 目录，Go 编译时通过 `go:embed` 嵌入到二进制中。

---

## 4. 配置

### 4.1 配置文件

复制示例配置并修改：

```bash
cp config.yaml.example config.yaml
```

完整配置项说明：

```yaml
server:
  port: 8080              # 服务端口
  dataDir: ./data         # 数据目录（数据库、索引等）
  readTimeout: 30s        # HTTP 读超时
  writeTimeout: 30s       # HTTP 写超时

db:
  driver: sqlite          # sqlite / mysql / postgres
  dsn: ./data/meta.db     # SQLite 文件路径
  # MySQL: root:password@tcp(127.0.0.1:3306)/pclsgu?charset=utf8mb4&parseTime=True&loc=Local
  # PostgreSQL: host=localhost user=postgres dbname=pclsgu sslmode=disable

storage:
  driver: fs              # 文件系统存储（目前仅支持 fs）
  path: ./data            # 存储根路径

jwt:
  secret: change-me-in-production  # JWT 签名密钥（生产环境必须修改）
  accessTtl: 2h           # Access Token 有效期
  refreshTtl: 168h        # Refresh Token 有效期（7天）

rateLimit:
  loginThreshold: 5       # 登录失败锁定阈值
  loginWindow: 5m         # 登录失败计数窗口
  globalRps: 100          # 全局每秒请求限制
  downloadRps: 50         # 下载每秒请求限制

audit:
  retentionDays: 180      # 审计日志保留天数
  siemEnabled: false      # 是否启用 SIEM Webhook
  siemWebhook: ""         # SIEM Webhook 地址

index:
  rebuildInterval: 5m     # 索引自动重建间隔
  formatVersion: 1        # 索引格式版本
  degradePolicy: db       # 索引不可用时降级策略（db=数据库查询）

metrics:
  aggregateInterval: 5m   # 指标聚合间隔

admin:
  superUsername: admin     # 超级管理员用户名
  superPassword: admin123  # 超级管理员密码（生产环境必须修改）

upload:
  pluginMaxSize: 209715200    # 插件包最大 200MB
  softwareMaxSize: 524288000  # 软件包最大 500MB
  fileMaxSize: 20971520       # 通用文件最大 20MB
  tokenRateLimit: 10          # Token 级限流（次/分钟）
  ipRateLimit: 20             # IP 级限流（次/分钟）
  rateLimitWindow: 1m         # 限流窗口
  pluginAllowedTypes: [".zip", ".tar.gz", ".tgz", ".7z", ".rar"]
  fileAllowedTypes: [".jpg", ".jpeg", ".png", ".gif", ".webp", ".pdf", ".doc", ".docx", ".txt", ".csv", ".xlsx"]
```

### 4.2 环境变量

所有配置项均可通过环境变量覆盖，前缀为 `PCL_`，层级用 `_` 分隔：

```bash
PCL_SERVER_PORT=9090
PCL_DB_DRIVER=postgres
PCL_DB_DSN=host=localhost user=postgres dbname=pclsgu sslmode=disable
PCL_JWT_SECRET=my-secret
PCL_ADMIN_SUPERUSERNAME=admin
PCL_ADMIN_SUPERPASSWORD=secure-password
```

---

## 5. 首次运行与初始化

### 5.1 初始化向导

首次启动时（检测到 `config.yaml` 不存在或 `.initialized` 标记文件缺失），服务自动进入初始化模式：

```bash
./pclsguserver -config config.yaml
```

浏览器访问 `http://localhost:8080/init`，按向导完成：
1. **数据库配置** — 选择 SQLite/MySQL/PostgreSQL，填写连接信息
2. **管理员设置** — 设置管理员用户名和密码
3. **JWT 密钥** — 设置 JWT 签名密钥
4. **存储路径** — 确认数据存储目录

初始化完成后，服务自动重启进入正常模式。

### 5.2 直接启动（已有配置文件）

如果已准备好 `config.yaml`：

```bash
./pclsguserver -config config.yaml
```

服务启动后：
- **插件市场**: `http://localhost:8080/`
- **管理后台**: `http://localhost:8080/admin`
- **API 文档**: `http://localhost:8080/swagger`
- **健康检查**: `http://localhost:8080/health`

---

## 6. 数据库配置

### 6.1 SQLite（默认，零配置）

无需额外安装，数据库文件自动创建在 `data/meta.db`。

### 6.2 MySQL

```yaml
db:
  driver: mysql
  dsn: root:password@tcp(127.0.0.1:3306)/pclsgu?charset=utf8mb4&parseTime=True&loc=Local
```

需提前创建数据库：

```sql
CREATE DATABASE pclsgu CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 6.3 PostgreSQL

```yaml
db:
  driver: postgres
  dsn: host=localhost user=postgres password=secret dbname=pclsgu sslmode=disable
```

需提前创建数据库：

```sql
CREATE DATABASE pclsgu;
```

---

## 7. 生产部署

### 7.1 Linux systemd 服务

```bash
# 1. 构建并上传
./deploy/deploy.sh root@<your-domain> /opt/pclsguserver

# 2. 或手动部署
scp pclsguserver-linux root@<your-domain>:/opt/pclsguserver/
scp config.yaml root@<your-domain>:/opt/pclsguserver/config.yaml
scp deploy/pclsguserver.service root@<your-domain>:/etc/systemd/system/

# 3. 启动服务
systemctl daemon-reload
systemctl enable pclsguserver
systemctl start pclsguserver

# 4. 查看状态
systemctl status pclsguserver
journalctl -u pclsguserver -f
```

systemd 服务配置文件位于 `deploy/pclsguserver.service`，默认使用 SQLite，数据目录 `/opt/pclsguserver/data`。

### 7.2 Nginx 反向代理

```bash
# 1. 安装 nginx
apt install nginx

# 2. 复制配置
cp deploy/nginx.conf /etc/nginx/conf.d/pclsguserver.conf

# 3. 配置 SSL 证书
mkdir -p /etc/nginx/certs
# 将 server.crt 和 server.key 放入 /etc/nginx/certs/

# 4. 重载 nginx
nginx -t && nginx -s reload
```

Nginx 配置特性：
- HTTP 自动跳转 HTTPS
- API 请求代理到后端 8080
- 管理后台上传接口允许 500MB body
- 启用 Gzip 压缩

### 7.3 Docker Compose 生产部署

```bash
# 1. 准备 SSL 证书
mkdir -p deploy/certs
cp <your-domain>.crt deploy/certs/server.crt
cp <your-domain>.key deploy/certs/server.key

# 2. 修改 config.yaml 中的安全配置
# - jwt.secret: 使用随机密钥
# - admin.superPassword: 使用强密码

# 3. 启动
docker-compose up -d

# 4. 查看日志
docker-compose logs -f app
```

---

## 8. API 端点概览

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
| GET | `/swagger` | Swagger UI |
| GET | `/docs/api.md` | Markdown API 文档 |

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
| GET | `/api/admin/software/versions` | 软件版本列表 |
| POST | `/api/admin/software/versions` | 发布软件版本 |
| GET | `/api/admin/plugins` | 插件列表 |
| POST | `/api/admin/plugins` | 提交插件 |
| GET | `/api/admin/plugins/reviewing` | 待审核插件 |
| POST | `/api/admin/plugins/:id/review` | 审核插件 |
| GET | `/api/admin/audit` | 审计日志 |
| GET | `/api/admin/metrics` | 统计指标 |
| GET | `/api/admin/upload-tokens` | 上传 Token 列表 |
| POST | `/api/admin/upload-tokens` | 创建上传 Token |
| PUT | `/api/admin/upload-tokens/:id/disable` | 禁用上传 Token |
| GET | `/api/admin/config` | 配置列表 |
| PUT | `/api/admin/config/:key` | 更新配置 |

---

## 9. 目录结构

```
pclsguserver/
├── cmd/server/          # 主入口
├── internal/
│   ├── apidoc/          # API 文档生成
│   ├── audit/           # 审计日志服务
│   ├── config/          # 配置加载
│   ├── httpapi/         # HTTP 路由与 Handler
│   ├── index/           # 内存索引引擎
│   ├── init/            # 首次运行初始化向导
│   ├── metrics/         # 统计指标服务
│   ├── middleware/      # 中间件（认证、限流、CORS）
│   ├── model/           # 数据模型定义
│   ├── pkg/             # 工具包（checksum、logger）
│   ├── repo/            # 数据库访问层
│   ├── service/         # 业务逻辑层
│   │   ├── auth/        # 认证服务
│   │   ├── config/      # 配置管理服务
│   │   ├── index/       # 索引管理服务
│   │   ├── plugin/      # 插件服务
│   │   ├── review/      # 审核服务
│   │   ├── update/      # 软件更新服务
│   │   ├── upload/      # 上传服务
│   │   └── user/        # 用户管理服务
│   └── store/           # 二进制文件存储
├── frontend/
│   ├── shared/          # 共享类型与 API 封装
│   ├── admin/           # 管理后台（Ant Design）
│   └── market/          # 插件市场（TailwindCSS）
├── web/                 # go:embed 前端资源
├── deploy/              # 部署配置（nginx、systemd）
├── config.yaml.example  # 配置示例
├── Dockerfile           # Docker 构建
├── docker-compose.yml   # Docker Compose
├── build.sh / build.bat # 构建脚本
└── go.mod               # Go 模块定义
```

---

## 10. 常见问题

### Q: 构建时 Go 依赖下载失败？

设置国内 Go 代理：

```bash
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GONOSUMDB=*
```

### Q: 前端构建时 esbuild 报错？

确保 Node.js 版本 >= 20，esbuild 版本 >= 0.24.0（已在 package.json 中配置）。

### Q: 如何切换数据库？

修改 `config.yaml` 中的 `db.driver` 和 `db.dsn`，重启服务即可。表结构会自动迁移。

### Q: 忘记管理员密码？

删除 `.initialized` 标记文件和 `config.yaml`，重新启动会进入初始化向导。或直接修改 `config.yaml` 中的 `admin.superPassword`，重启服务后新密码生效（仅对首次创建的管理员生效，已存在的管理员密码需通过数据库修改）。

### Q: 如何查看 API 文档？

启动服务后访问 `http://localhost:8080/swagger` 查看交互式 Swagger UI，或 `http://localhost:8080/docs/api.md` 查看 Markdown 格式文档。

### Q: 上传接口如何认证？

通过管理后台创建上传 Token（`POST /api/admin/upload-tokens`），获得明文 Token 后，在请求头中携带 `X-Upload-Token: utk_xxxx` 即可调用上传接口。Token 仅在创建时返回一次明文。