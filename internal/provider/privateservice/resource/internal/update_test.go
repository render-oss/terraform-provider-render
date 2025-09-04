package internal

import (
	"context"
	"testing"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	commontypes "terraform-provider-render/internal/provider/common/types"
	"terraform-provider-render/internal/provider/privateservice"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateServiceRequestFromModel_IncludesRuntime(t *testing.T) {
	ctx := context.Background()
	ownerID := "owner-123"

	plan := privateservice.PrivateServiceModel{
		Name: types.StringValue("test-service"),
		Plan: types.StringValue("starter"),
		RuntimeSource: &common.RuntimeSourceModel{
			Image: &common.ImageRuntimeSourceModel{
				ImageURL: commontypes.ImageURLStringValue{StringValue: types.StringValue("nginx:latest")},
			},
		},
		PullRequestPreviewsEnabled: types.BoolValue(false),
		PreDeployCommand:           types.StringNull(),
	}

	req, err := UpdateServiceRequestFromModel(ctx, plan, ownerID)
	require.NoError(t, err)

	// Extract the service details to verify runtime is set
	privateServiceDetails, err := req.ServiceDetails.AsPrivateServiceDetailsPATCH()
	require.NoError(t, err)

	// Verify Runtime field is set in service details
	assert.NotNil(t, privateServiceDetails.Runtime)
	assert.Equal(t, client.ServiceRuntimeImage, *privateServiceDetails.Runtime)
}

func TestUpdateServiceRequestFromModel_RuntimeNative(t *testing.T) {
	ctx := context.Background()
	ownerID := "owner-123"

	plan := privateservice.PrivateServiceModel{
		Name: types.StringValue("test-service"),
		Plan: types.StringValue("starter"),
		RuntimeSource: &common.RuntimeSourceModel{
			NativeRuntime: &common.NativeRuntimeModel{
				RepoURL: types.StringValue("https://github.com/test/repo"),
				Runtime: types.StringValue("node"),
			},
		},
		PullRequestPreviewsEnabled: types.BoolValue(false),
		PreDeployCommand:           types.StringNull(),
	}

	req, err := UpdateServiceRequestFromModel(ctx, plan, ownerID)
	require.NoError(t, err)

	// Extract the service details to verify runtime is set
	privateServiceDetails, err := req.ServiceDetails.AsPrivateServiceDetailsPATCH()
	require.NoError(t, err)

	// Verify Runtime field is set to native runtime (node in this case)
	assert.NotNil(t, privateServiceDetails.Runtime)
	assert.Equal(t, client.ServiceRuntimeNode, *privateServiceDetails.Runtime)
}

func TestUpdateServiceRequestFromModel_RuntimeDocker(t *testing.T) {
	ctx := context.Background()
	ownerID := "owner-123"

	plan := privateservice.PrivateServiceModel{
		Name: types.StringValue("test-service"),
		Plan: types.StringValue("starter"),
		RuntimeSource: &common.RuntimeSourceModel{
			Docker: &common.DockerRuntimeSourceModel{
				RepoURL: types.StringValue("https://github.com/test/repo"),
			},
		},
		PullRequestPreviewsEnabled: types.BoolValue(false),
		PreDeployCommand:           types.StringNull(),
	}

	req, err := UpdateServiceRequestFromModel(ctx, plan, ownerID)
	require.NoError(t, err)

	// Extract the service details to verify runtime is set
	privateServiceDetails, err := req.ServiceDetails.AsPrivateServiceDetailsPATCH()
	require.NoError(t, err)

	// Verify Runtime field is set to docker
	assert.NotNil(t, privateServiceDetails.Runtime)
	assert.Equal(t, client.ServiceRuntimeDocker, *privateServiceDetails.Runtime)
}