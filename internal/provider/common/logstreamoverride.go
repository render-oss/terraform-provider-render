package common

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-render/internal/client/logs"
)

var logStreamTypes = map[string]attr.Type{
	"send":     types.BoolType,
	"token":    types.StringType,
	"endpoint": types.StringType,
}

func LogStreamOverrideFromClient(client *logs.ResourceLogStreamSetting, diags diag.Diagnostics) types.Object {
	if client == nil {
		return types.ObjectNull(notificationTypes)
	}

	setting := logs.LogStreamSetting("send")
	if client.Setting != nil {
		setting = *client.Setting
	}

	endpoint := ""
	if client.Endpoint != nil {
		endpoint = *client.Endpoint
	}

	objectValue, objectDiags := types.ObjectValue(
		notificationTypes,
		map[string]attr.Value{
			"send":     types.BoolValue(setting == "send"),
			"endpoint": types.StringValue(endpoint),
		},
	)

	diags.Append(objectDiags...)

	return objectValue
}

func LogStreamOverrideToClient(model types.Object) (*logs.LogStreamResourceUpdate, error) {
	if model.IsNull() {
		return nil, nil
	}

	attrs := model.Attributes()

	var logStreamSetting logs.LogStreamSetting
	if attrs["send"] != nil && !attrs["send"].IsNull() && !attrs["send"].IsUnknown() {
		send, ok := attrs["send"].(types.Bool)
		if !ok {
			// This should never happen
			return nil, fmt.Errorf("unexpected type for send: %T", attrs["send"])
		}
		if send.ValueBool() {
			logStreamSetting = logs.LogStreamSettingSend
		} else {
			logStreamSetting = logs.LogStreamSettingDrop
		}
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
		LogStreamSetting:  logStreamSetting,
		LogStreamEndpoint: endpoint,
		LogStreamToken:    token,
	}, nil
}
