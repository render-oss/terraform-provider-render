package internal

import (
	"context"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
	"terraform-provider-render/internal/provider/staticsite"
)

func CreateServiceRequestFromModel(ctx context.Context, ownerID string, plan staticsite.StaticSiteModel) (client.CreateServiceJSONRequestBody, error) {
	var routeModels []client.RoutePost
	for _, route := range plan.Routes {
		routeType := common.ClientRouteType(route.Type.ValueString())

		routeModels = append(routeModels, client.RoutePost{
			Destination: route.Destination.ValueString(),
			Source:      route.Source.ValueString(),
			Type:        routeType,
		})
	}

	prPreviews := client.PullRequestPreviewsEnabledNo
	if plan.PullRequestPreviewsEnabled.ValueBool() {
		prPreviews = client.PullRequestPreviewsEnabledYes
	}

	// Handle IP allow list: omitted (null) -> send nil, otherwise send value
	var ipAllowList *[]client.CidrBlockAndDescription
	if !plan.IPAllowList.IsNull() && !plan.IPAllowList.IsUnknown() {
		list, err := common.ClientFromIPAllowList(plan.IPAllowList)
		if err != nil {
			return client.CreateServiceJSONRequestBody{}, err
		}
		ipAllowList = &list
	}

	staticSiteDetails := client.StaticSiteDetailsPOST{
		BuildCommand:               plan.BuildCommand.ValueStringPointer(),
		Headers:                    common.From(common.ModelToClientHeaderInput(plan.Headers)),
		IpAllowList:                ipAllowList,
		PublishPath:                plan.PublishPath.ValueStringPointer(),
		Previews:                   common.PreviewsObjectToPreviews(ctx, plan.Previews),
		PullRequestPreviewsEnabled: &prPreviews,
		Routes:                     &routeModels,
	}

	serviceDetails := &client.ServicePOST_ServiceDetails{}
	if err := serviceDetails.FromStaticSiteDetailsPOST(staticSiteDetails); err != nil {
		return client.CreateServiceJSONRequestBody{}, err
	}

	evs, err := common.EnvVarsToClient(plan.EnvVars)
	if err != nil {
		return client.CreateServiceJSONRequestBody{}, err
	}

	var createServiceBody = client.CreateServiceJSONRequestBody{
		EnvVars:        &evs,
		Name:           plan.Name.ValueString(),
		OwnerId:        ownerID,
		ServiceDetails: serviceDetails,
		Type:           client.StaticSite,
	}

	updateServiceGitRepoDeployConfigForCreate(plan, &createServiceBody)

	return createServiceBody, nil
}

func updateServiceGitRepoDeployConfigForCreate(plan staticsite.StaticSiteModel, body *client.CreateServiceJSONRequestBody) {
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
