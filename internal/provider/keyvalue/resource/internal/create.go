package internal

import (
	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	"terraform-provider-render/internal/provider/keyvalue"
)

func CreateKeyValueRequestFromModel(ownerID string, plan keyvalue.KeyValueModel) (client.CreateKeyValueJSONRequestBody, error) {
	ipAllowList, err := common.ClientFromIPAllowList(plan.IPAllowList)
	if err != nil {
		return client.CreateKeyValueJSONRequestBody{}, err
	}

	var maxMemoryPolicy client.MaxmemoryPolicy
	if plan.MaxMemoryPolicy.ValueString() != "" {
		maxMemoryPolicy = client.MaxmemoryPolicy(plan.MaxMemoryPolicy.ValueString())
	}

	var createKeyValueBody = client.CreateKeyValueJSONRequestBody{
		EnvironmentId:   plan.EnvironmentID.ValueStringPointer(),
		IpAllowList:     &ipAllowList,
		MaxmemoryPolicy: &maxMemoryPolicy,
		Name:            plan.Name.ValueString(),
		OwnerId:         ownerID,
		Plan:            client.KeyValuePlan(plan.Plan.ValueString()),
		Region:          plan.Region.ValueStringPointer(),
	}

	return createKeyValueBody, nil
}
