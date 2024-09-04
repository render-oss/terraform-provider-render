package resource

import (
	"context"

	"terraform-provider-render/internal/provider/types/resource"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func Schema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Provides a Render Private Service resource.",
		Attributes: map[string]schema.Attribute{
			"id":                            resource.ServiceID,
			"autoscaling":                   resource.Autoscaling,
			"runtime_source":                resource.RuntimeSource,
			"disk":                          resource.Disk,
			"environment_id":                resource.ResourceEnvironmentID,
			"name":                          resource.ServiceName,
			"slug":                          resource.Slug,
			"num_instances":                 resource.NumInstances,
			"plan":                          resource.Plan,
			"pre_deploy_command":            resource.PreDeployCommand,
			"previews":                      resource.Previews,
			"pull_request_previews_enabled": resource.PRPreviewsEnabled,
			"region":                        resource.Region,
			"root_directory":                resource.RootDirectory,
			"start_command":                 resource.StartCommand,
			"url":                           resource.ServiceURL,
			"max_shutdown_delay_seconds":    resource.MaxShutdownDelaySeconds,
			"env_vars":                      resource.EnvVars,
			"secret_files":                  resource.SecretFiles,
			"notification_override":         resource.NotificationOverride,
		},
	}
}
