package internal

import (
	"context"
	"testing"

	"terraform-provider-render/internal/client"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Issue #80: maintenance mode can only be configured for non-free tier services, so it
// must be omitted from the create payload on the free plan.
func TestCreateServiceRequestFromModel_OmitsMaintenanceModeOnFreePlan(t *testing.T) {
	ctx := context.Background()

	plan := webServiceModelForPlan(string(client.PlanFree))
	req, err := CreateServiceRequestFromModel(ctx, "owner-123", plan)
	require.NoError(t, err)

	details, err := req.ServiceDetails.AsWebServiceDetailsPOST()
	require.NoError(t, err)

	assert.Nil(t, details.MaintenanceMode, "maintenance_mode must be omitted for free-tier services")
}

func TestCreateServiceRequestFromModel_IncludesMaintenanceModeOnPaidPlan(t *testing.T) {
	ctx := context.Background()

	plan := webServiceModelForPlan(string(client.PlanStarter))
	req, err := CreateServiceRequestFromModel(ctx, "owner-123", plan)
	require.NoError(t, err)

	details, err := req.ServiceDetails.AsWebServiceDetailsPOST()
	require.NoError(t, err)

	require.NotNil(t, details.MaintenanceMode, "maintenance_mode must be sent for paid-tier services")
	assert.False(t, details.MaintenanceMode.Enabled)
}
