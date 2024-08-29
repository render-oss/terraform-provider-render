package internal

import (
	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/backgroundworker"
	"terraform-provider-render/internal/provider/common"
)

func CreateServiceRequestFromModel(ownerID string, plan backgroundWorker.BackgroundWorkerModel) (client.CreateServiceJSONRequestBody, error) {
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

	backgroundWorkerDetails := client.BackgroundWorkerDetailsPOST{
		Autoscaling:                autoscaling,
		Disk:                       common.DiskToClientCreate(plan.Disk),
		Runtime:                    client.ServiceRuntime(plan.RuntimeSource.Runtime()),
		EnvSpecificDetails:         envSpecificDetails,
		NumInstances:               &numInstances,
		Plan:                       &servicePlan,
		PreDeployCommand:           plan.PreDeployCommand.ValueStringPointer(),
		PullRequestPreviewsEnabled: &pullRequestPreviewsEnabled,
		Region:                     &region,
		MaxShutdownDelaySeconds:    common.ValueAsIntPointer(plan.MaxShutdownDelaySeconds),
	}

	serviceDetails := &client.ServicePOST_ServiceDetails{}
	if err := serviceDetails.FromBackgroundWorkerDetailsPOST(backgroundWorkerDetails); err != nil {
		return client.CreateServiceJSONRequestBody{}, err
	}

	envVars, err := common.EnvVarsToClient(plan.EnvVars)
	if err != nil {
		return client.CreateServiceJSONRequestBody{}, err
	}

	secretFiles := common.SecretFilesToClient(plan.SecretFiles)
	var createServiceBody = client.CreateServiceJSONRequestBody{
		Name:           plan.Name.ValueString(),
		OwnerId:        ownerID,
		EnvVars:        &envVars,
		RootDir:        plan.RootDirectory.ValueStringPointer(),
		SecretFiles:    &secretFiles,
		ServiceDetails: serviceDetails,
		Type:           client.BackgroundWorker,
	}

	if err := common.ApplyRuntimeSourceFieldsForCreate(plan.RuntimeSource, &createServiceBody); err != nil {
		return client.CreateServiceJSONRequestBody{}, err
	}

	return createServiceBody, nil
}
