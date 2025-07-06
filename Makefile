BINARY := golab

.PHONY: all
all: fix test build clean

.PHONY: fix
fix:
	@goimports -w -l .
	@gofumpt -w -l .
	@go vet ./...
	@echo "ok\tfix"

.PHONY: build
build:
	@GOOS=linux GOARCH=arm64 go build -o $(BINARY) ./cmd/golab/main.go
	@echo "ok\tbuild"

.PHONY: test
test:
	@go test -cover ./...

.PHONY: clean
clean:
	@rm -f $(BINARY)
	@echo "ok\tclean"
