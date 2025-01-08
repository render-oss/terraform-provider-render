package resource

import (
	"context"

	"terraform-provider-render/internal/provider/types/resource"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func Schema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Provides a Render Key / Value resource.",
		Attributes: map[string]schema.Attribute{
			"id":                  resource.ServiceID,
			"environment_id":      resource.ResourceEnvironmentID,
			"ip_allow_list":       resource.IPAllowList,
			"max_memory_policy":   resource.MaxMemoryPolicy,
			"name":                resource.ServiceName,
			"plan":                resource.RedisPlan,
			"region":              resource.Region,
			"connection_info":     resource.ConnectionInfo,
			"log_stream_override": resource.LogStreamOverride,
		},
	}
}
