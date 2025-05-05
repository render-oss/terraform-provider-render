package internal

import (
	"context"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	"terraform-provider-render/internal/provider/staticsite"
)

func UpdateServiceRequestFromModel(ctx context.Context, plan staticsite.StaticSiteModel) (client.UpdateServiceJSONRequestBody, error) {
	prPreviewEnabled := client.PullRequestPreviewsEnabledNo
	if plan.PullRequestPreviewsEnabled.ValueBool() {
		prPreviewEnabled = client.PullRequestPreviewsEnabledYes
	}

	var staticSiteDetails = client.StaticSiteDetailsPATCH{
		BuildCommand:               plan.BuildCommand.ValueStringPointer(),
		PublishPath:                plan.PublishPath.ValueStringPointer(),
		Previews:                   common.PreviewsObjectToPreviews(ctx, plan.Previews),
		PullRequestPreviewsEnabled: &prPreviewEnabled,
	}

	serviceDetails := &client.ServicePATCH_ServiceDetails{}
	if err := serviceDetails.FromStaticSiteDetailsPATCH(staticSiteDetails); err != nil {
		return client.UpdateServiceJSONRequestBody{}, err
	}

	var updateServiceBody = client.UpdateServiceJSONRequestBody{
		Name:           plan.Name.ValueStringPointer(),
		ServiceDetails: serviceDetails,
	}

	updateServiceGitRepoDeployConfigForUpdate(plan, &updateServiceBody)

	return updateServiceBody, nil
}

func updateServiceGitRepoDeployConfigForUpdate(plan staticsite.StaticSiteModel, body *client.UpdateServiceJSONRequestBody) {
	body.Repo = plan.RepoURL.ValueStringPointer()
	body.Branch = plan.Branch.ValueStringPointer()
	body.RootDir = plan.RootDirectory.ValueStringPointer()
	body.AutoDeploy = common.AutoDeployBoolToClient(plan.AutoDeploy)
	body.AutoDeployTrigger = common.StringToAutoDeployTrigger(plan.AutoDeployTrigger)
	body.BuildFilter = common.ClientBuildFilterForModel(plan.BuildFilter)
}
