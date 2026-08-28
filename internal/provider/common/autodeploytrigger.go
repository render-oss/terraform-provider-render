package common

import (
	"terraform-provider-render/internal/client/autodeploy"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func AutoDeployTriggerToString(trigger *autodeploy.AutoDeployTrigger) types.String {
	if trigger == nil {
		return types.StringNull()
	}

	return types.StringValue(string(*trigger))
}

func StringToAutoDeployTrigger(strTrigger types.String) *autodeploy.AutoDeployTrigger {
	if strTrigger.IsNull() || strTrigger.IsUnknown() {
		return nil
	}

	trigger := autodeploy.AutoDeployTrigger(strTrigger.ValueString())
	return &trigger
}

func BoolToAutoDeployTriggerString(autoDeploy bool) types.String {
	if autoDeploy {
		return types.StringValue(string(autodeploy.Commit))
	}
	return types.StringValue(string(autodeploy.Off))
}

func AutoDeployTriggerToBool(trigger autodeploy.AutoDeployTrigger) bool {
	return trigger != autodeploy.Off
}
