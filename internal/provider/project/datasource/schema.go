package datasource

import (
	"context"
	"terraform-provider-render/internal/provider/types/datasource"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func ProjectDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":           datasource.ProjectID,
			"name":         datasource.ProjectName,
			"environments": datasource.Environments,
		},
	}
}
