package common

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-render/internal/client"
)

func AutoDeployTriggerToString(trigger *client.AutoDeployTrigger) types.String {
	if trigger == nil {
		return types.StringNull()
	}

	return types.StringValue(string(*trigger))
}

func StringToAutoDeployTrigger(strTrigger types.String) *client.AutoDeployTrigger {
	if strTrigger.IsNull() || strTrigger.IsUnknown() {
		return nil
	}

	trigger := client.AutoDeployTrigger(strTrigger.ValueString())
	return &trigger
}

func BoolToAutoDeployTriggerString(autoDeploy bool) types.String {
	if autoDeploy {
		return types.StringValue(string(client.AutoDeployTriggerCommit))
	}
	return types.StringValue(string(client.AutoDeployTriggerOff))
}

func AutoDeployTriggerToBool(trigger client.AutoDeployTrigger) bool {
	if trigger == client.AutoDeployTriggerOff {
		return false
	}
	return true
}
