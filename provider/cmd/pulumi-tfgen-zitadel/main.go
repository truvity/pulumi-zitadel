package main

import (
	pftfgen "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/pf/tfgen"

	zitadel "github.com/truvity/pulumi-zitadel/provider"
)

func main() {
	pftfgen.MainWithMuxer("zitadel", zitadel.Provider())
}
