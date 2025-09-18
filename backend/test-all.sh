#!/bin/bash

# CyberEdge Backend Testing Script with Coverage
set -e

echo "🚀 开始运行 CyberEdge 后端完整测试套件..."

# 设置测试环境变量
export GIN_MODE=test
export JWT_SECRET=test-secret
export MYSQL_DSN="root:password@tcp(localhost:3306)/cyberedge_test?charset=utf8mb4&parseTime=True&loc=Local"

# 创建覆盖率目录
mkdir -p coverage

echo "📊 运行单元测试并生成覆盖率报告..."

# 运行所有测试并生成覆盖率报告
go test -v -race -coverprofile=coverage/coverage.out -covermode=atomic ./...

# 生成HTML覆盖率报告
echo "📈 生成HTML覆盖率报告..."
go tool cover -html=coverage/coverage.out -o coverage/coverage.html

# 显示覆盖率统计
echo "📋 覆盖率统计："
go tool cover -func=coverage/coverage.out

# 计算总体覆盖率
COVERAGE=$(go tool cover -func=coverage/coverage.out | grep total | awk '{print $3}')
echo "🎯 总体代码覆盖率: $COVERAGE"

# 覆盖率阈值检查
THRESHOLD="70.0%"
COVERAGE_NUM=$(echo $COVERAGE | sed 's/%//')
THRESHOLD_NUM=$(echo $THRESHOLD | sed 's/%//')

if (( $(echo "$COVERAGE_NUM >= $THRESHOLD_NUM" | bc -l) )); then
    echo "✅ 覆盖率测试通过！当前覆盖率 $COVERAGE >= 阈值 $THRESHOLD"
else
    echo "❌ 覆盖率测试失败！当前覆盖率 $COVERAGE < 阈值 $THRESHOLD"
    exit 1
fi

echo "🎉 所有测试完成！"
echo "📁 覆盖率报告已保存到: coverage/coverage.html"
echo "🌐 打开浏览器查看详细覆盖率报告: file://$(pwd)/coverage/coverage.html"