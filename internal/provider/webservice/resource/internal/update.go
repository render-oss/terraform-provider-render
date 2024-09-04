package internal

import (
	"context"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	"terraform-provider-render/internal/provider/webservice"
)

func UpdateServiceRequestFromModel(ctx context.Context, plan webservice.WebServiceModel, ownerID string) (client.UpdateServiceJSONRequestBody, error) {
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

	webServiceDetails := client.WebServiceDetailsPATCH{
		Plan:                       &servicePlan,
		EnvSpecificDetails:         envSpecificDetails,
		HealthCheckPath:            plan.HealthCheckPath.ValueStringPointer(),
		PreDeployCommand:           &preDeployCommand,
		Previews:                   common.PreviewsObjectToPreviews(ctx, plan.Previews),
		PullRequestPreviewsEnabled: &pullRequestPreviewsEnabled,
		MaxShutdownDelaySeconds:    common.ValueAsIntPointer(plan.MaxShutdownDelaySeconds),
		Runtime:                    common.From(client.ServiceRuntime(plan.RuntimeSource.Runtime())),
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
