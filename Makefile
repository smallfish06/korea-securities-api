.PHONY: build run test clean deps mock \
	kis-spec-fetch kis-spec-generate kis-spec-refresh kis-spec-check kis-spec-all \
	kiwoom-spec-fetch kiwoom-spec-generate kiwoom-spec-refresh kiwoom-spec-check kiwoom-spec-all

# Build the application
build:
	go build -o bin/krsec ./cmd/krsec

# Run the application
run:
	go run ./cmd/krsec -config config.yaml

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
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

# Lint code
lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Fetch latest documented KIS snapshot from portal (network required)
kis-spec-fetch:
	go run ./cmd/kis-specgen fetch --out pkg/kis/specs/documented_endpoints.json

# Generate KIS documented spec/type Go files from snapshot
kis-spec-generate:
	go run ./cmd/kis-specgen generate --in pkg/kis/specs/documented_endpoints.json --spec-out pkg/kis/specs/documented_specs_generated.go --types-out pkg/kis/specs/documented_endpoint_types_generated.go

# Refresh snapshot + regenerate KIS documented Go files
kis-spec-refresh:
	go run ./cmd/kis-specgen refresh --snapshot pkg/kis/specs/documented_endpoints.json --spec-out pkg/kis/specs/documented_specs_generated.go --types-out pkg/kis/specs/documented_endpoint_types_generated.go

# Verify generated KIS documented files are up to date
kis-spec-check:
	go run ./cmd/kis-specgen check --in pkg/kis/specs/documented_endpoints.json --spec-out pkg/kis/specs/documented_specs_generated.go --types-out pkg/kis/specs/documented_endpoint_types_generated.go

# Run full KIS spec workflow end-to-end
kis-spec-all: kis-spec-fetch kis-spec-generate kis-spec-refresh kis-spec-check

# Fetch latest documented Kiwoom snapshot from portal (network required)
kiwoom-spec-fetch:
	go run ./cmd/kiwoom-specgen fetch --out pkg/kiwoom/specs/documented_endpoints.json

# Generate Kiwoom documented spec Go file from snapshot
kiwoom-spec-generate:
	go run ./cmd/kiwoom-specgen generate --in pkg/kiwoom/specs/documented_endpoints.json --spec-out pkg/kiwoom/specs/documented_specs_generated.go --types-out pkg/kiwoom/specs/documented_endpoint_types_generated.go

# Refresh snapshot + regenerate Kiwoom documented Go files
kiwoom-spec-refresh:
	go run ./cmd/kiwoom-specgen refresh --snapshot pkg/kiwoom/specs/documented_endpoints.json --spec-out pkg/kiwoom/specs/documented_specs_generated.go --types-out pkg/kiwoom/specs/documented_endpoint_types_generated.go

# Verify generated Kiwoom documented files are up to date
kiwoom-spec-check:
	go run ./cmd/kiwoom-specgen check --in pkg/kiwoom/specs/documented_endpoints.json --spec-out pkg/kiwoom/specs/documented_specs_generated.go --types-out pkg/kiwoom/specs/documented_endpoint_types_generated.go

# Run full Kiwoom spec workflow end-to-end
kiwoom-spec-all: kiwoom-spec-fetch kiwoom-spec-generate kiwoom-spec-refresh kiwoom-spec-check
