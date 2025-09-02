package datasource

import (
	"context"

	"terraform-provider-render/internal/provider/types/datasource"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func Schema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Provides information about a Render Static Site.",
		Attributes: map[string]schema.Attribute{
			"id":                            datasource.ServiceID,
			"auto_deploy":                   datasource.AutoDeploy,
			"auto_deploy_trigger":           datasource.AutoDeployTrigger,
			"branch":                        datasource.Branch,
			"build_command":                 datasource.BuildCommand,
			"build_filter":                  datasource.BuildFilter,
			"custom_domains":                datasource.CustomDomains,
			"active_custom_domains":         datasource.ActiveCustomDomains,
			"environment_id":                datasource.ResourceEnvironmentID,
			"env_vars":                      datasource.EnvVars,
			"headers":                       datasource.Headers,
			"name":                          datasource.ServiceName,
			"slug":                          datasource.Slug,
			"notification_override":         datasource.NotificationOverride,
			"publish_path":                  datasource.PublishPath,
			"previews":                      datasource.Previews,
			"pull_request_previews_enabled": datasource.PRPreviewsEnabled,
			"repo_url":                      datasource.RepoURL,
			"root_directory":                datasource.RootDirectory,
			"url":                           datasource.ServiceURL,
			"routes":                        datasource.Routes,
		},
	}
}
