package resource

import (
	"context"

	"terraform-provider-render/internal/provider/types/resource"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func Schema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Provides a Render Static Site.",
		Attributes: map[string]schema.Attribute{
			"id":                            resource.ServiceID,
			"auto_deploy":                   resource.AutoDeploy,
			"auto_deploy_trigger":           resource.AutoDeployTrigger,
			"branch":                        resource.Branch,
			"build_command":                 resource.BuildCommand,
			"build_filter":                  resource.BuildFilter,
			"custom_domains":                resource.CustomDomains,
			"active_custom_domains":         resource.ActiveCustomDomains,
			"environment_id":                resource.ResourceEnvironmentID,
			"env_vars":                      resource.EnvVars,
			"headers":                       resource.Headers,
			"ip_allow_list":                 resource.IPAllowListOptional,
			"name":                          resource.ServiceName,
			"slug":                          resource.Slug,
			"notification_override":         resource.NotificationOverride,
			"publish_path":                  resource.PublishPath,
			"previews":                      resource.Previews,
			"pull_request_previews_enabled": resource.PRPreviewsEnabled,
			"repo_url":                      resource.RepoURL,
			"root_directory":                resource.RootDirectory,
			"url":                           resource.ServiceURL,
			"routes":                        resource.Routes,
		},
	}
}
