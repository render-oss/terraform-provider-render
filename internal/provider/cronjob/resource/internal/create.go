package internal

import (
	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	"terraform-provider-render/internal/provider/cronjob"
)

func CreateServiceRequestFromModel(ownerID string, plan cronJob.CronJobModel) (client.CreateServiceJSONRequestBody, error) {
	envSpecificDetails, err := buildEnvSpecificDetails(plan)
	if err != nil {
		return client.CreateServiceJSONRequestBody{}, err
	}

	servicePlan := client.PaidPlan(plan.Plan.ValueString())
	region := client.Region(plan.Region.ValueString())

	cronJobDetails := client.CronJobDetailsPOST{
		Env:                client.ServiceEnv(plan.RuntimeSource.Runtime()),
		EnvSpecificDetails: envSpecificDetails,
		Plan:               &servicePlan,
		Region:             &region,
		Schedule:           plan.Schedule.ValueString(),
	}

	serviceDetails := &client.ServicePOST_ServiceDetails{}
	if err := serviceDetails.FromCronJobDetailsPOST(cronJobDetails); err != nil {
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
		SecretFiles:    &secretFiles,
		ServiceDetails: serviceDetails,
		Type:           client.CronJob,
	}

	if err := common.ApplyRuntimeSourceFieldsForCreate(plan.RuntimeSource, &createServiceBody); err != nil {
		return client.CreateServiceJSONRequestBody{}, err
	}

	return createServiceBody, nil
}

func buildEnvSpecificDetails(plan cronJob.CronJobModel) (*client.EnvSpecificDetails, error) {
	envSpecificDetails := &client.EnvSpecificDetails{}

	switch plan.RuntimeSource.Runtime() {
	case string(client.ServiceEnvDocker):
		planDetails := plan.RuntimeSource.Docker
		dockerDetails := client.DockerDetails{
			DockerCommand:      plan.StartCommand.ValueString(),
			DockerContext:      planDetails.Context.ValueString(),
			DockerfilePath:     planDetails.DockerfilePath.ValueString(),
			RegistryCredential: &client.RegistryCredential{Id: planDetails.RegistryCredentialID.ValueString()},
		}
		if err := envSpecificDetails.FromDockerDetails(dockerDetails); err != nil {
			return nil, err
		}

	case string(client.ServiceEnvImage):
		dockerDetails := client.DockerDetails{
			DockerCommand: plan.StartCommand.ValueString(),
		}
		if plan.RuntimeSource.Image != nil {
			dockerDetails.RegistryCredential = &client.RegistryCredential{Id: plan.RuntimeSource.Image.RegistryCredentialID.ValueString()}
		}
		if err := envSpecificDetails.FromDockerDetails(dockerDetails); err != nil {
			return nil, err
		}

	default:
		planDetails := plan.RuntimeSource.NativeRuntime
		nativeEnvDetails := client.NativeEnvironmentDetails{
			BuildCommand: planDetails.BuildCommand.ValueString(),
			StartCommand: plan.StartCommand.ValueString(),
		}
		if err := envSpecificDetails.FromNativeEnvironmentDetails(nativeEnvDetails); err != nil {
			return nil, err
		}
	}

	return envSpecificDetails, nil
}
