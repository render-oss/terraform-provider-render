package redis

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-render/internal/provider/common"

	"terraform-provider-render/internal/client"
)

type RedisModel struct {
	Id              types.String `tfsdk:"id"`
	EnvironmentID   types.String `tfsdk:"environment_id"`
	IPAllowList     types.Set    `tfsdk:"ip_allow_list"`
	MaxMemoryPolicy types.String `tfsdk:"max_memory_policy"`
	Name            types.String `tfsdk:"name"`
	Plan            types.String `tfsdk:"plan"`
	Region          types.String `tfsdk:"region"`
}

func ModelForRedisResult(redis *client.Redis, diags diag.Diagnostics) *RedisModel {
	return &RedisModel{
		Id:              types.StringValue(redis.Id),
		EnvironmentID:   types.StringPointerValue(redis.EnvironmentId),
		IPAllowList:     common.IPAllowListFromClient(redis.IpAllowList, diags),
		MaxMemoryPolicy: types.StringValue(*redis.Options.MaxmemoryPolicy),
		Name:            types.StringValue(redis.Name),
		Plan:            types.StringValue(string(redis.Plan)),
		Region:          types.StringValue(string(redis.Region)),
	}
}
