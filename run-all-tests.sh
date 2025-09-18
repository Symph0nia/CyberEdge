#!/bin/bash

# CyberEdge Complete Testing Suite
# This script runs all tests: frontend unit tests, backend tests, and E2E tests

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 CyberEdge Complete Testing Suite${NC}"
echo "=========================================="

# Check if required tools are installed
check_dependencies() {
    echo -e "${YELLOW}🔍 检查依赖...${NC}"

    if ! command -v npm &> /dev/null; then
        echo -e "${RED}❌ npm 未安装${NC}"
        exit 1
    fi

    if ! command -v go &> /dev/null; then
        echo -e "${RED}❌ Go 未安装${NC}"
        exit 1
    fi

    echo -e "${GREEN}✅ 所有依赖已安装${NC}"
}

# Run frontend tests with coverage
run_frontend_tests() {
    echo -e "${YELLOW}🎯 运行前端测试 (单元测试)...${NC}"
    cd frontend

    echo "📦 安装前端依赖..."
    npm install --silent

    echo "🧪 运行前端单元测试..."
    npm run test:coverage

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ 前端测试通过${NC}"
    else
        echo -e "${RED}❌ 前端测试失败${NC}"
        exit 1
    fi

    cd ..
}

# Run backend tests with coverage
run_backend_tests() {
    echo -e "${YELLOW}🎯 运行后端测试...${NC}"
    cd backend

    echo "🧪 运行后端测试..."
    ./test-all.sh

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ 后端测试通过${NC}"
    else
        echo -e "${RED}❌ 后端测试失败${NC}"
        exit 1
    fi

    cd ..
}

# Run E2E tests (optional, requires browsers)
run_e2e_tests() {
    echo -e "${YELLOW}🎯 运行端到端测试...${NC}"
    cd frontend

    # Check if Playwright browsers are installed
    if npx playwright --version &> /dev/null; then
        echo "🌐 运行 E2E 测试..."
        npm run test:e2e

        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✅ E2E 测试通过${NC}"
        else
            echo -e "${YELLOW}⚠️  E2E 测试失败或跳过${NC}"
        fi
    else
        echo -e "${YELLOW}⚠️  Playwright 浏览器未安装，跳过 E2E 测试${NC}"
        echo "安装命令: npx playwright install"
    fi

    cd ..
}

# Generate coverage report summary
generate_coverage_summary() {
    echo -e "${BLUE}📊 生成覆盖率报告摘要...${NC}"

    # Frontend coverage
    if [ -f "frontend/coverage/coverage-summary.json" ]; then
        echo -e "${GREEN}前端覆盖率报告: frontend/coverage/index.html${NC}"
    fi

    # Backend coverage
    if [ -f "backend/coverage/coverage.html" ]; then
        echo -e "${GREEN}后端覆盖率报告: backend/coverage/coverage.html${NC}"
    fi

    # E2E report
    if [ -f "frontend/playwright-report/index.html" ]; then
        echo -e "${GREEN}E2E 测试报告: frontend/playwright-report/index.html${NC}"
    fi
}

# Main execution
main() {
    echo "开始时间: $(date)"

    check_dependencies

    # Run tests based on arguments
    if [ "$1" = "frontend" ]; then
        run_frontend_tests
    elif [ "$1" = "backend" ]; then
        run_backend_tests
    elif [ "$1" = "e2e" ]; then
        run_e2e_tests
    else
        # Run all tests
        run_frontend_tests
        run_backend_tests
        run_e2e_tests
    fi

    generate_coverage_summary

    echo -e "${GREEN}🎉 所有测试完成！${NC}"
    echo "结束时间: $(date)"
}

# Run main function with all arguments
main "$@"