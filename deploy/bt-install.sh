#!/bin/bash
set -e

#===============================================================
# pclsguserver 宝塔面板安装编译脚本
# 前提: 宝塔已安装 Go 和 Node.js
# 用法: bash bt-install.sh [项目源码目录]
#===============================================================

PROJECT_NAME="pclsguserver"
INSTALL_DIR="/opt/pclsguserver"
SRC_DIR="${1:-/www/wwwroot/pclsguserver}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info()  { echo -e "${GREEN}[INFO]${NC} $1"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

#===============================================================
# 1. 环境检查
#===============================================================
info "=== 1. 环境检查 ==="

if [ "$(id -u)" -ne 0 ]; then
  error "请使用 root 用户执行此脚本"
fi

# 宝塔 Go / Node.js 路径
BT_GO_PATH="/usr/local/btgo/bin"
BT_NODE_PATH="/www/server/nodejs/v24.22.1/bin"
export PATH="${BT_GO_PATH}:${BT_NODE_PATH}:${PATH}"

if ! command -v go &>/dev/null; then
  error "未找到 Go，请确认宝塔已安装 Go: ${BT_GO_PATH}"
fi
info "Go: $(go version)"

if ! command -v node &>/dev/null; then
  error "未找到 Node.js，请确认宝塔已安装 Node.js: ${BT_NODE_PATH}"
fi
info "Node.js: $(node -v)"

if ! command -v npm &>/dev/null; then
  error "未找到 npm"
fi
info "npm: $(npm -v)"

# 配置国内代理
info "配置 Go 代理"
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GONOSUMDB=*
go env -w GO111MODULE=on

npm config set registry https://registry.npmmirror.com

#===============================================================
# 2. 获取源码
#===============================================================
info "=== 2. 获取源码 ==="

if [ -d "$SRC_DIR" ] && [ -f "$SRC_DIR/go.mod" ]; then
  info "使用本地源码目录: $SRC_DIR"
  BUILD_DIR="$SRC_DIR"
else
  warn "源码目录不存在或无效: $SRC_DIR"
  info "从 Git 克隆源码"
  GIT_REPO="${GIT_REPO:-https://github.com/your-org/pclsguserver.git}"
  BUILD_DIR="/tmp/pclsguserver-build"
  rm -rf "$BUILD_DIR"
  git clone "$GIT_REPO" "$BUILD_DIR" || error "Git 克隆失败"
fi

cd "$BUILD_DIR"

#===============================================================
# 3. 构建前端
#===============================================================
info "=== 3. 构建前端 ==="

info "构建 shared"
cd frontend/shared
npm install --silent
cd "$BUILD_DIR"

info "构建 admin"
cd frontend/admin
npm install --silent
npm run build
cd "$BUILD_DIR"

info "构建 market"
cd frontend/market
npm install --silent
npm run build
cd "$BUILD_DIR"

info "前端构建完成"

#===============================================================
# 4. 构建 Go 二进制
#===============================================================
info "=== 4. 构建 Go 二进制 ==="

ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  GO_ARCH="amd64" ;;
  aarch64) GO_ARCH="arm64" ;;
  *)       error "不支持的架构: $ARCH" ;;
esac

CGO_ENABLED=0 GOOS=linux GOARCH="$GO_ARCH" \
  go build -o pclsguserver-linux ./cmd/server

info "Go 二进制构建完成: $(ls -lh pclsguserver-linux | awk '{print $5}')"

#===============================================================
# 5. 部署文件
#===============================================================
info "=== 5. 部署文件 ==="

mkdir -p "$INSTALL_DIR"/{data,logs,deploy/certs}

cp pclsguserver-linux "$INSTALL_DIR/"
chmod +x "$INSTALL_DIR/pclsguserver-linux"

if [ ! -f "$INSTALL_DIR/config.yaml" ]; then
  cp config.yaml.example "$INSTALL_DIR/config.yaml"
  warn "已复制示例配置，请修改 $INSTALL_DIR/config.yaml 中的安全配置"
else
  info "配置文件已存在，跳过"
fi

#===============================================================
# 6. 配置 systemd 服务
#===============================================================
info "=== 6. 配置 systemd 服务 ==="

cat > /etc/systemd/system/pclsguserver.service << 'UNIT'
[Unit]
Description=pclsguserver - Software Update & Plugin Marketplace Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/pclsguserver
ExecStart=/opt/pclsguserver/pclsguserver-linux -config /opt/pclsguserver/config.yaml
Restart=always
RestartSec=5
LimitNOFILE=65536

Environment=PCL_SERVER_PORT=8080
Environment=PCL_SERVER_DATADIR=/opt/pclsguserver/data
Environment=PCL_DB_DRIVER=sqlite
Environment=PCL_DB_DSN=/opt/pclsguserver/data/meta.db

StandardOutput=journal
StandardError=journal
SyslogIdentifier=pclsguserver

[Install]
WantedBy=multi-user.target
UNIT

systemctl daemon-reload
systemctl enable pclsguserver
systemctl restart pclsguserver

info "服务已启动"
systemctl status pclsguserver --no-pager || true

#===============================================================
# 7. 配置 Nginx 反向代理
#===============================================================
info "=== 7. 配置 Nginx 反向代理 ==="

# 宝塔 Nginx vhost 目录
BT_NGINX_VHOST="/www/server/panel/vhost/nginx"
if [ -d "$BT_NGINX_VHOST" ]; then
  NGINX_CONF_DIR="$BT_NGINX_VHOST"
  info "使用宝塔 Nginx vhost 目录: $NGINX_CONF_DIR"
else
  NGINX_CONF_DIR="/etc/nginx/conf.d"
  info "使用系统 Nginx 目录: $NGINX_CONF_DIR"
fi

mkdir -p "$NGINX_CONF_DIR"

cat > "${NGINX_CONF_DIR}/pclsguserver.conf" << 'NGINX'
upstream pclsguserver_backend {
    server 127.0.0.1:8080;
    keepalive 32;
}

server {
    listen 80;
    server_name _;

    location /api/admin/ {
        proxy_pass http://pclsguserver_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        client_max_body_size 500m;
        proxy_read_timeout 300s;
    }

    location /api/ {
        proxy_pass http://pclsguserver_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 120s;
        proxy_send_timeout 120s;
    }

    location /health {
        proxy_pass http://pclsguserver_backend;
        proxy_set_header Host $host;
    }

    location /index.json {
        proxy_pass http://pclsguserver_backend;
        proxy_set_header Host $host;
    }

    location /static/ {
        proxy_pass http://pclsguserver_backend;
        proxy_set_header Host $host;
    }

    location /admin {
        proxy_pass http://pclsguserver_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location / {
        proxy_pass http://pclsguserver_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml;
    gzip_min_length 1000;
}
NGINX

if command -v nginx &>/dev/null; then
  nginx -t && nginx -s reload
  info "Nginx 配置已生效"
else
  warn "Nginx 未安装，请在宝塔面板中安装 Nginx 后执行: nginx -s reload"
fi

#===============================================================
# 8. 验证
#===============================================================
info "=== 8. 验证服务 ==="

sleep 2
if curl -s http://localhost:8080/health | grep -q "ok\|alive\|healthy"; then
  info "服务健康检查通过"
else
  warn "健康检查未通过，请检查日志: journalctl -u pclsguserver -f"
fi

#===============================================================
# 完成
#===============================================================
echo ""
echo "=============================================="
echo "  pclsguserver 安装完成"
echo "=============================================="
echo ""
echo "  服务目录:   $INSTALL_DIR"
echo "  配置文件:   $INSTALL_DIR/config.yaml"
echo "  数据目录:   $INSTALL_DIR/data"
echo "  服务状态:   systemctl status pclsguserver"
echo "  服务日志:   journalctl -u pclsguserver -f"
echo "  管理后台:   http://<your-domain>/admin"
echo "  插件市场:   http://<your-domain>/"
echo "  初始化向导: http://<your-domain>/init"
echo ""
warn "请修改 config.yaml 中的 jwt.secret 和 admin.superPassword"
warn "如需 HTTPS，请在宝塔面板中配置 SSL 证书"
echo ""