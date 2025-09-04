package internal

import (
	"terraform-provider-render/internal/provider/common"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/cronjob"
)

func UpdateServiceRequestFromModel(plan cronJob.CronJobModel, ownerID string) (client.UpdateServiceJSONRequestBody, error) {
	envSpecificDetails, err := common.EnvSpecificDetailsForPATCH(plan.RuntimeSource, plan.StartCommand.ValueStringPointer())
	if err != nil {
		return client.UpdateServiceJSONRequestBody{}, err
	}

	servicePlan := client.PaidPlan(plan.Plan.ValueString())

	cronJobDetails := client.CronJobDetailsPATCH{
		Plan:               &servicePlan,
		EnvSpecificDetails: envSpecificDetails,
		Schedule:           plan.Schedule.ValueStringPointer(),
		Runtime:            common.From(client.ServiceRuntime(plan.RuntimeSource.Runtime())),
	}

	serviceDetails := &client.ServicePATCH_ServiceDetails{}
	if err := serviceDetails.FromCronJobDetailsPATCH(cronJobDetails); err != nil {
		return client.UpdateServiceJSONRequestBody{}, err
	}

	var image *client.Image
	if plan.RuntimeSource.Runtime() == string(client.ServiceRuntimeImage) {
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
