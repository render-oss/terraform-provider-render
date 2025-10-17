package datasource

import (
	"context"
	"terraform-provider-render/internal/provider/types/datasource"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func Schema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Provides information about a Render Web Service.",
		Attributes: map[string]schema.Attribute{
			"id":                            datasource.ServiceID,
			"autoscaling":                   datasource.Autoscaling,
			"custom_domains":                datasource.CustomDomains,
			"active_custom_domains":         datasource.ActiveCustomDomains,
			"runtime_source":                datasource.RuntimeSource,
			"disk":                          datasource.Disk,
			"environment_id":                datasource.ResourceEnvironmentID,
			"health_check_path":             datasource.HealthCheckPath,
			"name":                          datasource.ServiceName,
			"slug":                          datasource.Slug,
			"num_instances":                 datasource.NumInstances,
			"plan":                          datasource.Plan,
			"pre_deploy_command":            datasource.PreDeployCommand,
			"previews":                      datasource.Previews,
			"pull_request_previews_enabled": datasource.PRPreviewsEnabled,
			"root_directory":                datasource.RootDirectory,
			"start_command":                 datasource.StartCommand,
			"region":                        datasource.Region,
			"url":                           datasource.ServiceURL,
			"maintenance_mode":              datasource.MaintenanceMode,
			"max_shutdown_delay_seconds":    datasource.MaxShutdownDelaySeconds,
			"ip_allow_list":                 datasource.IPAllowList,
			"env_vars":                      datasource.EnvVars,
			"secret_files":                  datasource.SecretFiles,
			"notification_override":         datasource.NotificationOverride,
			"log_stream_override":           datasource.LogStreamOverride,
		},
	}
}
