package resource

import (
	"context"

	"terraform-provider-render/internal/provider/types/resource"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func Schema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description:         "Project to organize collections of environments and resources",
		MarkdownDescription: "[Project](https://docs.render.com/projects) to organize collections of environments and resources.",
		Attributes: map[string]schema.Attribute{
			"id":           resource.ProjectID,
			"name":         resource.ProjectName,
			"environments": resource.Environments,
		},
	}
}
