.PHONY: build run test clean deps mock

# Build the application
build:
	go build -o bin/kr-broker ./cmd/kr-broker

# Run the application
run:
	go run ./cmd/kr-broker -config config.yaml

# Run tests
test:
	go test -v ./...

# Generate mocks
mock:
	go run github.com/vektra/mockery/v3@v3.6.4 --config .mockery.yml

# Clean build artifacts
clean:
	rm -rf bin/

# Download dependencies
deps:
	go mod download
	go mod tidy

# Install development tools
dev-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Lint code
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...
