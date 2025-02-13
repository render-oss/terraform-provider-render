package keyvalue

import (
	"terraform-provider-render/internal/client/logs"
	"terraform-provider-render/internal/provider/common"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-render/internal/client"
)

type KeyValueModel struct {
	Id                types.String `tfsdk:"id"`
	EnvironmentID     types.String `tfsdk:"environment_id"`
	IPAllowList       types.Set    `tfsdk:"ip_allow_list"`
	MaxMemoryPolicy   types.String `tfsdk:"max_memory_policy"`
	Name              types.String `tfsdk:"name"`
	Plan              types.String `tfsdk:"plan"`
	Region            types.String `tfsdk:"region"`
	ConnectionInfo    types.Object `tfsdk:"connection_info"`
	LogStreamOverride types.Object `tfsdk:"log_stream_override"`
}

var connectionInfoTypes = map[string]attr.Type{
	"external_connection_string": types.StringType,
	"internal_connection_string": types.StringType,
	"cli_command":                types.StringType,
}

func connectionInfoFromClient(c *client.KeyValueConnectionInfo, diags diag.Diagnostics) types.Object {
	if c == nil {
		return types.ObjectNull(connectionInfoTypes)
	}

	objectValue, objectDiags := types.ObjectValue(
		connectionInfoTypes,
		map[string]attr.Value{
			"external_connection_string": types.StringValue(c.ExternalConnectionString),
			"internal_connection_string": types.StringValue(c.InternalConnectionString),
			"cli_command":                types.StringValue(c.CliCommand),
		},
	)

	diags.Append(objectDiags...)

	return objectValue
}

func ModelForKeyValueResult(kv *client.KeyValue, plan *KeyValueModel, connectionInfo *client.KeyValueConnectionInfo, logStreamOverride *logs.ResourceLogStreamSetting, diags diag.Diagnostics) *KeyValueModel {
	return &KeyValueModel{
		Id:                types.StringValue(kv.Id),
		EnvironmentID:     types.StringPointerValue(kv.EnvironmentId),
		IPAllowList:       common.IPAllowListFromClient(kv.IpAllowList, diags),
		MaxMemoryPolicy:   types.StringValue(*kv.Options.MaxmemoryPolicy),
		Name:              types.StringValue(kv.Name),
		Plan:              types.StringValue(string(kv.Plan)),
		Region:            types.StringValue(string(kv.Region)),
		ConnectionInfo:    connectionInfoFromClient(connectionInfo, diags),
		LogStreamOverride: common.LogStreamOverrideFromClient(logStreamOverride, plan.LogStreamOverride, diags),
	}
}
