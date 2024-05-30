package internal

import (
	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	"terraform-provider-render/internal/provider/redis"
)

func CreateRedisRequestFromModel(ownerID string, plan redis.RedisModel) (client.CreateRedisJSONRequestBody, error) {
	ipAllowList, err := common.ClientFromIPAllowList(plan.IPAllowList)
	if err != nil {
		return client.CreateRedisJSONRequestBody{}, err
	}

	var maxMemoryPolicy client.MaxmemoryPolicy
	if plan.MaxMemoryPolicy.ValueString() != "" {
		maxMemoryPolicy = client.MaxmemoryPolicy(plan.MaxMemoryPolicy.ValueString())
	}

	var createRedisBody = client.CreateRedisJSONRequestBody{
		EnvironmentId:   plan.EnvironmentID.ValueStringPointer(),
		IpAllowList:     &ipAllowList,
		MaxmemoryPolicy: &maxMemoryPolicy,
		Name:            plan.Name.ValueString(),
		OwnerId:         ownerID,
		Plan:            client.RedisPlan(plan.Plan.ValueString()),
		Region:          plan.Region.ValueStringPointer(),
	}

	return createRedisBody, nil
}
