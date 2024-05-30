package resource

import (
	"context"

	"terraform-provider-render/internal/provider/types/resource"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func Schema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Provides a Render Web Service resource.",
		Attributes: map[string]schema.Attribute{
			"id":                            resource.ServiceID,
			"autoscaling":                   resource.Autoscaling,
			"custom_domains":                resource.CustomDomains,
			"runtime_source":                resource.RuntimeSource,
			"disk":                          resource.Disk,
			"environment_id":                resource.ResourceEnvironmentID,
			"health_check_path":             resource.HealthCheckPath,
			"name":                          resource.ServiceName,
			"slug":                          resource.Slug,
			"num_instances":                 resource.NumInstances,
			"plan":                          resource.Plan,
			"pre_deploy_command":            resource.PreDeployCommand,
			"pull_request_previews_enabled": resource.PRPreviewsEnabled,
			"region":                        resource.Region,
			"root_directory":                resource.RootDirectory,
			"start_command":                 resource.StartCommand,
			"url":                           resource.ServiceURL,
			"env_vars":                      resource.EnvVars,
			"secret_files":                  resource.SecretFiles,
			"notification_override":         resource.NotificationOverride,
		},
	}
}
