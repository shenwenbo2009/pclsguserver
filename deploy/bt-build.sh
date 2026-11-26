#!/bin/bash
set -e

# pclsguserver 宝塔编译脚本
# 在当前目录执行，无需参数

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

info()  { echo -e "${GREEN}[INFO]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

# 宝塔环境路径
export PATH="/usr/local/btgo/bin:/www/server/nodejs/v24.22.1/bin:${PATH}"

# 配置国内代理
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GONOSUMDB=*
npm config set registry https://registry.npmmirror.com

# 检查环境
command -v go &>/dev/null || error "未找到 Go"
command -v node &>/dev/null || error "未找到 Node.js"
info "Go: $(go version)"
info "Node.js: $(node -v)"

# 检查当前目录是否为项目根目录
if [ ! -f "go.mod" ]; then
  error "当前目录不是项目根目录（未找到 go.mod）"
fi
ROOT_DIR="$(pwd)"
info "编译目录: $ROOT_DIR"

# 构建前端
info "=== 构建前端 ==="
(cd frontend/shared && npm install --silent)
(cd frontend/admin && npm install --silent && npm run build)
(cd frontend/market && npm install --silent && npm run build)
info "前端构建完成"

# 构建 Go 二进制
info "=== 构建 Go 二进制 ==="
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
CGO_ENABLED=0 GOOS=linux GOARCH="$ARCH" go build -o pclsguserver-linux ./cmd/server
info "构建完成: $(ls -lh pclsguserver-linux | awk '{print $5}')"