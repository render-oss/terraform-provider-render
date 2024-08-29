package privateservice

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
)

type PrivateServiceModel struct {
	Id                         types.String               `tfsdk:"id"`
	Autoscaling                *common.AutoscalingModel   `tfsdk:"autoscaling"`
	RuntimeSource              *common.RuntimeSourceModel `tfsdk:"runtime_source"`
	Disk                       *common.DiskModel          `tfsdk:"disk"`
	EnvironmentID              types.String               `tfsdk:"environment_id"`
	Name                       types.String               `tfsdk:"name"`
	Slug                       types.String               `tfsdk:"slug"`
	NumInstances               types.Int64                `tfsdk:"num_instances"`
	Plan                       types.String               `tfsdk:"plan"`
	PreDeployCommand           types.String               `tfsdk:"pre_deploy_command"`
	Previews                   types.Object               `tfsdk:"previews"`
	PullRequestPreviewsEnabled types.Bool                 `tfsdk:"pull_request_previews_enabled"`
	Region                     types.String               `tfsdk:"region"`
	RootDirectory              types.String               `tfsdk:"root_directory"`
	StartCommand               types.String               `tfsdk:"start_command"`
	Url                        types.String               `tfsdk:"url"`
	MaxShutdownDelaySeconds    types.Int64                `tfsdk:"max_shutdown_delay_seconds"`

	EnvVars     map[string]common.EnvVarModel     `tfsdk:"env_vars"`
	SecretFiles map[string]common.SecretFileModel `tfsdk:"secret_files"`

	NotificationOverride types.Object `tfsdk:"notification_override"`
}

func ModelForServiceResult(service *common.WrappedService, plan PrivateServiceModel, diags diag.Diagnostics) (*PrivateServiceModel, error) {
	details, err := service.ServiceDetails.AsPrivateServiceDetails()
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

	privateServiceModel := &PrivateServiceModel{
		Id:                         types.StringValue(service.Id),
		EnvironmentID:              types.StringPointerValue(service.EnvironmentId),
		Name:                       types.StringValue(service.Name),
		Slug:                       types.StringValue(service.Slug),
		NumInstances:               numInstances,
		Plan:                       types.StringValue(string(details.Plan)),
		PreDeployCommand:           types.StringPointerValue(preDeployCommand),
		Previews:                   common.PreviewsToPreviewsObject(details.Previews),
		PullRequestPreviewsEnabled: types.BoolValue(details.PullRequestPreviewsEnabled != nil && *details.PullRequestPreviewsEnabled == client.PullRequestPreviewsEnabledYes),
		Region:                     types.StringValue(string(details.Region)),
		RootDirectory:              types.StringValue(service.RootDir),
		Url:                        types.StringValue(details.Url),
		MaxShutdownDelaySeconds:    common.IntPointerAsValue(details.MaxShutdownDelaySeconds),

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
	privateServiceModel.RuntimeSource = runtimeSource

	startCommand, err := common.StartCommandForEnvSpecificDetails(details.EnvSpecificDetails, runtimeSource.Runtime())
	if err != nil {
		return nil, err
	}
	privateServiceModel.StartCommand = types.StringPointerValue(startCommand)

	return privateServiceModel, nil
}
