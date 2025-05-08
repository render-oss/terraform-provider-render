package common

import "terraform-provider-render/internal/client"

func AutoDeployBoolToClient(autoDeploy bool) client.AutoDeploy {
	if autoDeploy {
		return client.AutoDeployYes
	}

	return client.AutoDeployNo
}
