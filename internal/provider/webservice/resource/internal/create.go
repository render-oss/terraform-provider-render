package internal

import (
	"context"
	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	"terraform-provider-render/internal/provider/webservice"
)

func CreateServiceRequestFromModel(ctx context.Context, ownerID string, plan webservice.WebServiceModel) (client.CreateServiceJSONRequestBody, error) {
	envSpecificDetails, err := common.EnvSpecificDetailsForRuntimeSource(
		plan.RuntimeSource.Runtime(),
		plan.RuntimeSource,
		plan.StartCommand.ValueStringPointer(),
	)
	if err != nil {
		return client.CreateServiceJSONRequestBody{}, err
	}

	numInstances := int(plan.NumInstances.ValueInt64())
	if numInstances == 0 {
		numInstances = 1
	}

	servicePlan := client.PaidPlan(plan.Plan.ValueString())

	pullRequestPreviewsEnabled := client.PullRequestPreviewsEnabledNo
	if plan.PullRequestPreviewsEnabled.ValueBool() {
		pullRequestPreviewsEnabled = client.PullRequestPreviewsEnabledYes
	}

	region := client.Region(plan.Region.ValueString())

	autoscaling, err := common.AutoscalingRequest(plan.Autoscaling)
	if err != nil {
		return client.CreateServiceJSONRequestBody{}, err
	}

	// Handle IP allow list:
	// - Omitted (null) -> send nil (API uses default 0.0.0.0/0)
	// - Empty list -> send empty array (block all)
	// - With values -> send those values
	var ipAllowList *[]client.CidrBlockAndDescription
	if !plan.IPAllowList.IsNull() && !plan.IPAllowList.IsUnknown() {
		list, err := common.ClientFromIPAllowList(plan.IPAllowList)
		if err != nil {
			return client.CreateServiceJSONRequestBody{}, err
		}
		ipAllowList = &list
	}
	// If null/unknown, ipAllowList stays nil

	webServiceDetails := client.WebServiceDetailsPOST{
		Disk:                       common.DiskToClientCreate(plan.Disk),
		Runtime:                    common.ToClientRuntime(plan.RuntimeSource.Runtime()),
		EnvSpecificDetails:         envSpecificDetails,
		HealthCheckPath:            plan.HealthCheckPath.ValueStringPointer(),
		NumInstances:               &numInstances,
		Plan:                       &servicePlan,
		PreDeployCommand:           plan.PreDeployCommand.ValueStringPointer(),
		Previews:                   common.PreviewsObjectToPreviews(ctx, plan.Previews),
		PullRequestPreviewsEnabled: &pullRequestPreviewsEnabled,
		Region:                     &region,
		MaintenanceMode:            common.ToClientMaintenanceMode(plan.MaintenanceMode),
		MaxShutdownDelaySeconds:    common.ValueAsIntPointer(plan.MaxShutdownDelaySeconds),
		Autoscaling:                autoscaling,
		IpAllowList:                ipAllowList,
	}

	serviceDetails := &client.ServicePOST_ServiceDetails{}
	if err := serviceDetails.FromWebServiceDetailsPOST(webServiceDetails); err != nil {
		return client.CreateServiceJSONRequestBody{}, err
	}

	envVars, err := common.EnvVarsToClient(plan.EnvVars)
	if err != nil {
		return client.CreateServiceJSONRequestBody{}, err
	}

	secretFiles := common.SecretFilesToClient(plan.SecretFiles)
	var createServiceBody = client.CreateServiceJSONRequestBody{
		EnvVars:        &envVars,
		Name:           plan.Name.ValueString(),
		OwnerId:        ownerID,
		RootDir:        plan.RootDirectory.ValueStringPointer(),
		SecretFiles:    &secretFiles,
		ServiceDetails: serviceDetails,
		Type:           client.WebService,
	}

	if err := common.ApplyRuntimeSourceFieldsForCreate(plan.RuntimeSource, &createServiceBody); err != nil {
		return client.CreateServiceJSONRequestBody{}, err
	}

	return createServiceBody, nil
}
