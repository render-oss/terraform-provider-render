package common

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	commontypes "terraform-provider-render/internal/provider/common/types"

	"terraform-provider-render/internal/client"
)

const (
	CriteriaEnabled    = "enabled"
	CriteriaPercentage = "percentage"
)

var criteriaTypes = map[string]attr.Type{
	CriteriaEnabled:    types.BoolType,
	CriteriaPercentage: types.Int64Type,
}

func AutoscalingFromClient(autoscaling *client.AutoscalingConfig, diags diag.Diagnostics) *AutoscalingModel {
	if autoscaling == nil {
		return nil
	}

	cpuCriteria, cpuDiags := types.ObjectValue(criteriaTypes, map[string]attr.Value{
		CriteriaEnabled:    types.BoolValue(autoscaling.Criteria.Cpu.Enabled),
		CriteriaPercentage: types.Int64Value(int64(autoscaling.Criteria.Cpu.Percentage)),
	})
	diags = append(diags, cpuDiags...)
	if diags.HasError() {
		return nil
	}

	memoryCriteria, memoryDiags := types.ObjectValue(criteriaTypes, map[string]attr.Value{
		CriteriaEnabled:    types.BoolValue(autoscaling.Criteria.Memory.Enabled),
		CriteriaPercentage: types.Int64Value(int64(autoscaling.Criteria.Memory.Percentage)),
	})
	diags = append(diags, memoryDiags...)
	if diags.HasError() {
		return nil
	}

	return &AutoscalingModel{
		Criteria: &AutoscalingCriteriaModel{
			Cpu:    cpuCriteria,
			Memory: memoryCriteria,
		},
		Enabled: types.BoolValue(autoscaling.Enabled),
		Max:     types.Int64Value(int64(autoscaling.Max)),
		Min:     types.Int64Value(int64(autoscaling.Min)),
	}
}

func AutoscalingRequest(autoscaling *AutoscalingModel) (*client.AutoscalingConfig, error) {
	if autoscaling == nil {
		return nil, nil
	}

	var cpu client.AutoscalingCriteriaPercentage
	if !autoscaling.Criteria.Cpu.IsNull() && !autoscaling.Criteria.Cpu.IsUnknown() {
		cpuAttributes := autoscaling.Criteria.Cpu.Attributes()
		cpuEnabled, ok := cpuAttributes[CriteriaEnabled].(types.Bool)
		if !ok {
			return nil, fmt.Errorf("expected cpu %s to be a bool", CriteriaEnabled)
		}

		cpuPercentage, ok := cpuAttributes[CriteriaPercentage].(types.Int64)
		if !ok {
			return nil, fmt.Errorf("expected cpu %s to be an int64", CriteriaPercentage)
		}

		cpu = client.AutoscalingCriteriaPercentage{
			Enabled:    cpuEnabled.ValueBool(),
			Percentage: int(cpuPercentage.ValueInt64()),
		}
	}

	var memory client.AutoscalingCriteriaPercentage
	if !autoscaling.Criteria.Memory.IsNull() && !autoscaling.Criteria.Memory.IsUnknown() {
		memoryAttributes := autoscaling.Criteria.Memory.Attributes()
		memoryEnabled, ok := memoryAttributes[CriteriaEnabled].(types.Bool)
		if !ok {
			return nil, fmt.Errorf("expected memory %s to be a bool", CriteriaEnabled)
		}

		memoryPercentage, ok := memoryAttributes[CriteriaPercentage].(types.Int64)
		if !ok {
			return nil, fmt.Errorf("expected memory %s to be an int64", CriteriaPercentage)
		}
		memory = client.AutoscalingCriteriaPercentage{
			Enabled:    memoryEnabled.ValueBool(),
			Percentage: int(memoryPercentage.ValueInt64()),
		}
	}

	return &client.AutoscalingConfig{
		Criteria: client.AutoscalingCriteria{
			Cpu:    cpu,
			Memory: memory,
		},
		Enabled: autoscaling.Enabled.ValueBool(),
		Max:     int(autoscaling.Max.ValueInt64()),
		Min:     int(autoscaling.Min.ValueInt64()),
	}, nil
}

func RuntimeSourceFromClient(service *client.Service, env client.ServiceEnv, envDetails client.EnvSpecificDetails) (*RuntimeSourceModel, error) {
	runtimeSource := &RuntimeSourceModel{}
	if env == client.ServiceEnvImage {
		imageRuntime, err := ImageRuntimeSource(service, envDetails)
		if err != nil {
			return nil, err
		}

		runtimeSource.Image = imageRuntime

	} else if env == client.ServiceEnvDocker {
		dockerRuntime, err := DockerRuntimeSource(service, envDetails)
		if err != nil {
			return nil, err
		}

		runtimeSource.Docker = dockerRuntime
	} else {
		nativeRuntime, err := NativeRuntimeSource(service, env, envDetails)
		if err != nil {
			return nil, err
		}

		runtimeSource.NativeRuntime = nativeRuntime
	}

	return runtimeSource, nil
}

func NativeRuntimeSource(service *client.Service, env client.ServiceEnv, envDetails client.EnvSpecificDetails) (*NativeRuntimeModel, error) {
	nativeRuntime := &NativeRuntimeModel{}

	if service.Repo != nil {
		nativeRuntime.RepoURL = types.StringValue(*service.Repo)
	}

	nativeRuntime.AutoDeploy = types.BoolValue(service.AutoDeploy == client.AutoDeployYes)
	nativeRuntime.Branch = types.StringPointerValue(service.Branch)
	nativeRuntime.BuildFilter = BuildFilterModelForClient(service.BuildFilter)
	nativeRuntime.Runtime = types.StringValue(string(env))
	nativeEnvDetails, err := envDetails.AsNativeEnvironmentDetails()
	if err != nil {
		return nil, err
	}

	nativeRuntime.BuildCommand = types.StringValue(nativeEnvDetails.BuildCommand)

	return nativeRuntime, nil
}

func DockerRuntimeSource(service *client.Service, envDetails client.EnvSpecificDetails) (*DockerRuntimeSourceModel, error) {
	dockerDetails, err := envDetails.AsDockerDetails()
	if err != nil {
		return nil, err
	}

	docker := &DockerRuntimeSourceModel{
		AutoDeploy:     types.BoolValue(service.AutoDeploy == client.AutoDeployYes),
		Context:        types.StringValue(dockerDetails.DockerContext),
		DockerfilePath: types.StringValue(dockerDetails.DockerfilePath),
		RepoURL:        types.StringPointerValue(service.Repo),
		Branch:         types.StringPointerValue(service.Branch),
		BuildFilter:    BuildFilterModelForClient(service.BuildFilter),
	}

	if dockerDetails.RegistryCredential != nil {
		docker.RegistryCredentialID = types.StringValue(dockerDetails.RegistryCredential.Id)
	}

	return docker, nil

}

func ImageRuntimeSource(service *client.Service, envDetails client.EnvSpecificDetails) (*ImageRuntimeSourceModel, error) {
	image := &ImageRuntimeSourceModel{
		ImageURL: commontypes.ImageURLStringValue{StringValue: types.StringPointerValue(service.ImagePath)},
	}

	if service.RegistryCredential != nil {
		image.RegistryCredentialID = types.StringValue(service.RegistryCredential.Id)
	}

	return image, nil
}
