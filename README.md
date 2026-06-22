# Pulumi ZITADEL Provider

[![CI](https://github.com/truvity/pulumi-zitadel/actions/workflows/ci.yaml/badge.svg)](https://github.com/truvity/pulumi-zitadel/actions/workflows/ci.yaml)
[![Go Reference](https://pkg.go.dev/badge/github.com/truvity/pulumi-zitadel/sdk/go/zitadel.svg)](https://pkg.go.dev/github.com/truvity/pulumi-zitadel/sdk/go/zitadel)
[![Go Report Card](https://goreportcard.com/badge/github.com/truvity/pulumi-zitadel/sdk)](https://goreportcard.com/report/github.com/truvity/pulumi-zitadel/sdk)
[![License](https://img.shields.io/github/license/truvity/pulumi-zitadel)](LICENSE)

A Pulumi provider for [ZITADEL](https://zitadel.com), bridged from the official
[Terraform provider](https://github.com/zitadel/terraform-provider-zitadel) (v3.2.2).

## SDK Documentation

**→ [Go SDK usage guide](sdk/go/zitadel/README.md)** — installation, configuration, examples.

---

## Maintainer Guide

Everything below is for provider maintainers.

### Prerequisites

- [devbox](https://www.jetify.com/docs/devbox/) (provisions Go, Pulumi, golangci-lint, etc.)

```bash
devbox shell   # or let direnv auto-activate
```

### Building

```bash
# Full build: tfgen → schema → SDK → provider binary
just build

# Generate schema + Go SDK only (no provider binary)
just generate

# Build provider binary only (assumes generate already ran)
just provider
```

### Targets

| Target              | Description                                                  |
| ------------------- | ------------------------------------------------------------ |
| `just generate`     | Build tfgen, generate schema + bridge-metadata + Go SDK      |
| `just build`        | Full build (generate + provider binary + SDK compile check)  |
| `just check`        | Build + lint + vuln + verify-generate                        |
| `just verify-generate` | Fail if generated files are out of date                   |
| `just install`      | Install provider plugin locally for testing                  |
| `just lint`         | Lint provider module                                         |
| `just vuln`         | govulncheck on provider module                               |
| `just tidy`         | go mod tidy on all modules                                   |
| `just snapshot`     | GoReleaser snapshot (cross-platform binaries, no publish)    |
| `just test-integration` | Integration tests against real Zitadel Cloud            |

### Architecture

This is a **muxed provider** — the upstream ZITADEL Terraform provider serves two halves
behind a single protocol-v6 mux:

- **SDKv2 provider**: `zitadel.Provider()` (upgraded v5→v6 via `tf5to6server`)
- **Plugin Framework provider**: `zitadel.NewProviderPV6()`

The Pulumi bridge muxes them via `pf.MuxShimWithPF(ctx, sdkv2Shim, pfProvider)`.

Key implication: the tfgen binary must use `pf/tfgen.MainWithMuxer` (generates the `"mux"`
dispatch table in `bridge-metadata.json`), and the provider binary must use
`pf/tfbridge.MainWithMuxer` (reads the dispatch table at runtime).

### Upgrading the upstream provider

```bash
cd provider
go get github.com/zitadel/terraform-provider-zitadel/v2@vNEW
go mod tidy
cd ..
just generate   # regenerate schema + SDK
just build      # verify everything compiles
# Review changes in sdk/go/ and bridge-metadata.json
git add -A && git commit -m "chore(deps): bump upstream provider to vNEW"
```

### Upgrading the bridge

```bash
cd provider
go get github.com/pulumi/pulumi-terraform-bridge/v3@vX.Y.Z

# CRITICAL: update the replace directive to match the bridge's go.mod:
# https://github.com/pulumi/pulumi-terraform-bridge/blob/vX.Y.Z/go.mod
# Copy the exact: replace github.com/hashicorp/terraform-plugin-sdk/v2 => ...

go mod tidy
cd ..
just generate
just build
```

### Releasing

**Versioning strategy**: We mirror the upstream Terraform provider's semver. When upstream
releases v3.2.2, we tag `v3.2.2`. For bridge-only fixes on the same upstream version, use
pre-release suffixes: `v3.2.2-truvity.1`.

1. Ensure `just check` passes (CI enforces this)
2. Tag: `git tag v3.2.2 && git push origin v3.2.2`
3. CI builds cross-platform provider binaries and creates a GitHub Release
4. Consumers auto-download via `PluginDownloadURL: github://api.github.com/truvity/pulumi-zitadel`

### Module Layout

```
pulumi-zitadel/
├── provider/              # Provider module (bridge wiring)
│   ├── resources.go       # ProviderInfo: mux config, token mapping
│   ├── cmd/
│   │   ├── pulumi-resource-zitadel/  # Provider binary (MainWithMuxer)
│   │   │   ├── schema.json           # Generated Pulumi schema
│   │   │   └── bridge-metadata.json  # Generated mux dispatch table
│   │   └── pulumi-tfgen-zitadel/     # Schema generator (MainWithMuxer)
│   └── pkg/version/       # Version injected via ldflags
├── sdk/                   # SDK module (committed, tagged)
│   └── go/zitadel/        # Generated Go SDK (106 resources, 68 functions)
└── tests/integration/     # Automation API tests against real Zitadel
```

### Integration Tests

Tests use the same Zitadel Cloud Pro instance as zitadel-operator:

- **Credentials**: JWT key from system keyring (`service=zitadel-operator`, `username=jwt-key`)
- **Config**: `~/.config/zitadel-operator/config.yaml` (domain, port)
- **Backend**: Pulumi local file backend in `t.TempDir()` (no persistent state)

```bash
just test-integration
```
