package resource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"

	"terraform-provider-render/internal/provider/common/validators"
)

func RegistryCredentialResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Provides a Render Registry Credential resource. Used to create credentials for accessing private Docker registries.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for this credential",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"auth_token": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "The auth token to use when pulling the image",
				Validators:  []validator.String{validators.StringNotEmpty},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Descriptive name for this credential",
				Validators:  []validator.String{validators.StringNotEmpty},
			},
			"registry": schema.StringAttribute{
				Required:            true,
				Description:         "The registry to use this credential with. One of GITHUB, GITLAB, DOCKER.",
				MarkdownDescription: "The registry to use this credential with. One of `GITHUB`, `GITLAB`, `DOCKER`.",
				Validators: []validator.String{
					stringvalidator.OneOf(
						"GITHUB",
						"GITLAB",
						"DOCKER",
					),
				},
			},
			"username": schema.StringAttribute{
				Required:    true,
				Description: "The username associated with the credential",
				Validators:  []validator.String{validators.StringNotEmpty},
			},
		},
	}
}

type RegistryCredentialModel struct {
	AuthToken types.String `tfsdk:"auth_token"`
	Id        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Registry  types.String `tfsdk:"registry"`
	Username  types.String `tfsdk:"username"`
}
