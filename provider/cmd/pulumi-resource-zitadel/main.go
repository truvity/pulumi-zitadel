package main

import (
	"context"
	_ "embed"

	pftfbridge "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/pf/tfbridge"

	zitadel "github.com/truvity/pulumi-zitadel/provider"
)

//go:embed schema.json
var schema []byte

func main() {
	pftfbridge.MainWithMuxer(context.Background(), "zitadel", zitadel.Provider(), schema)
}
