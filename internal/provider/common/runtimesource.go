package common

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-render/internal/client"
	commontypes "terraform-provider-render/internal/provider/common/types"
)

type NativeRuntimeModel struct {
	AutoDeploy   types.Bool        `tfsdk:"auto_deploy"`
	Branch       types.String      `tfsdk:"branch"`
	BuildCommand types.String      `tfsdk:"build_command"`
	BuildFilter  *BuildFilterModel `tfsdk:"build_filter"`
	RepoURL      types.String      `tfsdk:"repo_url"`
	Runtime      types.String      `tfsdk:"runtime"`
}

type DockerRuntimeSourceModel struct {
	AutoDeploy           types.Bool        `tfsdk:"auto_deploy"`
	BuildFilter          *BuildFilterModel `tfsdk:"build_filter"`
	Context              types.String      `tfsdk:"context"`
	DockerfilePath       types.String      `tfsdk:"dockerfile_path"`
	RegistryCredentialID types.String      `tfsdk:"registry_credential_id"`
	RepoURL              types.String      `tfsdk:"repo_url"`
	Branch               types.String      `tfsdk:"branch"`
}

type ImageRuntimeSourceModel struct {
	ImageURL             commontypes.ImageURLStringValue `tfsdk:"image_url"`
	Tag                  types.String                    `tfsdk:"tag"`
	Digest               types.String                    `tfsdk:"digest"`
	RegistryCredentialID types.String                    `tfsdk:"registry_credential_id"`
}

type RuntimeSourceModel struct {
	Docker        *DockerRuntimeSourceModel `tfsdk:"docker"`
	NativeRuntime *NativeRuntimeModel       `tfsdk:"native_runtime"`
	Image         *ImageRuntimeSourceModel  `tfsdk:"image"`
}

func (m *RuntimeSourceModel) Runtime() string {
	if m.Docker != nil {
		return "docker"
	}
	if m.Image != nil {
		return "image"
	}
	if m.NativeRuntime != nil {
		return m.NativeRuntime.Runtime.ValueString()
	}

	return ""
}

func EnvSpecificDetailsForRuntimeSource(runtime string, deployConfig *RuntimeSourceModel, startCommand *string) (*client.EnvSpecificDetailsPOST, error) {
	envSpecificDetails := &client.EnvSpecificDetailsPOST{}

	switch runtime {
	case string(client.ServiceEnvDocker):
		planDetails := deployConfig.Docker
		dockerDetails := client.DockerDetailsPOST{
			DockerCommand:        startCommand,
			DockerContext:        planDetails.Context.ValueStringPointer(),
			DockerfilePath:       planDetails.DockerfilePath.ValueStringPointer(),
			RegistryCredentialId: planDetails.RegistryCredentialID.ValueStringPointer(),
		}
		if err := envSpecificDetails.FromDockerDetailsPOST(dockerDetails); err != nil {
			return nil, err
		}

	case string(client.ServiceEnvImage):
		dockerDetails := client.DockerDetailsPOST{
			DockerCommand: startCommand,
		}
		if err := envSpecificDetails.FromDockerDetailsPOST(dockerDetails); err != nil {
			return nil, err
		}

	default:
		if startCommand == nil {
			return nil, fmt.Errorf("missing start command for native runtime")
		}
		planDetails := deployConfig.NativeRuntime
		nativeEnvDetails := client.NativeEnvironmentDetailsPOST{
			BuildCommand: planDetails.BuildCommand.ValueString(),
			StartCommand: *startCommand,
		}
		if err := envSpecificDetails.FromNativeEnvironmentDetailsPOST(nativeEnvDetails); err != nil {
			return nil, err
		}
	}

	return envSpecificDetails, nil
}

func ApplyRuntimeSourceFieldsForCreate(runtimeSource *RuntimeSourceModel, createServiceBody *client.CreateServiceJSONRequestBody) error {
	if runtimeSource.NativeRuntime != nil {
		applyNativeRuntimeSourceFieldsForCreate(runtimeSource.NativeRuntime, createServiceBody)
		return nil
	} else if runtimeSource.Docker != nil {
		applyDockerRuntimeSourceFieldsForCreate(runtimeSource.Docker, createServiceBody)
		return nil
	} else if runtimeSource.Image != nil {
		applyImageRuntimeSourceFieldsForCreate(runtimeSource.Image, createServiceBody)
		return nil
	}

	return fmt.Errorf("missing runtime source configuration")
}

func ApplyRuntimeSourceFieldsForUpdate(runtimeSource *RuntimeSourceModel, updateServiceBody *client.UpdateServiceJSONRequestBody, ownerID string) error {
	if runtimeSource.NativeRuntime != nil {
		applyNativeEnvRuntimeSourceFieldsForUpdate(runtimeSource.NativeRuntime, updateServiceBody)
		return nil
	} else if runtimeSource.Docker != nil {
		applyDockerRuntimeSourceFieldsForUpdate(runtimeSource.Docker, updateServiceBody)
		return nil
	} else if runtimeSource.Image != nil {
		applyImageRuntimeSourceFieldsForUpdate(runtimeSource.Image, updateServiceBody, ownerID)
		return nil
	}

	return fmt.Errorf("missing runtime source configuration")
}

func EnvSpecificDetailsForPATCH(runtimeSource *RuntimeSourceModel, startCommand *string) (*client.EnvSpecificDetailsPATCH, error) {
	envSpecificDetails := &client.EnvSpecificDetailsPATCH{}

	switch runtimeSource.Runtime() {
	case string(client.ServiceEnvDocker):
		planDetails := runtimeSource.Docker
		dockerDetails := client.DockerDetailsPATCH{
			DockerCommand:        EmptyStringIfNil(startCommand),
			DockerContext:        planDetails.Context.ValueStringPointer(),
			DockerfilePath:       planDetails.DockerfilePath.ValueStringPointer(),
			RegistryCredentialId: planDetails.RegistryCredentialID.ValueStringPointer(),
		}
		if err := envSpecificDetails.FromDockerDetailsPATCH(dockerDetails); err != nil {
			return nil, err
		}

	case string(client.ServiceEnvImage):
		dockerDetails := client.DockerDetailsPATCH{
			DockerCommand: EmptyStringIfNil(startCommand),
		}
		if err := envSpecificDetails.FromDockerDetailsPATCH(dockerDetails); err != nil {
			return nil, err
		}
	default:
		if runtimeSource.NativeRuntime == nil {
			return nil, fmt.Errorf("missing native runtime details for runtime source")
		}

		planDetails := runtimeSource.NativeRuntime
		if planDetails == nil {
			return nil, fmt.Errorf("missing native runtime configuration")
		}

		nativeEnvDetails := client.NativeEnvironmentDetailsPATCH{
			BuildCommand: planDetails.BuildCommand.ValueStringPointer(),
			StartCommand: startCommand,
		}
		if err := envSpecificDetails.FromNativeEnvironmentDetailsPATCH(nativeEnvDetails); err != nil {
			return nil, err
		}
	}

	return envSpecificDetails, nil
}

func ImageURLForURLAndReference(url, tag, digest string) string {
	if tag != "" {
		return fmt.Sprintf("%s:%s", url, tag)
	}

	if digest != "" {
		return fmt.Sprintf("%s@%s", url, digest)
	}

	return url
}

func applyNativeRuntimeSourceFieldsForCreate(runtime *NativeRuntimeModel, body *client.CreateServiceJSONRequestBody) {
	body.Repo = runtime.RepoURL.ValueStringPointer()
	body.Branch = runtime.Branch.ValueStringPointer()
	body.AutoDeploy = From(AutoDeployBoolToClient(runtime.AutoDeploy.ValueBool()))
	body.BuildFilter = ClientBuildFilterForModel(runtime.BuildFilter)
}

func applyDockerRuntimeSourceFieldsForCreate(runtime *DockerRuntimeSourceModel, body *client.CreateServiceJSONRequestBody) {
	body.Repo = runtime.RepoURL.ValueStringPointer()
	body.Branch = runtime.Branch.ValueStringPointer()
	body.AutoDeploy = From(AutoDeployBoolToClient(runtime.AutoDeploy.ValueBool()))
	body.BuildFilter = ClientBuildFilterForModel(runtime.BuildFilter)
}

func applyImageRuntimeSourceFieldsForCreate(runtime *ImageRuntimeSourceModel, body *client.CreateServiceJSONRequestBody) {
	imagePath := ImageURLForURLAndReference(
		runtime.ImageURL.ValueString(),
		runtime.Tag.ValueString(),
		runtime.Digest.ValueString(),
	)
	body.Image = &client.Image{
		ImagePath:            imagePath,
		RegistryCredentialId: runtime.RegistryCredentialID.ValueStringPointer(),
	}
}

func applyNativeEnvRuntimeSourceFieldsForUpdate(runtimeSource *NativeRuntimeModel, body *client.UpdateServiceJSONRequestBody) {
	body.Repo = runtimeSource.RepoURL.ValueStringPointer()
	body.Branch = runtimeSource.Branch.ValueStringPointer()
	body.AutoDeploy = From(AutoDeployBoolToClient(runtimeSource.AutoDeploy.ValueBool()))
	body.BuildFilter = ClientBuildFilterForModel(runtimeSource.BuildFilter)
}

func applyDockerRuntimeSourceFieldsForUpdate(runtimeSource *DockerRuntimeSourceModel, updateServiceBody *client.UpdateServiceJSONRequestBody) {
	updateServiceBody.AutoDeploy = From(AutoDeployBoolToClient(runtimeSource.AutoDeploy.ValueBool()))
	updateServiceBody.BuildFilter = ClientBuildFilterForModel(runtimeSource.BuildFilter)
	updateServiceBody.Branch = runtimeSource.Branch.ValueStringPointer()
	updateServiceBody.Repo = runtimeSource.RepoURL.ValueStringPointer()

}

func applyImageRuntimeSourceFieldsForUpdate(runtimeSource *ImageRuntimeSourceModel, updateServiceBody *client.UpdateServiceJSONRequestBody, ownerID string) {
	imagePath := ImageURLForURLAndReference(
		runtimeSource.ImageURL.ValueString(),
		runtimeSource.Tag.ValueString(),
		runtimeSource.Digest.ValueString(),
	)

	updateServiceBody.Image = &client.Image{
		OwnerId:              ownerID,
		ImagePath:            imagePath,
		RegistryCredentialId: runtimeSource.RegistryCredentialID.ValueStringPointer(),
	}
}
