package datasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-render/internal/provider/types/datasource"
)

func EnvGroupDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Provides information about a Render Environment Group resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				Description:         "Unique identifier for this environment group",
				MarkdownDescription: "Unique identifier for this environment group",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				Description:         "Descriptive name for this environment group",
				MarkdownDescription: "Descriptive name for this environment group",
			},
			"environment_id": datasource.EnvironmentID,
			"env_vars":       datasource.EnvVars,
			"secret_files":   datasource.SecretFiles,
		},
	}
}

func EnvGroupLinkDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"env_group_id": schema.StringAttribute{
				Required:            true,
				Description:         "Unique identifier for the environment group",
				MarkdownDescription: "Unique identifier for the environment group",
			},
			"service_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				Description:         "List of service ids linked to the environment group",
				MarkdownDescription: "List of service ids linked to the environment group",
			},
		},
	}
}
