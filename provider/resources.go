package zitadel

import (
	"context"
	_ "embed"

	zitadel "github.com/zitadel/terraform-provider-zitadel/v2/zitadel"

	pf "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/pf/tfbridge"
	"github.com/pulumi/pulumi-terraform-bridge/v3/pkg/tfbridge"
	"github.com/pulumi/pulumi-terraform-bridge/v3/pkg/tfbridge/tokens"
	shimv2 "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/tfshim/sdk-v2"

	"github.com/truvity/pulumi-zitadel/provider/pkg/version"
)

//go:embed cmd/pulumi-resource-zitadel/bridge-metadata.json
var bridgeMetadata []byte

func Provider() tfbridge.ProviderInfo {
	prov := tfbridge.ProviderInfo{
		// Mux SDKv2 + Plugin Framework — mirrors upstream main.go.
		P: pf.MuxShimWithPF(context.Background(),
			shimv2.NewProvider(zitadel.Provider()),
			zitadel.NewProviderPV6(),
		),

		Name:        "zitadel",
		DisplayName: "ZITADEL",
		Publisher:   "Truvity",
		Version:     version.Version,
		Repository:  "https://github.com/truvity/pulumi-zitadel",

		GitHubOrg:               "zitadel",
		TFProviderModuleVersion: "v2",

		PluginDownloadURL: "github://api.github.com/truvity/pulumi-zitadel",
		MetadataInfo:      tfbridge.NewProviderMetadata(bridgeMetadata),

		Config: map[string]*tfbridge.SchemaInfo{},
	}

	prov.MustComputeTokens(tokens.SingleModule("zitadel_", "index",
		tokens.MakeStandard("zitadel")))
	prov.SetAutonaming(255, "-")
	prov.MustApplyAutoAliases()

	return prov
}
