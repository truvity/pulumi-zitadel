# Development commands for pulumi-zitadel bridged provider

# Disable go.work (parent workspace interferes with standalone module builds)
export GOWORK := "off"

PACK := "zitadel"
PROJECT := "github.com/truvity/pulumi-zitadel"
PROVIDER_PATH := "provider"
VERSION_PATH := PROVIDER_PATH + "/pkg/version.Version"
CODEGEN := "pulumi-tfgen-" + PACK
PROVIDER := "pulumi-resource-" + PACK
PROVIDER_VERSION := env("PROVIDER_VERSION", "0.0.1-dev")
WORKING_DIR := justfile_directory()

LDFLAGS := "-s -w -X " + PROJECT + "/" + VERSION_PATH + "=" + PROVIDER_VERSION

# Build the tfgen binary (generates schema + SDKs)
tfgen: ensure-dirs
    cd provider && go build -o ../bin/{{CODEGEN}} -ldflags "{{LDFLAGS}}" ./cmd/{{CODEGEN}}/

# Generate the Pulumi schema and bridge metadata
schema: tfgen
    ./bin/{{CODEGEN}} schema --out provider/cmd/{{PROVIDER}}

# Generate the Go SDK from the schema
generate: schema
    ./bin/{{CODEGEN}} go --out sdk/go/

# Build the Go SDK (compile check)
build-sdk: generate
    cd sdk && go build ./...

# Build the provider binary
provider: ensure-dirs
    cd provider && go build -o ../bin/{{PROVIDER}} -ldflags "{{LDFLAGS}}" ./cmd/{{PROVIDER}}/

# Build everything (provider + SDK)
build: provider build-sdk

# Install the provider plugin locally for testing
install: provider
    mkdir -p ~/.pulumi/plugins/resource-{{PACK}}-v{{PROVIDER_VERSION}}
    cp bin/{{PROVIDER}} ~/.pulumi/plugins/resource-{{PACK}}-v{{PROVIDER_VERSION}}/

# Run linters on the provider module
lint:
    cd provider && golangci-lint run ./...

# Run linters on the SDK module
lint-sdk:
    cd sdk && golangci-lint run ./...

# Run Go vulnerability check
vuln:
    cd provider && govulncheck ./...

# Run go mod tidy on both modules
tidy:
    cd provider && go mod tidy
    cd sdk && go mod tidy

# Clean build artifacts
clean:
    rm -rf bin/ dist/ .make/ .pulumi/

# Run all checks (build + lint + vuln)
check: build lint vuln

# Build a snapshot release locally (cross-platform provider binaries)
snapshot:
    goreleaser release --snapshot --clean

# Ensure output directories exist
[private]
ensure-dirs:
    mkdir -p bin
