package envgroup

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
)

type EnvGroupModel struct {
	Id            types.String                      `tfsdk:"id"`
	Name          types.String                      `tfsdk:"name"`
	EnvironmentID types.String                      `tfsdk:"environment_id"`
	EnvVars       map[string]common.EnvVarModel     `tfsdk:"env_vars"`
	SecretFiles   map[string]common.SecretFileModel `tfsdk:"secret_files"`
}

func ModelFromClient(envGroup *client.EnvGroup, planEnvVars map[string]common.EnvVarModel) EnvGroupModel {
	return EnvGroupModel{
		Id:            types.StringValue(envGroup.Id),
		Name:          types.StringValue(envGroup.Name),
		EnvironmentID: types.StringPointerValue(envGroup.EnvironmentId),
		EnvVars:       common.EnvVarsFromClient(&envGroup.EnvVars, planEnvVars),
		SecretFiles:   common.SecretFilesFromClient(&envGroup.SecretFiles),
	}
}

type EnvGroupLinkModel struct {
	EnvGroupId types.String `tfsdk:"env_group_id"`
	ServiceIds types.Set    `tfsdk:"service_ids"`
}

func LinkModelFromClient(envGroup *client.EnvGroup) (EnvGroupLinkModel, diag.Diagnostics) {
	// Use a map to deduplicate service IDs before creating the Set
	seen := make(map[string]bool)
	var serviceIdElements []attr.Value

	for _, link := range envGroup.ServiceLinks {
		if !seen[link.Id] {
			seen[link.Id] = true
			serviceIdElements = append(serviceIdElements, types.StringValue(link.Id))
		}
	}

	serviceIdsSet, diags := types.SetValue(types.StringType, serviceIdElements)
	if diags.HasError() {
		// failure is pretty unlikely here, but handle it gracefully just in case
		serviceIdsSet = types.SetNull(types.StringType)
	}

	return EnvGroupLinkModel{
		EnvGroupId: types.StringValue(envGroup.Id),
		ServiceIds: serviceIdsSet,
	}, diags
}
