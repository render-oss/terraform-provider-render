package internal

import (
	"testing"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	commontypes "terraform-provider-render/internal/provider/common/types"
	"terraform-provider-render/internal/provider/cronjob"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateServiceRequestFromModel_IncludesRuntime(t *testing.T) {
	ownerID := "owner-123"

	plan := cronJob.CronJobModel{
		Name:     types.StringValue("test-cronjob"),
		Plan:     types.StringValue("starter"),
		Schedule: types.StringValue("0 0 * * *"),
		RuntimeSource: &common.RuntimeSourceModel{
			Image: &common.ImageRuntimeSourceModel{
				ImageURL: commontypes.ImageURLStringValue{StringValue: types.StringValue("nginx:latest")},
			},
		},
	}

	req, err := UpdateServiceRequestFromModel(plan, ownerID)
	require.NoError(t, err)

	// Extract the service details to verify runtime is set
	cronJobDetails, err := req.ServiceDetails.AsCronJobDetailsPATCH()
	require.NoError(t, err)

	// Verify Runtime field is set in service details
	assert.NotNil(t, cronJobDetails.Runtime)
	assert.Equal(t, client.ServiceRuntimeImage, *cronJobDetails.Runtime)
}

func TestUpdateServiceRequestFromModel_RuntimeDocker(t *testing.T) {
	ownerID := "owner-123"

	plan := cronJob.CronJobModel{
		Name:     types.StringValue("test-cronjob"),
		Plan:     types.StringValue("starter"),
		Schedule: types.StringValue("0 0 * * *"),
		RuntimeSource: &common.RuntimeSourceModel{
			Docker: &common.DockerRuntimeSourceModel{
				RepoURL: types.StringValue("https://github.com/test/repo"),
			},
		},
	}

	req, err := UpdateServiceRequestFromModel(plan, ownerID)
	require.NoError(t, err)

	// Extract the service details to verify runtime is set
	cronJobDetails, err := req.ServiceDetails.AsCronJobDetailsPATCH()
	require.NoError(t, err)

	// Verify Runtime field is set to docker
	assert.NotNil(t, cronJobDetails.Runtime)
	assert.Equal(t, client.ServiceRuntimeDocker, *cronJobDetails.Runtime)
}