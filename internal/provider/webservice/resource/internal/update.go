package internal

import (
	"context"
	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	"terraform-provider-render/internal/provider/webservice"
)

func UpdateServiceRequestFromModel(ctx context.Context, plan webservice.WebServiceModel, state webservice.WebServiceModel, ownerID string) (client.UpdateServiceJSONRequestBody, error) {
	envSpecificDetails, err := common.EnvSpecificDetailsForPATCH(plan.RuntimeSource, plan.StartCommand.ValueStringPointer())
	if err != nil {
		return client.UpdateServiceJSONRequestBody{}, err
	}

	servicePlan := client.PaidPlan(plan.Plan.ValueString())

	pullRequestPreviewsEnabled := client.PullRequestPreviewsEnabledNo
	if plan.PullRequestPreviewsEnabled.ValueBool() {
		pullRequestPreviewsEnabled = client.PullRequestPreviewsEnabledYes
	}

	preDeployCommand := ""
	if plan.PreDeployCommand.ValueStringPointer() != nil && *plan.PreDeployCommand.ValueStringPointer() != "" {
		preDeployCommand = *plan.PreDeployCommand.ValueStringPointer()
	}

	// Handle IP allow list with state-aware logic:
	// - In state but not in plan (null) -> send default (0.0.0.0/0) to revert to API default
	// - Not in state and not in plan -> send nil (don't update)
	// - In plan with empty list -> send empty array (block all)
	// - In plan with values -> send those values
	var ipAllowList *[]client.CidrBlockAndDescription
	if !plan.IPAllowList.IsNull() && !plan.IPAllowList.IsUnknown() {
		// Field is configured in plan
		list, err := common.ClientFromIPAllowList(plan.IPAllowList)
		if err != nil {
			return client.UpdateServiceJSONRequestBody{}, err
		}
		ipAllowList = &list
	} else if !state.IPAllowList.IsNull() {
		// Field was in state but removed from plan -> revert to default (0.0.0.0/0 everywhere)
		ipAllowList = &common.AllowAllCIDRList
	}

	webServiceDetails := client.WebServiceDetailsPATCH{
		Plan:                       &servicePlan,
		EnvSpecificDetails:         envSpecificDetails,
		HealthCheckPath:            plan.HealthCheckPath.ValueStringPointer(),
		PreDeployCommand:           &preDeployCommand,
		Previews:                   common.PreviewsObjectToPreviews(ctx, plan.Previews),
		PullRequestPreviewsEnabled: &pullRequestPreviewsEnabled,
		MaintenanceMode:            common.ToClientMaintenanceMode(plan.MaintenanceMode),
		MaxShutdownDelaySeconds:    common.ValueAsIntPointer(plan.MaxShutdownDelaySeconds),
		Runtime:                    common.From(client.ServiceRuntime(plan.RuntimeSource.Runtime())),
		IpAllowList:                ipAllowList,
	}

	serviceDetails := &client.ServicePATCH_ServiceDetails{}
	if err := serviceDetails.FromWebServiceDetailsPATCH(webServiceDetails); err != nil {
		return client.UpdateServiceJSONRequestBody{}, err
	}

	var updateServiceBody = client.UpdateServiceJSONRequestBody{
		Name:           plan.Name.ValueStringPointer(),
		RootDir:        plan.RootDirectory.ValueStringPointer(),
		ServiceDetails: serviceDetails,
	}

	if err := common.ApplyRuntimeSourceFieldsForUpdate(plan.RuntimeSource, &updateServiceBody, ownerID); err != nil {
		return client.UpdateServiceJSONRequestBody{}, err
	}

	return updateServiceBody, nil
}
