BINARY := golab

.PHONY: all
all: fix build

.PHONY: fix
fix:
	@goimports -w -l .
	@gofumpt -w -l .
	@go vet ./...

.PHONY: build
build:
	@GOOS=linux GOARCH=arm64 go build -o $(BINARY) ./cmd/golab/main.go

.PHONY: test
test:
	@go test -cover ./...

.PHONY: clean
clean:
	@rm -f $(BINARY)
