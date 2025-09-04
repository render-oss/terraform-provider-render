package internal

import (
	"context"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	"terraform-provider-render/internal/provider/privateservice"
)

func UpdateServiceRequestFromModel(ctx context.Context, plan privateservice.PrivateServiceModel, ownerID string) (client.UpdateServiceJSONRequestBody, error) {
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

	privateServiceDetails := client.PrivateServiceDetailsPATCH{
		Plan:                       &servicePlan,
		EnvSpecificDetails:         envSpecificDetails,
		PreDeployCommand:           &preDeployCommand,
		Previews:                   common.PreviewsObjectToPreviews(ctx, plan.Previews),
		PullRequestPreviewsEnabled: &pullRequestPreviewsEnabled,
		MaxShutdownDelaySeconds:    common.ValueAsIntPointer(plan.MaxShutdownDelaySeconds),
		Runtime:                    common.From(client.ServiceRuntime(plan.RuntimeSource.Runtime())),
	}

	serviceDetails := &client.ServicePATCH_ServiceDetails{}
	if err := serviceDetails.FromPrivateServiceDetailsPATCH(privateServiceDetails); err != nil {
		return client.UpdateServiceJSONRequestBody{}, err
	}

	var image *client.Image
	if plan.RuntimeSource.Runtime() == string(client.ServiceEnvImage) {
		image = &client.Image{
			OwnerId:              ownerID,
			ImagePath:            plan.RuntimeSource.Image.ImageURL.ValueString(),
			RegistryCredentialId: plan.RuntimeSource.Image.RegistryCredentialID.ValueStringPointer(),
		}
	}

	var updateServiceBody = client.UpdateServiceJSONRequestBody{
		Image:          image,
		Name:           plan.Name.ValueStringPointer(),
		RootDir:        plan.RootDirectory.ValueStringPointer(),
		ServiceDetails: serviceDetails,
	}

	if err := common.ApplyRuntimeSourceFieldsForUpdate(plan.RuntimeSource, &updateServiceBody, ownerID); err != nil {
		return client.UpdateServiceJSONRequestBody{}, err
	}

	return updateServiceBody, nil
}
