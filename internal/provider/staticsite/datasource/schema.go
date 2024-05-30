package datasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"terraform-provider-render/internal/provider/types/datasource"
)

func Schema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Provides information about a Render Static Site.",
		Attributes: map[string]schema.Attribute{
			"id":                            datasource.ServiceID,
			"auto_deploy":                   datasource.AutoDeploy,
			"branch":                        datasource.Branch,
			"build_command":                 datasource.BuildCommand,
			"build_filter":                  datasource.BuildFilter,
			"custom_domains":                datasource.CustomDomains,
			"environment_id":                datasource.ResourceEnvironmentID,
			"env_vars":                      datasource.EnvVars,
			"headers":                       datasource.Headers,
			"name":                          datasource.ServiceName,
			"slug":                          datasource.Slug,
			"notification_override":         datasource.NotificationOverride,
			"publish_path":                  datasource.PublishPath,
			"pull_request_previews_enabled": datasource.PRPreviewsEnabled,
			"repo_url":                      datasource.RepoURL,
			"root_directory":                datasource.RootDirectory,
			"url":                           datasource.ServiceURL,
			"routes":                        datasource.Routes,
		},
	}
}
