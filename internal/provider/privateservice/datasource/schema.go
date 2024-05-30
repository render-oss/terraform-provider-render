package datasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"terraform-provider-render/internal/provider/types/datasource"
)

func Schema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Provides information about a Render Private Service.",
		Attributes: map[string]schema.Attribute{
			"id":                            datasource.ServiceID,
			"autoscaling":                   datasource.Autoscaling,
			"runtime_source":                datasource.RuntimeSource,
			"disk":                          datasource.Disk,
			"environment_id":                datasource.ResourceEnvironmentID,
			"name":                          datasource.ServiceName,
			"slug":                          datasource.Slug,
			"num_instances":                 datasource.NumInstances,
			"plan":                          datasource.Plan,
			"pre_deploy_command":            datasource.PreDeployCommand,
			"pull_request_previews_enabled": datasource.PRPreviewsEnabled,
			"region":                        datasource.Region,
			"root_directory":                datasource.RootDirectory,
			"start_command":                 datasource.StartCommand,
			"url":                           datasource.ServiceURL,
			"env_vars":                      datasource.EnvVars,
			"secret_files":                  datasource.SecretFiles,
			"notification_override":         datasource.NotificationOverride,
		},
	}
}
