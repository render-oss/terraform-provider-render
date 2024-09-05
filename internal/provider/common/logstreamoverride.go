package common

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-render/internal/client/logs"
)

var logStreamTypes = map[string]attr.Type{
	"setting":  types.StringType,
	"token":    types.StringType,
	"endpoint": types.StringType,
}

func LogStreamOverrideFromClient(client *logs.ResourceLogStreamSetting, plan types.Object, diags diag.Diagnostics) types.Object {
	if client == nil {
		return types.ObjectNull(logStreamTypes)
	}

	setting := logs.LogStreamSetting("send")
	if client.Setting != nil {
		setting = *client.Setting
	}

	endpoint := ""
	if client.Endpoint != nil {
		endpoint = *client.Endpoint
	}

	planAttrs := plan.Attributes()
	token := types.StringNull()
	if tkn, present := planAttrs["token"]; present {
		token = tkn.(types.String)
	}

	objectValue, objectDiags := types.ObjectValue(
		logStreamTypes,
		map[string]attr.Value{
			"setting":  types.StringValue(string(setting)),
			"endpoint": types.StringValue(endpoint),
			"token":    token,
		},
	)

	diags.Append(objectDiags...)

	return objectValue
}

func LogStreamOverrideToClient(model types.Object) (*logs.LogStreamResourceUpdate, error) {
	if model.IsNull() || model.IsUnknown() {
		return nil, nil
	}

	attrs := model.Attributes()

	var logStreamSetting logs.LogStreamSetting
	if attrs["setting"] != nil && !attrs["setting"].IsNull() && !attrs["setting"].IsUnknown() {
		str, ok := attrs["setting"].(types.String)
		if !ok {
			// This should never happen
			return nil, fmt.Errorf("unexpected type for setting: %T", attrs["setting"])
		}
		logStreamSetting = logs.LogStreamSetting(str.ValueString())
	}
	var endpoint *logs.LogStreamEndpoint
	if attrs["endpoint"] != nil && !attrs["endpoint"].IsNull() && !attrs["endpoint"].IsUnknown() {
		str, ok := attrs["endpoint"].(types.String)
		if !ok {
			// This should never happen
			return nil, fmt.Errorf("unexpected type for endpoint: %T", attrs["endpoint"])
		}
		endpoint = From(str.ValueString())
	}
	var token *logs.LogStreamEndpoint
	if attrs["token"] != nil && !attrs["token"].IsNull() && !attrs["token"].IsUnknown() {
		str, ok := attrs["token"].(types.String)
		if !ok {
			// This should never happen
			return nil, fmt.Errorf("unexpected type for token: %T", attrs["token"])
		}
		token = From(str.ValueString())
	}

	return &logs.LogStreamResourceUpdate{
		Setting:  &logStreamSetting,
		Endpoint: endpoint,
		Token:    token,
	}, nil
}
