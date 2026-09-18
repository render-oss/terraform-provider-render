package internal

import (
	"context"
	"testing"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	commontypes "terraform-provider-render/internal/provider/common/types"
	"terraform-provider-render/internal/provider/webservice"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func webServiceModelForPlan(servicePlan string) webservice.WebServiceModel {
	return webservice.WebServiceModel{
		Name:            types.StringValue("test-web-service"),
		Plan:            types.StringValue(servicePlan),
		MaintenanceMode: common.DefaultMaintenanceMode(),
		RuntimeSource: &common.RuntimeSourceModel{
			Image: &common.ImageRuntimeSourceModel{
				ImageURL: commontypes.ImageURLStringValue{StringValue: types.StringValue("nginx:latest")},
			},
		},
		PullRequestPreviewsEnabled: types.BoolValue(false),
		PreDeployCommand:           types.StringNull(),
	}
}

// Issue #80: maintenance mode can only be configured for non-free tier services, so it
// must be omitted from the update payload on the free plan or the API rejects the request.
func TestUpdateServiceRequestFromModel_OmitsMaintenanceModeOnFreePlan(t *testing.T) {
	ctx := context.Background()

	plan := webServiceModelForPlan(string(client.PlanFree))
	req, err := UpdateServiceRequestFromModel(ctx, plan, webservice.WebServiceModel{}, "owner-123")
	require.NoError(t, err)

	details, err := req.ServiceDetails.AsWebServiceDetailsPATCH()
	require.NoError(t, err)

	assert.Nil(t, details.MaintenanceMode, "maintenance_mode must be omitted for free-tier services")
}

func TestUpdateServiceRequestFromModel_IncludesMaintenanceModeOnPaidPlan(t *testing.T) {
	ctx := context.Background()

	plan := webServiceModelForPlan(string(client.PlanStarter))
	req, err := UpdateServiceRequestFromModel(ctx, plan, webservice.WebServiceModel{}, "owner-123")
	require.NoError(t, err)

	details, err := req.ServiceDetails.AsWebServiceDetailsPATCH()
	require.NoError(t, err)

	require.NotNil(t, details.MaintenanceMode, "maintenance_mode must be sent for paid-tier services")
	assert.False(t, details.MaintenanceMode.Enabled)
}
