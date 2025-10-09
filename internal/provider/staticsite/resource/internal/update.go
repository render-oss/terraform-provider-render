package internal

import (
	"context"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	"terraform-provider-render/internal/provider/staticsite"
)

func UpdateServiceRequestFromModel(ctx context.Context, plan staticsite.StaticSiteModel, state staticsite.StaticSiteModel) (client.UpdateServiceJSONRequestBody, error) {
	prPreviewEnabled := client.PullRequestPreviewsEnabledNo
	if plan.PullRequestPreviewsEnabled.ValueBool() {
		prPreviewEnabled = client.PullRequestPreviewsEnabledYes
	}

	// Handle IP allow list with state-aware logic:
	// - In state but not in plan (null) -> send default (0.0.0.0/0) to revert
	// - Not in state and not in plan -> send nil (don't update)
	// - In plan with empty list -> send empty array (block all)
	// - In plan with values -> send those values
	var ipAllowList *[]client.CidrBlockAndDescription
	if !plan.IPAllowList.IsNull() && !plan.IPAllowList.IsUnknown() {
		// Field is configured in plan
		list, err := common.ClientFromIPAllowList(plan.IPAllowList)
		if err != nil {
			return client.UpdateServiceJSONRequestBody{}, err
		}
		ipAllowList = &list
	} else if !state.IPAllowList.IsNull() {
		// Field was in state but removed from plan -> revert to default (0.0.0.0/0 everywhere)
		ipAllowList = &common.AllowAllCIDRList
	}

	var staticSiteDetails = client.StaticSiteDetailsPATCH{
		BuildCommand:               plan.BuildCommand.ValueStringPointer(),
		IpAllowList:                ipAllowList,
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
	body.AutoDeploy = common.From(common.AutoDeployBoolToClient(plan.AutoDeploy.ValueBool()))
	body.AutoDeployTrigger = common.StringToAutoDeployTrigger(plan.AutoDeployTrigger)
	if body.AutoDeployTrigger != nil {
		body.AutoDeploy = nil
	}
	body.BuildFilter = common.ClientBuildFilterForModel(plan.BuildFilter)
}
