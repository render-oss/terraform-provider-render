package webservice

import (
	"testing"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func wrappedServiceForMaintenanceMode(t *testing.T, servicePlan client.Plan, maintenanceMode *client.MaintenanceMode) *common.WrappedService {
	t.Helper()

	details := client.WebServiceDetails{
		Env:                client.ServiceEnvImage,
		Runtime:            client.ServiceRuntimeImage,
		EnvSpecificDetails: client.EnvSpecificDetails{},
		Plan:               servicePlan,
		Region:             client.Region("oregon"),
		NumInstances:       1,
		Url:                "https://test.onrender.com",
		MaintenanceMode:    maintenanceMode,
	}

	var serviceDetails client.Service_ServiceDetails
	require.NoError(t, serviceDetails.FromWebServiceDetails(details))

	imagePath := "nginx:latest"
	return &common.WrappedService{
		Service: &client.Service{
			Id:             "srv-1",
			Name:           "test-web-service",
			Slug:           "test-web-service",
			OwnerId:        "owner-123",
			ImagePath:      &imagePath,
			ServiceDetails: serviceDetails,
		},
	}
}

// Issue #80: the API omits maintenance_mode for free-tier services. Storing that as null
// would conflict with the schema's default object and cause a perpetual diff, so the read
// mapper must fall back to the default instead.
func TestModelForServiceResult_DefaultsMaintenanceModeWhenAPIOmitsIt(t *testing.T) {
	wrapped := wrappedServiceForMaintenanceMode(t, client.PlanFree, nil)

	plan := WebServiceModel{MaintenanceMode: common.DefaultMaintenanceMode()}
	model, err := ModelForServiceResult(wrapped, plan, diag.Diagnostics{})
	require.NoError(t, err)

	assert.False(t, model.MaintenanceMode.IsNull(), "maintenance_mode must not be null for free-tier services")
	assert.Equal(t, common.DefaultMaintenanceMode(), model.MaintenanceMode)
}

func TestModelForServiceResult_MapsMaintenanceModeWhenAPIReturnsIt(t *testing.T) {
	wrapped := wrappedServiceForMaintenanceMode(t, client.PlanStarter, &client.MaintenanceMode{
		Enabled: true,
		Uri:     "https://status.example.com",
	})

	plan := WebServiceModel{MaintenanceMode: common.DefaultMaintenanceMode()}
	model, err := ModelForServiceResult(wrapped, plan, diag.Diagnostics{})
	require.NoError(t, err)

	require.False(t, model.MaintenanceMode.IsNull())
	attrs := model.MaintenanceMode.Attributes()
	assert.Equal(t, "true", attrs["enabled"].String())
	assert.Equal(t, `"https://status.example.com"`, attrs["uri"].String())
}
