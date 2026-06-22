# Pulumi ZITADEL Provider

A Pulumi provider for [ZITADEL](https://zitadel.com), bridged from the official
[Terraform provider](https://github.com/zitadel/terraform-provider-zitadel).

## Go SDK

```bash
go get github.com/truvity/pulumi-zitadel/sdk/go/zitadel@latest
```

## Configuration

The provider accepts the same configuration as the upstream Terraform provider:

- `zitadel:domain` — ZITADEL instance domain
- `zitadel:insecure` — Use HTTP instead of HTTPS
- `zitadel:port` — Custom port
- `zitadel:token` — Access token (or use `jwt_profile_json`)
- `zitadel:jwt_profile_json` — JWT profile JSON for service account auth

## Usage

```go
package main

import (
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
    "github.com/truvity/pulumi-zitadel/sdk/go/zitadel"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        _, err := zitadel.NewOrg(ctx, "demo", &zitadel.OrgArgs{
            Name: pulumi.String("demo-org"),
        })
        return err
    })
}
```

## Plugin Installation

The plugin auto-installs via `PluginDownloadURL` when using the SDK. To manually install:

```bash
pulumi plugin install resource zitadel v0.1.0 \
  --server github://api.github.com/truvity/pulumi-zitadel
```

## Development

Prerequisites: [devbox](https://www.jetify.com/docs/devbox/)

```bash
# Enter devbox shell (or use direnv)
devbox shell

# Build everything
just build

# Install provider locally for testing
just install

# Run all checks
just check

# Build snapshot release
just snapshot
```

### Upgrading the upstream provider

```bash
cd provider
go get github.com/zitadel/terraform-provider-zitadel/v2@vNEW
go mod tidy
cd ..
just build
```

### Upgrading the bridge

```bash
cd provider
go get github.com/pulumi/pulumi-terraform-bridge/v3@vX.Y.Z
# Update the replace directive in go.mod to match the bridge's go.mod:
# https://github.com/pulumi/pulumi-terraform-bridge/blob/vX.Y.Z/go.mod
go mod tidy
cd ..
just build
```

## Architecture

This is a muxed provider — ZITADEL's Terraform provider serves two halves behind a
single protocol-v6 mux:

- SDKv2 provider: `zitadel.Provider()` (upgraded v5→v6 via `tf5to6server`)
- Plugin Framework provider: `zitadel.NewProviderPV6()`

The Pulumi bridge muxes them via `pf.MuxShimWithPF(ctx, sdkv2Shim, pfProvider)`.
