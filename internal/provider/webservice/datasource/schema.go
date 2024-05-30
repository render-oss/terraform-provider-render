package datasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"terraform-provider-render/internal/provider/types/datasource"
)

func Schema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Provides information about a Render Web Service.",
		Attributes: map[string]schema.Attribute{
			"id":                            datasource.ServiceID,
			"autoscaling":                   datasource.Autoscaling,
			"custom_domains":                datasource.CustomDomains,
			"runtime_source":                datasource.RuntimeSource,
			"disk":                          datasource.Disk,
			"environment_id":                datasource.ResourceEnvironmentID,
			"health_check_path":             datasource.HealthCheckPath,
			"name":                          datasource.ServiceName,
			"slug":                          datasource.Slug,
			"num_instances":                 datasource.NumInstances,
			"plan":                          datasource.Plan,
			"pre_deploy_command":            datasource.PreDeployCommand,
			"pull_request_previews_enabled": datasource.PRPreviewsEnabled,
			"root_directory":                datasource.RootDirectory,
			"start_command":                 datasource.StartCommand,
			"region":                        datasource.Region,
			"url":                           datasource.ServiceURL,
			"env_vars":                      datasource.EnvVars,
			"secret_files":                  datasource.SecretFiles,
			"notification_override":         datasource.NotificationOverride,
		},
	}
}
