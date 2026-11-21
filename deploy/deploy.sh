#!/bin/bash
set -e

# pclsguserver 部署脚本
# 用法: ./deploy/deploy.sh [user@host] [deploy_dir]
# 示例: ./deploy/deploy.sh root@192.168.1.100 /opt/pclsguserver

REMOTE="${1:-}"
DEPLOY_DIR="${2:-/opt/pclsguserver}"

if [ -z "$REMOTE" ]; then
  echo "用法: $0 <user@host> [deploy_dir]"
  echo "示例: $0 root@192.168.1.100 /opt/pclsguserver"
  exit 1
fi

echo "=== 1. 本地构建前端 ==="
cd frontend/shared && npm install --silent && cd ../..
cd frontend/admin && npm install --silent && npm run build && cd ../..
cd frontend/market && npm install --silent && npm run build && cd ../..

echo "=== 2. 本地构建 Go 二进制 ==="
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o pclsguserver-linux ./cmd/server

echo "=== 3. 创建远程目录 ==="
ssh "$REMOTE" "mkdir -p $DEPLOY_DIR/{data,logs,deploy/certs}"

echo "=== 4. 上传文件 ==="
scp pclsguserver-linux "$REMOTE:$DEPLOY_DIR/"
scp config.yaml.example "$REMOTE:$DEPLOY_DIR/config.yaml"
scp deploy/nginx.conf "$REMOTE:$DEPLOY_DIR/deploy/"
scp deploy/pclsguserver.service "$REMOTE:/etc/systemd/system/"

echo "=== 5. 启动服务 ==="
ssh "$REMOTE" << EOF
  cd $DEPLOY_DIR
  chmod +x pclsguserver-linux
  systemctl daemon-reload
  systemctl enable pclsguserver
  systemctl restart pclsguserver
  systemctl status pclsguserver --no-pager
EOF

echo "=== 6. 清理本地构建产物 ==="
rm -f pclsguserver-linux

echo "=== 部署完成 ==="
echo "访问: http://$REMOTE"
echo "管理后台: http://$REMOTE/admin"
echo "插件市场: http://$REMOTE/"