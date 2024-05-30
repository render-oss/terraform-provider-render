package webservice

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
)

type WebServiceModel struct {
	Id                         types.String               `tfsdk:"id"`
	Autoscaling                *common.AutoscalingModel   `tfsdk:"autoscaling"`
	CustomDomains              []common.CustomDomainModel `tfsdk:"custom_domains"`
	RuntimeSource              *common.RuntimeSourceModel `tfsdk:"runtime_source"`
	Disk                       *common.DiskModel          `tfsdk:"disk"`
	EnvironmentID              types.String               `tfsdk:"environment_id"`
	HealthCheckPath            types.String               `tfsdk:"health_check_path"`
	Name                       types.String               `tfsdk:"name"`
	Slug                       types.String               `tfsdk:"slug"`
	NumInstances               types.Int64                `tfsdk:"num_instances"`
	Plan                       types.String               `tfsdk:"plan"`
	PreDeployCommand           types.String               `tfsdk:"pre_deploy_command"`
	PullRequestPreviewsEnabled types.Bool                 `tfsdk:"pull_request_previews_enabled"`
	Region                     types.String               `tfsdk:"region"`
	RootDirectory              types.String               `tfsdk:"root_directory"`
	StartCommand               types.String               `tfsdk:"start_command"`
	Url                        types.String               `tfsdk:"url"`

	EnvVars     map[string]common.EnvVarModel     `tfsdk:"env_vars"`
	SecretFiles map[string]common.SecretFileModel `tfsdk:"secret_files"`

	NotificationOverride types.Object `tfsdk:"notification_override"`
}

func ModelForServiceResult(service *common.WrappedService, plan WebServiceModel, diags diag.Diagnostics) (*WebServiceModel, error) {
	details, err := service.ServiceDetails.AsWebServiceDetails()
	if err != nil {
		return nil, err
	}

	numInstances := types.Int64Value(int64(details.NumInstances))
	if plan.NumInstances.IsUnknown() || plan.NumInstances.IsNull() {
		numInstances = types.Int64Null()
	}

	preDeployCommand, err := common.PreDeployCommandForEnvSpecificDetails(details.EnvSpecificDetails)
	if err != nil {
		return nil, err
	}

	webServicesModel := &WebServiceModel{
		Id:                         types.StringValue(service.Id),
		CustomDomains:              common.CustomDomainClientsToCustomDomainModels(service.CustomDomains),
		EnvironmentID:              types.StringPointerValue(service.EnvironmentId),
		HealthCheckPath:            types.StringValue(details.HealthCheckPath),
		Name:                       types.StringValue(service.Name),
		Slug:                       types.StringValue(service.Slug),
		NumInstances:               numInstances,
		Plan:                       types.StringValue(string(details.Plan)),
		PreDeployCommand:           types.StringPointerValue(preDeployCommand),
		PullRequestPreviewsEnabled: types.BoolValue(details.PullRequestPreviewsEnabled == client.PullRequestPreviewsEnabledYes),
		Region:                     types.StringValue(string(details.Region)),
		RootDirectory:              types.StringValue(service.RootDir),
		Url:                        types.StringValue(details.Url),

		Autoscaling:          common.AutoscalingFromClient(details.Autoscaling, diags),
		Disk:                 common.DiskToDiskModel(details.Disk),
		EnvVars:              common.EnvVarsFromClientCursors(service.EnvVars, plan.EnvVars),
		SecretFiles:          common.SecretFilesFromClientCursors(service.SecretFiles),
		NotificationOverride: common.NotificationOverrideFromClient(service.NotificationOverride, diags),
	}

	runtimeSource, err := common.RuntimeSourceFromClient(service.Service, details.Env, details.EnvSpecificDetails)
	if err != nil {
		return nil, err
	}

	webServicesModel.RuntimeSource = runtimeSource

	startCommand, err := common.StartCommandForEnvSpecificDetails(details.EnvSpecificDetails, runtimeSource.Runtime())
	if err != nil {
		return nil, err
	}
	webServicesModel.StartCommand = types.StringPointerValue(startCommand)

	return webServicesModel, nil
}
