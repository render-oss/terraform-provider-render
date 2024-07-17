package internal

import (
	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	"terraform-provider-render/internal/provider/privateservice"
)

func UpdateServiceRequestFromModel(plan privateservice.PrivateServiceModel, ownerID string) (client.UpdateServiceJSONRequestBody, error) {
	envSpecificDetails, err := common.EnvSpecificDetailsForPATCH(plan.RuntimeSource, plan.StartCommand.ValueStringPointer())
	if err != nil {
		return client.UpdateServiceJSONRequestBody{}, err
	}

	servicePlan := client.PaidPlan(plan.Plan.ValueString())

	pullRequestPreviewsEnabled := client.PrivateServiceDetailsPATCHPullRequestPreviewsEnabledNo
	if plan.PullRequestPreviewsEnabled.ValueBool() {
		pullRequestPreviewsEnabled = client.PrivateServiceDetailsPATCHPullRequestPreviewsEnabledYes
	}

	preDeployCommand := ""
	if plan.PreDeployCommand.ValueStringPointer() != nil && *plan.PreDeployCommand.ValueStringPointer() != "" {
		preDeployCommand = *plan.PreDeployCommand.ValueStringPointer()
	}

	privateServiceDetails := client.PrivateServiceDetailsPATCH{
		Plan:                       &servicePlan,
		EnvSpecificDetails:         envSpecificDetails,
		PreDeployCommand:           &preDeployCommand,
		PullRequestPreviewsEnabled: &pullRequestPreviewsEnabled,
		MaxShutdownDelaySeconds:    common.IntPointerToRequest(plan.MaxShutdownDelaySeconds),
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
