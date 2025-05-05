package common

import (
	"terraform-provider-render/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/types"
) 

func AutoDeployBoolToClient(autoDeploy types.Bool) *client.AutoDeploy {
	if autoDeploy.IsNull() || autoDeploy.IsUnknown() {
		return nil
	}

	if autoDeploy.ValueBool() {
		return From(client.AutoDeployYes)
	}

	return From(client.AutoDeployNo)
}
