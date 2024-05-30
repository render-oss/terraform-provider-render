package common

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-render/internal/client"
)

type EnvVarModel struct {
	Value         types.String `tfsdk:"value"`
	GenerateValue types.Bool   `tfsdk:"generate_value"`
}

func EnvVarsToClient(evs map[string]EnvVarModel) (client.EnvVarInputArray, error) {
	if len(evs) == 0 {
		return nil, nil
	}

	var res client.EnvVarInputArray
	for k, v := range evs {
		evItem, err := EnvVarToClient(k, v)
		if err != nil {
			return nil, err
		}

		res = append(res, *evItem)
	}

	return res, nil
}

func EnvVarToClient(k string, v EnvVarModel) (*client.EnvVarInput, error) {
	evItem := &client.EnvVarInput{}

	if !v.Value.IsNull() && !v.Value.IsUnknown() {
		err := evItem.FromEnvVarKeyValue(client.EnvVarKeyValue{
			Key:   k,
			Value: v.Value.ValueString(),
		})
		if err != nil {
			return nil, err
		}
	} else if v.GenerateValue.ValueBool() {
		err := evItem.FromEnvVarKeyGenerateValue(client.EnvVarKeyGenerateValue{
			Key:           k,
			GenerateValue: true,
		})
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("env var %s has no value, either provide a value or set generate_value to true", k)
	}
	return evItem, nil
}

func EnvVarsFromClientCursors(evs *[]client.EnvVarWithCursor, planEVs map[string]EnvVarModel) map[string]EnvVarModel {
	res := map[string]EnvVarModel{}

	if evs == nil || len(*evs) == 0 {
		return nil
	}

	for _, ev := range *evs {
		res[ev.EnvVar.Key] = evFromClient(ev.EnvVar, planEVs)
	}

	return res
}

func EnvVarsFromClient(evs *[]client.EnvVar, planEVs map[string]EnvVarModel) map[string]EnvVarModel {
	res := map[string]EnvVarModel{}

	if evs == nil || len(*evs) == 0 {
		return nil
	}

	for _, ev := range *evs {
		model := evFromClient(ev, planEVs)
		res[ev.Key] = model
	}

	return res
}

func evFromClient(ev client.EnvVar, planEVs map[string]EnvVarModel) EnvVarModel {
	if planEVs == nil {
		planEVs = map[string]EnvVarModel{}
	}

	generateValue := types.BoolValue(false)
	if ev, ok := planEVs[ev.Key]; ok {
		generateValue = ev.GenerateValue
	}

	model := EnvVarModel{
		Value:         types.StringValue(ev.Value),
		GenerateValue: generateValue,
	}
	return model
}
