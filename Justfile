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

LDFLAGS := "-s -w -X " + PROJECT + "/" + VERSION_PATH + "=" + PROVIDER_VERSION

# Format all Go files (gofmt + goimports via golangci-lint)
fmt:
    cd provider && golangci-lint fmt ./...

# Build the tfgen binary (schema + SDK generator)
tfgen: ensure-dirs
    cd provider && go build -o ../bin/{{CODEGEN}} -ldflags "{{LDFLAGS}}" ./cmd/{{CODEGEN}}/

# Generate the Pulumi schema, bridge metadata (with mux dispatch table), and Go SDK
generate: tfgen
    ./bin/{{CODEGEN}} schema --out provider/cmd/{{PROVIDER}}
    ./bin/{{CODEGEN}} go --out sdk/go/

# Build the Go SDK (compile check)
build-sdk: generate
    cd sdk && go build ./...

# Build the provider binary
provider: ensure-dirs generate
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

# Run go mod tidy on all modules
tidy:
    cd provider && go mod tidy
    cd sdk && go mod tidy
    cd tests/integration && go mod tidy

# Verify generated files are committed (fails if generate produces uncommitted changes)
verify-generate: generate
    @echo "Checking for uncommitted generated files..."
    @git diff --exit-code -- provider/cmd/pulumi-resource-zitadel/schema.json provider/cmd/pulumi-resource-zitadel/bridge-metadata.json sdk/go/ || (echo "ERROR: Generated files are out of date. Run 'just generate' and commit." && exit 1)
    @echo "✅ Generated files are up to date."

# Run integration tests (requires real Zitadel + credentials in keyring)
test-integration: provider
    cd tests/integration && go test -tags=integration -v ./... -count=1 -timeout=120s

# Clean build artifacts
clean:
    rm -rf bin/ dist/ .make/ .pulumi/

# Run all checks (build + lint + vuln + verify-generate)
check: build lint vuln verify-generate

# Build a snapshot release locally (cross-platform provider binaries)
snapshot:
    goreleaser release --snapshot --clean

# Ensure output directories exist
[private]
ensure-dirs:
    mkdir -p bin
