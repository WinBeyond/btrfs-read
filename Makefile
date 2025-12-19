.PHONY: build test clean install fmt vet lint

# 项目配置
BINARY_NAME=btrfs-read
GO=go
GOFLAGS=-v

# 构建目录
BUILD_DIR=build

# 默认目标
all: build

# 构建 CLI 工具
build: fmt vet
	@echo "Building CLI tool..."
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/btrfs-read

# 运行测试
test:
	@echo "Running tests..."
	$(GO) test -v -race -coverprofile=coverage.out ./...

# 运行基准测试
bench:
	@echo "Running benchmarks..."
	$(GO) test -bench=. -benchmem ./...

# 代码覆盖率
coverage: test
	@echo "Generating coverage report..."
	$(GO) tool cover -html=coverage.out -o coverage.html

# 格式化代码
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# 静态检查
vet:
	@echo "Running go vet..."
	$(GO) vet ./...

# Lint (需要安装 golangci-lint)
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# 清理
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# 安装依赖
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download
	$(GO) mod tidy

# 安装工具
install: build
	@echo "Installing..."
	$(GO) install ./cmd/btrfs-read

# 运行示例
run-example: build
	@echo "Running example..."
	./$(BUILD_DIR)/$(BINARY_NAME) --help

# 创建测试镜像
create-test-image:
	@echo "Creating test image (requires root)..."
	@if [ ! -f tests/create-test-image.sh ]; then \
		echo "Error: tests/create-test-image.sh not found"; \
		exit 1; \
	fi
	@chmod +x tests/create-test-image.sh
	sudo ./tests/create-test-image.sh

# 运行集成测试
test-integration:
	@echo "Running integration tests..."
	@if [ ! -f tests/testdata/test.img ]; then \
		echo "Error: Test image not found. Run 'make create-test-image' first."; \
		exit 1; \
	fi
	$(GO) test -v ./tests/integration/... -coverprofile=coverage-integration.out

# 测试 CLI 工具
test-cli: build
	@echo "Testing CLI tool with test image..."
	@if [ ! -f tests/testdata/test.img ]; then \
		echo "Error: Test image not found. Run 'make create-test-image' first."; \
		exit 1; \
	fi
	./$(BUILD_DIR)/$(BINARY_NAME) ls tests/testdata/test.img /

# 帮助
help:
	@echo "Available targets:"
	@echo "  build              - Build CLI tool"
	@echo "  test               - Run all tests"
	@echo "  test-integration   - Run integration tests"
	@echo "  test-cli           - Test CLI tool with test image"
	@echo "  bench              - Run benchmarks"
	@echo "  coverage           - Generate coverage report"
	@echo "  fmt                - Format code"
	@echo "  vet                - Run go vet"
	@echo "  lint               - Run linter"
	@echo "  clean              - Clean build artifacts"
	@echo "  deps               - Download dependencies"
	@echo "  install            - Install binaries"
	@echo "  create-test-image  - Create test btrfs image (requires root)"
