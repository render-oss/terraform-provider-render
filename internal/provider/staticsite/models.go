package staticsite

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-render/internal/client"
	"terraform-provider-render/internal/provider/common"
)

type StaticSiteModel struct {
	Id                         types.String                  `tfsdk:"id"`
	AutoDeploy                 types.Bool                    `tfsdk:"auto_deploy"`
	AutoDeployTrigger          types.String                  `tfsdk:"auto_deploy_trigger"`
	Branch                     types.String                  `tfsdk:"branch"`
	BuildCommand               types.String                  `tfsdk:"build_command"`
	BuildFilter                *common.BuildFilterModel      `tfsdk:"build_filter"`
	EnvironmentID              types.String                  `tfsdk:"environment_id"`
	CustomDomains              []common.CustomDomainModel    `tfsdk:"custom_domains"`
	ActiveCustomDomains        types.Set                     `tfsdk:"active_custom_domains"`
	EnvVars                    map[string]common.EnvVarModel `tfsdk:"env_vars"`
	Headers                    []common.HeaderModel          `tfsdk:"headers"`
	IPAllowList                types.Set                     `tfsdk:"ip_allow_list"`
	Name                       types.String                  `tfsdk:"name"`
	Slug                       types.String                  `tfsdk:"slug"`
	NotificationOverride       types.Object                  `tfsdk:"notification_override"`
	PublishPath                types.String                  `tfsdk:"publish_path"`
	Previews                   types.Object                  `tfsdk:"previews"`
	PullRequestPreviewsEnabled types.Bool                    `tfsdk:"pull_request_previews_enabled"`
	RepoURL                    types.String                  `tfsdk:"repo_url"`
	RootDirectory              types.String                  `tfsdk:"root_directory"`
	Routes                     []common.RouteModel           `tfsdk:"routes"`
	Url                        types.String                  `tfsdk:"url"`
}

func ModelForServiceResult(service *common.WrappedStaticSite, state StaticSiteModel, diags diag.Diagnostics) (*StaticSiteModel, error) {
	details, err := service.ServiceDetails.AsStaticSiteDetails()
	if err != nil {
		return nil, err
	}

	var routes []common.RouteModel
	if service.Routes != nil && len(*service.Routes) > 0 {
		r, err := common.SortRoutesForPlan(
			state.Routes,
			common.ClientRoutesToRouteModels(*service.Routes),
		)
		if err != nil {
			return nil, err
		}
		routes = r
	}

	// Handle IP allow list: if not configured in state (null), keep it null
	// This prevents showing drift when API returns its default value
	ipAllowList := state.IPAllowList
	if !state.IPAllowList.IsNull() && details.IpAllowList != nil {
		ipAllowList = common.IPAllowListFromClient(*details.IpAllowList, diags)
	}

	staticSitesModel := &StaticSiteModel{
		Id:                   types.StringValue(service.Id),
		AutoDeploy:           types.BoolValue(service.AutoDeploy == client.AutoDeployYes),
		AutoDeployTrigger:    common.AutoDeployTriggerToString(service.AutoDeployTrigger),
		BuildFilter:          common.BuildFilterModelForClient(service.BuildFilter),
		CustomDomains:        common.CustomDomainClientsToCustomDomainModelsNonRedirecting(service.CustomDomains),
		ActiveCustomDomains:  common.CustomDomainSetFromClient(service.CustomDomains, diags),
		EnvironmentID:        types.StringPointerValue(service.EnvironmentId),
		Headers:              common.ClientHeadersToRouteModels(service.Headers),
		IPAllowList:          ipAllowList,
		Name:                 types.StringValue(service.Name),
		Slug:                 types.StringValue(service.Slug),
		NotificationOverride: common.NotificationOverrideFromClient(service.NotificationOverride, diags),
		RootDirectory:        types.StringValue(service.RootDir),
		Routes:               routes,
		EnvVars:              common.EnvVarsFromClientCursors(service.EnvVars, state.EnvVars),
	}

	applyGitBackedFields(service.Service, staticSitesModel, &details)

	return staticSitesModel, nil
}

func applyGitBackedFields(service *client.Service, model *StaticSiteModel, details *client.StaticSiteDetails) {
	model.BuildCommand = types.StringValue(details.BuildCommand)
	model.Previews = common.PreviewsToPreviewsObject(details.Previews)
	model.PullRequestPreviewsEnabled = types.BoolValue(details.PullRequestPreviewsEnabled != nil && *details.PullRequestPreviewsEnabled == client.PullRequestPreviewsEnabledYes)
	model.PublishPath = types.StringValue(details.PublishPath)
	model.Url = types.StringValue(details.Url)
	model.RepoURL = types.StringPointerValue(service.Repo)
	model.Branch = types.StringPointerValue(service.Branch)
}
