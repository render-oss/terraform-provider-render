package internal

import (
	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/backgroundworker"
	"terraform-provider-render/internal/provider/common"
)

func UpdateServiceRequestFromModel(plan backgroundWorker.BackgroundWorkerModel, ownerID string) (client.UpdateServiceJSONRequestBody, error) {
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

	backgroundWorkerDetails := client.BackgroundWorkerDetailsPATCH{
		Plan:                       &servicePlan,
		EnvSpecificDetails:         envSpecificDetails,
		PreDeployCommand:           &preDeployCommand,
		PullRequestPreviewsEnabled: &pullRequestPreviewsEnabled,
		MaxShutdownDelaySeconds:    common.ValueAsIntPointer(plan.MaxShutdownDelaySeconds),
	}

	serviceDetails := &client.ServicePATCH_ServiceDetails{}
	if err := serviceDetails.FromBackgroundWorkerDetailsPATCH(backgroundWorkerDetails); err != nil {
		return client.UpdateServiceJSONRequestBody{}, err
	}

	var image *client.Image
	if plan.RuntimeSource.Runtime() == string(client.ServiceEnvImage) {
		imagePath := common.ImageURLForURLAndReference(
			plan.RuntimeSource.Image.ImageURL.ValueString(),
			plan.RuntimeSource.Image.Tag.ValueString(),
			plan.RuntimeSource.Image.Digest.ValueString(),
		)

		image = &client.Image{
			OwnerId:              ownerID,
			ImagePath:            imagePath,
			RegistryCredentialId: plan.RuntimeSource.Image.RegistryCredentialID.ValueStringPointer(),
		}
	}

	var updateServiceBody = client.UpdateServiceJSONRequestBody{
		Image:          image,
		Name:           plan.Name.ValueStringPointer(),
		ServiceDetails: serviceDetails,
		RootDir:        plan.RootDirectory.ValueStringPointer(),
	}

	if err := common.ApplyRuntimeSourceFieldsForUpdate(plan.RuntimeSource, &updateServiceBody, ownerID); err != nil {
		return client.UpdateServiceJSONRequestBody{}, err
	}

	return updateServiceBody, nil
}
