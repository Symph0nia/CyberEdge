#!/bin/bash

set -e  # 遇到错误时停止

echo "🚀 CyberEdge 完整测试套件"
echo "=========================="
echo ""

# 检查依赖
echo "🔍 检查依赖..."

# 检查 Go
if ! command -v go &> /dev/null; then
    echo "❌ Go 未安装"
    exit 1
fi

# 检查 Node.js
if ! command -v node &> /dev/null; then
    echo "❌ Node.js 未安装"
    exit 1
fi

echo "✅ 依赖检查通过"
echo ""

# 运行后端测试
echo "🏗️  运行后端测试..."
echo "===================="
cd backend
if [ -f test.sh ]; then
    ./test.sh
else
    echo "📁 运行 Go 测试..."
    export GIN_MODE=test
    export JWT_SECRET=test-secret-for-testing
    go test -v ./pkg/...
fi

# 检查后端测试结果
if [ $? -ne 0 ]; then
    echo "❌ 后端测试失败"
    exit 1
fi

echo ""
echo "✅ 后端测试通过"
echo ""

# 运行前端测试
echo "🎨 运行前端测试..."
echo "=================="
cd ../frontend

# 检查是否已安装依赖
if [ ! -d "node_modules" ]; then
    echo "📦 安装前端依赖..."
    npm install
fi

echo "📁 运行前端测试..."
npm run test:run

# 检查前端测试结果
if [ $? -ne 0 ]; then
    echo "❌ 前端测试失败"
    exit 1
fi

echo ""
echo "✅ 前端测试通过"
echo ""

# 运行代码质量检查
echo "🔍 运行代码质量检查..."
echo "======================"

# 前端 lint
echo "📝 前端代码检查..."
npm run lint

# 回到根目录
cd ..

echo ""
echo "🎉 所有测试通过!"
echo "================"
echo ""
echo "✅ 后端单元测试: 通过"
echo "✅ 前端单元测试: 通过"
echo "✅ 代码质量检查: 通过"
echo ""
echo "🚀 项目已准备好部署!"