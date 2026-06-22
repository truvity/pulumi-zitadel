//go:build integration

package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/require"

	"github.com/truvity/pulumi-zitadel/sdk/go/zitadel"
)

// newStack creates a Pulumi automation stack with local file backend and provider config.
func newStack(t *testing.T, ctx context.Context, stackName string, program pulumi.RunFunc) auto.Stack {
	t.Helper()

	tmpDir := t.TempDir()
	backendURL := "file://" + tmpDir

	// Local file backend needs a passphrase for secrets encryption.
	os.Setenv("PULUMI_CONFIG_PASSPHRASE", "test")

	// Ensure locally-built provider binary is in PATH.
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	binDir := filepath.Join(repoRoot, "bin")
	os.Setenv("PATH", binDir+":"+os.Getenv("PATH"))

	stack, err := auto.UpsertStackInlineSource(ctx, stackName, "pulumi-zitadel-test", program,
		auto.Project(workspace.Project{
			Name:    "pulumi-zitadel-test",
			Runtime: workspace.NewProjectRuntimeInfo("go", nil),
			Backend: &workspace.ProjectBackend{URL: backendURL},
		}),
	)
	require.NoError(t, err)

	// Configure the provider.
	require.NoError(t, stack.SetConfig(ctx, "zitadel:domain", auto.ConfigValue{Value: cfg.Domain}))
	require.NoError(t, stack.SetConfig(ctx, "zitadel:insecure", auto.ConfigValue{Value: boolStr(cfg.Insecure)}))
	require.NoError(t, stack.SetConfig(ctx, "zitadel:jwtProfileJson", auto.ConfigValue{Value: jwtKey, Secret: true}))

	// Register cleanup.
	t.Cleanup(func() {
		_, err := stack.Destroy(ctx, optdestroy.Message("integration test cleanup"))
		if err != nil {
			t.Logf("warning: destroy failed: %v", err)
		}
	})

	return stack
}

func TestOrganizationCRUD(t *testing.T) {
	ctx := context.Background()

	program := func(ctx *pulumi.Context) error {
		org, err := zitadel.NewOrg(ctx, "test-org", &zitadel.OrgArgs{
			Name: pulumi.String("pulumi-test-org"),
		})
		if err != nil {
			return err
		}

		ctx.Export("orgId", org.ID())
		ctx.Export("orgName", org.Name)
		return nil
	}

	stack := newStack(t, ctx, "test-org-crud", program)

	// UP — create.
	upResult, err := stack.Up(ctx, optup.Message("integration test: create org"))
	require.NoError(t, err)

	orgID, ok := upResult.Outputs["orgId"]
	require.True(t, ok, "orgId output missing")
	require.NotEmpty(t, orgID.Value, "orgId should not be empty")

	orgName, ok := upResult.Outputs["orgName"]
	require.True(t, ok, "orgName output missing")
	require.Equal(t, "pulumi-test-org", orgName.Value)
	t.Logf("✅ Organization created: id=%s name=%s", orgID.Value, orgName.Value)

	// REFRESH — verify state matches remote.
	_, err = stack.Refresh(ctx)
	require.NoError(t, err)
	t.Log("✅ Refresh succeeded")

	// DESTROY — clean up.
	_, err = stack.Destroy(ctx, optdestroy.Message("integration test: destroy org"))
	require.NoError(t, err)
	t.Log("✅ Destroy succeeded")
}

func TestProjectCRUD(t *testing.T) {
	ctx := context.Background()

	program := func(ctx *pulumi.Context) error {
		org, err := zitadel.NewOrg(ctx, "test-project-org", &zitadel.OrgArgs{
			Name: pulumi.String("pulumi-test-project-org"),
		})
		if err != nil {
			return err
		}

		proj, err := zitadel.NewProject(ctx, "test-project", &zitadel.ProjectArgs{
			Name:                   pulumi.String("pulumi-test-project"),
			OrgId:                  org.ID(),
			HasProjectCheck:        pulumi.Bool(true),
			ProjectRoleAssertion:   pulumi.Bool(false),
			ProjectRoleCheck:       pulumi.Bool(false),
			PrivateLabelingSetting: pulumi.String("PRIVATE_LABELING_SETTING_UNSPECIFIED"),
		})
		if err != nil {
			return err
		}

		ctx.Export("orgId", org.ID())
		ctx.Export("projectId", proj.ID())
		ctx.Export("projectName", proj.Name)
		return nil
	}

	stack := newStack(t, ctx, "test-project-crud", program)

	upResult, err := stack.Up(ctx, optup.Message("integration test: create project"))
	require.NoError(t, err)

	projectID, ok := upResult.Outputs["projectId"]
	require.True(t, ok, "projectId output missing")
	require.NotEmpty(t, projectID.Value)

	projectName, ok := upResult.Outputs["projectName"]
	require.True(t, ok, "projectName output missing")
	require.Equal(t, "pulumi-test-project", projectName.Value)
	t.Logf("✅ Project created: id=%s name=%s", projectID.Value, projectName.Value)

	_, err = stack.Refresh(ctx)
	require.NoError(t, err)
	t.Log("✅ Refresh succeeded")

	_, err = stack.Destroy(ctx, optdestroy.Message("integration test: destroy project"))
	require.NoError(t, err)
	t.Log("✅ Destroy succeeded")
}

func TestMachineUserCRUD(t *testing.T) {
	ctx := context.Background()

	program := func(ctx *pulumi.Context) error {
		org, err := zitadel.NewOrg(ctx, "test-mu-org", &zitadel.OrgArgs{
			Name: pulumi.String("pulumi-test-mu-org"),
		})
		if err != nil {
			return err
		}

		mu, err := zitadel.NewMachineUser(ctx, "test-machine-user", &zitadel.MachineUserArgs{
			UserName:        pulumi.String("pulumi-test-sa"),
			Name:            pulumi.String("Pulumi Test Service Account"),
			OrgId:           org.ID(),
			AccessTokenType: pulumi.String("ACCESS_TOKEN_TYPE_JWT"),
		})
		if err != nil {
			return err
		}

		ctx.Export("orgId", org.ID())
		ctx.Export("userId", mu.ID())
		ctx.Export("userName", mu.UserName)
		return nil
	}

	stack := newStack(t, ctx, "test-machine-user", program)

	upResult, err := stack.Up(ctx, optup.Message("integration test: create machine user"))
	require.NoError(t, err)

	userID, ok := upResult.Outputs["userId"]
	require.True(t, ok, "userId output missing")
	require.NotEmpty(t, userID.Value)

	userName, ok := upResult.Outputs["userName"]
	require.True(t, ok, "userName output missing")
	require.Equal(t, "pulumi-test-sa", userName.Value)
	t.Logf("✅ Machine user created: id=%s userName=%s", userID.Value, userName.Value)

	_, err = stack.Refresh(ctx)
	require.NoError(t, err)
	t.Log("✅ Refresh succeeded")

	_, err = stack.Destroy(ctx, optdestroy.Message("integration test: destroy machine user"))
	require.NoError(t, err)
	t.Log("✅ Destroy succeeded")
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
