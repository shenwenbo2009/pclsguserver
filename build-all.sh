#!/bin/bash
set -e

#===============================================================
# pclsguserver 综合编译脚本
# 运行后选择构建类型
#===============================================================

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

info()  { echo -e "${GREEN}[INFO]${NC} $1"; }
warn()  { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }
step()  { echo -e "${BLUE}[STEP]${NC} $1"; }

#===============================================================
# 检查项目根目录
#===============================================================
if [ ! -f "go.mod" ]; then
  error "当前目录不是项目根目录（未找到 go.mod）"
fi
ROOT_DIR="$(pwd)"

#===============================================================
# 选择构建类型
#===============================================================
echo ""
echo -e "${CYAN}==============================================${NC}"
echo -e "${CYAN}  pclsguserver 编译脚本${NC}"
echo -e "${CYAN}==============================================${NC}"
echo ""
echo "  1) 系统环境编译          (使用系统 Go + Node.js)"
echo "  2) 宝塔面板环境编译      (使用宝塔安装的 Go + Node.js)"
echo "  3) Docker 构建           (使用 Docker 多阶段构建)"
echo "  4) 宝塔编译 + Docker 镜像 (宝塔环境编译后打包 Docker)"
echo ""
read -p "请选择构建类型 [1-4]: " BUILD_CHOICE

case "$BUILD_CHOICE" in
  1) USE_BT=false; USE_DOCKER=false ;;
  2) USE_BT=true;  USE_DOCKER=false ;;
  3) USE_BT=false; USE_DOCKER=true  ;;
  4) USE_BT=true;  USE_DOCKER=true  ;;
  *) error "无效选择: $BUILD_CHOICE" ;;
esac

echo ""
info "项目目录: $ROOT_DIR"

#===============================================================
# 环境配置
#===============================================================
if [ "$USE_BT" = true ]; then
  step "配置宝塔面板环境"
  BT_GO_PATH="/usr/local/btgo/bin"
  BT_NODE_PATH="/www/server/nodejs/v24.22.1/bin"
  export PATH="${BT_GO_PATH}:${BT_NODE_PATH}:${PATH}"

  go env -w GOPROXY=https://goproxy.cn,direct
  go env -w GONOSUMDB=*
  npm config set registry https://registry.npmmirror.com
  info "已配置宝塔环境路径和国内代理"
fi

#===============================================================
# Docker 构建（仅 Docker，无本地编译）
#===============================================================
if [ "$USE_DOCKER" = true ] && [ "$USE_BT" = false ]; then
  step "Docker 构建"
  command -v docker &>/dev/null || error "未找到 docker"

  IMAGE_NAME="pclsguserver"
  IMAGE_TAG="latest"

  info "构建 Docker 镜像: ${IMAGE_NAME}:${IMAGE_TAG}"
  docker build -t "${IMAGE_NAME}:${IMAGE_TAG}" .

  info "Docker 镜像构建完成"
  docker images "${IMAGE_NAME}:${IMAGE_TAG}"

  echo ""
  echo -e "${CYAN}==============================================${NC}"
  echo "  Docker 构建完成"
  echo -e "${CYAN}==============================================${NC}"
  echo "  镜像:    ${IMAGE_NAME}:${IMAGE_TAG}"
  echo "  运行:    docker run -d -p 8080:8080 ${IMAGE_NAME}:${IMAGE_TAG}"
  echo "  Compose: docker compose up -d"
  echo -e "${CYAN}==============================================${NC}"
  exit 0
fi

#===============================================================
# 本地编译（系统环境 或 宝塔环境）
#===============================================================
step "检查编译环境"
command -v go &>/dev/null || error "未找到 Go，请安装 Go 或选择宝塔环境"
command -v node &>/dev/null || error "未找到 Node.js，请安装 Node.js 或选择宝塔环境"
info "Go: $(go version)"
info "Node.js: $(node -v)"

#===============================================================
# 1. 构建前端
#===============================================================
step "1/3 构建前端"

info "构建 shared"
(cd frontend/shared && npm install --silent)

info "构建 admin"
(cd frontend/admin && npm install --silent && npm run build)

info "构建 market"
(cd frontend/market && npm install --silent && npm run build)

info "前端构建完成"

#===============================================================
# 2. 构建 Go 二进制
#===============================================================
step "2/3 构建 Go 二进制"

ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  GO_ARCH="amd64" ;;
  aarch64) GO_ARCH="arm64" ;;
  *)       warn "未识别架构: $ARCH，默认使用 amd64"; GO_ARCH="amd64" ;;
esac

OUTPUT_NAME="pclsguserver-linux"
if [ "$(uname -s)" = "Darwin" ]; then
  OUTPUT_NAME="pclsguserver-darwin"
elif [[ "$(uname -s)" == MINGW* ]] || [[ "$(uname -s)" == MSYS* ]]; then
  OUTPUT_NAME="pclsguserver.exe"
fi

CGO_ENABLED=0 GOOS=linux GOARCH="$GO_ARCH" \
  go build -o "$OUTPUT_NAME" ./cmd/server

info "Go 二进制构建完成: $(ls -lh "$OUTPUT_NAME" | awk '{print $5}') ($OUTPUT_NAME)"

#===============================================================
# 3. Docker 镜像构建（可选）
#===============================================================
if [ "$USE_DOCKER" = true ] && [ "$USE_BT" = true ]; then
  step "3/3 构建 Docker 镜像"
  command -v docker &>/dev/null || error "未找到 docker"

  IMAGE_NAME="pclsguserver"
  IMAGE_TAG="latest"

  info "构建 Docker 镜像: ${IMAGE_NAME}:${IMAGE_TAG}"
  docker build -t "${IMAGE_NAME}:${IMAGE_TAG}" .

  info "Docker 镜像构建完成"
  docker images "${IMAGE_NAME}:${IMAGE_TAG}"
else
  step "3/3 跳过 Docker 构建"
fi

#===============================================================
# 完成
#===============================================================
echo ""
echo -e "${CYAN}==============================================${NC}"
echo "  编译完成"
echo -e "${CYAN}==============================================${NC}"
echo "  二进制: ./$OUTPUT_NAME"
echo "  运行:   ./$OUTPUT_NAME -config config.yaml"
if [ "$USE_DOCKER" = true ]; then
  echo "  Docker: docker run -d -p 8080:8080 pclsguserver:latest"
  echo "  Compose: docker compose up -d"
fi
echo -e "${CYAN}==============================================${NC}"