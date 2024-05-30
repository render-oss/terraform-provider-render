package internal

import (
	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	"terraform-provider-render/internal/provider/redis"
)

func UpdateServiceRequestFromModel(plan redis.RedisModel) (client.UpdateRedisJSONRequestBody, error) {
	ipAllowList, err := common.ClientFromIPAllowList(plan.IPAllowList)
	if err != nil {
		return client.UpdateRedisJSONRequestBody{}, err
	}

	var maxMemoryPolicy client.MaxmemoryPolicy
	if plan.MaxMemoryPolicy.ValueString() != "" {
		maxMemoryPolicy = client.MaxmemoryPolicy(plan.MaxMemoryPolicy.ValueString())
	}

	var redisPlan client.RedisPlan
	if plan.Plan.ValueString() != "" {
		redisPlan = client.RedisPlan(plan.Plan.ValueString())
	}

	updateRedisRequest := client.UpdateRedisJSONRequestBody{
		IpAllowList:     &ipAllowList,
		MaxmemoryPolicy: &maxMemoryPolicy,
		Name:            plan.Name.ValueStringPointer(),
		Plan:            &redisPlan,
	}

	return updateRedisRequest, nil
}
