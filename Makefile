# Cyber Inspector Makefile

# 变量
BINARY_NAME_MASTER=cyber-inspector
BINARY_NAME_AGENT=cyber-agent
MAIN_MASTER_PATH=./cmd/master
MAIN_AGENT_PATH=./cmd/agent
CONFIG_PATH=./configs/config.yaml
DOCKER_IMAGE_MASTER=cyber-inspector/master
DOCKER_IMAGE_AGENT=cyber-inspector/agent
VERSION=2.0.0
LDFLAGS=-ldflags "-w -s -X main.AppVersion=${VERSION}"

# 默认目标
.PHONY: all
all: build

# 构建所有
.PHONY: build
build: build-master build-agent

# 构建 Master
.PHONY: build-master
build-master:
	@echo "Building ${BINARY_NAME_MASTER}..."
	go build ${LDFLAGS} -o bin/${BINARY_NAME_MASTER} ${MAIN_MASTER_PATH}

# 构建 Agent
.PHONY: build-agent
build-agent:
	@echo "Building ${BINARY_NAME_AGENT}..."
	go build ${LDFLAGS} -o bin/${BINARY_NAME_AGENT} ${MAIN_AGENT_PATH}

# 安装依赖
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# 运行 Master
.PHONY: run-master
run-master: build-master
	@echo "Running ${BINARY_NAME_MASTER}..."
	./bin/${BINARY_NAME_MASTER} --config=${CONFIG_PATH}

# 运行 Agent
.PHONY: run-agent
run-agent: build-agent
	@echo "Running ${BINARY_NAME_AGENT}..."
	./bin/${BINARY_NAME_AGENT}

# 测试
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# 清理
.PHONY: clean
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf logs/

# Docker 构建
.PHONY: docker-build
docker-build:
	@echo "Building Docker images..."
	docker build -f Dockerfile.master -t ${DOCKER_IMAGE_MASTER}:${VERSION} .
	docker build -f Dockerfile.agent -t ${DOCKER_IMAGE_AGENT}:${VERSION} .

# Docker 运行
.PHONY: docker-run
docker-run:
	@echo "Running with Docker Compose..."
	docker-compose up -d

# Docker 停止
.PHONY: docker-stop
docker-stop:
	@echo "Stopping Docker containers..."
	docker-compose down

# 数据库初始化
.PHONY: db-init
db-init:
	@echo "Initializing database..."
	mysql -u root -p < scripts/mysql/init/init.sql

# 生成 API 文档
.PHONY: swag
swag:
	@echo "Generating API documentation..."
	swag init -g cmd/master/main.go

# 代码格式化
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# 代码检查
.PHONY: lint
lint:
	@echo "Linting code..."
	golangci-lint run

# 帮助
.PHONY: help
help:
	@echo "Cyber Inspector Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build          Build all binaries"
	@echo "  build-master   Build master binary"
	@echo "  build-agent    Build agent binary"
	@echo "  deps           Download dependencies"
	@echo "  run-master     Run master server"
	@echo "  run-agent      Run agent server"
	@echo "  test           Run tests"
	@echo "  clean          Clean build artifacts"
	@echo "  docker-build   Build Docker images"
	@echo "  docker-run     Run with Docker Compose"
	@echo "  docker-stop    Stop Docker containers"
	@echo "  db-init        Initialize database"
	@echo "  swag           Generate API documentation"
	@echo "  fmt            Format code"
	@echo "  lint           Lint code"
	@echo "  help           Show this help message"
