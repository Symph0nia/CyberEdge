#!/bin/bash

echo "🧪 运行 CyberEdge 后端单元测试..."
echo "================================"

# 设置测试环境变量
export GIN_MODE=test
export JWT_SECRET=test-secret-for-testing

# 运行所有测试
echo "📁 运行所有 Go 测试..."
go test -v ./pkg/...

# 生成测试覆盖率报告
echo ""
echo "📊 生成测试覆盖率报告..."
go test -coverprofile=coverage.out ./pkg/...

# 显示覆盖率统计
if [ -f coverage.out ]; then
    echo ""
    echo "📈 测试覆盖率统计:"
    go tool cover -func=coverage.out | tail -1
fi

# 运行竞态条件检测
echo ""
echo "🏃 运行竞态条件检测..."
go test -race ./pkg/...

echo ""
echo "✅ 后端测试完成!"