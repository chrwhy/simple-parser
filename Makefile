.PHONY: test lint vet fmt tidy clean

# 运行测试（含竞态检测）
test:
	go test -v -race -count=1 ./...

# 运行 go vet
vet:
	go vet ./...

# 格式化代码
fmt:
	gofmt -s -w .
	goimports -w .

# 运行 golangci-lint
lint:
	golangci-lint run

# 整理依赖
tidy:
	go mod tidy

# 清理测试缓存
clean:
	go clean -testcache

# 全部检查
check: fmt vet lint test
