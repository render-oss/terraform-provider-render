package datasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func RegistryCredentialDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Provides information about a Render Registry Credential.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "Unique identifier for this credential",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Descriptive name for this credential",
			},
			"registry": schema.StringAttribute{
				Computed:    true,
				Description: "The registry to use this credential with. One of GITHUB, GITLAB, DOCKER, AWS_ECR, GOOGLE_ARTIFACT.",
			},
			"username": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "The username associated with the credential",
			},
		},
	}
}

type RegistryCredentialModel struct {
	Id       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Registry types.String `tfsdk:"registry"`
	Username types.String `tfsdk:"username"`
}
