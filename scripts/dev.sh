#!/bin/bash

# Development startup script for San11-trade

set -e

echo "======================================"
echo "  三国志11交易系统 - 开发环境启动"
echo "======================================"

# Check Go installation
if ! command -v go &> /dev/null; then
    echo "❌ Go未安装，请先安装Go 1.21+"
    exit 1
fi

# Check Node.js installation
if ! command -v node &> /dev/null; then
    echo "❌ Node.js未安装，请先安装Node.js 18+"
    exit 1
fi

echo "✅ 环境检查通过"
echo ""

# Start backend
echo "📦 启动后端服务..."
cd backend

if [ ! -f "go.sum" ]; then
    echo "   下载Go依赖..."
    go mod tidy
fi

# Create admin if not exists
echo "   创建管理员账号..."
go run ./cmd/server -create-admin -admin-user=admin -admin-pass=admin123 &
BACKEND_PID=$!

sleep 3
echo "✅ 后端服务已启动 (PID: $BACKEND_PID)"
echo "   API地址: http://localhost:8080"
echo ""

# Start frontend
echo "📦 启动前端服务..."
cd ../frontend

if [ ! -d "node_modules" ]; then
    echo "   安装npm依赖..."
    npm install
fi

npm run dev &
FRONTEND_PID=$!

sleep 5
echo "✅ 前端服务已启动 (PID: $FRONTEND_PID)"
echo "   访问地址: http://localhost:3000"
echo ""

echo "======================================"
echo "  系统启动完成！"
echo "======================================"
echo ""
echo "默认管理员账号: admin / admin123"
echo ""
echo "按 Ctrl+C 停止所有服务"

# Wait for both processes
wait $BACKEND_PID $FRONTEND_PID
