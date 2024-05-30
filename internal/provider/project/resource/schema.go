package resource

import (
	"context"

	"terraform-provider-render/internal/provider/types/resource"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func Schema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":           resource.ProjectID,
			"name":         resource.ProjectName,
			"environments": resource.Environments,
		},
	}
}
