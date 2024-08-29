package common

import (
	"terraform-provider-render/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AutoscalingModel struct {
	Criteria *AutoscalingCriteriaModel `tfsdk:"criteria"`
	Enabled  types.Bool                `tfsdk:"enabled"`
	Max      types.Int64               `tfsdk:"max"`
	Min      types.Int64               `tfsdk:"min"`
}

type AutoscalingCriteriaModel struct {
	Cpu    types.Object `tfsdk:"cpu"`
	Memory types.Object `tfsdk:"memory"`
}

type BuildFilterModel struct {
	IgnoredPaths []types.String `tfsdk:"ignored_paths"`
	Paths        []types.String `tfsdk:"paths"`
}

type DiskModel struct {
	SizeGB    types.Int64  `tfsdk:"size_gb"`
	MountPath types.String `tfsdk:"mount_path"`
	Name      types.String `tfsdk:"name"`
	ID        types.String `tfsdk:"id"`
}

type PreviewsModel struct {
	Generation types.String `tfsdk:"generation"`
}

func StartCommandForEnvSpecificDetails(serviceDetails client.EnvSpecificDetails, runtime string) (*string, error) {
	switch runtime {
	case string(client.ServiceEnvDocker), string(client.ServiceEnvImage):
		dockerDetails, err := serviceDetails.AsDockerDetails()
		if err != nil {
			return nil, err
		}

		if dockerDetails.DockerCommand == "" {
			return nil, nil
		}

		return &dockerDetails.DockerCommand, nil
	default:
		nativeEnvDetails, err := serviceDetails.AsNativeEnvironmentDetails()
		if err != nil {
			return nil, err
		}

		if nativeEnvDetails.StartCommand == "" {
			return nil, nil
		}

		return &nativeEnvDetails.StartCommand, nil
	}
}

func PreDeployCommandForEnvSpecificDetails(serviceDetails client.EnvSpecificDetails) (*string, error) {
	nativeEnvDetails, err := serviceDetails.AsNativeEnvironmentDetails()
	if err != nil {
		return nil, err
	}

	if nativeEnvDetails.PreDeployCommand != nil && *nativeEnvDetails.PreDeployCommand == "" {
		return nil, nil
	}

	return nativeEnvDetails.PreDeployCommand, nil
}
