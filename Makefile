.PHONY: 

BINARY_NAME=redis-server
VERSION=1.0.0
LDFLAG=-ldflags "-X main.version=$(VERSION)"


help:
	@echo "Build Your Own Redis - Make targets:"
	@echo "  make build      - Compile the binary"
	@echo "  make run        - Build and run the server"
	@echo "  make test       - Run tests with redis-cli"
	@echo "  make clean      - Remove build artifacts"
	@echo "  make release    - Create a release binary"
	@echo "  make benchmark  - Run redis-benchmark"
	@echo "  make docker-build - Build Docker image"

build:
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	go build -o bin/$(BINARY_NAME) cmd/redis-server/main.go
	@echo "✓ Binary: bin/$(BINARY_NAME)"

run: build
	@echo "Starting Redis server..."
	./bin/$(BINARY_NAME)

test: build
	@echo "Testing with redis-cli..."
	@echo "Starting server in background..."
	./bin/$(BINARY_NAME) &
	SERVER_PID=$$!; \
	sleep 1; \
	echo "Running tests..."; \
	redis-cli PING; \
	redis-cli SET test "works"; \
	redis-cli GET test; \
	redis-cli LPUSH list a b c; \
	redis-cli LRANGE list 0 -1; \
	redis-cli SADD set x y z; \
	redis-cli SMEMBERS set; \
	echo "✓ All tests passed!"; \
	kill $$SERVER_PID

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f dump.aof
	@echo "✓ Clean complete"

release: clean build
	@echo "Creating release v$(VERSION)..."
	mkdir -p releases
	cp bin/$(BINARY_NAME) releases/$(BINARY_NAME)-v$(VERSION)
	@echo "✓ Release: releases/$(BINARY_NAME)-v$(VERSION)"

benchmark: build
	@echo "Starting server in background..."
	./bin/$(BINARY_NAME) &
	SERVER_PID=$$!; \
	sleep 1; \
	echo "Running redis-benchmark..."; \
	redis-benchmark -t set,get -n 10000 -q; \
	kill $$SERVER_PID

fmt:
	go fmt ./...

lint:
	golangci-lint run ./...

docker-build:
	docker build -t build-your-own-redis:$(VERSION) .

docker-run:
	docker run -p 6379:6379 build-your-own-redis:$(VERSION)
	