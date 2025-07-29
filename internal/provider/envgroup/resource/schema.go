package resource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-render/internal/provider/common/validators"
	"terraform-provider-render/internal/provider/types/resource"
)

func EnvGroupResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Provides a Render Environment Group resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for this environment group",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Descriptive name for this environment group",
				Validators:  []validator.String{validators.StringNotEmpty},
			},
			"environment_id": resource.ResourceEnvironmentID,
			"env_vars":       resource.EnvVars,
			"secret_files":   resource.SecretFiles,
		},
	}
}

func EnvGroupLinkResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"env_group_id": schema.StringAttribute{
				Required:            true,
				Description:         "Unique identifier for the environment group",
				MarkdownDescription: "Unique identifier for the environment group",
			},
			"service_ids": schema.SetAttribute{
				ElementType:         types.StringType,
				Required:            true,
				Description:         "Set of service ids linked to the environment group",
				MarkdownDescription: "Set of service ids linked to the environment group",
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
		},
	}
}
