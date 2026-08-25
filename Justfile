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

# Example conversion inside tfgen shells out to the terraform converter
# plugin, which pulumi downloads LATEST when it is not installed —
# v1.4.0 reordered example properties and broke verify-generate on
# every open PR (2026-08-25) while warm local caches still held v1.3.0.
# Installing the exact version first makes generation deterministic;
# bump this pin and regenerate in the same commit.
CONVERTER_TERRAFORM_VERSION := "1.4.0"

LDFLAGS := "-s -w -X " + PROJECT + "/" + VERSION_PATH + "=" + PROVIDER_VERSION

# Format all Go files (gofmt + goimports via golangci-lint)
fmt:
    cd provider && golangci-lint fmt ./...

# Build the tfgen binary (schema + SDK generator)
tfgen: ensure-dirs
    cd provider && go build -o ../bin/{{CODEGEN}} -ldflags "{{LDFLAGS}}" ./cmd/{{CODEGEN}}/

# Generate the Pulumi schema, bridge metadata (with mux dispatch table), and Go SDK
generate: tfgen
    pulumi plugin install converter terraform {{CONVERTER_TERRAFORM_VERSION}}
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
# NOT `vuln`. A required check must be something a pull request can act
# on, and govulncheck is not: it went red here on 5 reachable
# standard-library advisories whose only fix is go1.26.6, which nixpkgs
# does not carry yet (INF-553), and every pull request in this repo
# became unmergeable on a finding nobody could act on -- including the
# one that would have fixed it.
#
# Coverage is unchanged: security.yaml still runs `just vuln` on every
# PR and daily. It simply cannot block a merge.
check: build lint verify-generate

# Build a snapshot release locally (cross-platform provider binaries)
snapshot:
    goreleaser release --snapshot --clean

# Ensure output directories exist
[private]
ensure-dirs:
    mkdir -p bin
