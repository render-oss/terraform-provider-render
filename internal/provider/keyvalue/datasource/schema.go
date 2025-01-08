package datasource

import (
	"context"

	"terraform-provider-render/internal/provider/types/datasource"
	"terraform-provider-render/internal/provider/types/resource"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func Schema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Provides information about a Render Key / Value instance.",
		Attributes: map[string]schema.Attribute{
			"id":                  datasource.ServiceID,
			"environment_id":      datasource.ResourceEnvironmentID,
			"ip_allow_list":       datasource.IPAllowList,
			"max_memory_policy":   datasource.MaxMemoryPolicy,
			"name":                datasource.ServiceName,
			"plan":                datasource.KeyValuePlan,
			"region":              datasource.Region,
			"connection_info":     datasource.ConnectionInfo,
			"log_stream_override": resource.LogStreamOverride,
		},
	}
}
