package resource

import (
	"context"

	"terraform-provider-render/internal/provider/types/resource"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func Schema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Provides a Render Cron Job resource.",
		Attributes: map[string]schema.Attribute{
			"id":                    resource.ServiceID,
			"runtime_source":        resource.RuntimeSource,
			"environment_id":        resource.ResourceEnvironmentID,
			"name":                  resource.ServiceName,
			"slug":                  resource.Slug,
			"plan":                  resource.Plan,
			"region":                resource.Region,
			"root_directory":        resource.RootDirectory,
			"schedule":              resource.CronJobSchedule,
			"start_command":         resource.StartCommand,
			"env_vars":              resource.EnvVars,
			"secret_files":          resource.SecretFiles,
			"notification_override": resource.NotificationOverride,
			"log_stream_override":   resource.LogStreamOverride,
		},
	}
}
