package internal

import (
	"context"
	"testing"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/backgroundworker"
	"terraform-provider-render/internal/provider/common"
	commontypes "terraform-provider-render/internal/provider/common/types"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateServiceRequestFromModel_IncludesRuntime(t *testing.T) {
	ctx := context.Background()
	ownerID := "owner-123"

	plan := backgroundWorker.BackgroundWorkerModel{
		Name: types.StringValue("test-worker"),
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
	backgroundWorkerDetails, err := req.ServiceDetails.AsBackgroundWorkerDetailsPATCH()
	require.NoError(t, err)

	// Verify Runtime field is set in service details
	assert.NotNil(t, backgroundWorkerDetails.Runtime)
	assert.Equal(t, client.ServiceRuntimeImage, *backgroundWorkerDetails.Runtime)
}

func TestUpdateServiceRequestFromModel_RuntimeDocker(t *testing.T) {
	ctx := context.Background()
	ownerID := "owner-123"

	plan := backgroundWorker.BackgroundWorkerModel{
		Name: types.StringValue("test-worker"),
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
	backgroundWorkerDetails, err := req.ServiceDetails.AsBackgroundWorkerDetailsPATCH()
	require.NoError(t, err)

	// Verify Runtime field is set to docker
	assert.NotNil(t, backgroundWorkerDetails.Runtime)
	assert.Equal(t, client.ServiceRuntimeDocker, *backgroundWorkerDetails.Runtime)
}
