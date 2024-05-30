package datasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"terraform-provider-render/internal/provider/types/datasource"
)

func Schema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Provides information about a Render Cron Job resource.",
		Attributes: map[string]schema.Attribute{
			"id":                    datasource.ServiceID,
			"runtime_source":        datasource.RuntimeSource,
			"environment_id":        datasource.ResourceEnvironmentID,
			"name":                  datasource.ServiceName,
			"slug":                  datasource.Slug,
			"plan":                  datasource.Plan,
			"region":                datasource.Region,
			"root_directory":        datasource.RootDirectory,
			"schedule":              datasource.CronJobSchedule,
			"start_command":         datasource.StartCommand,
			"env_vars":              datasource.EnvVars,
			"secret_files":          datasource.SecretFiles,
			"notification_override": datasource.NotificationOverride,
		},
	}
}
