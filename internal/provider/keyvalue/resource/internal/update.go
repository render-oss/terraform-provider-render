package internal

import (
	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	"terraform-provider-render/internal/provider/keyvalue"
)

func UpdateServiceRequestFromModel(plan keyvalue.KeyValueModel) (client.UpdateKeyvalueJSONRequestBody, error) {
	ipAllowList, err := common.ClientFromIPAllowList(plan.IPAllowList)
	if err != nil {
		return client.UpdateKeyvalueJSONRequestBody{}, err
	}

	var maxMemoryPolicy client.MaxmemoryPolicy
	if plan.MaxMemoryPolicy.ValueString() != "" {
		maxMemoryPolicy = client.MaxmemoryPolicy(plan.MaxMemoryPolicy.ValueString())
	}

	var keyValuePlan client.KeyValuePlan
	if plan.Plan.ValueString() != "" {
		keyValuePlan = client.KeyValuePlan(plan.Plan.ValueString())
	}

	updateKeyValueRequest := client.UpdateKeyvalueJSONRequestBody{
		IpAllowList:     &ipAllowList,
		MaxmemoryPolicy: &maxMemoryPolicy,
		Name:            plan.Name.ValueStringPointer(),
		Plan:            &keyValuePlan,
	}

	return updateKeyValueRequest, nil
}
