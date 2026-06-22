package main

import (
	"github.com/pulumi/pulumi-terraform-bridge/v3/pkg/tfgen"

	zitadel "github.com/truvity/pulumi-zitadel/provider"
	"github.com/truvity/pulumi-zitadel/provider/pkg/version"
)

func main() {
	tfgen.Main("zitadel", version.Version, zitadel.Provider())
}
