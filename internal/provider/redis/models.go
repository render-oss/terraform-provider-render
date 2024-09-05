package redis

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-render/internal/client/logs"
	"terraform-provider-render/internal/provider/common"

	"terraform-provider-render/internal/client"
)

type RedisModel struct {
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
	"redis_cli_command":          types.StringType,
}

func connectionInfoFromClient(c *client.RedisConnectionInfo, diags diag.Diagnostics) types.Object {
	if c == nil {
		return types.ObjectNull(connectionInfoTypes)
	}

	objectValue, objectDiags := types.ObjectValue(
		connectionInfoTypes,
		map[string]attr.Value{
			"external_connection_string": types.StringValue(c.ExternalConnectionString),
			"internal_connection_string": types.StringValue(c.InternalConnectionString),
			"redis_cli_command":          types.StringValue(c.RedisCLICommand),
		},
	)

	diags.Append(objectDiags...)

	return objectValue
}

func ModelForRedisResult(redis *client.Redis, plan *RedisModel, connectionInfo *client.RedisConnectionInfo, logStreamOverride *logs.ResourceLogStreamSetting, diags diag.Diagnostics) *RedisModel {
	return &RedisModel{
		Id:                types.StringValue(redis.Id),
		EnvironmentID:     types.StringPointerValue(redis.EnvironmentId),
		IPAllowList:       common.IPAllowListFromClient(redis.IpAllowList, diags),
		MaxMemoryPolicy:   types.StringValue(*redis.Options.MaxmemoryPolicy),
		Name:              types.StringValue(redis.Name),
		Plan:              types.StringValue(string(redis.Plan)),
		Region:            types.StringValue(string(redis.Region)),
		ConnectionInfo:    connectionInfoFromClient(connectionInfo, diags),
		LogStreamOverride: common.LogStreamOverrideFromClient(logStreamOverride, plan.LogStreamOverride, diags),
	}
}
