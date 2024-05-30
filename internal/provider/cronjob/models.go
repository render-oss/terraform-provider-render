package cronJob

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-render/internal/provider/common"
)

type CronJobModel struct {
	Id            types.String                      `tfsdk:"id"`
	RuntimeSource *common.RuntimeSourceModel        `tfsdk:"runtime_source"`
	EnvironmentID types.String                      `tfsdk:"environment_id"`
	Name          types.String                      `tfsdk:"name"`
	Slug          types.String                      `tfsdk:"slug"`
	Plan          types.String                      `tfsdk:"plan"`
	Region        types.String                      `tfsdk:"region"`
	RootDirectory types.String                      `tfsdk:"root_directory"`
	Schedule      types.String                      `tfsdk:"schedule"`
	StartCommand  types.String                      `tfsdk:"start_command"`
	EnvVars       map[string]common.EnvVarModel     `tfsdk:"env_vars"`
	SecretFiles   map[string]common.SecretFileModel `tfsdk:"secret_files"`

	NotificationOverride types.Object `tfsdk:"notification_override"`
}

func ModelForServiceResult(service *common.WrappedService, planEVs map[string]common.EnvVarModel, diags diag.Diagnostics) (*CronJobModel, error) {
	details, err := service.ServiceDetails.AsCronJobDetails()
	if err != nil {
		return nil, err
	}

	cronJobModel := &CronJobModel{
		Id: types.StringValue(service.Id),

		EnvironmentID:        types.StringPointerValue(service.EnvironmentId),
		Name:                 types.StringValue(service.Name),
		Slug:                 types.StringValue(service.Slug),
		Plan:                 types.StringValue(string(details.Plan)),
		Region:               types.StringValue(string(details.Region)),
		Schedule:             types.StringValue(details.Schedule),
		EnvVars:              common.EnvVarsFromClientCursors(service.EnvVars, planEVs),
		SecretFiles:          common.SecretFilesFromClientCursors(service.SecretFiles),
		NotificationOverride: common.NotificationOverrideFromClient(service.NotificationOverride, diags),
	}

	runtimeSource, err := common.RuntimeSourceFromClient(service.Service, details.Env, details.EnvSpecificDetails)
	if err != nil {
		return nil, err
	}

	cronJobModel.RuntimeSource = runtimeSource

	startCommand, err := common.StartCommandForEnvSpecificDetails(details.EnvSpecificDetails, runtimeSource.Runtime())
	if err != nil {
		return nil, err
	}
	cronJobModel.StartCommand = types.StringPointerValue(startCommand)

	return cronJobModel, nil
}
