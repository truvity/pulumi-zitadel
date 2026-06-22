package main

import (
	_ "embed"

	"github.com/pulumi/pulumi-terraform-bridge/v3/pkg/tfbridge"

	zitadel "github.com/truvity/pulumi-zitadel/provider"
	"github.com/truvity/pulumi-zitadel/provider/pkg/version"
)

//go:embed bridge-metadata.json
var bridgeMetadata []byte

func main() {
	tfbridge.Main("zitadel", version.Version, zitadel.Provider(), bridgeMetadata)
}
