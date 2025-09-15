#!/bin/bash

# CyberEdge 开发环境启动脚本
# 启动MySQL + 后端 + 前端

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的信息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查命令是否存在
check_command() {
    if ! command -v $1 &> /dev/null; then
        print_error "$1 未安装"
        exit 1
    fi
}

# 等待服务启动
wait_for_service() {
    local host=$1
    local port=$2
    local service_name=$3
    local max_attempts=30
    local attempt=1

    print_info "等待 $service_name 启动..."

    while [ $attempt -le $max_attempts ]; do
        if nc -z $host $port 2>/dev/null; then
            print_success "$service_name 已启动"
            return 0
        fi

        print_info "等待 $service_name 启动 ($attempt/$max_attempts)..."
        sleep 2
        attempt=$((attempt + 1))
    done

    print_error "$service_name 启动超时"
    return 1
}

# 清理函数
cleanup() {
    print_info "正在清理进程..."

    # 停止Docker容器
    if docker ps | grep -q cyberedge-mysql; then
        docker stop cyberedge-mysql > /dev/null 2>&1
        print_success "MySQL容器已停止"
    fi

    # 杀死后端进程
    pkill -f "cyberedge" > /dev/null 2>&1 || true

    # 杀死前端进程
    pkill -f "npm run serve" > /dev/null 2>&1 || true
    pkill -f "vue-cli-service serve" > /dev/null 2>&1 || true

    print_success "清理完成"
}

# 设置信号处理
trap cleanup EXIT INT TERM

print_info "启动 CyberEdge 开发环境..."

# 检查必要的命令
print_info "检查依赖..."
check_command docker
check_command go
check_command npm
check_command nc

# 1. 启动MySQL容器
print_info "启动MySQL容器..."

# 停止现有容器
docker stop cyberedge-mysql > /dev/null 2>&1 || true
docker rm cyberedge-mysql > /dev/null 2>&1 || true

# 启动新容器
docker run -d \
    --name cyberedge-mysql \
    -e MYSQL_ROOT_PASSWORD=password \
    -e MYSQL_DATABASE=cyberedge \
    -p 3306:3306 \
    mysql:8.0 \
    --character-set-server=utf8mb4 \
    --collation-server=utf8mb4_unicode_ci

# 等待MySQL启动
wait_for_service localhost 3306 "MySQL"

# 导入数据库schema
print_info "初始化数据库..."
sleep 5  # 等待MySQL完全准备好
docker exec -i cyberedge-mysql mysql -uroot -ppassword cyberedge < backend/schema.sql
print_success "数据库初始化完成"

# 2. 启动后端
print_info "启动后端服务..."
cd backend

# 设置环境变量
export MYSQL_DSN="root:password@tcp(localhost:3306)/cyberedge?charset=utf8mb4&parseTime=True&loc=Local"
export JWT_SECRET="your-super-secret-jwt-key-change-this-in-production"
export SESSION_SECRET="your-super-secret-session-key-change-this-in-production"
export PORT="31337"

# 编译并启动后端
go build -o cyberedge cmd/cyberedge.go
./cyberedge &
BACKEND_PID=$!

cd ..

# 等待后端启动
wait_for_service localhost 31337 "后端服务"

# 3. 启动前端
print_info "启动前端服务..."
cd frontend

# 安装依赖（如果需要）
if [ ! -d "node_modules" ]; then
    print_info "安装前端依赖..."
    npm install
fi

# 启动前端开发服务器
npm run serve &
FRONTEND_PID=$!

cd ..

# 等待前端启动
wait_for_service localhost 8080 "前端服务"

print_success "所有服务已启动！"
echo ""
print_info "服务地址："
echo "  前端: http://localhost:8080"
echo "  后端: http://localhost:31337"
echo "  MySQL: localhost:3306 (用户: root, 密码: password)"
echo ""
print_warning "按 Ctrl+C 停止所有服务"

# 保持脚本运行
wait