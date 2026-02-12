#!/bin/bash

# Production deployment script

set -e

echo "======================================"
echo "  三国志11交易系统 - 生产部署"
echo "======================================"

# Check Docker installation
if ! command -v docker &> /dev/null; then
    echo "❌ Docker未安装"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose未安装"
    exit 1
fi

echo "✅ Docker环境检查通过"
echo ""

# Set JWT secret if not provided
if [ -z "$JWT_SECRET" ]; then
    JWT_SECRET=$(openssl rand -base64 32)
    echo "⚠️  已自动生成JWT_SECRET，建议设置环境变量保存"
    export JWT_SECRET
fi

# Build and start
echo "📦 构建并启动服务..."
docker-compose up -d --build

echo ""
echo "======================================"
echo "  部署完成！"
echo "======================================"
echo ""
echo "访问地址: http://$(hostname -I | awk '{print $1}')"
echo ""
echo "查看日志: docker-compose logs -f"
echo "停止服务: docker-compose down"
echo ""
echo "首次使用请：
1. 创建管理员账号
   docker-compose exec backend ./server -create-admin -admin-user=admin -admin-pass=YOUR_PASSWORD

2. 登录管理后台导入Excel数据"
